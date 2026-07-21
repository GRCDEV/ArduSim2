package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type LogMessage struct {
	InstanceID string `json:"InstanceID"`
	ServiceID  string `json:"ServiceID"`
	Level      string `json:"Level"`
	Timestamp  string `json:"Timestamp"`
	Message    string `json:"Message"`
}

type UDPLogger struct {
	addr       *net.UDPAddr
	conn       *net.UDPConn
	logChan    chan LogMessage
	instanceID string
	level      string
}

func NewUDPLogger(level string, address string) *UDPLogger {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Warning: failed to resolve logger address %s: %v\n", address, err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Warning: failed to dial logger address %s: %v\n", address, err)
	}

	nodeId := os.Getenv("NODE_ID")
	if nodeId == "" {
		nodeId = "netsim"
	}

	logger := &UDPLogger{
		addr:       addr,
		conn:       conn,
		logChan:    make(chan LogMessage, 1024),
		instanceID: nodeId,
		level:      level,
	}

	go logger.startWorker()
	return logger
}

func (l *UDPLogger) startWorker() {
	if l.conn == nil {
		return
	}
	for msg := range l.logChan {
		data, err := json.Marshal(msg)
		if err == nil {
			l.conn.Write(data)
		}
	}
}

func (l *UDPLogger) log(level, msg string, fields ...any) {
	fullMsg := msg
	for _, f := range fields {
		fullMsg += fmt.Sprintf(" %v", f)
	}

	fmt.Printf("[%s] %s\n", level, fullMsg)

	select {
	case l.logChan <- LogMessage{
		InstanceID: l.instanceID,
		ServiceID:  "netsim",
		Level:      level,
		Timestamp:  time.Now().Format(time.RFC3339),
		Message:    fullMsg,
	}:
	default:
		// Drop message if channel is full
	}
}

func (l *UDPLogger) Info(msg string, fields ...any) {
	if l.level == "debug" || l.level == "info" {
		l.log("INFO", msg, fields...)
	}
}

func (l *UDPLogger) Warn(msg string, fields ...any) {
	if l.level != "error" {
		l.log("WARN", msg, fields...)
	}
}

func (l *UDPLogger) Error(msg string, fields ...any) {
	l.log("ERROR", msg, fields...)
}
