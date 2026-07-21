package domain

// AppConfig defines the connection parameters for the Local Broker and the Global Net Simulator
type AppConfig struct {
	BrokerIP   string `json:"broker_ip"`
	BrokerPort int    `json:"broker_port"`

	NetSimIP             string `json:"simulator_ip"`
	NetSimTelemetryPort  int    `json:"simulator_telemetry_port"`
	NetSimMessagesPort   int    `json:"simulator_messages_port"`

	LoggerIP   string `json:"logger_ip"`
	LoggerPort int    `json:"logger_port"`

	// Topics to subscribe from Local Broker
	SubTelemetryTopic string `json:"sub_telemetry_topic"`
	SubMessagesTopic  string `json:"sub_messages_topic"`
	SubLogsTopic      string `json:"sub_logs_topic"`

	// Topics to publish to Local Broker (coming from NetSim)
	UAVId   string
	SwarmId string
}

// SendedNetSimMessage envelope for sending/receiving data to the global network simulator
type SendedNetSimMessage struct {
	Topic   string                 `json:"topic"` // "telemetry" or "message"
	Source  string                 `json:"uav_id"`
	Payload map[string]interface{} `json:"payload"`
}

type ReceivedNetSimMessage struct {
	Source  string                 `json:"uav_id"`
	Payload map[string]interface{} `json:"payload"`
}
