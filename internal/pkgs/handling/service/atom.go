package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

type resourceAtom struct {
	kind    string
	pool    *domain.Units
	delay   time.Duration
	busy    bool
	waiting int
	sheet   domain.Board
	id      string
	x, y    float64
}

func (r *resourceAtom) Advance() (time.Duration, bool) {
	if r.busy {
		return r.delay, true
	}
	return 0, false
}

func (r *resourceAtom) Output() []engine.Message {
	return []engine.Message{{Port: "out", Payload: []byte{1}}}
}

func (r *resourceAtom) Internal() {
	r.busy = false
	if r.kind == "Release" || r.kind == "Service" || r.kind == "ResourceTaskEnd" || r.kind == "ResourceDetach" {
		r.pool.Free++
	}
	r.sheet.Leave(r.id)
	r.sheet.SetSeries(r.id, "engine", float64(r.pool.Free))
	if r.waiting > 0 && r.pool.Free > 0 && r.takes() {
		r.waiting--
		r.pool.Free--
		r.busy = true
		r.sheet.Enter(r.id, r.x, r.y)
	}
}

func (r *resourceAtom) External(time.Duration, engine.Message) {
	if r.kind == "Release" || r.kind == "ResourceTaskEnd" || r.kind == "ResourceDetach" {
		r.busy = true
		return
	}
	if !r.takes() {
		r.busy = true
		return
	}
	if r.pool.Free <= 0 || r.busy {
		r.waiting++
		r.sheet.SetSeries(r.id, "engine", float64(r.waiting))
		return
	}
	r.pool.Free--
	r.busy = true
	r.sheet.Enter(r.id, r.x, r.y)
}

func (r *resourceAtom) takes() bool {
	switch r.kind {
	case "Seize", "Service", "ResourceSendTo", "ResourceTaskStart", "ResourceAttach":
		return true
	default:
		return false
	}
}
