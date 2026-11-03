package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/domain"
	"time"
)

// stockStep 是存量积分的一步，发生在离散事件处理函数之外。
const stockStep = time.Second

func PlaceDynamics(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Stock", "Flow", "DynamicVariable", "Parameter":
		noteDynamic(w, el)
		return true, nil
	default:
		return false, nil
	}
}

func noteDynamic(w domain.Repository, el domain.Block) {
	w.MarkDynamics()
	w.MarkShell()
	switch canon(el.Type) {
	case "Stock":
		level := Number(el.Props["initial"], 0)
		w.Board().SetSeries(el.ID, "dynamics", level)
		rateKey := el.ID + "/rate"
		w.AppendStep(func(time.Duration) {
			level += w.Params()[rateKey] * stockStep.Seconds()
			w.Board().SetSeries(el.ID, "dynamics", level)
		})
	case "Flow":
		w.Params()[flowTarget(w, el)+"/rate"] = Number(el.Props["rate"], 0)
		w.Board().SetSeries(el.ID, "dynamics", Number(el.Props["rate"], 0))
	default:
		w.Params()[el.ID] = Number(el.Props["value"], Number(el.Props["initial"], 0))
		w.Board().SetSeries(el.ID, "dynamics", w.Params()[el.ID])
	}
}

func flowTarget(w domain.Repository, el domain.Block) string {
	for _, link := range w.Agent().Links {
		if link.From == el.ID {
			return link.To
		}
	}
	for _, other := range w.Agent().Elements {
		if canon(other.Type) == "Stock" {
			return other.ID
		}
	}
	return el.ID
}
