package domain

// TelemetryData represents the state of a UAV as reported by the network simulator.
type TelemetryData struct {
	NrGpsOnline int `mapstructure:"nr_gps_online" json:"nr_gps_online"`
	Position    struct {
		Heading     float64 `mapstructure:"heading" json:"heading"`
		Alt         float64 `mapstructure:"alt" json:"alt"`
		RelativeAlt float64 `mapstructure:"relative_alt" json:"relative_alt"`
		Lon         float64 `mapstructure:"lon" json:"lon"`
		Lat         float64 `mapstructure:"lat" json:"lat"`
	} `mapstructure:"position" json:"position"`
	Type       string `mapstructure:"type" json:"type"`
	Battery    int    `mapstructure:"battery" json:"battery"`
	Version    string `json:"version"`
	TimeBootMs uint64 `mapstructure:"time_boot_ms" json:"time_boot_ms"`
	Speed      struct {
		VX float64 `mapstructure:"vx" json:"vx"`
		VY float64 `mapstructure:"vy" json:"vy"`
		VZ float64 `mapstructure:"vz" json:"vz"`
	} `mapstructure:"speed" json:"speed"`
	Status     string `json:"status"`
	FlightMode string `mapstructure:"flight_mode" json:"flight_mode"`
}
