package main

import (
	"io"
	"log"
	"os"

	"mission/infrastructure"
	"mission/infrastructure/broker"
	"mission/infrastructure/config"
	"mission/infrastructure/parser"
	"mission/ports"
	"mission/usecase"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config.json>\n", os.Args[0])
	}

	configFile := os.Args[1]

	// Dependency Injection
	udpBroker := broker.NewUDPBroker()
	defer udpBroker.Close()

	fileLoader := config.NewFileLoader()
	kmlParser := parser.NewKmlParser()

	manager := usecase.NewMissionManager(udpBroker, fileLoader, kmlParser)

	if err := manager.Initialize(configFile); err != nil {
		log.Fatalf("Failed to initialize mission manager: %v", err)
	}

	setupLogger(udpBroker, manager.GetConfig().LogsTopic)

	manager.Run()
}

func setupLogger(broker ports.Broker, logsTopic string) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	outputs := []io.Writer{os.Stdout}

	if broker != nil && logsTopic != "" {
		brokerWriter := infrastructure.NewBrokerLogWriter(broker, logsTopic)
		outputs = append(outputs, brokerWriter)
	}

	multi := io.MultiWriter(outputs...)
	log.SetOutput(multi)
}
