package model

import "time"

type UAV struct {
	ID          string
	Position    Position
	BusyUntil   time.Time
	BufferUsed  int
	ChunkKey    ChunkKey
}

type ChunkKey struct {
	X int64
	Y int64
	Z int64
}
