package usecase

import (
	"encoding/json"
	"fmt"
	"netsim/domain/model"
	"time"
)

func (s *Simulator) SendMessages() {
	s.mu.Lock()
	now := time.Now()
	s.Spatial.PurgeExpiredTransmitters(now)

	// Dispatch pending
	if len(s.PendingMsgs) > 0 {
		newPendingMsgs := make(map[string][]model.Message)
		toDispatch := make(map[string][]model.Message)

		for recID, msgs := range s.PendingMsgs {
			detectCollisions(msgs)

			var keep []model.Message
			var dispatch []model.Message

			for _, m := range msgs {
				if now.Before(m.To) {
					// Still receiving, keep in pending
					keep = append(keep, m)
				} else {
					// Finished receiving
					if m.Overlapped {
						// Discard due to collision
						s.Logger.Info(fmt.Sprintf("Discarded message from %s to %s: collision detected", m.SenderID, recID), "timestamp", time.Now().Format(time.RFC3339Nano))
						if uav, ok := s.UAVs[recID]; ok {
							uav.BufferUsed -= len(m.Payload)
							if uav.BufferUsed < 0 {
								uav.BufferUsed = 0
							}
						}
					} else {
						// Ready to dispatch
						dispatch = append(dispatch, m)
						if uav, ok := s.UAVs[recID]; ok {
							uav.BufferUsed -= len(m.Payload)
							if uav.BufferUsed < 0 {
								uav.BufferUsed = 0
							}
						}
					}
				}
			}

			if len(keep) > 0 {
				newPendingMsgs[recID] = keep
			}
			if len(dispatch) > 0 {
				toDispatch[recID] = dispatch
			}
		}

		s.PendingMsgs = newPendingMsgs

		if len(toDispatch) > 0 {
			go s.dispatchToGateway(toDispatch)
		}
	}

	// Retry delayed
	oldDelayed := s.DelayedMsgs
	s.DelayedMsgs = nil

	for _, delayed := range oldDelayed {
		if now.Before(delayed.RetryAfter) {
			s.DelayedMsgs = append(s.DelayedMsgs, delayed)
		} else {
			// s.mu is locked, but EnqueueBroadcast locks it too. So we can't call EnqueueBroadcast directly while locked.
			// Instead, collect and call after unlock.
		}
	}
	
	var toRetry []model.Message
	for _, delayed := range oldDelayed {
		if !now.Before(delayed.RetryAfter) {
			toRetry = append(toRetry, delayed.Msg)
		}
	}
	s.mu.Unlock()

	for _, msg := range toRetry {
		s.EnqueueBroadcast(msg.SenderID, msg.Payload, msg.Retries+1, time.Now())
	}
}

func (s *Simulator) dispatchToGateway(pending map[string][]model.Message) {
	for recID, msgs := range pending {
		for _, msg := range msgs {
			deliver := map[string]any{
				"topic": "deliver",
				"payload": map[string]any{
					"target_uav_id": recID,
					"sender_id":     msg.SenderID,
					"payload":       msg.Payload,
				},
			}
			data, _ := json.Marshal(deliver)
			s.Sender.Send(data, nil)
			s.Logger.Info(fmt.Sprintf("Dispatched message from %s to %s via gateway", msg.SenderID, recID), "timestamp", time.Now().Format(time.RFC3339Nano))
		}
	}
}

func detectCollisions(msgs []model.Message) {
	for i := 0; i < len(msgs); i++ {
		for j := i + 1; j < len(msgs); j++ {
			if msgs[i].From.Before(msgs[j].To) && msgs[j].From.Before(msgs[i].To) {
				msgs[i].Overlapped = true
				msgs[j].Overlapped = true
			}
		}
	}
}
