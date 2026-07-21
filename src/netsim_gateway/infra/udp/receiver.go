package udp

import (
	"net"
	"netsim_gateway/ports/input"
	"netsim_gateway/ports/output"
)

type Receiver struct {
	logger output.Logger
	conn *net.UDPConn
}

func NewReceiver(conn *net.UDPConn, logger output.Logger) *Receiver {
	return &Receiver{conn: conn, logger: logger}
}

func (r *Receiver) Run(handler interface{ Handle(input.RawPacket) }) {
	buf := make([]byte, 65507)

	for {
		n, peerAddr, err := r.conn.ReadFromUDP(buf)
		if err != nil {
			r.logger.Error("Error reading UDP", "error", err)
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		handler.Handle(input.RawPacket{
			Data: data,
			Addr: peerAddr,
		})
	}
}
