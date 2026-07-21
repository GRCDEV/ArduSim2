package ports

import (
	"ui/internal/domain"
)

// ContainerRuntime handles the execution lifecycle of simulated containers.
type ContainerRuntime interface {
	BuildCompose(composePath string) error
	GenerateBuildManifest(simDir string, dockerHubUser string, isKubernetes bool, swarms []domain.Swarm, config domain.GeneralConfig) error
	BuildAllImages(simDir string, dockerHubUser string, isKubernetes bool, swarms []domain.Swarm, config domain.GeneralConfig) error
	StartCompose(composePath string) error
	StopCompose(composePath string) error
	StartKubernetes(manifestPath, dockerHubRepository, kubeConfigPath string) (loggerIP, gatewayIP string, err error)
	StopKubernetes(manifestPath, kubeConfigPath string) error
	CollectKubernetesLogs(simName, destDir string) error
}
