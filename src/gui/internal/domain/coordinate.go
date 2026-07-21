package domain

// Coordinate represents a point in 3D space.
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	Alt float64 `json:"alt"`
}
