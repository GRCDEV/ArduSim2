package usecases

import (
	"context"
	"io"

	"github.com/GRCDEV/ArduSim2/logger/ports"
)

// DownloadLogsUseCase handles the generation of the zip with all logs
type DownloadLogsUseCase struct {
	storage ports.LogStorage
}

// NewDownloadLogsUseCase creates a new DownloadLogsUseCase
func NewDownloadLogsUseCase(storage ports.LogStorage) *DownloadLogsUseCase {
	return &DownloadLogsUseCase{
		storage: storage,
	}
}

// Execute writes the zipped logs to the provided writer
func (uc *DownloadLogsUseCase) Execute(ctx context.Context, w io.Writer) error {
	return uc.storage.WriteZippedLogs(ctx, w)
}
