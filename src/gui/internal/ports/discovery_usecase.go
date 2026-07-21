package ports

import "ui/internal/domain"

// DiscoveryUseCase defines the business actions for service discovery.
type DiscoveryUseCase interface {
	GetAvailableServices() []domain.ServiceType
	GetAvailableMixers() []domain.ServiceType
	GetAvailableControllers() []domain.ServiceType
}
