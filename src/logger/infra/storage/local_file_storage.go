package storage

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/GRCDEV/ArduSim2/logger/domain"
)

type LocalFileStorage struct {
	baseDir      string
	maxLines     int // Limit before truncating 50%
	mu           sync.Mutex
	sanitizeRgx  *regexp.Regexp
	fileLocks    map[string]*sync.Mutex
	fileLocksMu  sync.Mutex
}

func NewLocalFileStorage(baseDir string, maxLines int) (*LocalFileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Pre-compile regex for sanitization (only alphanumeric, dash, underscore allowed)
	reg, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		return nil, err
	}

	return &LocalFileStorage{
		baseDir:     baseDir,
		maxLines:    maxLines,
		sanitizeRgx: reg,
		fileLocks:   make(map[string]*sync.Mutex),
	}, nil
}

func (s *LocalFileStorage) sanitize(input string) string {
	return s.sanitizeRgx.ReplaceAllString(input, "")
}

func (s *LocalFileStorage) buildFileName(instanceID, serviceID string) string {
	inst := s.sanitize(instanceID)
	serv := s.sanitize(serviceID)

	var nameParts []string
	if serv != "" {
		nameParts = append(nameParts, serv)
	}
	if inst != "" {
		nameParts = append(nameParts, inst)
	}

	if len(nameParts) == 0 {
		return "unknown.jsonl"
	}

	return strings.Join(nameParts, "_") + ".jsonl"
}

func (s *LocalFileStorage) getFileLock(filename string) *sync.Mutex {
	s.fileLocksMu.Lock()
	defer s.fileLocksMu.Unlock()
	if lock, exists := s.fileLocks[filename]; exists {
		return lock
	}
	lock := &sync.Mutex{}
	s.fileLocks[filename] = lock
	return lock
}

func (s *LocalFileStorage) SaveLog(ctx context.Context, log domain.LogMessage) error {
	filename := s.buildFileName(log.InstanceID, log.ServiceID)
	path := filepath.Join(s.baseDir, filename)

	lock := s.getFileLock(filename)
	lock.Lock()
	defer lock.Unlock()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = f.Write(append(data, '\n'))
	return err
}

func (s *LocalFileStorage) WriteZippedLogs(ctx context.Context, w io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock() // Use global lock for zip operation to avoid partial reads

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".jsonl") {
			continue
		}

		filePath := filepath.Join(s.baseDir, file.Name())

		// Read file
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Skipping file %s: %v\n", file.Name(), err)
			continue
		}

		// Create zip entry
		zw, err := zipWriter.Create(file.Name())
		if err != nil {
			f.Close()
			return err
		}

		if _, err := io.Copy(zw, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

func (s *LocalFileStorage) RotateLogs(ctx context.Context) error {
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".jsonl") {
			continue
		}

		s.truncateFileIfNeeded(file.Name())
	}
	return nil
}

func (s *LocalFileStorage) truncateFileIfNeeded(filename string) {
	lock := s.getFileLock(filename)
	lock.Lock()
	defer lock.Unlock()

	path := filepath.Join(s.baseDir, filename)
	
	// Read lines to memory (since files are small, e.g. <10MB)
	f, err := os.Open(path)
	if err != nil {
		return
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	f.Close()

	if len(lines) <= s.maxLines {
		return
	}

	// Truncate 50%
	keepIndex := len(lines) / 2
	linesToKeep := lines[keepIndex:]

	// Overwrite file
	tmpPath := path + ".tmp"
	tmpF, err := os.Create(tmpPath)
	if err != nil {
		return
	}

	writer := bufio.NewWriter(tmpF)
	for _, line := range linesToKeep {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
	tmpF.Close()

	os.Rename(tmpPath, path)
}
