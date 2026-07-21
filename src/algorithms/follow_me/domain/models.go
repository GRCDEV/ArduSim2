package domain

import (
	"math"
	"time"
)

// State represents the current state of the algorithm
type State int

const (
	IDLE State = iota
	RUNNING
	TAKEOFF
)

// FollowMeMessage is the message broadcast by the Master
type FollowMeMessage struct {
	Lat         float64 `mapstructure:"lat" json:"lat"`
	Lon         float64 `mapstructure:"lon" json:"lon"`
	Alt         float64 `mapstructure:"alt" json:"alt"`
	RelativeAlt float64 `mapstructure:"relative_alt" json:"relative_alt"`
	Heading     float64 `mapstructure:"heading" json:"heading"`
	Timestamp   int64   `mapstructure:"timestamp" json:"timestamp"`
	DestSwarmId string `json:"dest_swarm_id,omitempty"`
}

func (f FollowMeMessage) FromTelemetry(t Telemetry, dstSwarmId string) FollowMeMessage {
	f.Lat = t.Position.Lat
	f.Lon = t.Position.Lon
	f.Alt = t.Position.Alt
	f.RelativeAlt = t.Position.RelativeAlt
	f.Heading = t.Position.Heading
	f.Timestamp = time.Now().UnixMilli()
	f.DestSwarmId = dstSwarmId

	return f
}

type CommandMessage struct {
	Command     string `mapstructure:"command" json:"command"`
	Source      string `json:"source"`
	DestSwarmId string `json:"dest_swarm_id,omitempty"`
}

// Telemetry represents the own position of a UAV
type Telemetry struct {
	NrGPSOnline int `mapstructure:"nr_gps_online"`
	Position    struct {
		Lat         float64 `mapstructure:"lat"`
		Lon         float64 `mapstructure:"lon"`
		Alt         float64 `mapstructure:"alt"`
		RelativeAlt float64 `mapstructure:"relative_alt"`
		Heading     float64 `mapstructure:"heading"`
	} `mapstructure:"position"`
	Speed struct {
		Vx float64 `mapstructure:"vx"`
		Vy float64 `mapstructure:"vy"`
		Vz float64 `mapstructure:"vz"`
	} `mapstructure:"speed"`
}

type Config struct {
	Role                  string  `json:"role"`
	BrokerIP              string  `json:"broker_ip"`
	BrokerPort            int     `json:"broker_port"`
	TelemetryTopic        string  `json:"telemetry_topic"`
	SuggestionsTopic      string  `json:"suggestions_topic"`
	BroadcastTopic        string  `json:"broadcast_topic"`
	SubscriptionTopic     string  `json:"subscription_topic"`
	SlavesTakeoffAltitude float64 `json:"slaves_takeoff_altitude"`
	SendPeriodMs          int     `json:"send_period_ms"`
	LogsTopic             string  `json:"logs_topic"`
	ServiceID             string  `json:"service_id"`
	SwarmId           string
}

// Suggestion endpoints for uav_controller
const (
	ActionArm           = "Arm"
	ActionTakeoff       = "Takeoff"
	ActionMoveTo        = "MoveToPosition"
	ActionLand          = "Land"
	ActionSetFlightmode = "SetFlightmode"
)

// Suggestion represents a command sent to the uav_controller
type Suggestion struct {
	Endpoint   string  `json:"endpoint"`
	ServiceID  string  `json:"service_id"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Altitude   float64 `json:"altitude,omitempty"`
	Flightmode string  `json:"flightmode,omitempty"`
}

// Helper for coordinate conversion (simple approximation for meters to degrees)
const (
	EarthRadius = 6378137.0
)

func MetersToDegrees(metersLat, metersLon, lat float64) (float64, float64) {
	dLat := metersLat / EarthRadius
	dLon := metersLon / (EarthRadius * math.Cos(math.Pi*lat/180.0))
	return dLat * (180.0 / math.Pi), dLon * (180.0 / math.Pi)
}
