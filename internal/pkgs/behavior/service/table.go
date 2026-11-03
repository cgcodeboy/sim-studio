package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
)

func PlaceTable(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TableFunction", "Table", "InputTable":
	default:
		return false, nil
	}
	value := domain.Number(el.Props["value"], 0)
	w.Board().SetSeries(el.ID, "engine", value)
	w.Params()[el.ID] = value
	return true, nil
}
