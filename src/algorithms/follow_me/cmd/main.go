package main

import (
	"encoding/json"
	"follow_me/domain"
	"follow_me/infrastructure"
	"follow_me/infrastructure/broker"
	followme "follow_me/infrastructure/follow_me"
	"io"
	"log"
	"os"
)

func main() {
	log.Println("Initializing FollowMe Algorithm...")

	// Load configuration
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}

	byteValue, _ := io.ReadAll(configFile)
	var cfg domain.Config
	if err := json.Unmarshal(byteValue, &cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
	configFile.Close()
	cfg.SwarmId = os.Getenv("SWARM_ID")

	// Initialize components
	udpBroker := broker.NewUDPBroker()

	// Setup Logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	outputs := []io.Writer{os.Stdout}
	if cfg.LogsTopic != "" {
		// Import "follow_me/infrastructure" needed
	brokerWriter := infrastructure.NewBrokerLogWriter(udpBroker, cfg.LogsTopic)
	outputs = append(outputs, brokerWriter)
	}
	log.SetOutput(io.MultiWriter(outputs...))

	manager := followme.NewManager(cfg, udpBroker)

	// Start the algorithm
	if err := manager.Run(); err != nil {
		log.Fatalf("Error starting manager: %v", err)
	}
}
