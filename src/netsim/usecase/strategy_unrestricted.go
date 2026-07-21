package usecase

import (
	"fmt"
	"netsim/domain/model"
	"time"
)

// unrestrictedStrategy delivers messages to all UAVs with no restrictions:
// no CSMA, no retries, no distance/busy/buffer checks.
type unrestrictedStrategy struct{}

func (unrestrictedStrategy) Enqueue(s *Simulator, senderID, payload string, _ uint32, now time.Time) {
	sender, exists := s.UAVs[senderID]
	if !exists {
		return
	}

	count := 0
	for recID := range s.UAVs {
		if recID != senderID {
			s.deliverUnrestricted(senderID, recID, &sender.Position, payload, now)
			count++
		}
	}
	s.Logger.Info(fmt.Sprintf("Enqueuing broadcast from %s to %d receivers (unrestricted)", senderID, count), "timestamp", time.Now().Format(time.RFC3339Nano))
}

func (unrestrictedStrategy) HandlePeer(s *Simulator, senderID string, senderPos *model.Position, payload string, now time.Time) {
	for recID := range s.UAVs {
		if recID != senderID {
			s.deliverUnrestricted(senderID, recID, senderPos, payload, now)
		}
	}
}
