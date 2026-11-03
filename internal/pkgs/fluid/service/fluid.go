package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
	"time"
)

const stockStep = time.Second

func PlaceFluid(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "FluidSource", "Tank", "Valve", "ProcessTank", "FluidSplit", "FluidSelectOutput", "FluidDispose":
		w.AppendStep(fluidStep(w, el))
		w.MarkShell()
		w.MarkDynamics()
		return true, nil
	default:
		return false, nil
	}
}

func fluidStep(w domain.Repository, el domain.Block) func(time.Duration) {
	level := Number(el.Props["initial"], 0)
	key := el.ID + "/rate"
	w.Params()[key] = Number(el.Props["rate"], 0)
	w.Board().SetSeries(el.ID, "dynamics", level)
	return func(time.Duration) {
		level += w.Params()[key] * stockStep.Seconds()
		if level < 0 {
			level = 0
		}
		w.Board().SetSeries(el.ID, "dynamics", level)
	}
}
