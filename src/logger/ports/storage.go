package ports

import (
	"context"
	"io"

	"github.com/GRCDEV/ArduSim2/logger/domain"
)

// LogStorage defines how logs are saved and retrieved
type LogStorage interface {
	SaveLog(ctx context.Context, log domain.LogMessage) error
	WriteZippedLogs(ctx context.Context, w io.Writer) error
	RotateLogs(ctx context.Context) error
}
