package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/GRCDEV/ArduSim2/logger/domain"
	"github.com/GRCDEV/ArduSim2/logger/ports"
)

// LogProcessorUseCase handles incoming logs asynchronously
type LogProcessorUseCase struct {
	storage ports.LogStorage
	logChan chan domain.LogMessage
}

// NewLogProcessorUseCase creates a new LogProcessorUseCase and starts workers
func NewLogProcessorUseCase(storage ports.LogStorage, bufferSize int, workers int) *LogProcessorUseCase {
	lp := &LogProcessorUseCase{
		storage: storage,
		logChan: make(chan domain.LogMessage, bufferSize),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		go lp.worker()
	}

	return lp
}

// Enqueue adds a log to the processing queue. Enriches it with ReceivedAt.
func (lp *LogProcessorUseCase) Enqueue(log domain.LogMessage) {
	log.ReceivedAt = time.Now()

	// Non-blocking enqueue
	select {
	case lp.logChan <- log:
	default:
		// Buffer is full, drop the log to avoid blocking the UDP server
		fmt.Println("[LogProcessor] Warning: log buffer full, dropping log")
	}
}

// worker processes logs from the queue
func (lp *LogProcessorUseCase) worker() {
	for log := range lp.logChan {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := lp.storage.SaveLog(ctx, log)
		cancel()

		if err != nil {
			fmt.Printf("[LogProcessor] Error saving log: %v\n", err)
		}
	}
}

// Close closes the processing channel
func (lp *LogProcessorUseCase) Close() {
	close(lp.logChan)
}
