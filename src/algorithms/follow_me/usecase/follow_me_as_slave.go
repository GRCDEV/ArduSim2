package usecase

import (
	"fmt"
	"follow_me/domain"
	"follow_me/ports"
	"log"

	"github.com/go-viper/mapstructure/v2"
)

type FollowMeAsSlave struct {
	FollowMeBase
	tel            *domain.Telemetry
	alreadyTakeOff bool
}

func NewFollowMeAsSlave(
	cfg domain.Config,
	broker ports.CommunicationProvider,
) *FollowMeAsSlave {
	fm := &FollowMeAsSlave{
		FollowMeBase: FollowMeBase{
			Cfg:    cfg,
			Broker: broker,
			State:  domain.IDLE,
		},
	}
	fm.handler = fm
	return fm
}


func (f *FollowMeAsSlave) HandleTelemetryTopic(payload any) {
	if f.State == domain.IDLE {
		return
	}

	var tel domain.Telemetry
	if err := mapstructure.Decode(payload.(map[string]any), &tel); err != nil {
		log.Printf("Error unmarshalling telemetry: %v", err)
		return
	}

	f.tel = &tel

	if f.State == domain.TAKEOFF && f.tel.Position.RelativeAlt > f.Cfg.SlavesTakeoffAltitude {
		f.State = domain.RUNNING
	}
}

func (f *FollowMeAsSlave) OnStart() {
	f.State = domain.RUNNING
}

func (f *FollowMeAsSlave) OnPause() {
	f.State = domain.IDLE
}

func (f *FollowMeAsSlave) OnStop() {
	f.State = domain.IDLE
	f.Broker.Publish(f.Cfg.SuggestionsTopic, domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
		Endpoint: domain.ActionLand,
	})
	f.alreadyTakeOff = false
}

func (f *FollowMeAsSlave) HandleSubscriptionTopic(payload any) {
	if err := f.HandleCommand(payload); err == nil {
		return
	}
	f.handleMasterTelemetry(payload)
}


func (f *FollowMeAsSlave) handleMasterTelemetry(payload any) error {
	if f.State != domain.RUNNING {
		return nil
	}

	var masterTel domain.FollowMeMessage
	if err := mapstructure.Decode(payload.(map[string]any), &masterTel); err != nil {
		return err
	}

	if masterTel.Timestamp == 0 {
		return fmt.Errorf("void master telemetry")
	}

	isMasterFlying := masterTel.RelativeAlt >= 0.5
	imFlying := f.tel.Position.RelativeAlt >= 0.5

	if !isMasterFlying {
		if imFlying {
			f.Broker.Publish(
				f.Cfg.SuggestionsTopic,
				domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
					Endpoint: domain.ActionLand,
				},
			)
			f.alreadyTakeOff = false
		}
		return nil
	}

	if !imFlying && !f.alreadyTakeOff {
		f.takeOff()
		return nil
	}

	f.Broker.Publish(
		f.Cfg.SuggestionsTopic,
		domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
			Endpoint:  domain.ActionMoveTo,
			Latitude:  masterTel.Lat,
			Longitude: masterTel.Lon,
			Altitude:  masterTel.Alt,
		},
	)

	return nil
}

func (f *FollowMeAsSlave) takeOff() {
	f.alreadyTakeOff = true
	f.State = domain.TAKEOFF

	f.Broker.Publish(
		f.Cfg.SuggestionsTopic,
		domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
			Endpoint: domain.ActionArm,
		},
	)

	f.Broker.Publish(
		f.Cfg.SuggestionsTopic,
		domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
			Endpoint:   domain.ActionSetFlightmode,
			Flightmode: "GUIDED",
		},
	)

	log.Printf("Taking off to %.2f", f.Cfg.SlavesTakeoffAltitude)

	f.Broker.Publish(
		f.Cfg.SuggestionsTopic,
		domain.Suggestion{
	ServiceID: f.Cfg.ServiceID,
			Endpoint: domain.ActionTakeoff,
			Altitude: f.Cfg.SlavesTakeoffAltitude,
		},
	)
}
