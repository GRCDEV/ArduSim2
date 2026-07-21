package usecase

import (
	"fmt"
	"netsim/domain/model"
	"netsim/domain/service"
	"time"
)

func (s *Simulator) EnqueueBroadcast(senderID, payload string, retries uint32, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Strategy.Enqueue(s, senderID, payload, retries, now)
}

func (s *Simulator) HandlePeerBroadcast(senderID string, senderPos *model.Position, payload string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Strategy.HandlePeer(s, senderID, senderPos, payload, now)
}

// processReceiver delivers a message to a receiver applying CSMA/distance checks.
func (s *Simulator) processReceiver(senderID, receiverID string, senderPos *model.Position, payload string, now, busyUntil time.Time, txNs uint64) {
	receiver, exists := s.UAVs[receiverID]
	if !exists {
		return
	}

	if !service.PassDistanceCheck(senderPos, &receiver.Position, s.Config.MaxRangeM) {
		s.Logger.Info(fmt.Sprintf("Discarded message from %s to %s: distance check failed", senderID, receiverID), "timestamp", time.Now().Format(time.RFC3339Nano))
		return
	}

	if now.Before(receiver.BusyUntil) {
		s.Logger.Info(fmt.Sprintf("Discarded message from %s to %s: receiver busy", senderID, receiverID), "timestamp", time.Now().Format(time.RFC3339Nano))
		return
	}

	if receiver.BufferUsed+len(payload) > s.Config.BufferSizeBytes {
		s.Logger.Info(fmt.Sprintf("Discarded message from %s to %s: buffer full", senderID, receiverID), "timestamp", time.Now().Format(time.RFC3339Nano))
		return
	}

	msg := model.Message{
		SenderID: senderID,
		Payload:  payload,
		From:     now,
		To:       busyUntil,
		TxNs:     txNs,
		Retries:  0,
	}

	s.PendingMsgs[receiverID] = append(s.PendingMsgs[receiverID], msg)
	receiver.BufferUsed += len(payload)
	s.Logger.Info(fmt.Sprintf("Accepted message from %s to %s", senderID, receiverID), "timestamp", time.Now().Format(time.RFC3339Nano))
}

// deliverUnrestricted delivers a message directly with no distance/busy/buffer checks.
func (s *Simulator) deliverUnrestricted(senderID, receiverID string, senderPos *model.Position, payload string, now time.Time) {
	receiver, exists := s.UAVs[receiverID]
	if !exists {
		return
	}

	msg := model.Message{
		SenderID: senderID,
		Payload:  payload,
		From:     now,
		To:       now,
		TxNs:     0,
		Retries:  0,
	}

	s.PendingMsgs[receiverID] = append(s.PendingMsgs[receiverID], msg)
	_ = receiver // referenced to avoid "declared and not used" if compiler inlines
}
