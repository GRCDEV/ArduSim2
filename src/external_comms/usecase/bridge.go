package usecase

import (
	"log"

	"external_comms/domain"
	"external_comms/ports"
)

type GatewayBridge struct {
	config  *domain.AppConfig
	broker     ports.Broker
	netLink    ports.NetSimLink
	loggerLink ports.LoggerLink
}

func NewGatewayBridge(c *domain.AppConfig, b ports.Broker, n ports.NetSimLink, l ports.LoggerLink) *GatewayBridge {
	return &GatewayBridge{
		config:     c,
		broker:     b,
		netLink:    n,
		loggerLink: l,
	}
}

func (g *GatewayBridge) Run() {
	log.Println("Gateway Bridge AUX N Started...")

	brokerChan, err := g.broker.Listen()
	if err != nil {
		log.Fatalf("Broker listen failed: %v", err)
	}

	netChan, err := g.netLink.Listen()
	if err != nil {
		log.Fatalf("NetSim listen failed: %v", err)
	}

	for {
		select {
		case bMsg := <-brokerChan:
			g.handleInternalBrokerMessage(bMsg)
		case nMsg := <-netChan:
			g.handleExternalNetMessage(nMsg)
		}
	}
}

func (g *GatewayBridge) handleInternalBrokerMessage(msg ports.BrokerMessage) {
	// Internal to External
	switch msg.Topic {
	case g.config.SubTelemetryTopic:
		// Route internal telemetry to swarm
		extMsg := domain.SendedNetSimMessage{
			Topic:    "telemetry",
			Source:  g.config.UAVId,
			Payload: msg.Payload,
		}
		g.netLink.Send(extMsg)
		log.Printf("[Internal->External] Forwarded Telemetry")

	case g.config.SubMessagesTopic:
		// Route internal P2P message to swarm
		extMsg := domain.SendedNetSimMessage{
			Topic:    "broadcast",
			Source:  g.config.UAVId,
			Payload: msg.Payload,
		}
		g.netLink.Send(extMsg)
		log.Printf("[Internal->External] Forwarded Message: %v", msg.Payload)

	case g.config.SubLogsTopic:
		// Route internal logs directly to the Logger microservice
		if g.loggerLink != nil {
			if err := g.loggerLink.SendLog(msg.Payload); err != nil {
				log.Printf("[Internal->External] Failed to forward Log to logger: %v", err)
			}
		}
	}
}

func (g *GatewayBridge) handleExternalNetMessage(msg domain.ReceivedNetSimMessage) {
	// External to Internal
	// Ignore our own echo if NetSim broadcasts everything back
	log.Printf("[External->Internal] Forwarding message from %s: %v", msg.Source, msg.Payload)
	innerPayload, ok:= msg.Payload["payload"].(map[string]interface{})
	if !ok {
		return
	}
	
	if msg.Source == g.config.UAVId {
		return
	}
	if _, ok := msg.Payload["topic"]; !ok {
		return
	}

	if destUavId, ok := innerPayload["dest_uav_id"]; ok && destUavId != "" {
		if destUavIdStr, isStr := destUavId.(string); isStr && destUavIdStr != g.config.UAVId {
			return
		}
	}

	if destSwarmId, ok := innerPayload["dest_swarm_id"]; ok && destSwarmId != "" {
		if destSwarmIdStr, isStr := destSwarmId.(string); isStr && destSwarmIdStr != g.config.SwarmId {
			return
		}
	}

	g.broker.Publish(msg.Payload)
}
