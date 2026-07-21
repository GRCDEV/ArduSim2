package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GRCDEV/ArduSim2/logger/infra/config"
	"github.com/GRCDEV/ArduSim2/logger/infra/http"
	"github.com/GRCDEV/ArduSim2/logger/infra/storage"
	"github.com/GRCDEV/ArduSim2/logger/infra/udp"
	"github.com/GRCDEV/ArduSim2/logger/usecases"
)

func main() {
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	cfg := config.LoadConfig(configFile)

	udpAddr := cfg.UdpAddr
	httpAddr := cfg.HttpAddr
	logsDir := cfg.LogsDir
	maxLines := cfg.MaxLines
	bufferSize := cfg.BufferSize
	workersCount := cfg.WorkersCount
	truncationInterval := time.Duration(cfg.TruncationIntervalS) * time.Second

	fmt.Println("[Logger] Starting ArduSim2 Logger Microservice")

	// 1. Infra: Storage
	store, err := storage.NewLocalFileStorage(logsDir, maxLines)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}

	// 2. UseCases
	processUC := usecases.NewLogProcessorUseCase(store, bufferSize, workersCount)
	downloadUC := usecases.NewDownloadLogsUseCase(store)
	truncateUC := usecases.NewTruncateLogsUseCase(store, truncationInterval)

	// 3. Infra: Servers
	udpServer := udp.NewServer(udpAddr, processUC)
	httpServer := http.NewServer(httpAddr, downloadUC)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Background Process
	truncateUC.Start()

	// Start UDP Server
	go func() {
		if err := udpServer.Start(ctx); err != nil {
			fmt.Printf("UDP Server error: %v\n", err)
		}
	}()

	// Start HTTP Server
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			fmt.Printf("HTTP Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("[Logger] Shutting down server...")

	// Graceful shutdown
	cancel()          // Stops UDP read loop
	udpServer.Stop()  // Closes UDP connection
	processUC.Close() // Closes log channel
	truncateUC.Stop() // Stops truncation worker
	httpServer.Stop() // Gracefully stops HTTP server

	fmt.Println("[Logger] Server stopped cleanly")
}
