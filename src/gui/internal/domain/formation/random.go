package formation

import (
    "math/rand"
    "time"
)

// RandomFormation places UAVs in a larger jittered grid area.
type RandomFormation struct{}

func (f *RandomFormation) CalculateOffsets(numUAVs int, spacing float64) []Offset {
    offsets := make([]Offset, numUAVs)
    if numUAVs <= 0 {
        return offsets
    }

    // Seed the random number generator
    r := rand.New(rand.NewSource(time.Now().UnixNano()))

    // For Random, we assume a larger area (e.g., 2 times the area for minimum spacing)
    // and place drones randomly within it.
    // Length of the side of the square area.
    sideLength := float64(numUAVs) * spacing * 1.5

    for i := 0; i < numUAVs; i++ {
        // Center the random positions around (0,0)
        offsets[i] = Offset{
            X: (r.Float64() - 0.5) * sideLength,
            Y: (r.Float64() - 0.5) * sideLength,
        }
    }

    return offsets
}
