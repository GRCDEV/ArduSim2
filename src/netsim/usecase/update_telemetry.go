package usecase

import (
	"encoding/json"
	"netsim/domain/model"
	"netsim/domain/service"
	"time"
)

func (s *Simulator) UpdateUAVTelemetry(uavID string, telemetry model.TelemetryData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos := service.ToSimPosition(&telemetry.Position)
	newChunk := s.Spatial.GetChunkKey(&pos)

	var oldChunk *model.ChunkKey
	if existing, ok := s.UAVs[uavID]; ok {
		old := existing.ChunkKey
		oldChunk = &old
		existing.Position = pos
		existing.ChunkKey = newChunk
	} else {
		s.UAVs[uavID] = &model.UAV{
			ID:         uavID,
			Position:   pos,
			BusyUntil:  time.Now(),
			BufferUsed: 0,
			ChunkKey:   newChunk,
		}
	}

	s.Spatial.UpdateUAVChunk(uavID, oldChunk, newChunk)

	if raw, err := json.Marshal(telemetry); err == nil {
		s.Cache.Store(uavID, json.RawMessage(raw))
	}
}
