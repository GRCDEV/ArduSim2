package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"ui/internal/domain"
	"ui/internal/infrastructure/docker"
	"ui/internal/infrastructure/filesystem"
	"ui/internal/infrastructure/logger"
	"ui/internal/infrastructure/netsim"
	"ui/internal/infrastructure/wails"
	"ui/internal/ports"
	"ui/internal/usecases"
)

// App is the Wails binding layer. It delegates operations to use cases.
type App struct {
	ctx        context.Context
	simulation ports.SimulationUseCase
	discovery  ports.DiscoveryUseCase
	logs       ports.LogUseCase
	config     ports.ConfigUseCase
	ui         ports.UIBridge
}

// NewApp wires the application using Clean Architecture principles.
func NewApp() *App {
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("cannot determine working directory: %v", err))
	}

	// 1. Infrastructure (Adapters)
	repo := filesystem.NewFileRepository(workDir)
	bridge := wails.NewWailsBridge()
	runtime := docker.NewDockerRuntime(workDir, bridge)
	generator := usecases.NewManifestGenerator(workDir, docker.NewComposeBuilder, docker.NewKubernetesBuilder)
	// Read netsim_gateway config to get ports
	gwConfigPath := filepath.Join(workDir, "..", "netsim_gateway", "config.json")
	var gwCfg map[string]interface{}
	if data, err := os.ReadFile(gwConfigPath); err == nil {
		json.Unmarshal(data, &gwCfg)
	}

	subPort := 3002
	if p, ok := gwCfg["subscribers_port"].(float64); ok {
		subPort = int(p)
	}
	msgPort := 3001
	if p, ok := gwCfg["messages_port"].(float64); ok {
		msgPort = int(p)
	}

	subscriber := netsim.NewNetsimSubscriber(bridge, subPort, msgPort)
	loggerClient := logger.NewHttpLoggerClient()

	// 2. Use Cases (Interactors)
	simUC := usecases.NewSimulationInteractor(runtime, generator, subscriber, repo, bridge, loggerClient)
	discUC := usecases.NewDiscoveryInteractor(repo)
	logUC := usecases.NewLogInteractor(repo)
	configUC := usecases.NewConfigInteractor(repo, bridge)

	return &App{
		simulation: simUC,
		discovery:  discUC,
		logs:       logUC,
		config:     configUC,
		ui:         bridge,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.ui.SetContext(ctx)
}

func (a *App) shutdown() {
	a.simulation.StopSimulation()
}

// --- Wails Bindings ---

func (a *App) GetAvailableServices() []domain.ServiceType {
	return a.discovery.GetAvailableServices()
}

func (a *App) GetAvailableMixers() []domain.ServiceType {
	return a.discovery.GetAvailableMixers()
}

func (a *App) GetAvailableControllers() []domain.ServiceType {
	return a.discovery.GetAvailableControllers()
}

// Simulation
func (a *App) StartSimulation(swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error {
	a.config.PopulateDefaults(&config)
	return a.simulation.StartSimulation(a.ctx, swarms, config, isLocal)
}

func (a *App) BuildImages(swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error {
	a.config.PopulateDefaults(&config)
	return a.simulation.BuildImages(a.ctx, swarms, config, isLocal)
}

func (a *App) StopSimulation() {
	a.simulation.StopSimulation()
}

func (a *App) ExportSimulation(swarms []domain.Swarm, config domain.GeneralConfig) error {
	a.config.PopulateDefaults(&config)
	return a.simulation.ExportSimulation(a.ctx, swarms, config)
}

func (a *App) DownloadLogs() error {
	return a.simulation.DownloadLogs()
}

func (a *App) SendAlgorithmCommand(serviceId string, command string) error {
	return a.simulation.SendAlgorithmCommand(serviceId, command)
}

// Config
func (a *App) LoadSimulationConfig() (*domain.SimulationState, error) {
	return a.config.LoadSimulationConfig(a.ctx)
}

func (a *App) SaveSimulationConfig(swarms []domain.Swarm, config domain.GeneralConfig, mode string) (*domain.SimulationState, error) {
	return a.config.SaveSimulationConfig(swarms, config, mode)
}

func (a *App) DiscardCurrentRun(config domain.GeneralConfig) error {
	return a.config.DiscardCurrentRun(config)
}

func (a *App) SelectFile() (string, error) {
	return a.config.SelectFile(a.ctx)
}

func (a *App) SelectArduPilotInstance() (string, error) {
	return a.config.SelectArduPilotInstance(a.ctx)
}

func (a *App) SelectKubeConfig() (string, error) {
	return a.config.SelectKubeConfig(a.ctx)
}


func (a *App) GetKmlFirstCoordinate(path string) (*domain.Coordinate, error) {
	return a.config.GetKmlFirstCoordinate(path)
}

func (a *App) LoadFile(path string) (string, error) {
	return a.config.LoadFile(path)
}

// Logs
func (a *App) LoadLogEntries() ([]string, error) {
	return a.logs.LoadLogEntries(a.ctx)
}

func (a *App) SearchLogs(zipPath string, filter domain.LogFilter) ([]domain.LogMessage, error) {
	return a.logs.SearchLogs(zipPath, filter)
}
