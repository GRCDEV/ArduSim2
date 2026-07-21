package output

import "net"

type PacketSender interface {
	Send(data []byte, addr *net.UDPAddr) error
}
