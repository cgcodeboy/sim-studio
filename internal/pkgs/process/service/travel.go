package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	"math"
	"time"
)

// defaultSpeed 是每个模型分钟 1 个长度单位。
const defaultSpeed = 1

// Travel 用路径长度除以速度。速度缺省时用 defaultSpeed。
func Travel(length, speed float64) time.Duration {
	if speed <= 0 {
		speed = defaultSpeed
	}
	if length < 0 {
		length = 0
	}
	minutes := length / speed
	return time.Duration(minutes * float64(time.Minute))
}

// LengthOf 读取折线、length 属性，或元件宽高的直线距离。
func LengthOf(el domain.Block) float64 {
	if n := Number(el.Props["length"], 0); n > 0 {
		return n
	}
	if pts := pointsOf(el); len(pts) >= 2 {
		return Polyline(pts)
	}
	if el.W != 0 || el.H != 0 {
		return math.Hypot(el.W, el.H)
	}
	return 1
}

// SpeedOf 读取速度属性。
func SpeedOf(el domain.Block) float64 {
	speed := Number(el.Props["speed"], defaultSpeed)
	if speed <= 0 {
		return defaultSpeed
	}
	return speed
}

// Polyline 累加折线各段的长度。
func Polyline(pts [][2]float64) float64 {
	var total float64
	for i := 1; i < len(pts); i++ {
		dx := pts[i][0] - pts[i-1][0]
		dy := pts[i][1] - pts[i-1][1]
		total += math.Hypot(dx, dy)
	}
	return total
}

func pointsOf(el domain.Block) [][2]float64 {
	raw, _ := el.Props["points"].([]any)
	pts := make([][2]float64, 0, len(raw))
	for _, item := range raw {
		pair, ok := item.([]any)
		if !ok || len(pair) < 2 {
			continue
		}
		pts = append(pts, [2]float64{Number(pair[0], 0), Number(pair[1], 0)})
	}
	return pts
}
