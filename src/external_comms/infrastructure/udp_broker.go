package infrastructure

import (
	"encoding/json"
	"fmt"
	"net"

	"external_comms/ports"
)

type UDPBroker struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewUDPBroker() *UDPBroker {
	return &UDPBroker{}
}

func (b *UDPBroker) Connect(ip string, port int, createTopics []string) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	b.addr = serverAddr

	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}
	b.conn = conn

	for _, t := range createTopics {
		b.sendSubscription(t)
	}

	return nil
}

func (b *UDPBroker) sendSubscription(topic string) {
	msg := map[string]interface{}{
		"topic":        "$subscribe",
		"subscribe_to": topic,
	}
	b.publishJSON(msg)
}

func (b *UDPBroker) Publish(payload map[string]interface{}) error {
	return b.publishJSON(payload)
}

func (b *UDPBroker) publishJSON(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = b.conn.WriteToUDP(data, b.addr)
	return err
}

func (b *UDPBroker) Listen() (<-chan ports.BrokerMessage, error) {
	msgChan := make(chan ports.BrokerMessage, 100)

	go func() {
		defer close(msgChan)
		buffer := make([]byte, 4096)
		for {
			n, _, err := b.conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}

			var msg ports.BrokerMessage
			if err := json.Unmarshal(buffer[:n], &msg); err == nil {
				msgChan <- msg
			} else {
				fmt.Println("Error unmarshalling message:", err)
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
