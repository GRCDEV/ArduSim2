package infrastructure

import (
	"os"
	"strings"
	"time"

	"mission/ports"
)

type BrokerLogWriter struct {
	broker ports.Broker
	topic  string
}

func NewBrokerLogWriter(broker ports.Broker, topic string) *BrokerLogWriter {
	return &BrokerLogWriter{
		broker: broker,
		topic:  topic,
	}
}

func (w *BrokerLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))

	uavID := os.Getenv("UAV_ID")
	if uavID == "" {
		uavID = "unknown"
	}
	instanceID := "uav_" + uavID

	payload := map[string]interface{}{
		"InstanceID": instanceID,
		"ServiceID":  "mission",
		"Level":      "INFO",
		"Timestamp":  time.Now().Format(time.RFC3339Nano),
		"Message":    msg,
	}

	// Fire and forget
	w.broker.Publish(w.topic, payload)

	return len(p), nil
}
