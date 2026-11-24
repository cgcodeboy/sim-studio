package service

import (
	"math"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
)

func scaled(sheet domain.Board, d time.Duration) time.Duration {
	pace := sheet.Pace()
	if pace < 1 {
		pace = 1
	}
	out := time.Duration(float64(d) / pace)
	if d > 0 && out < time.Millisecond {
		return time.Millisecond
	}
	return out
}

func bodyID(msg engine.Message, fallback string) string {
	if len(msg.Payload) == 0 {
		return fallback
	}
	text := string(msg.Payload)
	if text == "" || text == "\x01" {
		return fallback
	}
	return text
}

func spanOf(el domain.Block) (x, y, z, yaw, length float64) {
	x, y, z, yaw = el.Px, el.Py, el.Pz, el.Yaw
	length = el.Len
	if length < 0.4 {
		length = 1.2
	}
	return
}

func place(x, y, z, yaw, along, side float64) (float64, float64, float64) {
	rx := along*math.Cos(yaw) - side*math.Sin(yaw)
	rz := along*math.Sin(yaw) + side*math.Cos(yaw)
	return x + rx, y, z + rz
}

func specOf(x, y, z, tx, ty, tz, heading, speed, seconds float64) [9]float64 {
	return [9]float64{x, y, z, tx, ty, tz, heading, speed, seconds}
}

func showStay(sheet domain.Board, id string, el domain.Block, action string, seconds float64) {
	x, y, z, yaw, length := spanOf(el)
	px, py, pz := place(x, y+0.78, z, yaw, 0, 0.55)
	if action == "leave" {
		px, py, pz = place(x, y+0.78, z, yaw, length/2, 0)
	}
	sheet.Begin(id, "box", action, specOf(px, py, pz, px, py, pz, yaw, 0, seconds))
}

func showRide(sheet domain.Board, id string, el domain.Block, seconds float64) {
	x, y, z, yaw, length := spanOf(el)
	ax, ay, az := place(x, y+0.78, z, yaw, -length/2, 0)
	bx, by, bz := place(x, y+0.78, z, yaw, length/2, 0)
	speed := 0.0
	if seconds > 0 {
		speed = length / seconds
	}
	sheet.Begin(id, "box", "ride", specOf(ax, ay, az, bx, by, bz, yaw, speed, seconds))
}
