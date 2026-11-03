package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
)

// populationLimit 是一次运行里的实例上限。
const populationLimit = 200

func PlacePopulation(w domain.Repository, el domain.Block) (bool, error) {
	if canon(el.Type) != "Population" {
		return false, nil
	}
	n := int(Number(el.Props["size"], 1))
	if n > populationLimit {
		return true, domain.ErrInvalid
	}
	w.Board().SetSeries(el.ID, "engine", float64(n))
	return true, nil
}
