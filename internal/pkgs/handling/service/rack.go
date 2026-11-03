package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

func PlaceRack(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "RackSystem":
		w.Board().SetSeries(el.ID, "engine", Number(el.Props["capacity"], 1))
		return true, nil
	case "RackStore", "Store", "RackPick", "Retrieve":
		return true, w.Add(el.ID, newRack(w, el))
	default:
		return false, nil
	}
}

func newRack(w domain.Repository, el domain.Block) *resourceAtom {
	name := Text(el.Props["storage"])
	if name == "" {
		name = el.ID
	}
	pool := w.EnsurePool(name, int(Number(el.Props["capacity"], 1)))
	kind := "Seize"
	if canon(el.Type) == "RackPick" || canon(el.Type) == "Retrieve" {
		kind = "Release"
	}
	return &resourceAtom{kind: kind, pool: pool, sheet: w.Board(), id: el.ID, x: el.X, y: el.Y}
}
