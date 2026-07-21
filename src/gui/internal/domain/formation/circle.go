package formation

import "math"

// CircleFormation places one UAV in the center and the rest in a circle around it.
type CircleFormation struct{}

func (f *CircleFormation) CalculateOffsets(numUAVs int, spacing float64) []Offset {
	offsets := make([]Offset, numUAVs)
	if numUAVs <= 0 {
		return offsets
	}

	// UAV 0 is always at the center (0,0)
	offsets[0] = Offset{X: 0, Y: 0}
	if numUAVs == 1 {
		return offsets
	}

	// For more than 1 UAV, calculate the minimum radius needed to keep spacing.
	// Arc length = S = angle * radius
	// chord length = L = 2 * R * sin(angle/2)
	// We want L = spacing
	angleOffset := 2.0 * math.Pi / float64(numUAVs-1)
	radius := spacing / (2.0 * math.Sin(angleOffset/2.0))

	// If there are few UAVs, radius could be smaller than spacing, 
	// but the original ArduSim scales it if needed.
	if numUAVs <= 7 && radius < spacing {
		radius = spacing
	}

	for i := 1; i < numUAVs; i++ {
		angle := float64(i-1) * angleOffset
		offsets[i] = Offset{
			X: radius * math.Cos(angle),
			Y: radius * math.Sin(angle),
		}
	}

	return offsets
}
