package main
import (
	"fmt"
	"netsim_gateway/infra/config"
	"netsim_gateway/infra/logger"
	"netsim_gateway/infra/udp"
	"netsim_gateway/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig("config.json")
	log := logger.NewUDPLogger(fmt.Sprintf("%s:%d", cfg.LoggerIp, cfg.LoggerPort))

	telemetryConn := udp.NewConnection(cfg.TelemetryPort, log)
	defer telemetryConn.Close()

	messagesConn := udp.NewConnection(cfg.MessagesPort, log)
	defer messagesConn.Close()

	subscribersConn := udp.NewConnection(cfg.SubscribersPort, log)
	defer subscribersConn.Close()

	netsimConn := udp.NewConnection(cfg.NetsimListenPort, log)
	defer netsimConn.Close()

	// We can use any of the uav connections to send back to the UAVs
	// For simplicity, we will just use the messagesConn as the uavSender
	uavSender := udp.NewSender(messagesConn, log)
	netsimSender := udp.NewSender(netsimConn, log)
	subscribersSender := udp.NewSender(subscribersConn, log)
	
	telemetryRecv := udp.NewReceiver(telemetryConn, log)
	messagesRecv := udp.NewReceiver(messagesConn, log)
	subscribersRecv := udp.NewReceiver(subscribersConn, log)
	netsimRecv := udp.NewReceiver(netsimConn, log)

	netsimAddrs := config.DiscoverNetsims(cfg.Addrs)
	gateway := usecase.NewGateway(usecase.Config{
		TelemetryPort:     cfg.TelemetryPort,
		MessagesPort:      cfg.MessagesPort,
		SubscribersPort:   cfg.SubscribersPort,
		NetsimListenPort:  cfg.NetsimListenPort,
		SnapshotIntervalS: cfg.SnapshotIntervalS,
	}, netsimAddrs, uavSender, netsimSender, subscribersSender, log)

	telemetryHandler := usecase.NewTelemetryHandler(gateway)
	messagesHandler := usecase.NewMessagesHandler(gateway)
	subscribersHandler := usecase.NewSubscribersHandler(gateway)
	netsimHandler := usecase.NewNetsimHandler(gateway)

	go telemetryRecv.Run(telemetryHandler)
	go messagesRecv.Run(messagesHandler)
	go subscribersRecv.Run(subscribersHandler)
	go netsimRecv.Run(netsimHandler)
	
	go uavSender.Run()
	go netsimSender.Run()
	go subscribersSender.Run()
	go usecase.RunAggregatedSnapshotEmitter(gateway)
	go func() {
		addrs := netsimAddrs
		for ; len(cfg.Addrs) > len(addrs); {
			addrs = config.DiscoverNetsims(cfg.Addrs)
			time.Sleep(1 * time.Second)
		}
		gateway.UpdateNetsims(addrs)
	}()

	log.Info("Gateway started", "telemetry_port", cfg.TelemetryPort, "messages_port", cfg.MessagesPort, "subscribers_port", cfg.SubscribersPort, "netsim_port", cfg.NetsimListenPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down Gateway...")
}
