package logger

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HttpLoggerClient struct{}

func NewHttpLoggerClient() *HttpLoggerClient {
	return &HttpLoggerClient{}
}

func (c *HttpLoggerClient) DownloadZip(remoteHost string) ([]byte, error) {
	parts := strings.Split(remoteHost, ":")
	remoteIp := ""
	if(len(parts) > 0) {
		remoteIp = parts[0]
	}
	url := "http://localhost:8080/api/logs/download"
	if remoteIp != "" && remoteIp != "localhost" && remoteIp != "127.0.0.1" {
		url = fmt.Sprintf("http://%s:8080/api/logs/download", remoteHost)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to reach logger service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("logger service returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
