package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
	"time"
)

const pedestrianStep = 100 * time.Millisecond

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
