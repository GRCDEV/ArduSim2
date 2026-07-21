package infrastructure

import (
	"encoding/json"
	"os"

	"external_comms/domain"
)

type FileLoader struct{}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (f *FileLoader) LoadAppConfig(filePath string) (*domain.AppConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &domain.AppConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
