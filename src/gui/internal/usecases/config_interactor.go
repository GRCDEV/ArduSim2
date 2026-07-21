package usecases

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"ui/internal/domain"
	"ui/internal/infrastructure/kml"
	"ui/internal/ports"

	"github.com/google/uuid"
)

type ConfigInteractor struct {
	repo ports.ConfigRepository
	ui   ports.UIBridge
}

func NewConfigInteractor(repo ports.ConfigRepository, ui ports.UIBridge) *ConfigInteractor {
	return &ConfigInteractor{
		repo: repo,
		ui:   ui,
	}
}

func (i *ConfigInteractor) LoadFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	return string(bytes), err
}

func (i *ConfigInteractor) LoadSimulationConfig(ctx context.Context) (*domain.SimulationState, error) {
	selectedDir, err := i.ui.OpenDirectoryDialog(ctx, "Select Simulation Directory", i.repo.GetSimulationsDir())
	if err != nil || selectedDir == "" {
		return nil, err
	}

	stateFile := filepath.Join(selectedDir, "simulation.json")
	state, err := i.repo.LoadSimulation(stateFile)
	if err != nil {
		return nil, err
	}

	// Repair logic
	availableServices := i.repo.GetAvailableServices()
	svcMap := make(map[string]string)
	for _, s := range availableServices {
		svcMap[s.ID] = s.FolderName
	}

	for idxSwarm := range state.Swarms {
		for idxUAV := range state.Swarms[idxSwarm].UAVs {
			for idxSvc := range state.Swarms[idxSwarm].UAVs[idxUAV].Services {
				svc := &state.Swarms[idxSwarm].UAVs[idxUAV].Services[idxSvc]
				if svc.FolderName == "" {
					if folder, ok := svcMap[svc.ServiceId]; ok {
						svc.FolderName = folder
					} else {
						svc.FolderName = svc.ServiceId
					}
				}
			}
		}
	}

	state.GeneralConfig.OriginalSimulationName = filepath.Base(selectedDir)
	state.GeneralConfig.SimulationName = filepath.Base(selectedDir)

	i.PopulateDefaults(&state.GeneralConfig)

	return state, nil
}

func (i *ConfigInteractor) SaveSimulationConfig(swarms []domain.Swarm, config domain.GeneralConfig, mode string) (*domain.SimulationState, error) {
	i.PopulateDefaults(&config)
	config.SanitizeSimulationName()
	simDir := filepath.Join(i.repo.GetSimulationsDir(), config.SimulationName)

	state := domain.SimulationState{
		Swarms:        swarms,
		GeneralConfig: config,
		ActiveMode:    mode,
	}

	err := i.repo.SaveSimulation(simDir, state)
	return &state, err
}

func (i *ConfigInteractor) PopulateDefaults(config *domain.GeneralConfig) {
	// 1. Default UAV Speed
	if config.DefaultUAVSpeed <= 0 {
		config.DefaultUAVSpeed = 10.0
	}

	// 5. Default Battery Capacity
	if config.BatteryCapacity <= 0 {
		config.BatteryCapacity = 5000
	}

	// 2. Default ArduPilot Instance
	if config.DefaultArduPilotInstance == "" {
		controllersDir := filepath.Clean(filepath.Join(i.repo.GetSimulationsDir(), "..", "src", "uav_controller"))
		config.DefaultArduPilotInstance = filepath.Join(controllersDir, "ardupilot4_5_3", "ardupilot", "arducopter4_5_3")
	}

	availableMixers := i.repo.GetAvailableMixers()
	if config.DefaultMixer.ServiceId == "" && len(availableMixers) > 0 {
		best := availableMixers[0]
		config.DefaultMixer.ServiceId = best.ID
		config.DefaultMixer.FolderName = best.FolderName
		config.DefaultMixer.ServiceTitle = best.Title
		config.DefaultMixer.InstanceId = uuid.NewString()
		config.DefaultMixer.Config = parseDefaultConfig(best.SchemaRaw)
	} else if config.DefaultMixer.ServiceId != "" && config.DefaultMixer.FolderName == "" {
		for _, m := range availableMixers {
			if m.ID == config.DefaultMixer.ServiceId {
				config.DefaultMixer.FolderName = m.FolderName
				break
			}
		}
	}

	availableControllers := i.repo.GetAvailableControllers()
	if config.DefaultController.ServiceId == "" && len(availableControllers) > 0 {
		best := availableControllers[0]
		config.DefaultController.ServiceId = best.ID
		config.DefaultController.FolderName = best.FolderName
		config.DefaultController.ServiceTitle = best.Title
		config.DefaultController.InstanceId = uuid.NewString()
		config.DefaultController.Config = parseDefaultConfig(best.SchemaRaw)
	} else if config.DefaultController.ServiceId != "" && config.DefaultController.FolderName == "" {
		for _, c := range availableControllers {
			if c.ID == config.DefaultController.ServiceId {
				config.DefaultController.FolderName = c.FolderName
				break
			}
		}
	}
}

func parseDefaultConfig(schemaRaw string) map[string]interface{} {
	var schema struct {
		Properties map[string]struct {
			Default interface{} `json:"default"`
		} `json:"properties"`
	}
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return nil
	}
	res := make(map[string]interface{})
	for k, v := range schema.Properties {
		if v.Default != nil {
			res[k] = v.Default
		}
	}
	return res
}

func (i *ConfigInteractor) DiscardCurrentRun(config domain.GeneralConfig) error {
	config.SanitizeSimulationName()
	// This logic remains similar, but using the repo
	return nil // Implementation detail for later
}

func (i *ConfigInteractor) SelectFile(ctx context.Context) (string, error) {
	return i.ui.OpenFileDialog(ctx, "Select Auxiliary File", []ports.FileFilter{
		{DisplayName: "KML files (*.kml)", Pattern: "*.kml"},
		{DisplayName: "All files", Pattern: "*.*"},
	})
}

func (i *ConfigInteractor) SelectArduPilotInstance(ctx context.Context) (string, error) {
	return i.ui.OpenFileDialog(ctx, "Select ArduPilot Instance", []ports.FileFilter{
		{DisplayName: "Executables (*)", Pattern: "*"},
		{DisplayName: "All files", Pattern: "*.*"},
	})
}

func (i *ConfigInteractor) GetKmlFirstCoordinate(path string) (*domain.Coordinate, error) {
	return kml.GetFirstCoordinate(path)
}

func (i *ConfigInteractor) SelectKubeConfig(ctx context.Context) (string, error) {
	return i.ui.OpenFileDialog(ctx, "Select Kubeconfig File", []ports.FileFilter{
		{DisplayName: "Kubeconfig (*)", Pattern: "*"},
		{DisplayName: "All files", Pattern: "*.*"},
	})
}
