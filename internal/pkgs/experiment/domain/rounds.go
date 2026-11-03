package domain

// RoundLimit 是顺序坐标搜索的轮数上限。
const RoundLimit = 30

// Rounds 按实验类型产生顺序重跑的参数值。不并行。
func Rounds(kind string, min, max, step float64) []float64 {
	if step <= 0 {
		step = 1
	}
	if kind == "optimization" || kind == "rl" {
		return coordinate(min, max)
	}
	if kind != "variation" {
		return nil
	}
	var out []float64
	for x := min; x <= max && len(out) < RoundLimit; x += step {
		out = append(out, x)
	}
	if len(out) == 0 {
		out = append(out, min)
	}
	return out
}

func coordinate(min, max float64) []float64 {
	out := make([]float64, 0, RoundLimit)
	span := max - min
	if span == 0 {
		span = 1
	}
	for i := 0; i < RoundLimit; i++ {
		out = append(out, min+span*float64(i)/float64(RoundLimit))
	}
	return out
}
