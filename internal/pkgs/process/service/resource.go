package service

import (
	"math"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
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
	pool := w.EnsurePool(el.ID, int(Number(el.Props["capacity"], 1)))
	pool.X, pool.Y, pool.Z, pool.Yaw = el.Px, el.Py, el.Pz, el.Yaw
	return pool
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
		sheet: w.Board(), id: el.ID, el: el, x: el.X, y: el.Y,
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
	el      domain.Block
	body    string
	x, y    float64
}

func (r *resourceAtom) Advance() (time.Duration, bool) {
	if r.busy {
		return scaled(r.sheet, r.span()), true
	}
	return 0, false
}

func (r *resourceAtom) span() time.Duration {
	if r.delay > r.walk() {
		return r.delay
	}
	return r.walk()
}

func (r *resourceAtom) walk() time.Duration {
	x, y, z, _, _ := spanOf(r.el)
	dx, dy, dz := r.pool.X-x, r.pool.Y-y, r.pool.Z-z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	return time.Duration(dist / 1.3 * float64(time.Second))
}

func (r *resourceAtom) Output() []engine.Message {
	if r.body == "" {
		r.body = r.id
	}
	return []engine.Message{{Port: "out", Payload: []byte(r.body)}}
}

func (r *resourceAtom) Internal() {
	r.busy = false
	if r.kind == "Release" || r.kind == "Service" || r.kind == "ResourceTaskEnd" || r.kind == "ResourceDetach" {
		r.pool.Free++
	}
	r.sheet.Leave(r.id)
	r.sheet.Hide("person:" + r.id)
	r.sheet.SetSeries(r.id, "engine", float64(r.pool.Free))
	if r.waiting > 0 && r.pool.Free > 0 && r.takes() {
		r.waiting--
		r.pool.Free--
		r.busy = true
		r.sheet.Enter(r.id, r.x, r.y)
	}
}

func (r *resourceAtom) External(_ time.Duration, msg engine.Message) {
	r.body = bodyID(msg, r.id)
	if r.kind == "Release" || r.kind == "ResourceTaskEnd" || r.kind == "ResourceDetach" {
		r.busy = true
		r.pose()
		return
	}
	if !r.takes() {
		r.busy = true
		r.pose()
		return
	}
	if r.pool.Free <= 0 || r.busy {
		r.waiting++
		showStay(r.sheet, r.body, r.el, "wait", 0)
		r.sheet.SetSeries(r.id, "engine", float64(r.waiting))
		return
	}
	r.pool.Free--
	r.busy = true
	r.sheet.Enter(r.id, r.x, r.y)
	r.pose()
}

func (r *resourceAtom) pose() {
	showStay(r.sheet, r.body, r.el, "work", r.delay.Seconds())
	x, y, z, yaw, _ := spanOf(r.el)
	dx, dy, dz := place(x, y, z, yaw, 0, 0.55)
	seconds := r.walk().Seconds()
	r.sheet.Begin("person:"+r.id, "person", "walk", specOf(r.pool.X, r.pool.Y, r.pool.Z, dx, dy, dz, yaw, 1.3, seconds))
}

func (r *resourceAtom) takes() bool {
	switch r.kind {
	case "Seize", "Service", "ResourceSendTo", "ResourceTaskStart", "ResourceAttach":
		return true
	default:
		return false
	}
}
