package usecase

import (
	"encoding/json"
	"fmt"
	"net"
	"netsim_gateway/domain/model"
	"netsim_gateway/domain/service"
	"netsim_gateway/ports/input"
	"netsim_gateway/ports/output"
	"sync"
	"time"
)

type Config struct {
	TelemetryPort     int
	MessagesPort      int
	SubscribersPort   int
	NetsimListenPort  int
	SnapshotIntervalS int
	NetsimDiscovery   DiscoveryConfig
}

type DiscoveryConfig struct {
	Mode                string
	Addrs               []string
	SwarmService        string
	RediscoverIntervalS int
}

type Gateway struct {
	Config         Config
	netsimAddrs    []*net.UDPAddr
	uavRegistry    sync.Map // string (uavID) -> *net.UDPAddr
	telemetryCache sync.Map // string (uavID) -> json.RawMessage
	uiSubscribers  sync.Map // string (addr.String()) -> *net.UDPAddr
	uavSender         output.PacketSender
	netsimSender      output.PacketSender
	subscribersSender output.PacketSender
	logger            output.Logger
	mu                sync.RWMutex
}

func NewGateway(cfg Config, netsims []*net.UDPAddr, uavSender, netsimSender, subscribersSender output.PacketSender, logger output.Logger) *Gateway {
	return &Gateway{
		Config:            cfg,
		netsimAddrs:       netsims,
		uavSender:         uavSender,
		netsimSender:      netsimSender,
		subscribersSender: subscribersSender,
		logger:            logger,
	}
}

func (g *Gateway) UpdateNetsims(addrs []*net.UDPAddr) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.netsimAddrs = addrs
}

type InboundMsg struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

type TelemetryPayload struct {
	UAVID string `json:"uav_id"`
}

func (g *Gateway) HandleTelemetry(pkt input.RawPacket) {
	var payload map[string]any
	if err := json.Unmarshal(pkt.Data, &payload); err == nil {
		if uavID, ok := payload["uav_id"].(string); ok && uavID != "" {
			g.uavRegistry.Store(uavID, pkt.Addr)

			g.mu.RLock()
			netsimCount := len(g.netsimAddrs)
			var target *net.UDPAddr
			if netsimCount > 0 {
				idx := service.ConsistentHash(uavID, netsimCount)
				target = g.netsimAddrs[idx]
			}
			g.mu.RUnlock()

			if target != nil {
				wrapped := map[string]any{
					"topic":   "uav_telemetry",
					"payload": payload,
				}
				if data, err := json.Marshal(wrapped); err == nil {
					g.netsimSender.Send(data, target)
					g.logger.Info(fmt.Sprintf("Forwarded telemetry from %s to %s", uavID, target.String()))
				}
			}
		}
	}
}

func (g *Gateway) HandleMessages(pkt input.RawPacket) {
	var payload map[string]any
	if err := json.Unmarshal(pkt.Data, &payload); err == nil {
		if uavID, ok := payload["uav_id"].(string); ok {
			if uavID != "" {
				g.uavRegistry.Store(uavID, pkt.Addr)
			}

			if uavID == "" {
				if data, err := json.Marshal(payload); err == nil {
					g.uavRegistry.Range(func(key, value any) bool {
						addr := value.(*net.UDPAddr)
						g.uavSender.Send(data, addr)
						return true
					})
				}
				return
			}

			g.mu.RLock()
			netsimCount := len(g.netsimAddrs)
			var target *net.UDPAddr
			if netsimCount > 0 {
				idx := service.ConsistentHash(uavID, netsimCount)
				target = g.netsimAddrs[idx]
			}
			g.mu.RUnlock()

			if target != nil {
				wrapped := map[string]any{
					"topic":   "uav_broadcast",
					"payload": payload,
				}
				if data, err := json.Marshal(wrapped); err == nil {
					g.netsimSender.Send(data, target)
					g.logger.Info(fmt.Sprintf("Forwarded broadcast from %s to %s", uavID, target.String()), "timestamp", time.Now().Format(time.RFC3339Nano))
				}
			}
		}
	}
}

func (g *Gateway) HandleSubscribers(pkt input.RawPacket) {
	g.uiSubscribers.Store(pkt.Addr.String(), pkt.Addr)
	g.logger.Info("New UI subscriber", "addr", pkt.Addr.String())
}

type DeliverMessage struct {
	TargetUAVID string `json:"target_uav_id"`
	SenderID    string `json:"sender_id"`
	Payload     string `json:"payload"`
}

type BroadcastNotify struct {
	SenderID       string          `json:"sender_id"`
	SenderPosition map[string]any  `json:"sender_position"`
	Payload        json.RawMessage `json:"payload"`
}

func (g *Gateway) HandleNetsimMessage(pkt input.RawPacket) {
	var msg InboundMsg
	if err := json.Unmarshal(pkt.Data, &msg); err != nil {
		return
	}

	switch msg.Topic {
	case "snapshot":
		var snap model.TelemetrySnapshot
		if err := json.Unmarshal(msg.Payload, &snap); err == nil {
			for uavID, telemetry := range snap.UAVs {
				g.telemetryCache.Store(uavID, telemetry)
			}
		}
	case "deliver":
		var deliver DeliverMessage
		if err := json.Unmarshal(msg.Payload, &deliver); err == nil {
			if val, ok := g.uavRegistry.Load(deliver.TargetUAVID); ok {
				addr := val.(*net.UDPAddr)
				var parsedPayload map[string]interface{}
				if err := json.Unmarshal([]byte(deliver.Payload), &parsedPayload); err == nil {
					formatted := map[string]interface{}{
						"uav_id":  deliver.SenderID,
						"payload": parsedPayload,
					}
					data, _ := json.Marshal(formatted)
					g.uavSender.Send(data, addr)
					g.logger.Info(fmt.Sprintf("Delivered message from %s to %s", deliver.SenderID, deliver.TargetUAVID), "timestamp", time.Now().Format(time.RFC3339Nano))
				}
			}
		}
	case "broadcast_notify":
		var notify BroadcastNotify
		if err := json.Unmarshal(msg.Payload, &notify); err == nil {
			uiMsg := map[string]any{
				"topic": "broadcast",
				"payload": map[string]any{
					"sender_id": notify.SenderID,
					"payload":   notify.Payload,
				},
			}
			if uiData, err := json.Marshal(uiMsg); err == nil {
				g.uiSubscribers.Range(func(key, value any) bool {
					addr := value.(*net.UDPAddr)
					g.subscribersSender.Send(uiData, addr)
					return true
				})
			}

			g.mu.RLock()
			for _, netsimAddr := range g.netsimAddrs {
				if netsimAddr.String() != pkt.Addr.String() {
					peerMsg := map[string]any{
						"topic": "peer_broadcast",
						"payload": map[string]any{
							"sender_id":       notify.SenderID,
							"sender_position": notify.SenderPosition,
							"payload":         notify.Payload,
							"origin_node":     pkt.Addr.String(),
						},
					}
					if peerData, err := json.Marshal(peerMsg); err == nil {
						g.netsimSender.Send(peerData, netsimAddr)
					}
				}
			}
			g.mu.RUnlock()
		}
	}
}

type TelemetryHandler struct{ gateway *Gateway }
func (h *TelemetryHandler) Handle(pkt input.RawPacket) { h.gateway.HandleTelemetry(pkt) }

type MessagesHandler struct{ gateway *Gateway }
func (h *MessagesHandler) Handle(pkt input.RawPacket) { h.gateway.HandleMessages(pkt) }

type SubscribersHandler struct{ gateway *Gateway }
func (h *SubscribersHandler) Handle(pkt input.RawPacket) { h.gateway.HandleSubscribers(pkt) }

type NetsimHandler struct{ gateway *Gateway }
func (h *NetsimHandler) Handle(pkt input.RawPacket) { h.gateway.HandleNetsimMessage(pkt) }

func NewTelemetryHandler(g *Gateway) input.UAVMessageHandler       { return &TelemetryHandler{gateway: g} }
func NewMessagesHandler(g *Gateway) input.UAVMessageHandler        { return &MessagesHandler{gateway: g} }
func NewSubscribersHandler(g *Gateway) input.UAVMessageHandler     { return &SubscribersHandler{gateway: g} }
func NewNetsimHandler(g *Gateway) input.NetsimMessageHandler       { return &NetsimHandler{gateway: g} }
