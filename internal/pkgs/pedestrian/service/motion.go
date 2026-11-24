package service

import (
	"math"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
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
		length = 2
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

func blocked(sheet domain.Board, id string, x, z float64) bool {
	for _, spot := range sheet.Spots() {
		if spot.ID == id {
			continue
		}
		if math.Hypot(spot.X-x, spot.Z-z) < 0.6 {
			return true
		}
	}
	return false
}
