package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

func PlaceResource(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "ResourcePool":
		poolOf(w, el)
		w.Board().SetSeries(el.ID, "engine", Number(el.Props["capacity"], 1))
		return true, nil
	case "Seize", "Release", "Service", "ResourceSendTo", "ResourceTaskStart", "ResourceTaskEnd", "ResourceAttach", "ResourceDetach":
		return true, w.Add(el.ID, newResource(w, el))
	default:
		return false, nil
	}
}

func poolOf(w domain.Repository, el domain.Block) *domain.Units {
	return w.EnsurePool(el.ID, int(Number(el.Props["capacity"], 1)))
}

func newResource(w domain.Repository, el domain.Block) *resourceAtom {
	name := Text(el.Props["resource"])
	if name == "" {
		name = "resourcePool"
	}
	pool := w.EnsurePool(name, 1)
	delay := time.Duration(0)
	if canon(el.Type) == "Service" {
		delay = Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 1))
	}
	if canon(el.Type) == "ResourceSendTo" {
		delay = Travel(LengthOf(el), SpeedOf(el))
	}
	return &resourceAtom{
		kind: canon(el.Type), pool: pool, delay: delay,
		sheet: w.Board(), id: el.ID, x: el.X, y: el.Y,
	}
}

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
