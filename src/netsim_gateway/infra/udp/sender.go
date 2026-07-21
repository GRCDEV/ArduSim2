package udp

import (
	"net"
	"netsim_gateway/ports/output"
)

type DispatchJob struct {
	Data []byte
	Addr *net.UDPAddr
}

type Sender struct {
	conn   *net.UDPConn
	jobs   chan DispatchJob
	logger output.Logger
}

func NewSender(conn *net.UDPConn, logger output.Logger) *Sender {
	return &Sender{
		conn:   conn,
		jobs:   make(chan DispatchJob, 1024),
		logger: logger,
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
		if job.Addr != nil {
			s.conn.WriteToUDP(job.Data, job.Addr)
		}
	}
}
