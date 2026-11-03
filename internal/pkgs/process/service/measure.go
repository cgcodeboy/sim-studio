package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

func PlaceMeasure(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "TimeMeasureStart", "TimeMeasureEnd", "RestrictedAreaStart", "RestrictedAreaEnd":
		return true, w.Add(el.ID, newMeasure(w, el))
	default:
		return false, nil
	}
}

type measureAtom struct {
	kind  string
	cap   int
	held  int
	busy  bool
	seen  int
	sheet domain.Board
	id    string
	area  *domain.Gate
}

func newMeasure(w domain.Repository, el domain.Block) *measureAtom {
	key := Text(el.Props["area"])
	if key == "" {
		key = "area"
	}
	area := w.EnsureArea(key, int(Number(el.Props["capacity"], 1)))
	return &measureAtom{
		kind: canon(el.Type), cap: int(Number(el.Props["capacity"], 1)),
		sheet: w.Board(), id: el.ID, area: area,
	}
}

func (m *measureAtom) Advance() (time.Duration, bool) {
	if m.busy {
		return 0, true
	}
	return 0, false
}

func (m *measureAtom) Output() []engine.Message {
	return []engine.Message{{Port: "out", Payload: []byte{1}}}
}

func (m *measureAtom) Internal() {
	m.busy = false
	if m.kind == "RestrictedAreaEnd" && m.area.Used > 0 {
		m.area.Used--
	}
	m.sheet.SetSeries(m.id, "engine", float64(m.area.Used))
}

func (m *measureAtom) External(time.Duration, engine.Message) {
	if m.kind == "RestrictedAreaStart" {
		if m.area.Cap > 0 && m.area.Used >= m.area.Cap {
			m.sheet.SetSeries(m.id, "engine", float64(m.area.Used))
			return
		}
		m.area.Used++
	}
	if m.kind == "TimeMeasureStart" || m.kind == "TimeMeasureEnd" {
		m.seen++
		m.sheet.SetSeries(m.id, "engine", float64(m.seen))
	}
	m.busy = true
}
