package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	"time"
)

func PlaceEvent(w domain.Repository, el domain.Block) (bool, error) {
	if canon(el.Type) != "Event" {
		return false, nil
	}
	when := domain.Lasting(domain.Text(el.Props["time"]), domain.Number(el.Props["time"], 1))
	w.MarkShell()
	w.AppendStep(func(now time.Duration) {
		if now >= when {
			w.Board().SetSeries(el.ID, "engine", now.Seconds())
		}
	})
	return true, nil
}
