package model

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type TelemetryPosition struct {
	Heading     float64 `json:"heading"`
	Alt         float64 `json:"alt"`
	RelativeAlt float64 `json:"relative_alt"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
}

type TelemetrySpeed struct {
	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`
	Vz float64 `json:"vz"`
}

type TelemetryData struct {
	NrGpsOnline int               `json:"nr_gps_online"`
	Position    TelemetryPosition `json:"position"`
	Type        string            `json:"type"`
	Battery     int               `json:"battery"`
	Version     string            `json:"version"`
	TimeBootMs  uint64            `json:"time_boot_ms"`
	Speed       TelemetrySpeed    `json:"speed"`
	Status      string            `json:"status"`
	FlightMode  string            `json:"flight_mode"`
}
