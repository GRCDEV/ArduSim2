package main

import (
	"io"
	"log"
	"os"
	"time"

	"application/infrastructure/broker"
	"application/infrastructure/config"
	"application/infrastructure/uav"
	"application/ports"
	"application/usecase"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config.json>\n", os.Args[0])
	}

	configFile := os.Args[1]

	// Dependency Injection Loaders
	log.Println("Starting Application initialization...")
	log.Println("Loading Configuration...")
	fileLoader := config.NewFileLoader()
	appConfig, err := fileLoader.LoadAppConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if appConfig.MixWindowMs == 0 {
		appConfig.MixWindowMs = 200 // Default to 5Hz sampling
	}

	// Infra
	log.Println("Initializing UDP Broker...")
	udpBroker := broker.NewUDPBroker()
	defer udpBroker.Close()

	for {
		if err := udpBroker.Connect(appConfig.BrokerIP, appConfig.BrokerPort, []string{
			appConfig.SuggestionsTopic,
			appConfig.GlobalCommands,
			appConfig.TelemetryTopic,
		}); err != nil {
			log.Printf("Failed to connect to Broker: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	log.Println("Successfully connected to UDP Broker.")

	setupLogger(udpBroker, appConfig.LogsTopic)
	log.Println("Logger configured. Logging to stdout and broker.")

	log.Println("Establishing direct UDP link to UAV...")
	var uavLink *uav.DirectUAVLink
	for {
		uavLink, err = uav.NewDirectUAVLink(appConfig.UAVControllerIP, appConfig.UAVControllerPort, appConfig.UAVTelemetryPort)
		if err != nil {
			log.Printf("Failed to establish direct UDP link to UAV: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer uavLink.Close()
	log.Println("Successfully connected to UAV telemetry and command ports.")

	log.Println("Initializing Movement Mixer use case...")
	mixer := usecase.NewMovementMixer(udpBroker, uavLink, appConfig)

	mixer.Run()
}

func setupLogger(b ports.Broker, logsTopic string) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	outputs := []io.Writer{os.Stdout}

	// Add our custom BrokerLogWriter if broker is connected
	if b != nil && logsTopic != "" {
		brokerWriter := broker.NewBrokerLogWriter(b, logsTopic)
		outputs = append(outputs, brokerWriter)
	}

	multi := io.MultiWriter(outputs...)
	log.SetOutput(multi)
}
