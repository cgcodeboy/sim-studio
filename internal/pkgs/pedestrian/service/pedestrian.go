package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
	"time"
)

// pedestrianStep 是行人外壳的一步，不进入离散事件处理函数。
const pedestrianStep = 100 * time.Millisecond

func PlacePed(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "PedSource", "PedEnter":
		w.Board().SetMover(el.ID, el.X, el.Y, 0)
		w.Board().SetSeries(el.ID, "engine", 1)
		return true, nil
	case "PedGoTo", "PedWait", "PedService", "PedSelectOutput", "PedExit", "PedSink", "PedGroupAssemble", "PedGroupDisassemble":
		w.AppendStep(pedStep(w, el))
		w.MarkShell()
		return true, nil
	default:
		return false, nil
	}
}

func pedStep(w domain.Repository, el domain.Block) func(time.Duration) {
	speed := SpeedOf(el)
	x, y := el.X, el.Y
	return func(now time.Duration) {
		_ = now
		step := speed * pedestrianStep.Minutes()
		x += step
		if gap(w, el.ID, x, y) < 1 {
			x -= step
		}
		w.Board().SetMover(el.ID, x, y, 0)
		w.Board().SetSeries(el.ID, "engine", x)
	}
}

func gap(w domain.Repository, id string, x, y float64) float64 {
	best := 1e9
	for _, mover := range w.Board().Spots() {
		if mover.ID == id {
			continue
		}
		dx := mover.X - x
		dy := mover.Y - y
		d := dx*dx + dy*dy
		if d < best {
			best = d
		}
	}
	if best > 1 {
		return 2
	}
	return 0
}
