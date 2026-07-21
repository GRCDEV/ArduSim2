package usecases

import (
	"ui/internal/domain"
	"ui/internal/ports"
)

type DiscoveryInteractor struct {
	repo ports.ConfigRepository
}

func NewDiscoveryInteractor(repo ports.ConfigRepository) *DiscoveryInteractor {
	return &DiscoveryInteractor{repo: repo}
}

func (i *DiscoveryInteractor) GetAvailableServices() []domain.ServiceType {
	return i.repo.GetAvailableServices()
}

func (i *DiscoveryInteractor) GetAvailableMixers() []domain.ServiceType {
	return i.repo.GetAvailableMixers()
}

func (i *DiscoveryInteractor) GetAvailableControllers() []domain.ServiceType {
	return i.repo.GetAvailableControllers()
}
