package domain

import "time"

// LogMessage represents the internal log structure
type LogMessage struct {
	InstanceID string    `json:"InstanceID"`
	ServiceID  string    `json:"ServiceID"`
	Level      string    `json:"Level"`
	Timestamp  string    `json:"Timestamp"` // As it comes from client (could be string or int, let's use string for simplicity or parse it)
	Message    string    `json:"Message"`
	EventID    string    `json:"EventID,omitempty"`
	ReceivedAt time.Time `json:"ReceivedAt"` // Enriched by the logger
}
