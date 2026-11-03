package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	"time"
)

func PlaceChart(w domain.Repository, el domain.Block) (bool, error) {
	if canon(el.Type) != "Statechart" && canon(el.Type) != "State" {
		return false, nil
	}
	state := domain.Text(el.Props["state"])
	if state == "" {
		state = "idle"
	}
	next := domain.Text(el.Props["next"])
	after := domain.Lasting(domain.Text(el.Props["after"]), domain.Number(el.Props["after"], 0))
	w.Board().SetSeries(el.ID, "engine", 0)
	w.MarkShell()
	w.AppendStep(func(now time.Duration) {
		if next != "" && after > 0 && now >= after {
			state = next
		}
		w.Params()[el.ID+"/state"] = 1
		_ = state
		w.Board().SetSeries(el.ID, "engine", 1)
	})
	return true, nil
}
