package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ui/internal/domain"
	"ui/internal/ports"
)

func (g *ManifestGenerator) buildServiceResources(svc domain.DeployedService, writer *ResourceWriter, mutator func(map[string]interface{}), suffix string) (string, []domain.VolumeMount, ports.ResourceLimits) {
	var extraVolumes []domain.VolumeMount
	cfg := make(map[string]interface{})
	for k, v := range svc.Config {
		cfg[k] = v
	}

	if schema, err := g.getServiceSchema(svc.FolderName); err == nil {
		if props, ok := schema["properties"].(map[string]interface{}); ok {
			for key, val := range props {
				prop, ok := val.(map[string]interface{})
				if !ok {
					continue
				}
				if format, ok := prop["format"].(string); ok && (format == "kml" || format == "file") {
					if srcPath, ok := cfg[key].(string); ok && srcPath != "" {
						ext := filepath.Ext(srcPath)
						if data, err := os.ReadFile(srcPath); err == nil {
							hash := sha256.Sum256(data)
							hashStr := hex.EncodeToString(hash[:])[:8]
							hostName := fmt.Sprintf("%s_%s_%s%s", svc.ServiceId, key, hashStr, ext)
							if err := os.WriteFile(filepath.Join(writer.outputDir, hostName), data, os.ModePerm); err == nil {
								containerPath := fmt.Sprintf("/app/%s%s", key, ext)
								cfg[key] = containerPath
								extraVolumes = append(extraVolumes, domain.VolumeMount{
									HostPath:      hostName,
									ContainerPath: containerPath,
								})
							}
						}
					}
				}
			}
		}
	}

	if mutator != nil {
		mutator(cfg)
	}

	svcFile, _ := writer.Write(svc.ServiceId+"_config", cfg, suffix)

	var algoLimits ports.ResourceLimits
	if schema, err := g.getServiceSchema(svc.FolderName); err == nil {
		algoLimits = g.ParseResourceLimits(schema)
	}

	return svcFile, extraVolumes, algoLimits
}

func (g *ManifestGenerator) getServiceSchema(folderName string) (map[string]interface{}, error) {
	searchDirs := []string{g.algorithmsDir, g.mixersDir, g.controllersDir}
	for _, dir := range searchDirs {
		schemaPath := filepath.Join(dir, folderName, "schema.json")
		if rawData, err := os.ReadFile(schemaPath); err == nil {
			var schema map[string]interface{}
			if err := json.Unmarshal(rawData, &schema); err == nil {
				return schema, nil
			}
		}
	}
	return nil, fmt.Errorf("schema not found for %s", folderName)
}

func (g *ManifestGenerator) generateUAVParams(swarmID string, uav domain.UAV, config domain.GeneralConfig, resDir string) (string, error) {
	// Resolve the base copter.parm from the selected controller's folder.
	controller := g.resolveController(uav, config)
	baseParmPath := ""
	if controller.FolderName != "" {
		candidate := filepath.Join(g.controllersDir, controller.FolderName, "ardupilot", "copter.parm")
		if _, err := os.Stat(candidate); err == nil {
			baseParmPath = candidate
		}
	}
	if baseParmPath == "" {
		// Fallback: legacy hardcoded path
		baseParmPath = filepath.Join(g.projectRoot, "..", "uav_controller", "ardupilot4_5_3", "ardupilot", "copter.parm")
	}
	content, _ := os.ReadFile(baseParmPath)
	params := string(content)
	if !strings.HasSuffix(params, "\n") {
		params += "\n"
	}
	if !config.LoggingEnabled {
		params += "LOG_BITMASK 0\n"
	}
	uavBattery := config.BatteryCapacity
	if uavBattery <= 0 {
		uavBattery = 5000
	}
	if uav.BatteryCapacity != nil {
		uavBattery = *uav.BatteryCapacity
	}
	params += fmt.Sprintf("BATT_CAPACITY %d\n", uavBattery)
	params += fmt.Sprintf("FS_BATT_MAH %d\n", uavBattery*20/100)
	params += "FS_BATT_ENABLE 2\n"
	params += "BATT_MONITOR 4\n"
	if config.WindEnabled {
		params += fmt.Sprintf("SIM_WIND_DIR %.2f\n", config.WindDirection)
		params += fmt.Sprintf("SIM_WIND_SPD %.2f\n", config.WindSpeed)
	}
	uavSpeed := config.DefaultUAVSpeed
	if uavSpeed <= 0 {
		uavSpeed = 10.0
	}
	if uav.Speed != nil {
		uavSpeed = *uav.Speed
	}
	params += fmt.Sprintf("WPNAV_SPEED %d\n", int(uavSpeed*100))
	params += fmt.Sprintf("WPNAV_SPEED_UP %d\n", int(uavSpeed*100))
	params += fmt.Sprintf("WPNAV_SPEED_DN %d\n", int(uavSpeed*100))

	fileName := fmt.Sprintf("swarm_%s_uav_%s_params.param", swarmID, uav.ID)
	_ = os.WriteFile(filepath.Join(resDir, fileName), []byte(params), 0644)
	return fileName, nil
}

func (g *ManifestGenerator) writeTemplateConfig(baseName, templatePath string, overrides map[string]interface{}, writer *ResourceWriter, mutator func(map[string]interface{}), suffix string) (string, error) {
	cfg := make(map[string]interface{})
	if rawData, err := os.ReadFile(templatePath); err == nil {
		_ = json.Unmarshal(rawData, &cfg)
	}
	for key, value := range overrides {
		cfg[key] = value
	}
	if mutator != nil {
		mutator(cfg)
	}
	return writer.Write(baseName, cfg, suffix)
}

func (g *ManifestGenerator) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// ParseResourceLimits extracts ram_limit and cpu_limit from a generic config map.
func (g *ManifestGenerator) ParseResourceLimits(cfg map[string]interface{}) ports.ResourceLimits {
	var limits ports.ResourceLimits

	if val, ok := cfg["ram_limit"]; ok {
		if str, ok := val.(string); ok {
			limits.RAM = str
		}
	}

	if val, ok := cfg["cpu_limit"]; ok {
		switch v := val.(type) {
		case float64:
			limits.CPU = v
		case float32:
			limits.CPU = float64(v)
		case int:
			limits.CPU = float64(v)
		case int64:
			limits.CPU = float64(v)
		}
	}

	return limits
}

// LoadRawConfig reads a JSON file into a generic map, useful for parsing
// resource limits before templating or modifications.
func (g *ManifestGenerator) LoadRawConfig(path string) map[string]interface{} {
	cfg := make(map[string]interface{})
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	return cfg
}
