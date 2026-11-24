package service

import (
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

func PlaceRack(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "RackSystem":
		pool := w.EnsurePool(el.ID, int(Number(el.Props["capacity"], 1)))
		pool.X, pool.Y, pool.Z, pool.Yaw = el.Px, el.Py, el.Pz, el.Yaw
		w.Board().SetSeries(el.ID, "engine", float64(pool.Free))
		return true, nil
	case "RackStore", "Store", "RackPick", "Retrieve":
		return true, w.Add(el.ID, newRack(w, el))
	default:
		return false, nil
	}
}

type rackAtom struct {
	sheet domain.Board
	el    domain.Block
	pool  *domain.Units
	out   bool
	body  string
	busy  bool
}

func newRack(w domain.Repository, el domain.Block) *rackAtom {
	name := Text(el.Props["storage"])
	if name == "" {
		name = el.ID
	}
	out := canon(el.Type) == "RackPick" || canon(el.Type) == "Retrieve"
	return &rackAtom{sheet: w.Board(), el: el, pool: w.EnsurePool(name, 1), out: out}
}

func (r *rackAtom) Advance() (time.Duration, bool) {
	if !r.busy {
		return 0, false
	}
	return scaled(r.sheet, 800*time.Millisecond), true
}

func (r *rackAtom) Output() []engine.Message {
	if r.body == "" {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(r.body)}}
}

func (r *rackAtom) Internal() {
	r.busy = false
	if r.out {
		r.pool.Free++
	}
	r.sheet.SetSeries(r.el.ID, "engine", float64(r.pool.Free))
}

func (r *rackAtom) External(_ time.Duration, msg engine.Message) {
	r.body = bodyID(msg, r.el.ID)
	if !r.out && r.pool.Free <= 0 {
		r.sheet.AddFlow(r.el.ID, 0, 0, 1)
		r.busy = false
		r.body = ""
		return
	}
	if r.out && r.pool.Free >= r.poolCap() {
		r.busy = false
		return
	}
	if !r.out {
		r.pool.Free--
	}
	r.busy = true
	r.show()
}

func (r *rackAtom) poolCap() int {
	if r.pool.Free < 0 {
		return 0
	}
	return r.pool.Free + 1
}

func (r *rackAtom) show() {
	x, y, z, yaw, _ := spanOf(r.el)
	slot := float64(r.pool.Free) * 0.5
	px, py, pz := place(x, y+0.4+slot, z, yaw, 0, 0.4)
	dx, dy, dz := place(x, y+0.78, z, yaw, 0, 0)
	if r.out {
		r.sheet.Begin(r.body, "box", "pick", specOf(px, py, pz, dx, dy, dz, yaw, 0.5, 0.8))
		return
	}
	r.sheet.Begin(r.body, "box", "store", specOf(dx, dy, dz, px, py, pz, yaw, 0.5, 0.8))
}
