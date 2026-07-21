package domain

type ActionType string

const (
	ActionMoveToPosition     ActionType = "MoveToPosition"
	ActionMoveByVector       ActionType = "MoveByVector"
	ActionSetFlightmode      ActionType = "SetFlightmode"
	ActionLand               ActionType = "Land"
	ActionTakeoff            ActionType = "Takeoff"
	ActionRotate             ActionType = "Rotate"
	ActionArm                ActionType = "Arm"
	ActionDisarm             ActionType = "Disarm"
	ActionRecoverControl     ActionType = "RecoverControl"
	ActionSetMessageInterval ActionType = "SetMessageInterval"
	ActionRequestMessage     ActionType = "RequestMessage"
)

// Suggestion represents a JSON command payload arriving from an algorithm
type Suggestion struct {
	Endpoint  ActionType `json:"endpoint"`
	ServiceID string     `json:"service_id,omitempty"`

	// PositionData
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`

	// VelocityData
	VX float64 `json:"vx,omitempty"`
	VY float64 `json:"vy,omitempty"`
	VZ float64 `json:"vz,omitempty"`

	// RotationData
	Yaw       float64 `json:"yaw,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
	Direction int     `json:"direction,omitempty"`
	Relative  int     `json:"relative,omitempty"`

	// Configuration
	Flightmode string `json:"flightmode,omitempty"`
	MessageID  int    `json:"messageID,omitempty"`
}

// AppConfig is the initialization struct pointing to the controller and the broker
type AppConfig struct {
	BrokerIP          string `json:"broker_ip"`
	BrokerPort        int    `json:"broker_port"`
	UAVControllerIP   string `json:"uav_controller_ip"`
	UAVControllerPort int    `json:"uav_controller_port"`
	UAVTelemetryPort  int    `json:"uav_telemetry_port"`

	TelemetryTopic   string `json:"telemetry_topic"`
	SuggestionsTopic string `json:"suggestions_topic"`
	GlobalCommands   string `json:"global_commands"`
	LogsTopic        string `json:"logs_topic"`

	MixWindowMs      int      `json:"mix_window_ms"` // The time window (ms) to collect and merge suggestions
	ServicesPriority []string `json:"services_priority"`
}

// IsStructural returns true for state-changing or sequence-critical commands
func (s Suggestion) IsStructural() bool {
	switch s.Endpoint {
	case ActionArm, ActionDisarm, ActionSetFlightmode, ActionTakeoff, ActionLand, ActionRecoverControl:
		return true
	default:
		return false
	}
}

// IsMovement returns true for continuous vector-based commands that can be 'slotted'
func (s Suggestion) IsMovement() bool {
	switch s.Endpoint {
	case ActionMoveToPosition, ActionMoveByVector, ActionRotate:
		return true
	default:
		return false
	}
}

// IsCritical returns true for commands that should abort current movement in the window
func (s Suggestion) IsCritical() bool {
	if s.Endpoint == ActionLand || s.Flightmode == "RTL" || s.Flightmode == "BRAKE" || s.Flightmode == "LAND" {
		return true
	}
	return false
}
