package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	UdpAddr             string `json:"udp_addr"`
	HttpAddr            string `json:"http_addr"`
	LogsDir             string `json:"logs_dir"`
	MaxLines            int    `json:"max_lines"`
	BufferSize          int    `json:"buffer_size"`
	WorkersCount        int    `json:"workers_count"`
	TruncationIntervalS int    `json:"truncation_interval_s"`
}

func LoadConfig(path string) Config {
	file, err := os.Open(path)
	cfg := Config{
		UdpAddr:             "0.0.0.0:5000",
		HttpAddr:            "0.0.0.0:8080",
		LogsDir:             "./data/logs",
		MaxLines:            10000,
		BufferSize:          5000,
		WorkersCount:        5,
		TruncationIntervalS: 10,
	}

	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&cfg)
	}

	return cfg
}
