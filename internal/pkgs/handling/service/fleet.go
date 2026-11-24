package service

import (
	"math"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

const fleetLimit = 10

func PlaceFleet(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TransporterFleet", "TransporterControl":
		if int(Number(el.Props["capacity"], 1)) > fleetLimit {
			return true, domain.ErrInvalid
		}
		pool := w.EnsurePool("transporter", int(Number(el.Props["capacity"], 1)))
		pool.X, pool.Y, pool.Z, pool.Yaw = el.Px, el.Py, el.Pz, el.Yaw
		w.Board().SetSeries(el.ID, "engine", float64(pool.Free))
		return true, nil
	case "SeizeTransporter", "ReleaseTransporter", "MoveByTransporter":
		return true, w.Add(el.ID, newHaul(w, el))
	default:
		return false, nil
	}
}

type haul struct {
	sheet domain.Board
	el    domain.Block
	pool  *domain.Units
	mode  string
	state string
	body  string
	queue []string
	left  time.Duration
}

func newHaul(w domain.Repository, el domain.Block) *haul {
	return &haul{
		sheet: w.Board(), el: el, mode: canon(el.Type),
		pool: w.EnsurePool("transporter", 1),
	}
}

func (h *haul) Advance() (time.Duration, bool) {
	if h.state == "" {
		return 0, false
	}
	return h.left, true
}

func (h *haul) Output() []engine.Message {
	if h.body == "" || h.state != h.hand() {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(h.body)}}
}

func (h *haul) hand() string {
	switch h.mode {
	case "ReleaseTransporter":
		return "lower"
	case "MoveByTransporter":
		return "carry"
	default:
		return "lift"
	}
}

func (h *haul) Internal() {
	if h.state == "wait" {
		if h.pool.Free <= 0 {
			return
		}
		h.pool.Free--
		h.state = "drive"
		h.arm()
		return
	}
	switch h.state {
	case "drive":
		h.state = "lift"
	case "lift":
		h.state = ""
	case "carry":
		h.state = ""
	case "lower":
		h.state = "back"
	case "back":
		h.state = ""
		h.sheet.Hide("fork")
	}
	if h.state == "" {
		h.body = ""
		if len(h.queue) > 0 {
			h.body = h.queue[0]
			h.queue = h.queue[1:]
			h.begin()
		}
		return
	}
	h.arm()
}

func (h *haul) External(_ time.Duration, msg engine.Message) {
	id := bodyID(msg, h.el.ID)
	if h.state != "" {
		h.queue = append(h.queue, id)
		return
	}
	h.body = id
	h.begin()
}

func (h *haul) begin() {
	if h.mode == "SeizeTransporter" && h.pool.Free <= 0 {
		h.state = "wait"
		h.left = scaled(h.sheet, 200*time.Millisecond)
		h.holdBox(0.78)
		return
	}
	if h.mode == "SeizeTransporter" {
		h.pool.Free--
		h.state = "drive"
	} else if h.mode == "MoveByTransporter" {
		h.state = "carry"
	} else {
		h.state = "lower"
	}
	h.arm()
}

func (h *haul) arm() {
	home := [3]float64{h.pool.X, h.pool.Y, h.pool.Z}
	door := h.at(false)
	exit := h.at(true)
	switch h.state {
	case "drive":
		h.left = h.travel(home, door, 2)
		h.fork(home, door, 2)
		h.holdBox(0.78)
	case "lift":
		h.left = scaled(h.sheet, 1200*time.Millisecond)
		h.sheet.Begin(h.body, "box|fork", "lift", specOf(door[0], h.el.Py+0.78, door[2], door[0], h.el.Py+0.8, door[2], h.el.Yaw, 0, 1.2))
	case "carry":
		h.left = h.travel(door, exit, 1.4)
		h.fork(door, exit, 1.4)
		h.sheet.Begin(h.body, "box|fork", "carry", specOf(door[0], h.el.Py+0.8, door[2], exit[0], h.el.Py+0.8, exit[2], h.el.Yaw, 1.4, h.nominal(door, exit, 1.4)))
	case "lower":
		h.left = scaled(h.sheet, 1200*time.Millisecond)
		h.sheet.Begin(h.body, "box|fork", "drop", specOf(door[0], h.el.Py+0.8, door[2], door[0], h.el.Py+0.78, door[2], h.el.Yaw, 0, 1.2))
	case "back":
		h.left = h.travel(door, home, 2)
		h.fork(door, home, 2)
		h.pool.Free++
	default:
		h.left = scaled(h.sheet, 200*time.Millisecond)
	}
}

func (h *haul) holdBox(height float64) {
	dx, dz := h.at(false)[0], h.at(false)[2]
	h.sheet.Begin(h.body, "box", "wait", specOf(dx, h.el.Py+height, dz, dx, h.el.Py+height, dz, h.el.Yaw, 0, 0.2))
}

func (h *haul) at(front bool) [3]float64 {
	x, y, z, yaw, length := spanOf(h.el)
	along := -length / 2
	if front {
		along = length / 2
	}
	px, py, pz := place(x, y, z, yaw, along, 0)
	return [3]float64{px, py, pz}
}

func (h *haul) travel(from, to [3]float64, speed float64) time.Duration {
	return scaled(h.sheet, time.Duration(h.nominal(from, to, speed)*float64(time.Second)))
}

func (h *haul) nominal(from, to [3]float64, speed float64) float64 {
	dist := math.Sqrt((to[0]-from[0])*(to[0]-from[0]) + (to[1]-from[1])*(to[1]-from[1]) + (to[2]-from[2])*(to[2]-from[2]))
	if speed <= 0 {
		speed = 1.4
	}
	seconds := dist / speed
	if seconds < 0.35 {
		seconds = 0.35
	}
	return seconds
}

func (h *haul) fork(from, to [3]float64, speed float64) {
	seconds := h.nominal(from, to, speed)
	h.sheet.Begin("fork", "fork", h.state, specOf(from[0], from[1], from[2], to[0], to[1], to[2], h.el.Yaw, speed, seconds))
}
