package domain

import (
	"sync"
	"time"
)

// Series 是样本里的一条命名数值，自带来源。
type Series struct {
	Name   string  `json:"name" bson:"name"`
	Value  float64 `json:"value" bson:"value"`
	Source string  `json:"source" bson:"source"`
}

// Mover 是同一帧里的位置。
type Mover struct {
	ID      string  `json:"id" bson:"id"`
	X       float64 `json:"x" bson:"x"`
	Y       float64 `json:"y" bson:"y"`
	Z       float64 `json:"z" bson:"z"`
	Heading float64 `json:"heading" bson:"heading"`
	Busy    bool    `json:"busy" bson:"busy"`
}

// Camera 是开始时抄入的姿态，求解不移动它。
type Camera struct {
	X      float64 `json:"x" bson:"x"`
	Y      float64 `json:"y" bson:"y"`
	Z      float64 `json:"z" bson:"z"`
	Target string  `json:"target,omitempty" bson:"target,omitempty"`
}

// Sample 是一帧结果。存量序列的来源不会是 engine。
type Sample struct {
	Time      float64            `json:"time" bson:"time"`
	Arrived   float64            `json:"arrived" bson:"arrived"`
	Queued    float64            `json:"queued" bson:"queued"`
	InProcess float64            `json:"inProcess" bson:"inProcess"`
	Exited    float64            `json:"exited" bson:"exited"`
	Overflow  float64            `json:"overflow" bson:"overflow"`
	Source    string             `json:"source" bson:"source"`
	Counts    map[string]float64 `json:"counts,omitempty" bson:"counts,omitempty"`
	Series    []Series           `json:"series,omitempty" bson:"series,omitempty"`
	Movers    []Mover            `json:"movers" bson:"movers"`
	Camera    *Camera            `json:"camera,omitempty" bson:"camera,omitempty"`
	Reward    []float64          `json:"reward,omitempty" bson:"reward,omitempty"`
}

// Sheet 保存同一次运行的计数和位置。处理函数只通过它的方法改数。
type Sheet struct {
	mu       sync.Mutex
	arrived  float64
	queued   float64
	inProc   float64
	exited   float64
	overflow float64
	when     time.Duration
	unit     string
	series   map[string]Series
	counts   map[string]float64
	movers   map[string]Mover
	camera   *Camera
	reward   []float64
}

// NewSheet 准备一份空样本。
func NewSheet(unit string) *Sheet {
	return &Sheet{
		unit:   unit,
		series: map[string]Series{},
		counts: map[string]float64{},
		movers: map[string]Mover{},
	}
}

// SetSeries 写入一条带来源的数值。dynamics 不会被标成 engine。
func (s *Sheet) SetSeries(name, source string, value float64) {
	if source != "engine" && source != "dynamics" && source != "experiment" {
		source = "engine"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.series[name] = Series{Name: name, Value: value, Source: source}
	s.counts[name] = value
}

// AddFlow 累加到达、离开或溢出。
func (s *Sheet) AddFlow(id string, arrived, exited, overflow float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.arrived += arrived
	s.exited += exited
	s.overflow += overflow
	if id != "" {
		s.counts[id] += arrived + exited
		s.series[id] = Series{Name: id, Value: s.counts[id], Source: "engine"}
	}
}

// SetQueued 把排队人数设成当前队长。
func (s *Sheet) SetQueued(id string, n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queued = float64(n)
	s.counts[id] = float64(n)
	s.series[id] = Series{Name: id, Value: float64(n), Source: "engine"}
}

// Enter 记录一个正在处理的对象及其位置。
func (s *Sheet) Enter(id string, x, y float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inProc++
	s.counts[id] = 1
	s.movers[id] = Mover{ID: id, X: x, Y: y, Busy: true}
	s.series[id] = Series{Name: id, Value: 1, Source: "engine"}
}

// Leave 结束一个正在处理的对象。
func (s *Sheet) Leave(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inProc > 0 {
		s.inProc--
	}
	s.counts[id] = 0
	delete(s.movers, id)
}

// SetMover 更新外壳步进算出的位置，来源仍由序列自己保持。
func (s *Sheet) SetMover(id string, x, y, heading float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.movers[id] = Mover{ID: id, X: x, Y: y, Heading: heading, Busy: true}
}

// SetCamera 抄入摄像机姿态。
func (s *Sheet) SetCamera(cam Camera) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.camera = &cam
}

// SetReward 记下实验各轮的目标。
func (s *Sheet) SetReward(values []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reward = append([]float64(nil), values...)
}

// Spots 返回当前各对象的位置，供间距计算读取。
func (s *Sheet) Spots() []struct {
	ID string
	X  float64
	Y  float64
} {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]struct {
		ID string
		X  float64
		Y  float64
	}, 0, len(s.movers))
	for _, mover := range s.movers {
		out = append(out, struct {
			ID string
			X  float64
			Y  float64
		}{ID: mover.ID, X: mover.X, Y: mover.Y})
	}
	return out
}

// SetWhen 记下外壳已经推进的模型时间。
func (s *Sheet) SetWhen(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.when = d
}

// Frame 拼出一帧。有离散事件时主体来源是 engine，存量仍留在自己的序列里。
func (s *Sheet) Frame(now time.Duration) Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now > s.when {
		s.when = now
	}
	movers := make([]Mover, 0, len(s.movers))
	for _, mover := range s.movers {
		movers = append(movers, mover)
	}
	series := make([]Series, 0, len(s.series))
	source := ""
	for _, item := range s.series {
		if item.Source == "dynamics" {
			item.Source = "dynamics"
		}
		series = append(series, item)
		if item.Source == "engine" {
			source = "engine"
		} else if source == "" {
			source = item.Source
		}
	}
	if source == "" {
		source = "engine"
	}
	counts := map[string]float64{}
	for key, value := range s.counts {
		counts[key] = value
	}
	return Sample{
		Time: unitTime(s.when, s.unit), Arrived: s.arrived, Queued: s.queued,
		InProcess: s.inProc, Exited: s.exited, Overflow: s.overflow,
		Source: source, Counts: counts, Series: series, Movers: movers,
		Camera: s.camera, Reward: append([]float64(nil), s.reward...),
	}
}

func unitTime(d time.Duration, unit string) float64 {
	switch unit {
	case "秒":
		return d.Seconds()
	case "小时":
		return d.Hours()
	default:
		return d.Minutes()
	}
}
