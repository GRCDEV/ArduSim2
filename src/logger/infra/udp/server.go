package udp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/GRCDEV/ArduSim2/logger/domain"
	"github.com/GRCDEV/ArduSim2/logger/usecases"
)

type Server struct {
	address   string
	processor *usecases.LogProcessorUseCase
	conn      *net.UDPConn
}

func NewServer(address string, processor *usecases.LogProcessorUseCase) *Server {
	return &Server{
		address:   address,
		processor: processor,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		return err
	}

	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	fmt.Printf("[UDP Server] Listening on %s\n", s.address)

	// Buffer for reading incoming UDP packets
	buf := make([]byte, 65535)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, _, err := s.conn.ReadFromUDP(buf)
				if err != nil {
					// Ignore errors when connection is closed
					select {
					case <-ctx.Done():
						return
					default:
						fmt.Printf("[UDP Server] Error reading: %v\n", err)
						continue
					}
				}

				var logMsg domain.LogMessage
				if err := json.Unmarshal(buf[:n], &logMsg); err != nil {
					fmt.Printf("[UDP Server] Error parsing JSON: %v\n", err)
					continue
				}

				s.processor.Enqueue(logMsg)
			}
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
