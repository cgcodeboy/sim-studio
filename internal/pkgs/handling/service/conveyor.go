package service

import (
	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

func PlaceConveyor(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Conveyor", "Convey", "ConveyorEnter":
		return true, w.Add(el.ID, newBelt(w.Board(), el))
	default:
		return false, nil
	}
}

type belt struct {
	delayAtom
}

func newBelt(sheet domain.Board, el domain.Block) *belt {
	atom := newDelay(sheet, el)
	atom.delay = Travel(LengthOf(el), SpeedOf(el))
	atom.open = true
	return &belt{delayAtom: *atom}
}

func (b *belt) Output() []engine.Message {
	return b.delayAtom.Output()
}
