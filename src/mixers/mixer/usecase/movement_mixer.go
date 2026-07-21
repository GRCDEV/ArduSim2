package usecase

import (
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"

	"application/domain"
	"application/ports"
)

type MovementMixer struct {
	broker ports.Broker
	uav    ports.UAVLink
	config *domain.AppConfig

	mu           sync.Mutex
	suggestions  []domain.Suggestion
}

func NewMovementMixer(b ports.Broker, u ports.UAVLink, c *domain.AppConfig) *MovementMixer {
	return &MovementMixer{
		broker: b,
		uav:    u,
		config: c,
	}
}

func (m *MovementMixer) Run() {
	msgChan, err := m.broker.Listen()
	if err != nil {
		log.Fatalf("Failed to listen on broker: %v", err)
	}

	ticker := time.NewTicker(time.Duration(m.config.MixWindowMs) * time.Millisecond)
	defer ticker.Stop()

	log.Println("Movement Mixer Started...")

	// Telemetry Bridge
	telemetryChan, err := m.uav.ListenTelemetry()
	if err != nil {
		log.Printf("Failed to listen for telemetry: %v", err)
	}

	for {
		select {
		case tel := <-telemetryChan:
			log.Printf("Forwarding UAV telemetry to broker topic: %s", m.config.TelemetryTopic)
			m.broker.Publish(m.config.TelemetryTopic, tel)
		case msg := <-msgChan:
			if msg.Topic == m.config.SuggestionsTopic {
				m.handleSuggestionArrival(msg.Payload)
			}
		case <-ticker.C:
			m.evaluateAndMixWindow()
		}
	}
}

func (m *MovementMixer) getPriority(serviceID string) int {
	for i, id := range m.config.ServicesPriority {
		if id == serviceID {
			return i
		}
	}
	return 999999 // Lowest priority if not in list
}

func (m *MovementMixer) handleSuggestionArrival(payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	var sug domain.Suggestion
	if err := json.Unmarshal(data, &sug); err == nil {
		log.Printf("Received suggestion: %s (Priority: %v)", sug.Endpoint, sug.IsCritical())
		m.mu.Lock()
		m.suggestions = append(m.suggestions, sug)
		m.mu.Unlock()
	} else {
		log.Printf("Failed to unmarshal suggestion: %v", err)
	}
}

func (m *MovementMixer) evaluateAndMixWindow() {
	m.mu.Lock()
	currentBatch := m.suggestions
	m.suggestions = nil // reset window
	m.mu.Unlock()

	if len(currentBatch) == 0 {
		return
	}

	log.Printf("Processing %d suggestions in current mix window", len(currentBatch))

	bestMoveToPriority := 9999999
	lastMoveToIdx := -1
	
	bestMoveByVecPriority := 9999999
	lastMoveByVecIdx := -1
	
	bestRotatePriority := 9999999
	lastRotateIdx := -1

	// For structural/priority commands, we keep all of them to respect sequential logic
	structuralIndices := []int{}
	
	// Flag to detect if a critical stopping command (Priority) is present
	hasCritical := false

	for i, req := range currentBatch {
		prio := m.getPriority(req.ServiceID)
		if req.IsStructural() {
			structuralIndices = append(structuralIndices, i)
			if req.IsCritical() {
				hasCritical = true
			}
		} else if req.IsMovement() {
			switch req.Endpoint {
			case domain.ActionMoveToPosition:
				if prio <= bestMoveToPriority {
					bestMoveToPriority = prio
					lastMoveToIdx = i
				}
			case domain.ActionMoveByVector:
				if prio <= bestMoveByVecPriority {
					bestMoveByVecPriority = prio
					lastMoveByVecIdx = i
				}
			case domain.ActionRotate:
				if prio <= bestRotatePriority {
					bestRotatePriority = prio
					lastRotateIdx = i
				}
			}
		} else {
			// Complementary commands (SetMessageInterval, etc.) are treated as structural (sequential)
			structuralIndices = append(structuralIndices, i)
		}
	}

	// If a critical command (Land/Brake/RTL) is present, discard all movements in this window
	if hasCritical {
		lastMoveToIdx = -1
		lastMoveByVecIdx = -1
		lastRotateIdx = -1
		log.Printf("Critical Priority detected: Discarding movement vectors in this window.")
	}

	// Collect all winning indices
	finalIndices := make([]int, 0)
	if lastMoveToIdx >= 0 {
		finalIndices = append(finalIndices, lastMoveToIdx)
	}
	if lastMoveByVecIdx >= 0 {
		finalIndices = append(finalIndices, lastMoveByVecIdx)
	}
	if lastRotateIdx >= 0 {
		finalIndices = append(finalIndices, lastRotateIdx)
	}
	finalIndices = append(finalIndices, structuralIndices...)

	// Sort indices to respect order of arrival, but primarily service priority
	sort.SliceStable(finalIndices, func(i, j int) bool {
		idxI := finalIndices[i]
		idxJ := finalIndices[j]
		prioI := m.getPriority(currentBatch[idxI].ServiceID)
		prioJ := m.getPriority(currentBatch[idxJ].ServiceID)
		if prioI != prioJ {
			return prioI < prioJ
		}
		return idxI < idxJ
	})

	// Execute suggestions in order
	for _, idx := range finalIndices {
		req := currentBatch[idx]
		log.Printf("Executing Suggestion (Idx %d): %v", idx, req.Endpoint)
		m.uav.SendSuggestion(req)

		// Broadcast stopping events if necessary
		if req.IsCritical() {
			m.broker.Publish(m.config.GlobalCommands, map[string]interface{}{
				"command": "stop",
			})
		}
	}
}
