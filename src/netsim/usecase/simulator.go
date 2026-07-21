package usecase

import (
	"encoding/json"
	"fmt"
	"netsim/domain/model"
	"netsim/domain/service"
	"netsim/ports/input"
	"netsim/ports/output"
	"sync"
	"time"
)

type Config struct {
	LossMode          string
	BufferSizeBytes   int
	CsmaRangeM        float64
	MaxCsmaRetries    uint32
	ChunkSizeM        float64
	ChunkRadius       int64
	MaxRangeM         float64
	SnapshotIntervalS int
}

type Simulator struct {
	Config      Config
	UAVs        map[string]*model.UAV
	DelayedMsgs []model.DelayedMessage
	PendingMsgs map[string][]model.Message
	Spatial     *service.SpatialGrid
	Sender      output.PacketSender
	Logger      output.Logger
	Cache       sync.Map // string -> json.RawMessage
	mu          sync.Mutex
	NodeID      string
	Strategy    BroadcastStrategy
}

func NewSimulator(nodeID string, cfg Config, spatial *service.SpatialGrid, sender output.PacketSender, logger output.Logger) *Simulator {
	service.InitLossFunction(cfg.LossMode)

	var strategy BroadcastStrategy
	if cfg.LossMode == "unrestricted" {
		strategy = unrestrictedStrategy{}
	} else {
		strategy = csmaStrategy{}
	}

	return &Simulator{
		Config:      cfg,
		UAVs:        make(map[string]*model.UAV),
		PendingMsgs: make(map[string][]model.Message),
		Spatial:     spatial,
		Sender:      sender,
		Logger:      logger,
		NodeID:      nodeID,
		Strategy:    strategy,
	}
}

type GatewayToNetsimMsg struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

type UAVTelemetryPayload struct {
	UAVID   string              `json:"uav_id"`
	Payload model.TelemetryData `json:"payload"`
}

type UAVBroadcastPayload struct {
	UAVID   string          `json:"uav_id"`
	Payload json.RawMessage `json:"payload"`
}

type PeerBroadcastPayload struct {
	SenderID       string          `json:"sender_id"`
	SenderPosition model.Position  `json:"sender_position"`
	Payload        json.RawMessage `json:"payload"`
	OriginNode     string          `json:"origin_node"`
}

func (s *Simulator) Handle(pkt input.RawPacket) {
	var msg GatewayToNetsimMsg
	if err := json.Unmarshal(pkt.Data, &msg); err != nil {
		s.Logger.Error("Failed to parse packet", "error", err)
		return
	}

	switch msg.Topic {
	case "uav_telemetry":
		var tel UAVTelemetryPayload
		if err := json.Unmarshal(msg.Payload, &tel); err == nil {
			s.UpdateUAVTelemetry(tel.UAVID, tel.Payload)
			s.Logger.Info(fmt.Sprintf("Processed telemetry from UAV %s", tel.UAVID))
		}
	case "uav_broadcast":
		var bcast UAVBroadcastPayload
		if err := json.Unmarshal(msg.Payload, &bcast); err == nil {
			payloadStr := string(bcast.Payload)
			s.EnqueueBroadcast(bcast.UAVID, payloadStr, 0, pkt.ReceivedAt)
			s.NotifyGatewayOfBroadcast(bcast.UAVID, payloadStr)
			s.Logger.Info(fmt.Sprintf("Processed broadcast from UAV %s", bcast.UAVID), "timestamp", time.Now().Format(time.RFC3339Nano))
		}
	case "peer_broadcast":
		var peerBcast PeerBroadcastPayload
		if err := json.Unmarshal(msg.Payload, &peerBcast); err == nil {
			if peerBcast.OriginNode != s.NodeID {
				payloadStr := string(peerBcast.Payload)
				s.HandlePeerBroadcast(peerBcast.SenderID, &peerBcast.SenderPosition, payloadStr, pkt.ReceivedAt)
			}
		}
	}
}

func (s *Simulator) NotifyGatewayOfBroadcast(senderID, payload string) {
	s.mu.Lock()
	uav, exists := s.UAVs[senderID]
	s.mu.Unlock()

	if !exists {
		return
	}

	notify := map[string]any{
		"topic": "broadcast_notify",
		"payload": map[string]any{
			"sender_id":       senderID,
			"sender_position": uav.Position,
			"payload":         json.RawMessage(payload), // Will be raw string or json string, let's keep it generic
		},
	}
	data, _ := json.Marshal(notify)
	s.Sender.Send(data, nil)
}
