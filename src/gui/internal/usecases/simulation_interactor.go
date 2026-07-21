package usecases

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"ui/internal/domain"
	"ui/internal/infrastructure/util"
	"ui/internal/ports"
)

type SimulationInteractor struct {
	runtime      ports.ContainerRuntime
	generator    *ManifestGenerator
	subscriber   ports.TelemetrySubscriber
	repo         ports.ConfigRepository
	ui           ports.UIBridge
	logger       ports.LoggerClient
	session      *activeSession
}

func NewSimulationInteractor(
	runtime ports.ContainerRuntime,
	generator *ManifestGenerator,
	subscriber ports.TelemetrySubscriber,
	repo ports.ConfigRepository,
	ui ports.UIBridge,
	logger ports.LoggerClient,
) *SimulationInteractor {
	return &SimulationInteractor{
		runtime:      runtime,
		generator:    generator,
		subscriber:   subscriber,
		repo:         repo,
		ui:           ui,
		logger:       logger,
		session:      newActiveSession(),
	}
}

func (i *SimulationInteractor) StartSimulation(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error {
	config.SanitizeSimulationName()
	simDir := filepath.Join(i.repo.GetSimulationsDir(), config.SimulationName)
	
	// Delete previous uav_logs before starting
	os.RemoveAll(filepath.Join(simDir, "uav_logs"))

	composePath, err := i.generator.Generate(swarms, config, isLocal, simDir)
	if err != nil {
		return fmt.Errorf("prepare simulation: %w", err)
	}

	i.session.clear()
	i.session.ctx = ctx
	i.session.composePath = composePath
	i.session.simulationName = config.SimulationName
	
	var uavIDs []string
	for _, swarm := range swarms {
		for _, uav := range swarm.UAVs {
			uavIDs = append(uavIDs, fmt.Sprintf("swarm_%s_uav_%s", swarm.ID, uav.ID))
		}
	}
	i.session.uavIDs = uavIDs
	i.session.loggingEnabled = config.LoggingEnabled

	if isLocal {
		if err := i.runtime.StartCompose(composePath); err != nil {
			return err
		}

		i.subscriber.SetExpectedFleet(i.session.uavIDs)

		// Algorithm tracking
		var allUAVs []domain.UAV
		for _, swarm := range swarms {
			allUAVs = append(allUAVs, swarm.UAVs...)
		}
		algoIDs := util.CollectAlgorithmIDs(allUAVs)
		for _, id := range algoIDs {
			i.session.algorithmIDs[id] = true
		}

		i.subscriber.SetOnFinish(func() {
			for _, algo := range algoIDs {
				if err := i.SendAlgorithmCommand(algo, "stop"); err != nil {
					fmt.Printf("[interactor] failed to stop algorithm %s: %v\n", algo, err)
				}
			}
		})
	} else {
		// KUBERNETES MODE
		i.session.kubernetesManifestPath = composePath
		i.session.dockerHubUser = config.DockerHubRepository
		i.session.kubeConfigPath = config.KubeConfigPath

		loggerIP, gatewayIP, err := i.runtime.StartKubernetes(composePath, config.DockerHubRepository, config.KubeConfigPath)
		if err != nil {
			return err
		}
		
		i.session.loggerIP = loggerIP
		i.session.gatewayIP = gatewayIP

		i.subscriber.SetRemoteAddr(gatewayIP)
		i.subscriber.SetExpectedFleet(i.session.uavIDs)
	}
	go i.subscriber.Start(ctx)

	return nil
}

func (i *SimulationInteractor) BuildImages(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig, isLocal bool) error {
	config.SanitizeSimulationName()
	simDir := filepath.Join(i.repo.GetSimulationsDir(), config.SimulationName)

	// Build all known images and all local algorithms, independent of simulation size/mode
	return i.runtime.BuildAllImages(simDir, config.DockerHubRepository, !isLocal, swarms, config)
}


func (i *SimulationInteractor) StopSimulation() {
	if i.session.kubernetesManifestPath != "" {
		_ = i.runtime.CollectKubernetesLogs(i.session.simulationName, "")
		_ = i.runtime.StopKubernetes(i.session.kubernetesManifestPath, i.session.kubeConfigPath)
	} else if i.session.composePath != "" {
		_ = i.runtime.StopCompose(i.session.composePath)
	}
	i.session.clear()
}

func (i *SimulationInteractor) SendAlgorithmCommand(serviceId string, command string) error {
	payload := map[string]interface{}{
		"topic":   "algo/" + serviceId,
		"payload": map[string]interface{}{
			"command": command,
		},
	}
	if err := i.subscriber.SendGlobalBroadcast(payload); err != nil {
		return err
	}

	i.ui.EmitEvent("netsim:message", map[string]string{
		"source":  "[Global] ",
		"label":   fmt.Sprintf("[Global] Command '%s' sent to service: %s", command, serviceId),
	})

	if command == "stop" && len(i.session.algorithmIDs) > 0 {
		if i.session.markStopped(serviceId) {
			i.subscriber.NotifyUserStoppedAll()
		}
	}
	return nil
}

func (i *SimulationInteractor) DownloadLogs() error {
	if !i.session.isRunning() && i.session.simulationName == "" {
		return fmt.Errorf("no active simulation")
	}
	simDir := filepath.Join(i.repo.GetSimulationsDir(), i.session.simulationName)
	logsDir := filepath.Join(simDir, "logs")
	os.MkdirAll(logsDir, 0755)

	loggerZipBytes, err := i.logger.DownloadZip(i.session.loggerHost())
	if err != nil {
		return fmt.Errorf("fetch logger zip: %w", err)
	}

	uavLogDirs := map[string]string{}
	if i.session.loggingEnabled {
		for _, uavID := range i.session.uavIDs {
			dir := filepath.Join(simDir, "uav_logs", uavID)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				uavLogDirs[uavID] = dir
			}
		}
	}

	timestamp := time.Now().Format("20060102_150405")
	outPath := filepath.Join(logsDir, fmt.Sprintf("logs_%s.zip", timestamp))
	return buildCombinedZip(loggerZipBytes, uavLogDirs, outPath)
}

func (i *SimulationInteractor) ExportSimulation(ctx context.Context, swarms []domain.Swarm, config domain.GeneralConfig) error {
	config.SanitizeSimulationName()
	simDir := filepath.Join(i.repo.GetSimulationsDir(), config.SimulationName)

	if _, err := i.generator.Generate(swarms, config, true, simDir); err != nil {
		return fmt.Errorf("prepare export (local): %w", err)
	}
	if _, err := i.generator.Generate(swarms, config, false, simDir); err != nil {
		return fmt.Errorf("prepare export (kubernetes): %w", err)
	}

	if err := i.runtime.GenerateBuildManifest(simDir, config.DockerHubRepository, false, swarms, config); err != nil {
		return fmt.Errorf("prepare export (build file): %w", err)
	}

	defaultFilename := fmt.Sprintf("ArduSim_Export_%s.zip", config.SimulationName)
	savePath, err := i.ui.SaveFileDialog(ctx, "Save Export as ZIP", defaultFilename, []ports.FileFilter{{DisplayName: "ZIP Archive", Pattern: "*.zip"}})
	if err != nil {
		return err
	}
	if savePath == "" {
		return nil // cancelled
	}

	if err := zipDirectory(simDir, savePath); err != nil {
		return fmt.Errorf("zip simulation directory: %w", err)
	}

	i.ui.EmitEvent("simulation:log", "[Export] Simulation exported successfully to "+savePath)
	return nil
}

func zipDirectory(sourceDir, destZip string) error {
	outFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return nil
		}

		relPath = filepath.ToSlash(relPath)
		fileInfo, err := d.Info()
		if err != nil {
			return nil
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return nil
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
