package service

import (
	"strconv"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
)

func PlaceRoad(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "CarSource", "CarEnter", "CarMoveTo", "CarExit", "CarDispose":
		return true, w.Add(el.ID, newCar(w.Board(), el))
	default:
		return false, nil
	}
}

type carRun struct {
	sheet domain.Board
	el    domain.Block
	mode  string
	body  string
	born  int
	left  int
	busy  bool
	hold  bool
	delay time.Duration
}

func newCar(sheet domain.Board, el domain.Block) *carRun {
	_, _, _, _, length := spanOf(el)
	mode := canon(el.Type)
	delay := time.Duration(length / 10 * float64(time.Second))
	if mode == "CarSource" || mode == "CarEnter" {
		delay = Every(Text(el.Props["rate"]), Number(el.Props["rate"], 1))
		if Text(el.Props["rate"]) == "" && Number(el.Props["rate"], 0) == 0 {
			delay = 2 * time.Second
		}
	}
	if mode == "CarExit" || mode == "CarDispose" {
		delay = 300 * time.Millisecond
	}
	return &carRun{sheet: sheet, el: el, mode: mode, delay: delay, left: Limit(el.Props["limit"])}
}

func (c *carRun) Advance() (time.Duration, bool) {
	if c.mode == "CarSource" || c.mode == "CarEnter" {
		if c.left == 0 {
			return 0, false
		}
		return scaled(c.sheet, c.delay), true
	}
	if !c.busy {
		return 0, false
	}
	if c.hold {
		return scaled(c.sheet, 200*time.Millisecond), true
	}
	return scaled(c.sheet, c.delay), true
}

func (c *carRun) Output() []engine.Message {
	if c.mode == "CarExit" || c.mode == "CarDispose" || c.hold {
		return nil
	}
	if c.mode == "CarSource" || c.mode == "CarEnter" {
		c.born++
		c.body = c.el.ID + "#" + strconv.Itoa(c.born)
	}
	if c.body == "" {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(c.body)}}
}

func (c *carRun) Internal() {
	if c.mode == "CarSource" || c.mode == "CarEnter" {
		if c.left > 0 {
			c.left--
		}
		c.sheet.AddFlow(c.el.ID, 1, 0, 0)
		return
	}
	if c.hold {
		if c.blocked() {
			return
		}
		c.hold = false
		c.show()
		return
	}
	if (c.mode == "CarExit" || c.mode == "CarDispose") && c.body != "" {
		c.sheet.Hide(c.body)
		c.sheet.AddFlow(c.el.ID, 0, 1, 0)
	}
	c.busy = false
	c.body = ""
}

func (c *carRun) External(_ time.Duration, msg engine.Message) {
	if c.mode == "CarSource" || c.mode == "CarEnter" {
		return
	}
	c.body = bodyID(msg, c.el.ID)
	c.busy = true
	if c.mode == "CarMoveTo" && c.blocked() {
		c.hold = true
		c.mark(0)
		return
	}
	c.show()
}

func (c *carRun) blocked() bool {
	_, _, z, yaw, length := spanOf(c.el)
	x, _, pz := place(c.el.Px, 0, z, yaw, length/2)
	return tooClose(c.sheet, c.body, x, pz)
}

func (c *carRun) mark(seconds float64) {
	x, _, z, yaw, _ := spanOf(c.el)
	c.sheet.Begin(c.body, "car", "wait", specOf(x, 0, z, x, 0, z, yaw, 0, seconds))
}

func (c *carRun) show() {
	x, _, z, yaw, length := spanOf(c.el)
	ax, _, az := place(x, 0, z, yaw, -length/2)
	bx, _, bz := place(x, 0, z, yaw, length/2)
	seconds := c.delay.Seconds()
	speed := 10.0
	action := "drive"
	if c.mode != "CarMoveTo" {
		ax, az, bx, bz = x, z, x, z
		speed = 0
		action = "leave"
	}
	c.sheet.Begin(c.body, "car", action, specOf(ax, 0, az, bx, 0, bz, yaw, speed, seconds))
}
