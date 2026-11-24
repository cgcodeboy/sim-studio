package service

import (
	"math"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
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
	if length < 1 {
		length = 12
	}
	return
}

func place(x, z, yaw, along float64) (float64, float64) {
	return x + along*math.Cos(yaw), z + along*math.Sin(yaw)
}

func specOf(x, y, z, tx, ty, tz, heading, speed, seconds float64) [9]float64 {
	return [9]float64{x, y, z, tx, ty, tz, heading, speed, seconds}
}
