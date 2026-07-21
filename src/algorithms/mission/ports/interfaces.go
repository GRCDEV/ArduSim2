package ports

import "mission/domain"

// ConfigLoader is an interface for reading the application configuration.
type ConfigLoader interface {
	LoadAppConfig(filePath string) (*domain.AppConfig, error)
}

// MissionParser is an interface for parsing mission description files.
type MissionParser interface {
	ParseMission(filePath string) ([]domain.Waypoint, error)
}

// Broker is an interface for pub/sub operations.
type Broker interface {
	Connect(ip string, port int, subTopic, telemetryTopic string) error
	Publish(topic string, payload map[string]interface{}) error
	Listen() (<-chan domain.BrokerMessage, error)
	Close() error
}
