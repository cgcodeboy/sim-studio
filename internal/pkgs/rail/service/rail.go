package service

import (
	"strconv"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
)

func PlaceRail(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TrainSource", "TrainMoveTo", "TrainDispose":
		return true, w.Add(el.ID, newTrain(w, el))
	default:
		return false, nil
	}
}

type train struct {
	sheet domain.Board
	el    domain.Block
	mode  string
	gate  *domain.Turnout
	body  string
	born  int
	left  int
	busy  bool
	hold  bool
	delay time.Duration
}

func newTrain(w domain.Repository, el domain.Block) *train {
	_, _, _, _, length := spanOf(el)
	mode := canon(el.Type)
	delay := time.Duration(length / 15 * float64(time.Second))
	if mode == "TrainSource" {
		delay = Every(Text(el.Props["rate"]), Number(el.Props["rate"], 1))
		if Text(el.Props["rate"]) == "" && Number(el.Props["rate"], 0) == 0 {
			delay = 3 * time.Second
		}
	}
	if mode == "TrainDispose" {
		delay = 300 * time.Millisecond
	}
	return &train{
		sheet: w.Board(), el: el, mode: mode, delay: delay,
		left: Limit(el.Props["limit"]), gate: switchOf(w, el),
	}
}

func switchOf(w domain.Repository, el domain.Block) *domain.Turnout {
	name := Text(el.Props["switch"])
	if name == "" {
		name = "switch"
	}
	return w.EnsureSwitch(name)
}

func (t *train) Advance() (time.Duration, bool) {
	if t.mode == "TrainSource" {
		if t.left == 0 {
			return 0, false
		}
		return scaled(t.sheet, t.delay), true
	}
	if !t.busy {
		return 0, false
	}
	if t.hold {
		return scaled(t.sheet, 200*time.Millisecond), true
	}
	return scaled(t.sheet, t.delay), true
}

func (t *train) Output() []engine.Message {
	if t.mode == "TrainDispose" || t.hold {
		return nil
	}
	if t.mode == "TrainSource" {
		t.born++
		t.body = t.el.ID + "#" + strconv.Itoa(t.born)
	}
	if t.body == "" {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(t.body)}}
}

func (t *train) Internal() {
	if t.mode == "TrainSource" {
		if t.left > 0 {
			t.left--
		}
		t.sheet.AddFlow(t.el.ID, 1, 0, 0)
		return
	}
	if t.hold {
		if t.gate.Busy {
			return
		}
		t.hold = false
		t.gate.Busy = true
		t.show()
		return
	}
	if t.mode == "TrainMoveTo" {
		t.gate.Busy = false
	}
	if t.mode == "TrainDispose" && t.body != "" {
		t.sheet.Hide(t.body)
		t.sheet.AddFlow(t.el.ID, 0, 1, 0)
	}
	t.busy = false
	t.body = ""
}

func (t *train) External(_ time.Duration, msg engine.Message) {
	if t.mode == "TrainSource" {
		return
	}
	t.body = bodyID(msg, t.el.ID)
	t.busy = true
	if t.mode == "TrainMoveTo" && t.gate.Busy {
		t.hold = true
		t.mark()
		return
	}
	if t.mode == "TrainMoveTo" {
		t.gate.Busy = true
	}
	t.show()
}

func (t *train) mark() {
	x, _, z, yaw, _ := spanOf(t.el)
	t.sheet.Begin(t.body, "train", "wait", specOf(x, 0.4, z, x, 0.4, z, yaw, 0, 0.2))
}

func (t *train) show() {
	x, _, z, yaw, length := spanOf(t.el)
	ax, az := place(x, z, yaw, -length/2)
	bx, bz := place(x, z, yaw, length/2)
	seconds := t.delay.Seconds()
	speed := 15.0
	action := "ride"
	if t.mode != "TrainMoveTo" {
		ax, az, bx, bz = x, z, x, z
		speed = 0
		action = "leave"
	}
	t.sheet.Begin(t.body, "train", action, specOf(ax, 0.4, az, bx, 0.4, bz, yaw, speed, seconds))
}
