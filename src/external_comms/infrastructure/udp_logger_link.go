package infrastructure

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type UDPLoggerLink struct {
	ip    string
	port  int
	conn  *net.UDPConn
	queue []map[string]interface{}
	mu    sync.Mutex
}

func NewUDPLoggerLink(ip string, port int) *UDPLoggerLink {
	link := &UDPLoggerLink{
		ip:    ip,
		port:  port,
		queue: make([]map[string]interface{}, 0),
	}
	_ = link.connect() // try to connect initially, ignore error
	return link
}

func (u *UDPLoggerLink) connect() error {
	if u.conn != nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", u.ip, u.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	u.conn = conn
	return nil
}

func (u *UDPLoggerLink) SendLog(payload map[string]interface{}) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Guardar siempre el nuevo mensaje en la cola
	u.queue = append(u.queue, payload)

	if err := u.connect(); err != nil {
		return fmt.Errorf("queued message (logger disconnected: %v)", err)
	}

	var remaining []map[string]interface{}
	connected := true

	for _, p := range u.queue {
		if !connected {
			// Si la conexión se perdió, guardamos el resto
			remaining = append(remaining, p)
			continue
		}

		data, err := json.Marshal(p)
		if err != nil {
			continue // descartamos payloads mal formados
		}

		_, err = u.conn.Write(data)
		if err != nil {
			// Fallo al enviar, cerramos para reintentar la próxima vez
			u.conn.Close()
			u.conn = nil
			connected = false
			remaining = append(remaining, p)
		}
	}

	// La cola ahora contiene solo los que no se han podido enviar
	u.queue = remaining

	if !connected {
		return fmt.Errorf("queued some messages (logger connection lost during send)")
	}
	return nil
}

func (u *UDPLoggerLink) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.conn != nil {
		err := u.conn.Close()
		u.conn = nil
		return err
	}
	return nil
}
