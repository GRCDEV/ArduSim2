package input

import "net"

type RawPacket struct {
	Data []byte
	Addr *net.UDPAddr
}

type UAVMessageHandler interface {
	Handle(pkt RawPacket)
}

type NetsimMessageHandler interface {
	Handle(pkt RawPacket)
}
