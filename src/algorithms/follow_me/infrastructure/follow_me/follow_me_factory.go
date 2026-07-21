package followme

import (
	"follow_me/domain"
	"follow_me/ports"
	"follow_me/usecase"
)

func NewManager(cfg domain.Config, broker ports.CommunicationProvider) ports.FollowMeManager {
	switch cfg.Role {
	case "master":
		return usecase.NewFollowMeAsMaster(cfg, broker)
	case "slave":
		return usecase.NewFollowMeAsSlave(cfg, broker)
	default:
		panic("Invalid role")
	}
}
