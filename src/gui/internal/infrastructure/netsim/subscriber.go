package netsim

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"ui/internal/domain"
	"ui/internal/ports"

	"github.com/go-viper/mapstructure/v2"
)


type NetsimSubscriber struct {
	mu              sync.Mutex
	ui              ports.UIBridge
	expectedUAVs    int
	receivedUAVs    map[string]bool
	finishedUAVs    map[string]bool
	onFinish        func()
	ready           bool
	finished        bool
	subscribersAddr string
	messagesAddr    string
	subscribersPort int
	messagesPort    int
}

func NewNetsimSubscriber(ui ports.UIBridge, subPort, msgPort int) *NetsimSubscriber {
	return &NetsimSubscriber{
		ui:              ui,
		subscribersPort: subPort,
		messagesPort:    msgPort,
		subscribersAddr: fmt.Sprintf("127.0.0.1:%d", subPort),
		messagesAddr:    fmt.Sprintf("127.0.0.1:%d", msgPort),
	}
}

func (s *NetsimSubscriber) SetRemoteAddr(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribersAddr = fmt.Sprintf("%s:%d", ip, s.subscribersPort)
	s.messagesAddr = fmt.Sprintf("%s:%d", ip, s.messagesPort)
}

func (s *NetsimSubscriber) SetExpectedFleet(uavIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expectedUAVs = len(uavIDs)
	s.receivedUAVs = make(map[string]bool)
	s.finishedUAVs = make(map[string]bool)
	s.ready = false
	s.finished = false
}

func (s *NetsimSubscriber) SetOnFinish(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onFinish = fn
}

func (s *NetsimSubscriber) SendGlobalBroadcast(payload interface{}) error {
	remoteUDPAddr, err := net.ResolveUDPAddr("udp", s.messagesAddr)
	if err != nil {
		return err
	}

	packet := map[string]interface{}{
		"uav_id":  "",
		"payload": payload,
	}

	marshaled, _ := json.Marshal(packet)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.WriteToUDP(marshaled, remoteUDPAddr)
	return err
}

func (s *NetsimSubscriber) NotifyUserStoppedAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished {
		return
	}
	s.finished = true
	s.ui.EmitEvent("simulation:finished", nil)
}

func (s *NetsimSubscriber) Start(ctx context.Context) {
	subscribersUDPAddr, _ := net.ResolveUDPAddr("udp", s.subscribersAddr)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		fmt.Println("Error listening on UDP", err)
		return
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	msgReceivedChan := make(chan any, 1)
	msgReceived := false

	// Subscribe in a loop until we get a message or timeout
	go func() {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-msgReceivedChan:
				return
			case <-ctx.Done():
				return
			default:
			}

			// Just ping the subscribers port to register our address
			p := []byte("{}")
			_, _ = conn.WriteToUDP(p, subscribersUDPAddr)
			
			time.Sleep(5 * time.Second)
		}
	}()

	buffer := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		if !msgReceived {
			msgReceived = true
			msgReceivedChan <- struct{}{}
		}

		var msg struct {
			Topic   string                 `json:"topic"`
			UavID   string                 `json:"uav_id"`
			Payload map[string]interface{} `json:"payload"`
		}
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			continue
		}

		switch msg.Topic {
		case "telemetry_snapshot":
			uavsRaw, ok := msg.Payload["uavs"].(map[string]interface{})
			if !ok {
				continue
			}

			snapshotMap := make(map[string]interface{})

			s.mu.Lock()
			for id, raw := range uavsRaw {
				var t domain.TelemetryData
				err := mapstructure.WeakDecode(raw, &t)
				if err != nil {
					fmt.Printf("DECODE ERROR for UAV %s: %v\n", id, err)
				}
				snapshotMap[id] = t
				
				if t.NrGpsOnline > 0 && !s.ready {
					s.receivedUAVs[id] = true
				}
			}

			if !s.ready {
				keys := make([]string, 0, len(s.receivedUAVs))
				for k := range s.receivedUAVs {
					keys = append(keys, k)
				}
				if len(keys) == s.expectedUAVs {
					s.ready = true
					s.ui.EmitEvent("simulation:ready", nil)
				}
			}
			s.mu.Unlock()

			if(s.ready) {
				s.ui.EmitEvent("telemetry_snapshot", map[string]interface{}{
					"uavs": snapshotMap,
				})
			}
		case "broadcast":
			s.handleBroadcastPacket(msg)
		}
	}
}

func (s *NetsimSubscriber) handleBroadcastPacket(msg struct {
	Topic   string                 `json:"topic"`
	UavID   string                 `json:"uav_id"`
	Payload map[string]interface{} `json:"payload"`
}) {
	senderID, _ := msg.Payload["sender_id"].(string)
	innerPayloadRaw, ok := msg.Payload["payload"]
	if !ok {
		return
	}

	var innerPayload map[string]interface{}
	switch v := innerPayloadRaw.(type) {
	case string:
		_ = json.Unmarshal([]byte(v), &innerPayload)
	case map[string]interface{}:
		innerPayload = v
	}

	if innerPayload == nil {
		return
	}

	var service, command string
	if senderID == "" {
		topic, _ := innerPayload["topic"].(string)
		service = s.serviceNameFromTopic(topic)
		if p, ok := innerPayload["payload"].(map[string]interface{}); ok {
			command, _ = p["command"].(string)
		}
	} else {
		command, _ = innerPayload["command"].(string)
		source, _ := innerPayload["source"].(string)
		service = source
	}

	if command == "" {
		return
	}

	sourceStr := s.formatSource(senderID)
	logLine := fmt.Sprintf("%s%s: %s", sourceStr, service, command)

	s.ui.EmitEvent("netsim:message", map[string]string{
		"source":  sourceStr,
		"label":   logLine,
	})

	if command == "finish" {
		s.registerFinish(senderID)
	}
}

func (s *NetsimSubscriber) serviceNameFromTopic(name string) string {
	if name == "" {
		return "unknown"
	}
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

func (s *NetsimSubscriber) formatSource(uavID string) string {
	if uavID == "" {
		return ""
	}
	return "(UAV-" + uavID + ") "
}

func (s *NetsimSubscriber) registerFinish(uavID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished {
		return
	}
	if uavID != "" {
		s.finishedUAVs[uavID] = true
	}
	keys := make([]string, 0, len(s.finishedUAVs))
	for k := range s.finishedUAVs {
		keys = append(keys, k)
	}
	if len(keys) == s.expectedUAVs {
		s.finished = true
		s.ui.EmitEvent("simulation:finished", nil)
		if s.onFinish != nil {
			go s.onFinish()
		}
	}
}

