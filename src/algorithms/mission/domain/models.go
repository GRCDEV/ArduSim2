package domain

// Waypoint represents a geographic location parsed from a mission file.
type Waypoint struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

// State represents the current state of the mission algorithm.
type State int

const (
	IDLE State = iota
	TAKEOFF
	FLYING
	PAUSED
	LANDING
	FINISHED
)

type MissionEnd string

const (
	Unmodified MissionEnd = "unmodified"
	LandMode   MissionEnd = "land"
	RTLMode    MissionEnd = "rtl"
)

type YawStrategy string

const (
	FixedStrategy               YawStrategy = "Fixed"
	FaceNextWPStrategy          YawStrategy = "Face next Waypoint"
	FaceNextWPExceptRTLStrategy YawStrategy = "Face next Waypoint except RTL"
	FaceAlongGPSCourseStrategy  YawStrategy = "Face along GPS course"
)

// AppConfig represents the JSON execution configuration.
type AppConfig struct {
	MissionFile                     string      `json:"mission_file"`
	BrokerIP                        string      `json:"broker_ip"`
	BrokerPort                      int         `json:"broker_port"`
	SubscriptionTopic               string      `json:"subscription_topic"`
	PublishTopic                    string      `json:"publish_topic"`
	TelemetryTopic                  string      `json:"telemetry_topic"`
	DistanceToWaypointReached       float64     `json:"distance_to_waypoint_reached"`
	MinimumWaypointRelativeAltitude float64     `json:"minimum_waypoint_relative_altitude"`
	WaypointsRelativeAltitude       float64     `json:"waypoints_relative_altitude"`
	ExternalMessagesTopic           string      `json:"external_messages_topic"`
	MissionEnd                      MissionEnd  `json:"mission_end"`
	FinalAltitudeForRTL             float64     `json:"final_altitude_for_rtl"`
	InputMissionDelay               float64     `json:"input_mission_delay"`
	OverrideIncludedAltitudeValues  bool        `json:"override_include_altitude_values"`
	OverrideIncludedYawValues       bool        `json:"override_include_yaw_values"`
	YawValue                        YawStrategy `json:"yaw_value"`
	RelativeMovement                bool        `json:"relative_movement"`
	LogsTopic                       string      `json:"logs_topic"`
	ServiceID                       string      `json:"service_id"`
}

// BrokerMessage represents a message received from or sent to the broker.
type BrokerMessage struct {
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}
