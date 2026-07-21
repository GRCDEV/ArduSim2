package ports

import "application/domain"

// Broker is a generic pub/sub interface (Internal to Docker Network)
type Broker interface {
	Connect(ip string, port int, createTopics []string) error
	Publish(topic string, payload map[string]interface{}) error
	Listen() (<-chan BrokerMessage, error)
	Close() error
}

type BrokerMessage struct {
	Topic   string `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

// UAVLink directly interfaces via UDP with ArduPilot uav_controller
type UAVLink interface {
	SendSuggestion(suggestion domain.Suggestion) error
	ListenTelemetry() (<-chan map[string]interface{}, error)
	Close() error
}

// ConfigLoader is an interface for reading JSON configs
type ConfigLoader interface {
	LoadAppConfig(filePath string) (*domain.AppConfig, error)
}
