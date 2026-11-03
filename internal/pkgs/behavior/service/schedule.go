package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	"time"
)

func PlaceSchedule(w domain.Repository, el domain.Block) (bool, error) {
	if canon(el.Type) != "Schedule" {
		return false, nil
	}
	value := domain.Number(el.Props["value"], 1)
	w.MarkShell()
	w.AppendStep(func(now time.Duration) {
		_ = now
		w.Params()[el.ID] = value
		w.Board().SetSeries(el.ID, "engine", value)
	})
	return true, nil
}
