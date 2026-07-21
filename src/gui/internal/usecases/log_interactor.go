package usecases

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ui/internal/domain"
	"ui/internal/ports"
)

type LogInteractor struct {
	repo ports.ConfigRepository
}

func NewLogInteractor(repo ports.ConfigRepository) *LogInteractor {
	return &LogInteractor{repo: repo}
}

func (i *LogInteractor) SearchLogs(zipPath string, filter domain.LogFilter) ([]domain.LogMessage, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	var allLogs []domain.LogMessage
	filterText := strings.ToLower(filter.SearchText)

	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".jsonl") {
			continue
		}
		
		rc, err := f.Open()
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			var msg domain.LogMessage
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}

			// Apply filters
			if filter.InstanceID != "" && !strings.Contains(strings.ToLower(msg.InstanceID), strings.ToLower(filter.InstanceID)) {
				continue
			}
			if filter.ServiceID != "" && !strings.Contains(strings.ToLower(msg.ServiceID), strings.ToLower(filter.ServiceID)) {
				continue
			}
			if filter.Level != "" && !strings.EqualFold(msg.Level, filter.Level) {
				continue
			}
			if filter.EventID != "" && !strings.Contains(strings.ToLower(msg.EventID), strings.ToLower(filter.EventID)) {
				continue
			}
			if filterText != "" && !strings.Contains(strings.ToLower(msg.Message), filterText) {
				continue
			}

			allLogs = append(allLogs, msg)
		}
		rc.Close()
	}

	// Sort logs chronologically (oldest first, so newest at the bottom as requested).
	// Use ReceivedAt as tiebreaker to preserve arrival order for logs with equal Timestamp.
	sort.SliceStable(allLogs, func(a, b int) bool {
		ta := allLogs[a].Timestamp
		tb := allLogs[b].Timestamp
		if ta.Equal(tb) {
			return allLogs[a].ReceivedAt.Before(allLogs[b].ReceivedAt)
		}
		return ta.Before(tb)
	})

	return allLogs, nil
}

func (i *LogInteractor) LoadLogEntries(ctx context.Context) ([]string, error) {
	baseDir := i.repo.GetSimulationsDir()
	var entries []string

	entriesMap := make(map[string]bool)

	// We expect logs to be in simulations/<sim_name>/logs/logs_*.zip
	entriesFiles, err := filepath.Glob(filepath.Join(baseDir, "*", "logs", "*.zip"))
	if err == nil {
		for _, file := range entriesFiles {
			entriesMap[file] = true
		}
	}

	for k := range entriesMap {
		entries = append(entries, k)
	}

	// Sort entries reverse chronologically so newest zip is at the top
	sort.Sort(sort.Reverse(sort.StringSlice(entries)))

	return entries, nil
}

func buildCombinedZip(loggerZipBytes []byte, uavLogDirs map[string]string, outPath string) error {
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create combined zip: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// 1. Copy original logger zip contents
	if len(loggerZipBytes) > 0 {
		loggerReader, err := zip.NewReader(bytes.NewReader(loggerZipBytes), int64(len(loggerZipBytes)))
		if err != nil {
			return fmt.Errorf("read logger zip: %w", err)
		}

		for _, file := range loggerReader.File {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			header := file.FileHeader
			writer, err := zipWriter.CreateHeader(&header)
			if err != nil {
				rc.Close()
				return err
			}
			if _, err := io.Copy(writer, rc); err != nil {
				rc.Close()
				return err
			}
			rc.Close()
		}
	}

	// 2. Add ArduPilot logs
	for uavID, dir := range uavLogDirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}

			// Clean path for ZIP
			relPath = filepath.ToSlash(relPath)
			zipPath := fmt.Sprintf("ardupilot/%s/%s", uavID, relPath)

			fileInfo, err := d.Info()
			if err != nil {
				return nil
			}

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return nil
			}
			header.Name = zipPath
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
		if err != nil {
			fmt.Printf("[interactor] error walking uav_logs for %s: %v\n", uavID, err)
		}
	}

	return nil
}
