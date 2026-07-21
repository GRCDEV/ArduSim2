package usecase

import (
	"encoding/json"
	"fmt"
	"netsim/ports/output"
	"time"
)

func RunSnapshotEmitter(sim *Simulator, sender output.PacketSender, intervalS int, logger output.Logger) {
	ticker := time.NewTicker(time.Duration(intervalS) * time.Second)
	for range ticker.C {
		snapshot := make(map[string]json.RawMessage)
		sim.Cache.Range(func(key, value any) bool {
			snapshot[key.(string)] = value.(json.RawMessage)
			return true
		})

		if len(snapshot) > 0 {
			msg := map[string]any{
				"topic": "snapshot",
				"payload": map[string]any{
					"node_id": sim.NodeID,
					"uavs":    snapshot,
				},
			}
			if data, err := json.Marshal(msg); err == nil {
				sender.Send(data, nil)
				logger.Info(fmt.Sprintf("Emitted telemetry snapshot with %d UAVs", len(snapshot)))
			} else {
				logger.Error("Failed to marshal snapshot", "error", err)
			}
		}
	}
}
