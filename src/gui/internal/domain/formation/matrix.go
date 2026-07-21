package formation

// MatrixFormation places UAVs in an expanding grid pattern.
type MatrixFormation struct{}

func (f *MatrixFormation) CalculateOffsets(numUAVs int, spacing float64) []Offset {
	offsets := make([]Offset, numUAVs)
	if numUAVs <= 0 {
		return offsets
	}

	// UAV 0 is at (0,0) center.
	offsets[0] = Offset{X: 0, Y: 0}

	// For more UAVs, use a simple spiral grid placement.
	// Step distance 1, 1, 2, 2, 3, 3, etc.
	// Direction: East, North, West, South.
	x, y := 0, 0
	dx, dy := 1, 0
	stepSize := 1
	stepCount := 0
	dirCount := 0

	for i := 1; i < numUAVs; i++ {
		x += dx
		y += dy
		offsets[i] = Offset{X: float64(x) * spacing, Y: float64(y) * spacing}

		stepCount++
		if stepCount == stepSize {
			stepCount = 0
			// Turn 90 degrees CCW
			dx, dy = -dy, dx
			dirCount++
			if dirCount == 2 {
				dirCount = 0
				stepSize++
			}
		}
	}

	return offsets
}
