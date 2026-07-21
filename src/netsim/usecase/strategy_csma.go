package usecase

import (
	"fmt"
	"netsim/domain/model"
	"netsim/domain/service"
	"time"
)

// csmaStrategy applies CSMA/CA with retries and distance checks.
// Used for "realistic" and "fixed_range" loss modes.
type csmaStrategy struct{}

func (csmaStrategy) Enqueue(s *Simulator, senderID, payload string, retries uint32, now time.Time) {
	sender, exists := s.UAVs[senderID]
	if !exists {
		return
	}

	payloadLen := len(payload)
	txNs := service.CalculateTxTimeNs(payloadLen)
	busyUntil := now.Add(time.Duration(txNs))

	if now.Before(sender.BusyUntil) {
		if retries >= s.Config.MaxCsmaRetries {
			s.Logger.Warn("Max retries exceeded for sender", "sender", senderID, "timestamp", time.Now().Format(time.RFC3339Nano))
		} else {
			s.Logger.Info(fmt.Sprintf("Delayed message from %s: sender busy", senderID), "timestamp", time.Now().Format(time.RFC3339Nano))
			s.DelayedMsgs = append(s.DelayedMsgs, model.DelayedMessage{
				Msg: model.Message{
					SenderID: senderID,
					Payload:  payload,
					Retries:  retries,
				},
				RetryAfter: now.Add(time.Millisecond * time.Duration(1<<retries)),
			})
		}
		return
	}

	if service.HasNearTransmitters(s.Spatial, &sender.Position, s.Config.CsmaRangeM, now) {
		if retries >= s.Config.MaxCsmaRetries {
			s.Logger.Warn("Max retries exceeded (CSMA) for sender", "sender", senderID, "timestamp", time.Now().Format(time.RFC3339Nano))
		} else {
			s.Logger.Info(fmt.Sprintf("Delayed message from %s: near transmitters (CSMA)", senderID), "timestamp", time.Now().Format(time.RFC3339Nano))
			s.DelayedMsgs = append(s.DelayedMsgs, model.DelayedMessage{
				Msg: model.Message{
					SenderID: senderID,
					Payload:  payload,
					Retries:  retries,
				},
				RetryAfter: now.Add(time.Millisecond * time.Duration(1<<retries)),
			})
		}
		return
	}

	s.Spatial.RecordTransmitter(senderID, sender.Position, busyUntil)
	sender.BusyUntil = busyUntil

	receiverIDs := s.Spatial.GetNearbyUAVIDs(sender.ChunkKey, s.Config.ChunkRadius, senderID)
	s.Logger.Info(fmt.Sprintf("Enqueuing broadcast from %s to %d receivers", senderID, len(receiverIDs)), "timestamp", time.Now().Format(time.RFC3339Nano))
	for _, recID := range receiverIDs {
		s.processReceiver(senderID, recID, &sender.Position, payload, now, busyUntil, txNs)
	}
}

func (csmaStrategy) HandlePeer(s *Simulator, senderID string, senderPos *model.Position, payload string, now time.Time) {
	payloadLen := len(payload)
	txNs := service.CalculateTxTimeNs(payloadLen)
	busyUntil := now.Add(time.Duration(txNs))

	centerChunk := s.Spatial.GetChunkKey(senderPos)
	receiverIDs := s.Spatial.GetNearbyUAVIDs(centerChunk, s.Config.ChunkRadius, "")
	for _, recID := range receiverIDs {
		if recID != senderID {
			s.processReceiver(senderID, recID, senderPos, payload, now, busyUntil, txNs)
		}
	}
}
