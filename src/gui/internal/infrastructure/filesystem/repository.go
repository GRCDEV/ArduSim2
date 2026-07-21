package filesystem

import (
	"encoding/json"
	"os"
	"path/filepath"

	"ui/internal/domain"
)

type FileRepository struct {
	algorithmsDir  string
	mixersDir      string
	controllersDir string
	simulationsDir string
}

func NewFileRepository(projectRoot string) *FileRepository {
	base := filepath.Clean(projectRoot)
	return &FileRepository{
		algorithmsDir:  filepath.Join(base, "..", "algorithms"),
		mixersDir:      filepath.Join(base, "..", "mixers"),
		controllersDir: filepath.Join(base, "..", "controllers"),
		simulationsDir: filepath.Join(base, "..", "..", "simulations"),
	}
}

func (r *FileRepository) GetAvailableServices() []domain.ServiceType {
	return r.scanServices(r.algorithmsDir)
}

func (r *FileRepository) GetAvailableMixers() []domain.ServiceType {
	return r.scanServices(r.mixersDir)
}

func (r *FileRepository) GetAvailableControllers() []domain.ServiceType {
	return r.scanServices(r.controllersDir)
}

func (r *FileRepository) scanServices(dir string) []domain.ServiceType {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	services := make([]domain.ServiceType, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		schemaPath := filepath.Join(dir, entry.Name(), "schema.json")
		rawData, err := os.ReadFile(schemaPath)
		if err != nil {
			continue
		}

		var schema struct {
			ServiceID string `json:"service_id"`
			Title     string `json:"title"`
		}
		if err := json.Unmarshal(rawData, &schema); err != nil {
			continue
		}

		svcID := schema.ServiceID
		if svcID == "" {
			svcID = entry.Name()
		}

		services = append(services, domain.ServiceType{
			ID:         svcID,
			FolderName: entry.Name(),
			Title:      schema.Title,
			SchemaRaw:  string(rawData),
		})
	}
	return services
}

func (r *FileRepository) SaveSimulation(simDir string, state domain.SimulationState) error {
	if err := os.MkdirAll(simDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	statePath := filepath.Join(simDir, "simulation.json")
	return os.WriteFile(statePath, data, 0644)
}

func (r *FileRepository) GetFiles(srcDir string, ext string) ([]string, error) {
	paths := []string{}
	children, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		path := filepath.Join(srcDir, child.Name())
		if !child.IsDir() {
			paths = append(paths, path)
			continue
		}
		
		childPaths, err := r.GetFiles(path, ext)
		if err != nil {
			return nil, err
		}
		paths = append(paths, childPaths...)
	}

	return paths, nil
}

func (r *FileRepository) ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (r *FileRepository) LoadSimulation(stateFile string) (*domain.SimulationState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state domain.SimulationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (r *FileRepository) GetSimulationsDir() string {
	return r.simulationsDir
}
