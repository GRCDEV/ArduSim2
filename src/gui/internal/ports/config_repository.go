package ports

import "ui/internal/domain"

// ConfigRepository handles persistence of simulation data and service discovery.
type ConfigRepository interface {
	GetAvailableServices() []domain.ServiceType
	GetAvailableMixers() []domain.ServiceType
	GetAvailableControllers() []domain.ServiceType
	SaveSimulation(simDir string, state domain.SimulationState) error
	LoadSimulation(stateFile string) (*domain.SimulationState, error)
	GetSimulationsDir() string
	ReadFile(filePath string) (string, error)
	GetFiles(srcDir string, ext string) ([]string, error)
}
