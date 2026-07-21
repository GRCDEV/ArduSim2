package udp

import (
	"net"
	"netsim/ports/output"
)

func NewConnection(port int, logger output.Logger) *net.UDPConn {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		panic(err)
	}
	return conn
}