package udp

import (
	"net"
	"netsim/ports/input"
	"netsim/ports/output"
	"time"
)

type Receiver struct {
	conn *net.UDPConn
	logger output.Logger
}

func NewReceiver(conn *net.UDPConn, logger output.Logger) *Receiver {
	return &Receiver{conn: conn, logger: logger}
}

func (r *Receiver) Run(handler input.MessageHandler) {
	buf := make([]byte, 65507)

	for {
		n, peerAddr, err := r.conn.ReadFromUDP(buf)
		receivedAt := time.Now()
		if err != nil {
			r.logger.Error("Error reading UDP", "error", err)
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		handler.Handle(input.RawPacket{
			Data:       data,
			Addr:       peerAddr,
			ReceivedAt: receivedAt,
		})
	}
}
