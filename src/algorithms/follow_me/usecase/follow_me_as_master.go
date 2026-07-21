package usecase

import (
	"follow_me/domain"
	chronjob "follow_me/infrastructure/chron-job"
	"follow_me/ports"
	"log"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

type FollowMeAsMaster struct {
	FollowMeBase
	lastTelemetry  *domain.Telemetry
	chronosned     *chronjob.ChronJob
	alreadyTakeOff bool
}

func NewFollowMeAsMaster(cfg domain.Config, broker ports.CommunicationProvider) *FollowMeAsMaster {
	fm := &FollowMeAsMaster{
		FollowMeBase: FollowMeBase{
			Cfg:    cfg,
			Broker: broker,
			State:  domain.IDLE,
		},
	}
	fm.handler = fm
	fm.chronosned = chronjob.NewChronJob(fm.sendTelemetry, time.Duration(cfg.SendPeriodMs)*time.Millisecond)

	return fm
}

func (f *FollowMeAsMaster) HandleTelemetryTopic(payload any) {
	if f.State == domain.IDLE { return }

	var tel domain.Telemetry
	if err := mapstructure.Decode(payload.(map[string]any), &tel); err != nil {
		log.Printf("Error unmarshalling telemetry: %v", err)
		return
	}

	if tel.Position.RelativeAlt > 0.5 {
		f.alreadyTakeOff = true
	}

	if !f.isLanding(&tel) {
		f.lastTelemetry = &tel
		return
	}

	f.State = domain.IDLE
	f.Broker.Publish(f.Cfg.SuggestionsTopic, domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
		Endpoint: domain.ActionLand,
	})
	f.Broker.Publish(f.Cfg.BroadcastTopic, domain.CommandMessage{
		Command:     "finish",
		Source:      "followme",
		DestSwarmId: f.Cfg.SwarmId,
	})
	f.chronosned.Stop()

	f.lastTelemetry = &tel
	f.sendTelemetry()
}

func (f *FollowMeAsMaster) isLanding(current *domain.Telemetry) bool {
	if !f.alreadyTakeOff {
		return false
	}
	if current.Position.RelativeAlt > 0.5 {
		return false
	}
	if f.lastTelemetry == nil {
		return false
	}
	return f.lastTelemetry.Position.RelativeAlt > current.Position.RelativeAlt
}


func (f *FollowMeAsMaster) OnStart() {
	f.State = domain.RUNNING
	f.chronosned.Start()
}

func (f *FollowMeAsMaster) OnPause() {
	f.State = domain.IDLE
	f.chronosned.Pause()
}

func (f *FollowMeAsMaster) OnStop() {
	f.State = domain.IDLE
	f.Broker.Publish(f.Cfg.SuggestionsTopic, domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
		Endpoint: domain.ActionLand,
	})
	f.chronosned.Stop()
}

func (f *FollowMeAsMaster) sendTelemetry() {
	f.Broker.Publish(
		f.Cfg.BroadcastTopic,
		map[string]any{
			"topic":         f.Cfg.SubscriptionTopic,
			"payload":       domain.FollowMeMessage{}.FromTelemetry(*f.lastTelemetry, f.Cfg.SwarmId),
		},
	)
}
