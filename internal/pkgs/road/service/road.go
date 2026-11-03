package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
	"time"
)

func PlaceRoad(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "CarSource", "CarEnter", "CarExit", "CarDispose":
		w.Board().SetMover(el.ID, el.X, el.Y, 0)
		w.Board().SetSeries(el.ID, "engine", 1)
		return true, nil
	case "CarMoveTo":
		w.AppendStep(carStep(w, el))
		w.MarkShell()
		return true, nil
	default:
		return false, nil
	}
}

func carStep(w domain.Repository, el domain.Block) func(time.Duration) {
	speed := SpeedOf(el)
	x := el.X
	return func(time.Duration) {
		ahead := gap(w, el.ID, x+1, el.Y)
		pace := speed
		if ahead < 2 {
			pace = speed / 2
		}
		x += pace * pedestrianStep.Minutes()
		w.Board().SetMover(el.ID, x, el.Y, 0)
		w.Board().SetSeries(el.ID, "engine", x)
	}
}
