package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ResourceWriter writes JSON config files into an output directory using
// content-addressed naming to deduplicate identical configs across UAVs.
type ResourceWriter struct {
	outputDir string
}

// NewResourceWriter creates a ResourceWriter targeting the given directory.
// The directory must already exist before calling Write.
func NewResourceWriter(outputDir string) *ResourceWriter {
	return &ResourceWriter{outputDir: outputDir}
}

// Write serialises content to a JSON file named "<baseName>_<hash8><suffix>.json".
// If a file with the same content hash already exists it is reused.
// Returns the base filename (not the full path) so callers can reference it
// in Docker Compose volume mounts.
func (w *ResourceWriter) Write(baseName string, content map[string]interface{}, suffix string) (string, error) {
	encoded, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config %q: %w", baseName, err)
	}

	hash := sha256.Sum256(encoded)
	shortHash := hex.EncodeToString(hash[:])[:8]

	fileName := fmt.Sprintf("%s_%s%s.json", baseName, shortHash, suffix)
	filePath := filepath.Join(w.outputDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, encoded, 0644); err != nil {
			return "", fmt.Errorf("write config %q: %w", fileName, err)
		}
	}

	return fileName, nil
}
