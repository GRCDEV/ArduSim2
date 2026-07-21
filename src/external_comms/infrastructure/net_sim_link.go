package infrastructure

import (
	"encoding/json"
	"fmt"
	"net"

	"external_comms/domain"
)

type UDPNetSimLink struct {
	conn          *net.UDPConn
	telemetryAddr *net.UDPAddr
	messagesAddr  *net.UDPAddr
}

func NewUDPNetSimLink(ip string, telemetryPort, messagesPort int) (*UDPNetSimLink, error) {
	telemetryAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, telemetryPort))
	if err != nil {
		return nil, err
	}

	messagesAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, messagesPort))
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	return &UDPNetSimLink{
		conn:          conn,
		telemetryAddr: telemetryAddr,
		messagesAddr:  messagesAddr,
	}, nil
}

func (n *UDPNetSimLink) Send(msg domain.SendedNetSimMessage) error {
	valid_msg := map[string]interface{}{
		"payload": msg.Payload,
		"uav_id":  msg.Source,
	}
	data, err := json.Marshal(valid_msg)
	if err != nil {
		return err
	}

	var targetAddr *net.UDPAddr
	if msg.Topic == "telemetry" {
		targetAddr = n.telemetryAddr
	} else {
		targetAddr = n.messagesAddr
	}

	_, err = n.conn.WriteToUDP(data, targetAddr)
	return err
}

func (n *UDPNetSimLink) Listen() (<-chan domain.ReceivedNetSimMessage, error) {
	msgChan := make(chan domain.ReceivedNetSimMessage, 100)

	go func() {
		defer close(msgChan)
		buffer := make([]byte, 4096)
		for {
			bytesRead, _, err := n.conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}

			var msg domain.ReceivedNetSimMessage
			if err := json.Unmarshal(buffer[:bytesRead], &msg); err == nil {
				msgChan <- msg
			} else {
				fmt.Println(err.Error())
			}
		}
	}()

	return msgChan, nil
}

func (n *UDPNetSimLink) Close() error {
	if n.conn != nil {
		return n.conn.Close()
	}
	return nil
}
