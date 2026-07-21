package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/GRCDEV/ArduSim2/logger/ports"
)

// TruncateLogsUseCase handles background log rotation
type TruncateLogsUseCase struct {
	storage  ports.LogStorage
	interval time.Duration
	stopChan chan struct{}
}

// NewTruncateLogsUseCase creates a new TruncateLogsUseCase
func NewTruncateLogsUseCase(storage ports.LogStorage, interval time.Duration) *TruncateLogsUseCase {
	return &TruncateLogsUseCase{
		storage:  storage,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins the background truncation process
func (uc *TruncateLogsUseCase) Start() {
	go func() {
		ticker := time.NewTicker(uc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), uc.interval)
				err := uc.storage.RotateLogs(ctx)
				cancel()
				if err != nil {
					fmt.Printf("[TruncateLogs] Error rotating logs: %v\n", err)
				}
			case <-uc.stopChan:
				return
			}
		}
	}()
}

// Stop stops the background process
func (uc *TruncateLogsUseCase) Stop() {
	close(uc.stopChan)
}
