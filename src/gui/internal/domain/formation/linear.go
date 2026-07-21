package formation

// LinearFormation places all UAVs in a straight line with spacing between them.
type LinearFormation struct{}

func (f *LinearFormation) CalculateOffsets(numUAVs int, spacing float64) []Offset {
	offsets := make([]Offset, numUAVs)
	centerIndex := float64(numUAVs-1) / 2.0

	for i := 0; i < numUAVs; i++ {
		offsets[i] = Offset{
			X: (float64(i) - centerIndex) * spacing,
			Y: 0.0,
		}
	}
	return offsets
}
