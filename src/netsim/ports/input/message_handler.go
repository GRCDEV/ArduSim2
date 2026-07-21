package input

import (
	"net"
	"time"
)

type RawPacket struct {
	Data       []byte
	Addr       *net.UDPAddr
	ReceivedAt time.Time
}

type MessageHandler interface {
	Handle(pkt RawPacket)
}
