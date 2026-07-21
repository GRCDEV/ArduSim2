package service

import (
	"math"
	"math/rand"
	"netsim/domain/model"
)

const metersPerLatDeg = 111120.0

func ToSimPosition(telPos *model.TelemetryPosition) model.Position {
	metersPerLonDeg := metersPerLatDeg * math.Cos(telPos.Lat*math.Pi/180.0)
	return model.Position{
		X: telPos.Lon * metersPerLonDeg,
		Y: telPos.Lat * metersPerLatDeg,
		Z: telPos.Alt,
	}
}

var PassDistanceCheck func(senderPos, receiverPos *model.Position, maxRangeM float64) bool = PassDistanceCheckRealistic

func InitLossFunction(mode string) {
	switch mode {
	case "realistic":
		PassDistanceCheck = PassDistanceCheckRealistic
	case "fixed_range":
		PassDistanceCheck = PassDistanceCheckFixedRange
	case "unrestricted":
		PassDistanceCheck = func(senderPos, receiverPos *model.Position, maxRangeM float64) bool {
			return PassDistanceCheckUnrestricted()
		}
	default:
		// Default to realistic
		PassDistanceCheck = PassDistanceCheckRealistic
	}
}

// PassDistanceCheckRealistic probabilistic model
func PassDistanceCheckRealistic(senderPos, receiverPos *model.Position, maxRangeM float64) bool {
	dx := receiverPos.X - senderPos.X
	dy := receiverPos.Y - senderPos.Y
	dz := receiverPos.Z - senderPos.Z
	dSq := dx*dx + dy*dy + dz*dz

	if dSq > maxRangeM*maxRangeM {
		return false
	}

	d := math.Sqrt(dSq)
	lossProb := 5.335e-7*d*d + 3.395e-5*d
	return rand.Float64() > lossProb
}

func PassDistanceCheckFixedRange(senderPos, receiverPos *model.Position, maxRangeM float64) bool {
	dx := receiverPos.X - senderPos.X
	dy := receiverPos.Y - senderPos.Y
	dz := receiverPos.Z - senderPos.Z
	dSq := dx*dx + dy*dy + dz*dz

	return dSq <= maxRangeM*maxRangeM
}

func PassDistanceCheckUnrestricted() bool {
	return true
}
