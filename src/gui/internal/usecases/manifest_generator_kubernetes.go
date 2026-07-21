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

func sanitizeName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return s
}

func makeVolumeMounts(mounts map[string]string) ([]ports.KubeVolumeMount, []ports.KubeVolume) {
	var vms []ports.KubeVolumeMount
	var vs []ports.KubeVolume
	for mountPath, file := range mounts {
		volName := "vol-" + sanitizeName(file)
		cmName := "cm-" + sanitizeName(file)
		vms = append(vms, ports.KubeVolumeMount{Name: volName, MountPath: mountPath, SubPath: file})
		vs = append(vs, ports.KubeVolume{Name: volName, ConfigMap: cmName})
	}
	return vms, vs
}

func deduplicateVolumes(vols []ports.KubeVolume) []ports.KubeVolume {
	seen := make(map[string]bool)
	var out []ports.KubeVolume
	for _, v := range vols {
		if !seen[v.Name] {
			seen[v.Name] = true
			out = append(out, v)
		}
	}
	return out
}

func (g *ManifestGenerator) buildKubernetesManifests(swarms []domain.Swarm, config domain.GeneralConfig, resDir, simDir string) (string, error) {
	writer := NewResourceWriter(resDir)
	builder := g.newKubernetesBuilder()

	netsimInstances := normalizeNetsimInstances(config.NetsimInstances)
	netsimAddrs := make([]string, netsimInstances)
	globalMapping := map[string]string{
		"netsim_gateway": "netsim-gateway",
		"logger":         "logger",
	}
	for i := 1; i <= netsimInstances; i++ {
		netsimAddrs[i-1] = fmt.Sprintf("netsim-%d:3000", i)
		globalMapping[fmt.Sprintf("netsim_%d", i)] = fmt.Sprintf("netsim-%d", i)
	}

	gwCfg := g.LoadRawConfig(g.netsimGatewayConfig)
	gwMutator := g.createK8sMutator(globalMapping, "netsim-gateway")
	gwFile, _ := g.writeTemplateConfig("netsim_gateway_config", g.netsimGatewayConfig, nil, writer, gwMutator, "_k8s")
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

	nsOverrides := g.buildNetsimOverrides(config)
	nsMutator := g.createK8sMutator(globalMapping, "netsim")
	nsFile, _ := g.writeTemplateConfig("netsim_config", g.netsimConfig, nsOverrides, writer, nsMutator, "_k8s")

	loggerMutator := g.createK8sMutator(globalMapping, "logger")
	loggerFile, _ := g.writeTemplateConfig("logger_config", g.loggerConfig, nil, writer, loggerMutator, "_k8s")

	var uavDeployments []uavDep

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

			var arduPilotHostPath string
			absPath, err := filepath.Abs(arduPilotInstance)
			if err == nil {
				arduPilotHostPath = absPath
			} else {
				arduPilotHostPath = arduPilotInstance
			}

			paramFile, _ := g.generateUAVParams(swarm.ID, uav, config, resDir)
			deps := g.buildKubernetesUAV(swarm.ID, uav, paramFile, arduPilotHostPath, writer, config, offsets[i], swarm)
			uavDeployments = append(uavDeployments, deps...)
		}
	}

	// We now generate a ConfigMap to support remote deployments.
	// We will mount the entire 'resources' directory using a ConfigMap volume.

	vmsGW, vsGW := makeVolumeMounts(map[string]string{"/config.json": gwFile})
	builder.AddDeployment("netsim-gateway", []ports.KubeContainer{
		{
			Name:  "gateway",
			Image: g.getImageName("netsim_gateway", config.DockerHubRepository),
			Ports: []ports.KubePort{
				{ContainerPort: telPort, Protocol: "UDP"},
				{ContainerPort: msgPort, Protocol: "UDP"},
				{ContainerPort: subPort, Protocol: "UDP"},
			},
			Env:          map[string]string{"ADDRS": strings.Join(netsimAddrs, ",")},
			VolumeMounts: vmsGW,
		},
	}, vsGW, "", true)

	for i := 1; i <= netsimInstances; i++ {
		vmsNS, vsNS := makeVolumeMounts(map[string]string{"/config.json": nsFile})
		builder.AddDeployment(fmt.Sprintf("netsim-%d", i), []ports.KubeContainer{
			{
				Name:         "netsim",
				Image:        g.getImageName("netsim", config.DockerHubRepository),
				Env:          map[string]string{"NODE_ID": fmt.Sprintf("netsim_%d", i)},
				VolumeMounts: vmsNS,
			},
		}, vsNS, "", false)
	}

	vmsLog, vsLog := makeVolumeMounts(map[string]string{"/app/config.json": loggerFile})
	builder.AddDeployment("logger", []ports.KubeContainer{
		{
			Name:  "logger",
			Image: g.getImageName("logger", config.DockerHubRepository),
			Ports: []ports.KubePort{
				{ContainerPort: 5000, Protocol: "UDP"},
				{ContainerPort: 8080, Protocol: "TCP"},
			},
			VolumeMounts: vmsLog,
		},
	}, vsLog, "", true)

	for _, dep := range uavDeployments {
		builder.AddDeployment(dep.name, dep.containers, dep.volumes, dep.nodeLabel, false)
	}

	cmBuilder := g.newKubernetesBuilder()
	entries, err := os.ReadDir(resDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".json" || ext == ".param" || ext == ".kml" {
					content, err := os.ReadFile(filepath.Join(resDir, entry.Name()))
					if err == nil {
						cmBuilder.AddConfigMap("cm-"+sanitizeName(entry.Name()), map[string]string{entry.Name(): string(content)})
					}
				}
			}
		}
	}
	cmPath := filepath.Join(simDir, "configmaps.yaml")
	_ = os.WriteFile(cmPath, []byte(cmBuilder.Build()), 0644)

	composePath := filepath.Join(simDir, "kubernetes.yaml")
	_ = os.WriteFile(composePath, []byte(builder.Build()), 0644)
	return composePath, nil
}

type uavDep struct {
	name       string
	containers []ports.KubeContainer
	volumes    []ports.KubeVolume
	nodeLabel  string
}

func (g *ManifestGenerator) buildKubernetesUAV(swarmID string, uav domain.UAV, paramFileName, arduPilotHostPath string, writer *ResourceWriter, config domain.GeneralConfig, offset formation.Offset, swarm domain.Swarm) []uavDep {
	var deployments []uavDep
	var mainContainers []ports.KubeContainer

	uavNodeLabel := ""
	if uav.NodeLabel != nil && *uav.NodeLabel != "" {
		uavNodeLabel = *uav.NodeLabel
	}

	mainPodName := fmt.Sprintf("swarm-%s-uav-%s", swarmID, uav.ID)
	uavMapping := map[string]string{
		"netsim_gateway":       "netsim-gateway",
		"logger":               "logger",
		"communication_module": mainPodName,
		"external_comms":       mainPodName,
	}
	netsimInstances := normalizeNetsimInstances(config.NetsimInstances)
	for i := 1; i <= netsimInstances; i++ {
		uavMapping[fmt.Sprintf("netsim_%d", i)] = fmt.Sprintf("netsim-%d", i)
	}

	mixer := g.resolveMixer(uav, config)
	if mixer.NodeLabel != "" && mixer.NodeLabel != uavNodeLabel {
		uavMapping["mixer"] = fmt.Sprintf("swarm-%s-uav-%s-mixer", swarmID, uav.ID)
	} else {
		uavMapping["mixer"] = mainPodName
	}

	controller := g.resolveController(uav, config)
	if controller.NodeLabel != "" && controller.NodeLabel != uavNodeLabel {
		uavMapping["uav_controller"] = fmt.Sprintf("swarm-%s-uav-%s-uav-controller", swarmID, uav.ID)
	} else {
		uavMapping["uav_controller"] = mainPodName
	}

	for _, svc := range uav.Services {
		if svc.NodeLabel != "" && svc.NodeLabel != uavNodeLabel {
			uavMapping[svc.ServiceId] = fmt.Sprintf("swarm-%s-uav-%s-%s", swarmID, uav.ID, strings.ToLower(svc.ServiceId))
		} else {
			uavMapping[svc.ServiceId] = mainPodName
		}
	}

	uavEnv := map[string]string{
		"UAV_ID":   uav.ID,
		"SWARM_ID": swarmID,
	}

	mainContainers = append(mainContainers, ports.KubeContainer{
		Name:  "communication-module",
		Image: g.getImageName("communication_module", config.DockerHubRepository),
		Env:   uavEnv,
	})

	var mainVolumes []ports.KubeVolume

	appFile, _, _ := g.buildServiceResources(mixer, writer, g.createK8sMutator(uavMapping, uavMapping["mixer"]), "_k8s")
	vmsMixer, vsMixer := makeVolumeMounts(map[string]string{"/app/config.json": appFile})
	mixerContainer := ports.KubeContainer{
		Name:         "mixer",
		Image:        g.getImageName(mixer.FolderName, config.DockerHubRepository),
		Env:          uavEnv,
		VolumeMounts: vmsMixer,
	}
	if mixer.NodeLabel != "" && mixer.NodeLabel != uavNodeLabel {
		deployments = append(deployments, uavDep{
			name:       fmt.Sprintf("swarm-%s-uav-%s-mixer", swarmID, uav.ID),
			containers: []ports.KubeContainer{mixerContainer},
			volumes:    vsMixer,
			nodeLabel:  mixer.NodeLabel,
		})
	} else {
		mainContainers = append(mainContainers, mixerContainer)
		mainVolumes = append(mainVolumes, vsMixer...)
	}

	ucMutator := g.createK8sMutator(uavMapping, uavMapping["uav_controller"])
	ucFile, _, _ := g.buildServiceResources(controller, writer, ucMutator, "_k8s")
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

	vmsCtrl, vsCtrl := makeVolumeMounts(map[string]string{
		"/app/config.json": ucFile,
		"/app/copter.parm": paramFileName,
	})

	serviceId := controller.ServiceId
	if serviceId == "" {
		serviceId = controller.FolderName
	}
	uavControllerImage := serviceId
	if arduPilotHostPath != "" {
		binName := filepath.Base(arduPilotHostPath)
		uavControllerImage = fmt.Sprintf("%s_%s", serviceId, strings.ToLower(strings.ReplaceAll(binName, ".", "_")))
	}

	controllerContainer := ports.KubeContainer{
		Name:         "uav-controller",
		Image:        g.getImageName(uavControllerImage, config.DockerHubRepository),
		Env:          ctrlEnv,
		VolumeMounts: vmsCtrl,
	}
	if controller.NodeLabel != "" && controller.NodeLabel != uavNodeLabel {
		deployments = append(deployments, uavDep{
			name:       fmt.Sprintf("swarm-%s-uav-%s-uav-controller", swarmID, uav.ID),
			containers: []ports.KubeContainer{controllerContainer},
			volumes:    vsCtrl,
			nodeLabel:  controller.NodeLabel,
		})
	} else {
		mainContainers = append(mainContainers, controllerContainer)
		mainVolumes = append(mainVolumes, vsCtrl...)
	}

	ecMutator := g.createK8sMutator(uavMapping, mainPodName)
	ecFile, _ := g.writeTemplateConfig("external_comms_config", g.externalCommsConfig, nil, writer, ecMutator, "_k8s")
	vmsEC, vsEC := makeVolumeMounts(map[string]string{"/app/config.json": ecFile})
	mainContainers = append(mainContainers, ports.KubeContainer{
		Name:         "external-comms",
		Image:        g.getImageName("external_comms", config.DockerHubRepository),
		Env:          uavEnv,
		VolumeMounts: vmsEC,
	})
	mainVolumes = append(mainVolumes, vsEC...)

	for _, svc := range uav.Services {
		svcFile, extraVolumes, _ := g.buildServiceResources(svc, writer, g.createK8sMutator(uavMapping, uavMapping[svc.ServiceId]), "_k8s")

		mounts := map[string]string{"/app/config.json": svcFile}
		for _, v := range extraVolumes {
			mounts[v.ContainerPath] = v.HostPath
		}
		vmsSvc, vsSvc := makeVolumeMounts(mounts)

		svcContainer := ports.KubeContainer{
			Name:         svc.ServiceId,
			Image:        g.getImageName(svc.ServiceId, config.DockerHubRepository),
			Env:          uavEnv,
			VolumeMounts: vmsSvc,
		}

		if svc.NodeLabel != "" && svc.NodeLabel != uavNodeLabel {
			deployments = append(deployments, uavDep{
				name:       fmt.Sprintf("swarm-%s-uav-%s-%s", swarmID, uav.ID, strings.ToLower(svc.ServiceId)),
				containers: []ports.KubeContainer{svcContainer},
				volumes:    vsSvc,
				nodeLabel:  svc.NodeLabel,
			})
		} else {
			mainContainers = append(mainContainers, svcContainer)
			mainVolumes = append(mainVolumes, vsSvc...)
		}
	}

	deployments = append(deployments, uavDep{
		name:       fmt.Sprintf("swarm-%s-uav-%s", swarmID, uav.ID),
		containers: mainContainers,
		volumes:    deduplicateVolumes(mainVolumes),
		nodeLabel:  uavNodeLabel,
	})

	return deployments
}

func (g *ManifestGenerator) createK8sMutator(mapping map[string]string, currentDeployment string) func(map[string]interface{}) {
	return func(cfg map[string]interface{}) {
		var traverse func(data interface{}) interface{}
		traverse = func(data interface{}) interface{} {
			switch v := data.(type) {
			case string:
				if targetDep, ok := mapping[v]; ok {
					if targetDep == currentDeployment {
						return "127.0.0.1"
					}
					return targetDep
				}
				return v
			case map[string]interface{}:
				for key, val := range v {
					v[key] = traverse(val)
				}
				return v
			case []interface{}:
				for i, val := range v {
					v[i] = traverse(val)
				}
				return v
			default:
				return v
			}
		}

		for k, v := range cfg {
			cfg[k] = traverse(v)
		}
	}
}
