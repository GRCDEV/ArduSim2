package ports

import (
	"context"
	"ui/internal/domain"
)

// SimulationUseCase defines the business actions related to the simulation lifecycle.
type SimulationUseCase interface {
	StartSimulation(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error
	BuildImages(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error
	StopSimulation()
	SendAlgorithmCommand(serviceId string, command string) error
	DownloadLogs() error
	ExportSimulation(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig) error
}
