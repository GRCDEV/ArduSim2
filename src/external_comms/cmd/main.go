package main

import (
	"io"
	"log"
	"os"
	"time"

	"external_comms/infrastructure"
	"external_comms/ports"
	"external_comms/usecase"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config.json>\n", os.Args[0])
	}
	configFile := os.Args[1]

	fileLoader := infrastructure.NewFileLoader()
	config, err := fileLoader.LoadAppConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config.UAVId = os.Getenv("UAV_ID")
	config.SwarmId = os.Getenv("SWARM_ID")

	broker := infrastructure.NewUDPBroker()
	defer broker.Close()

	for {
		if err := broker.Connect(config.BrokerIP, config.BrokerPort, []string{
			config.SubTelemetryTopic,
			config.SubMessagesTopic,
			config.SubLogsTopic,
		}); err != nil {
			log.Printf("Failed to connect broker: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	var netLink *infrastructure.UDPNetSimLink
	for {
		netLink, err = infrastructure.NewUDPNetSimLink(config.NetSimIP, config.NetSimTelemetryPort, config.NetSimMessagesPort)
		if err != nil {
			log.Printf("Failed to connect NetSim: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer netLink.Close()

	var loggerLink *infrastructure.UDPLoggerLink
	if config.LoggerIP != "" && config.LoggerPort != 0 {
		loggerLink = infrastructure.NewUDPLoggerLink(config.LoggerIP, config.LoggerPort)
		defer loggerLink.Close()
		log.Printf("Logger UDP link initialized for %s:%d (queues logs if disconnected)", config.LoggerIP, config.LoggerPort)
	}

	setupLogger(loggerLink)

	bridge := usecase.NewGatewayBridge(config, broker, netLink, loggerLink)
	bridge.Run()
}

func setupLogger(loggerLink ports.LoggerLink) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	outputs := []io.Writer{os.Stdout}

	if loggerLink != nil {
		directWriter := infrastructure.NewDirectLogWriter(loggerLink)
		outputs = append(outputs, directWriter)
	}

	multi := io.MultiWriter(outputs...)
	log.SetOutput(multi)
}
