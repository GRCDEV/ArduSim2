package broker

import (
	"os"
	"strings"
	"time"

	"application/ports"
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
		"ServiceID":  "mixer",
		"Level":      "INFO",
		"Timestamp":  time.Now().Format(time.RFC3339Nano),
		"Message":    msg,
	}

	// This is a fire-and-forget publish to not block the main application thread
	w.broker.Publish(w.topic, payload)

	return len(p), nil
}
