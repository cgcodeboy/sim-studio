package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

// fleetLimit 是一次运行里的车辆上限。
const fleetLimit = 10

func PlaceFleet(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TransporterFleet", "TransporterControl":
		if int(Number(el.Props["capacity"], 1)) > fleetLimit {
			return true, domain.ErrInvalid
		}
		w.Board().SetSeries(el.ID, "engine", Number(el.Props["capacity"], 1))
		return true, nil
	case "SeizeTransporter", "ReleaseTransporter", "MoveByTransporter":
		return true, w.Add(el.ID, newHaul(w.Board(), el))
	default:
		return false, nil
	}
}

func newHaul(sheet domain.Board, el domain.Block) *delayAtom {
	atom := newDelay(sheet, el)
	atom.open = true
	if canon(el.Type) == "MoveByTransporter" {
		atom.delay = Travel(LengthOf(el), SpeedOf(el))
	}
	return atom
}
