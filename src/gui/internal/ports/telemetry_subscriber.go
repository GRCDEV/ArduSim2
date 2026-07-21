package ports

import "context"

// TelemetrySubscriber handles receiving telemetry from the network simulator.
type TelemetrySubscriber interface {
	Start(ctx context.Context)
	SetRemoteAddr(addr string)
	SetExpectedFleet(uavIDs []string)
	SetOnFinish(fn func())
	SendGlobalBroadcast(payload interface{}) error
	NotifyUserStoppedAll()
}
