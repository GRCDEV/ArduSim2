package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"mission/domain"
)

type UDPBroker struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewUDPBroker() *UDPBroker {
	return &UDPBroker{}
}

func (b *UDPBroker) Connect(ip string, port int, subTopic, telemetryTopic string) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}

	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}

	b.conn = conn
	b.addr = serverAddr

	b.sendSubscription(subTopic)
	b.sendSubscription(telemetryTopic)

	return nil
}

func (b *UDPBroker) sendSubscription(topic string) {
	msg := map[string]interface{}{
		"topic":        "$subscribe",
		"subscribe_to": topic,
	}
	b.publishJSON(msg)
}

func (b *UDPBroker) Publish(topic string, payload map[string]interface{}) error {
	return b.publishJSON(map[string]interface{}{
		"topic":   topic,
		"payload": payload,
	})
}

func (b *UDPBroker) publishJSON(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	if b.conn == nil {
		return fmt.Errorf("udp broker not connected")
	}
	
	_, err = b.conn.WriteToUDP(data, b.addr)
	return err
}

func (b *UDPBroker) Listen() (<-chan domain.BrokerMessage, error) {
	msgChan := make(chan domain.BrokerMessage)

	go func() {
		defer close(msgChan)
		buffer := make([]byte, 4096)
		for {
			n, _, err := b.conn.ReadFromUDP(buffer)
			if err != nil {
				return // Connection closed
			}

			var msg domain.BrokerMessage
			if err := json.Unmarshal(buffer[:n], &msg); err == nil {
				msgChan <- msg
			} else {
				log.Printf("Error unmarshalling message: %v", err)
			}
		}
	}()

	return msgChan, nil
}

func (b *UDPBroker) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
