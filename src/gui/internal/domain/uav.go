package domain

// UAV groups a UAV identifier with its set of deployed services.
type UAV struct {
	ID                string            `json:"id"`
	Services          []DeployedService `json:"services"`
	Mixer             *DeployedService  `json:"mixer,omitempty"`
	Controller        *DeployedService  `json:"controller,omitempty"`
	Speed             *float64          `json:"speed,omitempty"`
	BatteryCapacity   *int              `json:"batteryCapacity,omitempty"`
	HomeOverride      *Coordinate       `json:"homeOverride,omitempty"`
	ArduPilotInstance *string           `json:"arduPilotInstance,omitempty"`
	NodeLabel         *string           `json:"nodeLabel,omitempty"`
}
