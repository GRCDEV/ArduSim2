package ports

import "external_comms/domain"

// Broker handles pub/sub events locally (Docker network)
type Broker interface {
	Connect(ip string, port int, createTopics []string) error
	Publish(payload map[string]interface{}) error
	Listen() (<-chan BrokerMessage, error)
	Close() error
}

type BrokerMessage struct {
	Topic   string `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

// NetSimLink handles raw UDP communication with the external Network Simulator
type NetSimLink interface {
	Send(msg domain.SendedNetSimMessage) error
	Listen() (<-chan domain.ReceivedNetSimMessage, error)
	Close() error
}

// LoggerLink handles raw UDP communication with the Logger microservice
type LoggerLink interface {
	SendLog(payload map[string]interface{}) error
	Close() error
}
