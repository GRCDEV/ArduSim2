package model

import "encoding/json"

type TelemetrySnapshot struct {
	NodeID string                     `json:"node_id"`
	UAVs   map[string]json.RawMessage `json:"uavs"`
}
