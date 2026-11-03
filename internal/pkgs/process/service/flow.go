package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

func PlaceFlow(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Source":
		return true, w.Add(el.ID, newSource(w.Board(), el))
	case "Sink":
		return true, w.Add(el.ID, &sink{sheet: w.Board(), id: el.ID})
	case "Queue":
		return true, w.Add(el.ID, newQueue(w.Board(), el))
	case "Delay", "Hold", "Wait", "Enter", "Exit", "PlainTransfer", "MoveTo":
		return true, w.Add(el.ID, newDelay(w.Board(), el))
	case "SelectOutput", "SelectOutput5", "SelectOutputIn", "SelectOutputOut":
		return true, w.Add(el.ID, newPick(w.Board(), el))
	default:
		return false, nil
	}
}

type source struct {
	interval time.Duration
	left     int
	sheet    domain.Board
	id       string
}

func newSource(sheet domain.Board, el domain.Block) *source {
	text := Text(el.Props["rate"])
	return &source{
		interval: Every(text, Number(el.Props["rate"], 1)),
		left:     Limit(el.Props["limit"]),
		sheet:    sheet,
		id:       el.ID,
	}
}

func (s *source) Advance() (time.Duration, bool) {
	if s.left == 0 {
		return 0, false
	}
	return s.interval, true
}

func (s *source) Output() []engine.Message {
	return []engine.Message{{Port: "out", Payload: []byte{1}}}
}

func (s *source) Internal() {
	if s.left > 0 {
		s.left--
	}
	s.sheet.AddFlow(s.id, 1, 0, 0)
}

func (s *source) External(time.Duration, engine.Message) {}

type queue struct {
	cap   int
	n     int
	sheet domain.Board
	id    string
}

func newQueue(sheet domain.Board, el domain.Block) *queue {
	q := &queue{cap: int(Number(el.Props["capacity"], 15)), sheet: sheet, id: el.ID}
	sheet.SetSeries(el.ID, "engine", 0)
	return q
}

func (q *queue) Advance() (time.Duration, bool) {
	if q.n > 0 {
		return 0, true
	}
	return 0, false
}

func (q *queue) Output() []engine.Message {
	if q.n <= 0 {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte{1}}}
}

func (q *queue) Internal() {
	if q.n > 0 {
		q.n--
	}
	q.sheet.SetQueued(q.id, q.n)
}

func (q *queue) External(time.Duration, engine.Message) {
	if q.cap > 0 && q.n >= q.cap {
		q.sheet.AddFlow(q.id, 0, 0, 1)
		return
	}
	q.n++
	q.sheet.SetQueued(q.id, q.n)
}

type sink struct {
	sheet domain.Board
	id    string
}

func (s *sink) Advance() (time.Duration, bool) { return 0, false }
func (s *sink) Output() []engine.Message       { return nil }
func (s *sink) Internal()                      {}
func (s *sink) External(time.Duration, engine.Message) {
	s.sheet.AddFlow(s.id, 0, 1, 0)
}

type delayAtom struct {
	delay time.Duration
	busy  bool
	open  bool
	sheet domain.Board
	id    string
	x, y  float64
}

func newDelay(sheet domain.Board, el domain.Block) *delayAtom {
	d := Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 1))
	switch canon(el.Type) {
	case "Enter", "Exit":
		d = 0
	case "PlainTransfer", "MoveTo":
		d = Travel(LengthOf(el), SpeedOf(el))
	case "Hold":
		d = 0
	}
	open := true
	if canon(el.Type) == "Hold" {
		v, _ := Eval(Text(el.Props["condition"]), Env{})
		open = v > 0
	}
	return &delayAtom{delay: d, open: open, sheet: sheet, id: el.ID, x: el.X, y: el.Y}
}

func (h *delayAtom) Advance() (time.Duration, bool) {
	if h.busy {
		return h.delay, true
	}
	return 0, false
}

func (h *delayAtom) Output() []engine.Message {
	return []engine.Message{{Port: "out", Payload: []byte{1}}}
}

func (h *delayAtom) Internal() {
	h.busy = false
	h.sheet.Leave(h.id)
}

func (h *delayAtom) External(time.Duration, engine.Message) {
	if !h.open || h.busy {
		h.sheet.AddFlow(h.id, 0, 0, 1)
		return
	}
	h.busy = true
	h.sheet.Enter(h.id, h.x, h.y)
}

type pick struct {
	sheet domain.Board
	id    string
	text  string
	ready bool
	port  string
}

func newPick(sheet domain.Board, el domain.Block) *pick {
	return &pick{sheet: sheet, id: el.ID, text: Text(el.Props["condition"]), port: "out"}
}

func (p *pick) Advance() (time.Duration, bool) {
	if p.ready {
		return 0, true
	}
	return 0, false
}

func (p *pick) Output() []engine.Message {
	return []engine.Message{{Port: p.port, Payload: []byte{1}}}
}

func (p *pick) Internal() {
	p.ready = false
	p.sheet.AddFlow(p.id, 0, 1, 0)
}

func (p *pick) External(time.Duration, engine.Message) {
	p.ready = true
	v, err := Eval(p.text, Env{})
	p.port = "out"
	if err == nil && v > 0 {
		p.port = "out1"
	}
}
