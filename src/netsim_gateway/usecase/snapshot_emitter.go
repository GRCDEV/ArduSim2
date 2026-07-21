package usecase

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func RunAggregatedSnapshotEmitter(gateway *Gateway) {
	ticker := time.NewTicker(time.Duration(gateway.Config.SnapshotIntervalS) * time.Second)
	for range ticker.C {
		snapshot := make(map[string]json.RawMessage)
		gateway.telemetryCache.Range(func(key, value any) bool {
			snapshot[key.(string)] = value.(json.RawMessage)
			return true
		})

		if len(snapshot) > 0 {
			msg := map[string]any{
				"topic": "telemetry_snapshot",
				"payload": map[string]any{
					"uavs": snapshot,
				},
			}
			if data, err := json.Marshal(msg); err == nil {
				gateway.uiSubscribers.Range(func(key, value any) bool {
					addr := value.(*net.UDPAddr)
					gateway.subscribersSender.Send(data, addr)
					return true
				})
				gateway.logger.Info(fmt.Sprintf("Emitted telemetry snapshot with %d UAVs", len(snapshot)))
			} else {
				gateway.logger.Error("Failed to marshal snapshot", "error", err)
			}
		}
	}
}
