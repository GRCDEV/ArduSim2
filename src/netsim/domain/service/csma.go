package service

import (
	"netsim/domain/model"
	"time"
)

func HasNearTransmitters(spatial *SpatialGrid, senderPos *model.Position, csmaRangeM float64, now time.Time) bool {
	rangeSq := csmaRangeM * csmaRangeM
	for _, state := range spatial.Transmitters {
		if state.BusyUntil.After(now) {
			dx := state.Pos.X - senderPos.X
			dy := state.Pos.Y - senderPos.Y
			dz := state.Pos.Z - senderPos.Z
			if (dx*dx + dy*dy + dz*dz) <= rangeSq {
				return true
			}
		}
	}
	return false
}

func CalculateTxTimeNs(payloadLen int) uint64 {
	return 20_000 + 4_000*uint64((payloadLen+61)/3)
}
