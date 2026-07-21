package service

import (
	"math"
	"netsim/domain/model"
	"time"
)

type SpatialGrid struct {
	chunkSizeM   float64
	Chunks       map[model.ChunkKey][]string
	Transmitters map[string]TransmitterState
}

type TransmitterState struct {
	Pos       model.Position
	BusyUntil time.Time
}

func NewSpatialGrid(chunkSizeM float64) *SpatialGrid {
	return &SpatialGrid{
		chunkSizeM:   chunkSizeM,
		Chunks:       make(map[model.ChunkKey][]string),
		Transmitters: make(map[string]TransmitterState),
	}
}

func (s *SpatialGrid) GetChunkKey(pos *model.Position) model.ChunkKey {
	return model.ChunkKey{
		X: int64(math.Floor(pos.X / s.chunkSizeM)),
		Y: int64(math.Floor(pos.Y / s.chunkSizeM)),
		Z: int64(math.Floor(pos.Z / s.chunkSizeM)),
	}
}

func (s *SpatialGrid) UpdateUAVChunk(uavID string, oldChunk *model.ChunkKey, newChunk model.ChunkKey) {
	chunkChanged := oldChunk == nil || *oldChunk != newChunk

	if oldChunk != nil && chunkChanged {
		s.removeFromChunk(uavID, *oldChunk)
	}

	// Only insert if chunk changed or it's the first registration.
	// Avoids duplicates when the UAV stays in the same chunk across telemetry updates.
	if chunkChanged {
		s.addToChunk(uavID, newChunk)
	}
}

func (s *SpatialGrid) removeFromChunk(uavID string, chunk model.ChunkKey) {
	ids, ok := s.Chunks[chunk]
	if !ok {
		return
	}

	var filtered []string
	for _, id := range ids {
		if id != uavID {
			filtered = append(filtered, id)
		}
	}

	if len(filtered) == 0 {
		delete(s.Chunks, chunk)
	} else {
		s.Chunks[chunk] = filtered
	}
}

func (s *SpatialGrid) addToChunk(uavID string, chunk model.ChunkKey) {
	// Guard against duplicates: only append if not already present.
	for _, id := range s.Chunks[chunk] {
		if id == uavID {
			return
		}
	}
	s.Chunks[chunk] = append(s.Chunks[chunk], uavID)
}

func (s *SpatialGrid) GetNearbyUAVIDs(center model.ChunkKey, radius int64, excludeID string) []string {
	var result []string
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			for dz := -radius; dz <= radius; dz++ {
				key := model.ChunkKey{X: center.X + dx, Y: center.Y + dy, Z: center.Z + dz}
				if ids, ok := s.Chunks[key]; ok {
					for _, id := range ids {
						if id != excludeID {
							result = append(result, id)
						}
					}
				}
			}
		}
	}
	return result
}

func (s *SpatialGrid) RecordTransmitter(uavID string, pos model.Position, busyUntil time.Time) {
	s.Transmitters[uavID] = TransmitterState{
		Pos:       pos,
		BusyUntil: busyUntil,
	}
}

func (s *SpatialGrid) PurgeExpiredTransmitters(now time.Time) {
	for id, state := range s.Transmitters {
		if state.BusyUntil.Before(now) || state.BusyUntil.Equal(now) {
			delete(s.Transmitters, id)
		}
	}
}
