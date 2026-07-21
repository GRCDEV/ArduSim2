package ports

import (
	"context"
	"ui/internal/domain"
)

type ConfigUseCase interface {
	LoadSimulationConfig(ctx context.Context) (*domain.SimulationState, error)
	SaveSimulationConfig(swarms []domain.Swarm, config domain.GeneralConfig, mode string) (*domain.SimulationState, error)
	DiscardCurrentRun(config domain.GeneralConfig) error
	SelectFile(ctx context.Context) (string, error)
	SelectArduPilotInstance(ctx context.Context) (string, error)
	SelectKubeConfig(ctx context.Context) (string, error)
	GetKmlFirstCoordinate(path string) (*domain.Coordinate, error)
	LoadFile(path string) (string, error)
	PopulateDefaults(config *domain.GeneralConfig)
}
