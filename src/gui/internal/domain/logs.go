package domain

import "time"

// LogMessage represents a single log entry emitted by a microservice.
type LogMessage struct {
	InstanceID string    `json:"InstanceID"`
	ServiceID  string    `json:"ServiceID"`
	Level      string    `json:"Level"`
	Timestamp  time.Time `json:"Timestamp"`
	Message    string    `json:"Message"`
	EventID    string    `json:"EventID,omitempty"`
	ReceivedAt time.Time `json:"ReceivedAt"` // Enriched by the logger service
}

// LogFilter defines the criteria for filtering logs in the viewer.
type LogFilter struct {
	InstanceID string `json:"InstanceID"`
	ServiceID  string `json:"ServiceID"`
	Level      string `json:"Level"`
	EventID    string `json:"EventID"`
	SearchText string `json:"SearchText"` // Searches within the Message field
}
