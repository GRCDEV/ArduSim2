package util

import "math"

const (
	// EarthRadius is the mean radius of Earth in meters.
	EarthRadius = 6371000.0
)

// AddOffset converts a metric (X, Y) offset from a given (Lat, Lon) origin
// into new absolute (Lat, Lon) coordinates using a spherical approximation.
// dx: offset in meters to the east (longitude axis)
// dy: offset in meters to the north (latitude axis)
func AddOffset(lat, lon float64, dx, dy float64) (float64, float64) {
	// Offset in radians
	dLat := dy / EarthRadius
	dLon := dx / (EarthRadius * math.Cos(math.Pi*lat/180.0))

	// Resulting coordinates in degrees
	newLat := lat + dLat*180.0/math.Pi
	newLon := lon + dLon*180.0/math.Pi

	return newLat, newLon
}
