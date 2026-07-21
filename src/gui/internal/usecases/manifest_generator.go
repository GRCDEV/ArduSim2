package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"ui/internal/domain"
	"ui/internal/ports"
)

type ManifestGenerator struct {
	projectRoot         string
	externalCommsConfig string
	netsimGatewayConfig string
	netsimConfig        string
	loggerConfig        string
	algorithmsDir       string
	mixersDir           string
	controllersDir      string
	newComposeBuilder   ports.ComposeBuilderFactory
	newKubernetesBuilder ports.KubernetesBuilderFactory
}

func NewManifestGenerator(
	projectRoot string,
	composeBuilderFactory ports.ComposeBuilderFactory,
	kubernetesBuilderFactory ports.KubernetesBuilderFactory,
) *ManifestGenerator {
	base := filepath.Clean(projectRoot)
	return &ManifestGenerator{
		projectRoot:          base,
		externalCommsConfig:  filepath.Join(base, "..", "external_comms", "config.json"),
		netsimGatewayConfig:  filepath.Join(base, "..", "netsim_gateway", "config.json"),
		netsimConfig:         filepath.Join(base, "..", "netsim", "config.json"),
		loggerConfig:         filepath.Join(base, "..", "logger", "config.json"),
		algorithmsDir:        filepath.Join(base, "..", "algorithms"),
		mixersDir:            filepath.Join(base, "..", "mixers"),
		controllersDir:       filepath.Join(base, "..", "controllers"),
		newComposeBuilder:    composeBuilderFactory,
		newKubernetesBuilder: kubernetesBuilderFactory,
	}
}

func (g *ManifestGenerator) Generate(swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool, simDir string) (string, error) {
	resDir := filepath.Join(simDir, "resources")
	_ = os.RemoveAll(resDir)
	if err := os.MkdirAll(resDir, 0755); err != nil {
		return "", fmt.Errorf("create simulation dirs: %w", err)
	}

	if isLocal {
		return g.buildLocalCompose(swarms, config, resDir, simDir)
	}
	return g.buildKubernetesManifests(swarms, config, resDir, simDir)
}

func normalizeNetsimInstances(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func (g *ManifestGenerator) getImageName(image string, user string) string {
	if user != "" {
		return fmt.Sprintf("%s:%s", user, image)
	}
	return image
}

func (g *ManifestGenerator) buildNetsimOverrides(config domain.GeneralConfig) map[string]interface{} {
	if config.NetsimMode == "" {
		return nil
	}
	overrides := map[string]interface{}{
		"loss_mode": config.NetsimMode,
	}
	if config.NetsimMode == "fixed_range" && config.NetsimMaxRangeM != nil {
		overrides["max_range_m"] = *config.NetsimMaxRangeM
	}
	return overrides
}

func (g *ManifestGenerator) resolveMixer(uav domain.UAV, config domain.GeneralConfig) domain.DeployedService {
	if uav.Mixer != nil {
		return *uav.Mixer
	}
	return config.DefaultMixer
}

func (g *ManifestGenerator) resolveController(uav domain.UAV, config domain.GeneralConfig) domain.DeployedService {
	if uav.Controller != nil {
		return *uav.Controller
	}
	return config.DefaultController
}


