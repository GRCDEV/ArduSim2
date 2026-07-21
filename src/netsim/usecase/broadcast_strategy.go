package usecase

import (
	"netsim/domain/model"
	"time"
)

// BroadcastStrategy defines how a broadcast is dispatched to receivers.
// The implementation is chosen at startup based on config and never changes.
type BroadcastStrategy interface {
	// Enqueue handles a broadcast originating from a local UAV.
	Enqueue(s *Simulator, senderID, payload string, retries uint32, now time.Time)
	// HandlePeer handles a broadcast forwarded from another netsim node.
	HandlePeer(s *Simulator, senderID string, senderPos *model.Position, payload string, now time.Time)
}
