package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
)

func PlaceMessage(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Connection", "AgentLink":
		w.Board().SetSeries(el.ID, "engine", float64(len(w.Agent().Links)))
		return true, nil
	default:
		return false, nil
	}
}
