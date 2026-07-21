package config

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
)

type Config struct {
	TelemetryPort     int      `json:"telemetry_port"`
	MessagesPort      int      `json:"messages_port"`
	SubscribersPort   int      `json:"subscribers_port"`
	NetsimListenPort  int      `json:"netsim_listen_port"`
	SnapshotIntervalS int      `json:"snapshot_interval_s"`
	LoggerIp          string   `json:"logger_ip"`
	LoggerPort        int      `json:"logger_port"`
	Addrs             []string 
}

func LoadConfig(path string) Config {
	file, err := os.Open(path)
	addrsStr := os.Getenv("ADDRS")
	cfg := Config{
		TelemetryPort:     3000,
		MessagesPort:      3001,
		SubscribersPort:   3002,
		NetsimListenPort:  3003,
		SnapshotIntervalS: 1,
		LoggerIp:          "logger",
		LoggerPort:        5000,
		Addrs:             strings.Split(addrsStr, ","),
	}

	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&cfg)
	}

	return cfg
}

func DiscoverNetsims(addrs []string) []*net.UDPAddr {
	var resolved []*net.UDPAddr
	for _, addrStr := range addrs {
		if addr, err := net.ResolveUDPAddr("udp", addrStr); err == nil {
			resolved = append(resolved, addr)
		} else {
			log.Printf("Failed to resolve %s: %v\n", addrStr, err)
		}
	}
	return resolved
}
