package uav

import (
	"encoding/json"
	"fmt"
	"net"

	"application/domain"
)

type DirectUAVLink struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewDirectUAVLink(uavIP string, sendPort int, recvPort int) (*DirectUAVLink, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", uavIP, sendPort))
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", recvPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	return &DirectUAVLink{
		conn: conn,
		addr: serverAddr,
	}, nil
}

func (l *DirectUAVLink) SendSuggestion(s domain.Suggestion) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	
	_, err = l.conn.WriteToUDP(data, l.addr)
	return err
}

func (l *DirectUAVLink) ListenTelemetry() (<-chan map[string]interface{}, error) {
	msgChan := make(chan map[string]interface{})
	go func() {
		buf := make([]byte, 2048)
		for {
			n, _, err := l.conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			
			var payload map[string]interface{}
			if err := json.Unmarshal(buf[:n], &payload); err == nil {
				msgChan <- payload
			}
		}
	}()
	return msgChan, nil
}

func (l *DirectUAVLink) Close() error {
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}
