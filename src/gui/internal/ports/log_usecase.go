package ports

import (
	"context"
	"ui/internal/domain"
)

type LogUseCase interface {
	LoadLogEntries(ctx context.Context) ([]string, error)
	SearchLogs(zipPath string, filter domain.LogFilter) ([]domain.LogMessage, error)
}
