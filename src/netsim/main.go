package main
import (
	"fmt"
	"netsim/domain/service"
	"netsim/infra/config"
	"netsim/infra/logger"
	"netsim/infra/udp"
	"netsim/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig("config.json")
	log := logger.NewUDPLogger(cfg.Level, fmt.Sprintf("%s:%d", cfg.LoggerIp, cfg.LoggerPort))

	conn := udp.NewConnection(cfg.ListenPort, log)
	defer conn.Close()

	sender := udp.NewSender(conn, fmt.Sprintf("%s:%d", cfg.GatewayIp, cfg.GatewayPort), log)
	receiver := udp.NewReceiver(conn, log)

	nodeId := os.Getenv("NODE_ID")
	simCfg := cfg.SimulationConfig()
	spatial := service.NewSpatialGrid(simCfg.ChunkSizeM)
	sim := usecase.NewSimulator(nodeId, simCfg, spatial, sender, log)

	go receiver.Run(sim)
	go sender.Run()
	go usecase.RunSnapshotEmitter(sim, sender, simCfg.SnapshotIntervalS, log)

	go func() {
		flushInterval := cfg.FlushIntervalMs
		if flushInterval <= 0 {
			flushInterval = 1
		}
		ticker := time.NewTicker(time.Millisecond * time.Duration(flushInterval))
		for range ticker.C {
			sim.SendMessages()
		}
	}()

	log.Info("Netsim started", "node_id", nodeId, "port", cfg.ListenPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down Netsim...")
}
