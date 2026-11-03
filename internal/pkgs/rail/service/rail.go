package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

const pedestrianStep = 100 * time.Millisecond

func PlaceRail(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TrainSource", "TrainDispose":
		return true, w.Add(el.ID, newDelay(w.Board(), el))
	case "TrainMoveTo":
		return true, w.Add(el.ID, &train{delayAtom: *newDelay(w.Board(), el), gate: switchOf(w, el)})
	default:
		return false, nil
	}
}

func switchOf(w domain.Repository, el domain.Block) *domain.Turnout {
	name := Text(el.Props["switch"])
	if name == "" {
		name = "switch"
	}
	return w.EnsureSwitch(name)
}

type train struct {
	delayAtom
	gate   *domain.Turnout
	moving bool
}

func (t *train) Advance() (time.Duration, bool) {
	if t.moving {
		return t.delay, true
	}
	if t.busy {
		return pedestrianStep, true
	}
	return 0, false
}

func (t *train) External(elapsed time.Duration, msg engine.Message) {
	t.delayAtom.External(elapsed, msg)
	if !t.busy {
		return
	}
	if t.gate.Busy {
		return
	}
	t.gate.Busy = true
	t.moving = true
	t.delay = Travel(LengthOf(domain.Block{Props: map[string]any{}, W: t.x, H: t.y}), defaultSpeed)
}

func (t *train) Internal() {
	if t.moving {
		t.gate.Busy = false
		t.moving = false
	}
	t.delayAtom.Internal()
}
