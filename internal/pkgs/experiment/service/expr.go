package service

import (
	"strconv"
	"strings"
	"time"
)

// Env 是求值时能看见的量。Draw 返回 0 到 1，为空时用 0.5，结果可重复。
type Env struct {
	Now    float64
	Draw   func() float64
	Params map[string]float64
}

// Number 从属性里取出一个数。字符串里只认数字和小数点。
func Number(v any, fallback float64) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		return digits(n, fallback)
	default:
		return fallback
	}
}

// Text 读取字符串属性。
func Text(v any) string {
	s, _ := v.(string)
	return s
}

// Every 把「每秒 20 个」这类到达速率换成间隔。
func Every(text string, rate float64) time.Duration {
	if rate <= 0 {
		rate = 1
	}
	d := time.Duration(float64(unitOf(text)) / rate)
	if d < time.Millisecond {
		return time.Millisecond
	}
	return d
}

// Lasting 把「1 秒」这类停留换成时长。缺省按分钟。
func Lasting(text string, n float64) time.Duration {
	if n <= 0 {
		n = 1
	}
	return time.Duration(n * float64(unitOf(text)))
}

// Horizon 把实验停止时间换成模型时长。缺省 100 个时间单位。
func Horizon(unit string, stop any) time.Duration {
	return Lasting(unit, Number(stop, 100))
}

// Limit 读取到达上限。空和「无限」都不封顶。
func Limit(v any) int {
	text := Text(v)
	if text == "" || strings.Contains(text, "无限") {
		return -1
	}
	return int(Number(v, -1))
}

// Eval 计算短表达式。不执行导入、类和循环。
func Eval(text string, env Env) (float64, error) {
	text = strings.TrimSpace(text)
	switch text {
	case "", "true":
		return 1, nil
	case "false":
		return 0, nil
	case "time":
		return env.Now, nil
	}
	if v, ok := env.Params[text]; ok {
		return v, nil
	}
	if v, err := strconv.ParseFloat(text, 64); err == nil {
		return v, nil
	}
	if name, args, ok := callOf(text); ok {
		return dist(name, args, env)
	}
	if i := strings.IndexAny(text, "+-"); i > 0 {
		left, err := Eval(text[:i], env)
		if err != nil {
			return 0, err
		}
		right, err := Eval(text[i+1:], env)
		if err != nil {
			return 0, err
		}
		if text[i] == '-' {
			return left - right, nil
		}
		return left + right, nil
	}
	return digits(text, 0), nil
}

func dist(name string, args []float64, env Env) (float64, error) {
	draw := 0.5
	if env.Draw != nil {
		draw = env.Draw()
	}
	switch name {
	case "uniform":
		if len(args) < 2 {
			return 0, errExpr
		}
		return args[0] + (args[1]-args[0])*draw, nil
	case "exponential", "normal":
		if len(args) < 1 {
			return 0, errExpr
		}
		return args[0], nil
	case "triangular":
		if len(args) < 3 {
			return 0, errExpr
		}
		return args[1], nil
	default:
		return 0, errExpr
	}
}

func callOf(text string) (string, []float64, bool) {
	open := strings.IndexByte(text, '(')
	if open <= 0 || !strings.HasSuffix(text, ")") {
		return "", nil, false
	}
	name := strings.TrimSpace(text[:open])
	body := strings.TrimSuffix(text[open+1:], ")")
	var args []float64
	if strings.TrimSpace(body) == "" {
		return name, args, true
	}
	for _, part := range strings.Split(body, ",") {
		v, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return "", nil, false
		}
		args = append(args, v)
	}
	return name, args, true
}

func unitOf(text string) time.Duration {
	if strings.Contains(text, "秒") {
		return time.Second
	}
	if strings.Contains(text, "小时") {
		return time.Hour
	}
	return time.Minute
}

func digits(s string, fallback float64) float64 {
	var b []byte
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			b = append(b, byte(r))
		}
	}
	if len(b) == 0 {
		return fallback
	}
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return fallback
	}
	return f
}

var errExpr = errExprType{}

type errExprType struct{}

func (errExprType) Error() string { return "expr" }
