package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ui/internal/domain"
	"ui/internal/domain/formation"
	"ui/internal/infrastructure/util"
	"ui/internal/ports"
)

func (g *ManifestGenerator) buildLocalCompose(swarms []domain.Swarm, config domain.GeneralConfig, resDir, simDir string) (string, error) {
	writer := NewResourceWriter(resDir)
	builder := g.newComposeBuilder()
	pool := newSubnetPool()

	netsimInstances := normalizeNetsimInstances(config.NetsimInstances)
	netsimAddrs := make([]string, netsimInstances)
	for i := 1; i <= netsimInstances; i++ {
		netsimAddrs[i-1] = fmt.Sprintf("netsim_%d:3000", i)
	}

	gwCfg := g.LoadRawConfig(g.netsimGatewayConfig)
	gwLimits := g.ParseResourceLimits(gwCfg)
	gwFile, _ := g.writeTemplateConfig("netsim_gateway_config", g.netsimGatewayConfig, nil, writer, nil, "")

	telPort, msgPort, subPort := 3000, 3001, 3002
	if p, ok := gwCfg["telemetry_port"].(float64); ok {
		telPort = int(p)
	}
	if p, ok := gwCfg["messages_port"].(float64); ok {
		msgPort = int(p)
	}
	if p, ok := gwCfg["subscribers_port"].(float64); ok {
		subPort = int(p)
	}

	gwEnv := make(map[string]string)
	if config.VerboseLogging {
		gwEnv["DEBUG"] = "true"
	}
	if len(netsimAddrs) > 0 {
		gwEnv["ADDRS"] = strings.Join(netsimAddrs, ",")
	}
	
	builder.AddNetwork("air", "10.9.0.0/24")
	builder.AddService(ports.ComposeService{
		Name:          "netsim_gateway",
		Image:         "netsim_gateway",
		ContainerName: "netsim_gateway",
		ExtraHosts:    []string{"host.docker.internal:host-gateway"},
		Ports: []string{
			fmt.Sprintf("%d:%d/udp", telPort, telPort),
			fmt.Sprintf("%d:%d/udp", msgPort, msgPort),
			fmt.Sprintf("%d:%d/udp", subPort, subPort),
		},
		Environment: gwEnv,
		Volumes:     []string{fmt.Sprintf("./resources/%s:/app/config.json", gwFile)},
		Networks:    map[string][]string{"air": nil},
		Limits:      gwLimits,
	})

	nsCfg := g.LoadRawConfig(g.netsimConfig)
	nsLimits := g.ParseResourceLimits(nsCfg)
	nsOverrides := g.buildNetsimOverrides(config)
	nsFile, _ := g.writeTemplateConfig("netsim_config", g.netsimConfig, nsOverrides, writer, nil, "")

	for i := 1; i <= netsimInstances; i++ {
		nsEnv := map[string]string{"NODE_ID": fmt.Sprintf("netsim_%d", i)}
		if config.VerboseLogging {
			nsEnv["DEBUG"] = "true"
		}
		builder.AddService(ports.ComposeService{
			Name:          fmt.Sprintf("netsim_%d", i),
			Image:         "netsim",
			ContainerName: fmt.Sprintf("netsim_%d", i),
			DependsOn:     []string{"netsim_gateway"},
			Environment:   nsEnv,
			Volumes:       []string{fmt.Sprintf("./resources/%s:/app/config.json", nsFile)},
			Networks:      map[string][]string{"air": nil},
			Limits:        nsLimits,
		})
	}

	loggerCfg := g.LoadRawConfig(g.loggerConfig)
	loggerLimits := g.ParseResourceLimits(loggerCfg)
	loggerFile, _ := g.writeTemplateConfig("logger_config", g.loggerConfig, nil, writer, nil, "")

	builder.AddService(ports.ComposeService{
		Name:          "logger",
		Image:         "logger",
		ContainerName: "logger",
		Ports:         []string{"5000:5000/udp", "8080:8080/tcp"},
		Volumes:       []string{fmt.Sprintf("./resources/%s:/app/config.json", loggerFile)},
		Networks:      map[string][]string{"air": nil},
		Limits:        loggerLimits,
	})

	if config.LoggingEnabled {
		for _, swarm := range swarms {
			for _, uav := range swarm.UAVs {
				logDir := filepath.Join(simDir, "uav_logs", fmt.Sprintf("swarm_%s_uav_%s", swarm.ID, uav.ID))
				os.MkdirAll(logDir, 0755)
			}
		}
	}

	for _, swarm := range swarms {
		f := formation.GetFormation(swarm.GroundFormation)
		offsets := f.CalculateOffsets(len(swarm.UAVs), swarm.FormationSpacing)

		for i, uav := range swarm.UAVs {
			arduPilotInstance := config.DefaultArduPilotInstance
			if uav.ArduPilotInstance != nil && *uav.ArduPilotInstance != "" {
				arduPilotInstance = *uav.ArduPilotInstance
			}
			if arduPilotInstance == "" {
				arduPilotInstance = filepath.Join(g.projectRoot, "..", "uav_controller", "ardupilot4_5_3", "ardupilot", "arducopter4_5_3")
			}
			
			arduPilotInstanceFile := filepath.Base(arduPilotInstance)
			destPath := filepath.Join(resDir, arduPilotInstanceFile)
			_ = g.copyFile(arduPilotInstance, destPath)

			paramFile, _ := g.generateUAVParams(swarm.ID, uav, config, resDir)
			
			uavServices := g.buildComposeUAV(swarm.ID, uav, paramFile, arduPilotInstanceFile, writer, config, offsets[i], swarm, pool, builder)
			for _, svc := range uavServices {
				builder.AddService(svc)
			}
		}
	}

	composePath := filepath.Join(simDir, "docker-compose.yaml")
	_ = os.WriteFile(composePath, []byte(builder.Build()), 0644)
	return composePath, nil
}

func (g *ManifestGenerator) buildComposeUAV(swarmID string, uav domain.UAV, paramFileName, arduPilotInstanceFile string, writer *ResourceWriter, config domain.GeneralConfig, offset formation.Offset, swarm domain.Swarm, pool *subnetPool, builder ports.ComposeBuilder) []ports.ComposeService {
	nContainers := 4 + len(uav.Services)
	subnet := pool.Next(nContainers)
	uavNet := fmt.Sprintf("swarm_net_%s_uav_%s", swarmID, uav.ID)
	builder.AddNetwork(uavNet, subnet)

	var services []ports.ComposeService
	uavEnv := map[string]string{
		"UAV_ID":   uav.ID,
		"SWARM_ID": swarmID,
	}
	if config.VerboseLogging {
		uavEnv["DEBUG"] = "true"
	}

	commName := fmt.Sprintf("swarm_%s_uav_%s_communication_module", swarmID, uav.ID)
	services = append(services, ports.ComposeService{
		Name:          commName,
		Image:         "communication_module",
		ContainerName: commName,
		Environment:   uavEnv,
		Networks:      map[string][]string{uavNet: {"communication_module"}},
	})

	mixer := g.resolveMixer(uav, config)
	appFile, _, appLimits := g.buildServiceResources(mixer, writer, nil, "")
	mixerName := fmt.Sprintf("swarm_%s_uav_%s_mixer", swarmID, uav.ID)
	services = append(services, ports.ComposeService{
		Name:          mixerName,
		Image:         mixer.FolderName,
		ContainerName: mixerName,
		DependsOn:     []string{commName, fmt.Sprintf("swarm_%s_uav_%s_controller", swarmID, uav.ID)},
		Environment:   uavEnv,
		Volumes:       []string{fmt.Sprintf("./resources/%s:/app/config.json", appFile)},
		Networks:      map[string][]string{uavNet: {"mixer"}},
		Limits:        appLimits,
	})

	controller := g.resolveController(uav, config)
	ucFile, _, ucLimits := g.buildServiceResources(controller, writer, nil, "")
	var homeLat, homeLon float64
	if uav.HomeOverride != nil {
		homeLat, homeLon = uav.HomeOverride.Lat, uav.HomeOverride.Lon
	} else {
		homeLat, homeLon = util.AddOffset(swarm.FormationCenterLat, swarm.FormationCenterLon, offset.X, offset.Y)
	}
	
	ctrlEnv := map[string]string{
		"UAV_HOME_LOCATION": fmt.Sprintf("%f,%f,0,0", homeLat, homeLon),
		"UAV_ID":            uav.ID,
		"SWARM_ID":          swarmID,
	}
	if config.VerboseLogging {
		ctrlEnv["DEBUG"] = "true"
	}
	ctrlMounts := []string{
		fmt.Sprintf("./resources/%s:/app/config.json", ucFile),
		fmt.Sprintf("./resources/%s:/app/copter.parm", paramFileName),
	}
	if config.LoggingEnabled {
		ctrlMounts = append(ctrlMounts, fmt.Sprintf("./uav_logs/swarm_%s_uav_%s/:/app/logs/", swarmID, uav.ID))
	}
	
	ctrlName := fmt.Sprintf("swarm_%s_uav_%s_controller", swarmID, uav.ID)
	
	// Format the image name based on the binary
	serviceId := controller.ServiceId
	if serviceId == "" {
		serviceId = controller.FolderName
	}
	uavControllerImage := serviceId
	if arduPilotInstanceFile != "" {
		uavControllerImage = fmt.Sprintf("%s_%s", serviceId, strings.ToLower(strings.ReplaceAll(arduPilotInstanceFile, ".", "_")))
	}
	
	services = append(services, ports.ComposeService{
		Name:          ctrlName,
		Image:         uavControllerImage,
		ContainerName: ctrlName,
		DependsOn:     []string{commName},
		Environment:   ctrlEnv,
		Volumes:       ctrlMounts,
		Networks:      map[string][]string{uavNet: {"uav_controller"}},
		Limits:        ucLimits,
	})

	ecCfg := g.LoadRawConfig(g.externalCommsConfig)
	ecLimits := g.ParseResourceLimits(ecCfg)
	ecFile, _ := g.writeTemplateConfig("external_comms_config", g.externalCommsConfig, nil, writer, nil, "")
	
	ecName := fmt.Sprintf("swarm_%s_uav_%s_external_comms", swarmID, uav.ID)
	services = append(services, ports.ComposeService{
		Name:          ecName,
		Image:         "external_comms",
		ContainerName: ecName,
		DependsOn:     []string{commName, "netsim_gateway"},
		Environment:   uavEnv,
		Volumes:       []string{fmt.Sprintf("./resources/%s:/app/config.json", ecFile)},
		Networks:      map[string][]string{uavNet: {"external_comms"}, "air": nil},
		Limits:        ecLimits,
	})

	for _, svc := range uav.Services {
		svcFile, extraVolumes, algoLimits := g.buildServiceResources(svc, writer, nil, "")
		
		svcMounts := []string{fmt.Sprintf("./resources/%s:/app/config.json", svcFile)}
		for _, v := range extraVolumes {
			svcMounts = append(svcMounts, fmt.Sprintf("./resources/%s:%s", v.HostPath, v.ContainerPath))
		}
		
		svcName := fmt.Sprintf("swarm_%s_uav_%s_%s", swarmID, uav.ID, svc.ServiceId)
		services = append(services, ports.ComposeService{
			Name:          svcName,
			Image:         svc.ServiceId,
			ContainerName: svcName,
			DependsOn:     []string{commName},
			Environment:   uavEnv,
			Volumes:       svcMounts,
			Networks:      map[string][]string{uavNet: {svc.ServiceId}},
			Limits:        algoLimits,
		})
	}

	return services
}
