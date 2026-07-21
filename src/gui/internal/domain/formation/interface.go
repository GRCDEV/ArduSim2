package formation

import (
	"strings"
)

// Offset represents a position in metric units relative to a formation center.
type Offset struct {
	X float64 // meters (east/longitude axis)
	Y float64 // meters (north/latitude axis)
}

// Formation defines the Strategy interface for calculating UAV positions.
type Formation interface {
	CalculateOffsets(numUAVs int, spacing float64) []Offset
}

// GetFormation returns the appropriate Formation strategy based on the layout ID.
func GetFormation(layout string) Formation {
	switch strings.ToUpper(layout) {
	case "LINEAR":
		return &LinearFormation{}
	case "MATRIX":
		return &MatrixFormation{}
	case "CIRCLE":
		return &CircleFormation{}
	case "RANDOM":
		return &RandomFormation{}
	default:
		// Default to Linear if layout is unrecognized or empty.
		return &LinearFormation{}
	}
}
