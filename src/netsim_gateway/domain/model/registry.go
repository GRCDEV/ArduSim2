package model

import "net"

type UAVEntry struct {
	ID   string
	Addr *net.UDPAddr
}

type Subscriber struct {
	Addr   *net.UDPAddr
	Topics []string
}
