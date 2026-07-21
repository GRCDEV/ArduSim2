package udp

import (
	"net"
	"netsim/ports/output"
	"time"
)

type DispatchJob struct {
	Data []byte
	Addr *net.UDPAddr
}

type Sender struct {
	conn       *net.UDPConn
	targetAddr *net.UDPAddr
	jobs       chan DispatchJob
	logger     output.Logger
}

func NewSender(conn *net.UDPConn, targetAddrStr string, logger output.Logger) *Sender {
	var addr *net.UDPAddr
	var err error
	for {
		addr, err = net.ResolveUDPAddr("udp", targetAddrStr)
		if err != nil {
			logger.Error("Failed to resolve target address", "address", targetAddrStr, "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break	
	}

	return &Sender{
		conn:       conn,
		targetAddr: addr,
		jobs:       make(chan DispatchJob, 1024),
		logger:     logger,
	}
}

func (s *Sender) Send(data []byte, addr *net.UDPAddr) error {
	select {
	case s.jobs <- DispatchJob{Data: data, Addr: addr}:
	default:
		s.logger.Warn("Sender channel full, dropping packet")
	}
	return nil
}

func (s *Sender) Run() {
	if s.conn == nil {
		return
	}
	for job := range s.jobs {
		target := job.Addr
		if target == nil {
			target = s.targetAddr
		}
		if target != nil {
			s.conn.WriteToUDP(job.Data, target)
		}
	}
}
