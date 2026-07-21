package tests

import (
	"testing"
	"time"

	"mission/domain"
	"mission/usecase"
)

// MockBroker implements ports.Broker
type MockBroker struct {
	PublishedMessages []domain.BrokerMessage
	msgChan           chan domain.BrokerMessage
}

func NewMockBroker() *MockBroker {
	return &MockBroker{
		msgChan: make(chan domain.BrokerMessage, 10),
	}
}

func (m *MockBroker) Connect(ip string, port int, subTopic, telemetryTopic string) error {
	return nil
}

func (m *MockBroker) Publish(topic string, payload map[string]interface{}) error {
	m.PublishedMessages = append(m.PublishedMessages, domain.BrokerMessage{
		Topic:   topic,
		Payload: payload,
	})
	return nil
}

func (m *MockBroker) Listen() (<-chan domain.BrokerMessage, error) {
	return m.msgChan, nil
}

func (m *MockBroker) Close() error {
	close(m.msgChan)
	return nil
}

func (m *MockBroker) SimulateMessage(topic string, cmd string) {
	payload := map[string]interface{}{
		"command": cmd,
	}
	m.msgChan <- domain.BrokerMessage{
		Topic:   topic,
		Payload: payload,
	}
}

func (m *MockBroker) SimulateTelemetry(topic string, lat, lon, alt float64) {
	m.msgChan <- domain.BrokerMessage{
		Topic: topic,
		Payload: map[string]interface{}{
			"lat":          lat,
			"lon":          lon,
			"relative_alt": alt,
		},
	}
}

// MockConfigLoader implements ports.ConfigLoader
type MockConfigLoader struct{}

func (m *MockConfigLoader) LoadAppConfig(filePath string) (*domain.AppConfig, error) {
	return &domain.AppConfig{
		BrokerIP:                        "127.0.0.1",
		BrokerPort:                      3000,
		SubscriptionTopic:               "algo/mission",
		PublishTopic:                    "uav/1/cmd",
		TelemetryTopic:                  "uav/1/telemetry",
		DistanceToWaypointReached:       5.0,
		MinimumWaypointRelativeAltitude: 10.0,
		WaypointsRelativeAltitude:       15.0,
		ExternalMessagesTopic:           "external/messages",
	}, nil
}

// MockMissionParser implements ports.MissionParser
type MockMissionParser struct{}

func (m *MockMissionParser) ParseMission(filePath string) ([]domain.Waypoint, error) {
	return []domain.Waypoint{
		{Latitude: 39.4816, Longitude: -0.3492, Altitude: 0},
		{Latitude: 39.4797, Longitude: -0.3424, Altitude: 10},
		{Latitude: 39.4834, Longitude: -0.3407, Altitude: 15},
	}, nil
}

func TestMissionManager(t *testing.T) {
	mockBroker := NewMockBroker()
	mockConfig := &MockConfigLoader{}
	mockParser := &MockMissionParser{}

	manager := usecase.NewMissionManager(mockBroker, mockConfig, mockParser)

	err := manager.Initialize("dummy.json")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Run in background since it listens to channel
	go manager.Run()
	time.Sleep(100 * time.Millisecond) // Give it time to start

	// Phase 1: Send "start" (Change State)
	mockBroker.SimulateMessage("algo/mission", "start")
	time.Sleep(100 * time.Millisecond)

	// Check if Arm, GUIDED, and Takeoff were published
	if len(mockBroker.PublishedMessages) < 3 {
		t.Fatalf("Expected at least 3 messages (Arm, Guided, Takeoff), got %d", len(mockBroker.PublishedMessages))
	}

	takeoffMsg := mockBroker.PublishedMessages[2]
	if takeoffMsg.Payload["endpoint"] != "Takeoff" || takeoffMsg.Payload["altitude"] != 10.0 {
		t.Errorf("Expected Takeoff to 10.0, got %v", takeoffMsg.Payload)
	}

	// Phase 2: Send telemetry tracking altitude
	mockBroker.PublishedMessages = []domain.BrokerMessage{} // clear
	mockBroker.SimulateTelemetry("uav/1/telemetry", 39.4816, -0.3492, 10.0)
	time.Sleep(100 * time.Millisecond)

	if len(mockBroker.PublishedMessages) < 1 {
		t.Fatalf("Expected MoveToPosition message, got 0")
	}
	moveMsg := mockBroker.PublishedMessages[0]
	if moveMsg.Payload["endpoint"] != "MoveToPosition" {
		t.Errorf("Expected MoveToPosition, got %v", moveMsg.Payload)
	}

	// Phase 3: Send telemetry reaching waypoint 2
	mockBroker.PublishedMessages = []domain.BrokerMessage{} // clear
	// Waypoint 2 is: {Latitude: 39.4834, Longitude: -0.3407}
	mockBroker.SimulateTelemetry("uav/1/telemetry", 39.4834, -0.3407, 15.0)
	time.Sleep(100 * time.Millisecond)

	if len(mockBroker.PublishedMessages) < 1 {
		t.Fatalf("Expected Land message, got 0")
	}
	landMsg := mockBroker.PublishedMessages[0]
	if landMsg.Payload["endpoint"] != "Land" {
		t.Errorf("Expected Land, got %v", landMsg.Payload)
	}

	// Phase 4: Send telemetry for landing
	mockBroker.PublishedMessages = []domain.BrokerMessage{} // clear
	mockBroker.SimulateTelemetry("uav/1/telemetry", 39.4834, -0.3407, 0.2) // near ground
	time.Sleep(100 * time.Millisecond)

	if len(mockBroker.PublishedMessages) < 1 {
		t.Fatalf("Expected external messages finish command, got 0")
	}
	finishMsg := mockBroker.PublishedMessages[0]
	if finishMsg.Topic != "external/messages" {
		t.Errorf("Expected topic external/messages, got %s", finishMsg.Topic)
	}
	if finishMsg.Payload["command"] != "finish" {
		t.Errorf("Expected command finish, got %v", finishMsg.Payload)
	}
}
