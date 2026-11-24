package service

import (
	"strconv"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
)

func PlaceFlow(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Source":
		return true, w.Add(el.ID, newSource(w.Board(), el))
	case "Sink":
		return true, w.Add(el.ID, &sink{sheet: w.Board(), id: el.ID, el: el})
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
	born     int
	token    string
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
	return scaled(s.sheet, s.interval), true
}

func (s *source) Output() []engine.Message {
	s.born++
	s.token = s.id + "#" + strconv.Itoa(s.born)
	return []engine.Message{{Port: "out", Payload: []byte(s.token)}}
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
	items []string
	ready time.Duration
	hold  time.Duration
	sheet domain.Board
	id    string
	el    domain.Block
}

func newQueue(sheet domain.Board, el domain.Block) *queue {
	q := &queue{
		cap: int(Number(el.Props["capacity"], 15)), hold: 900 * time.Millisecond,
		sheet: sheet, id: el.ID, el: el,
	}
	sheet.SetSeries(el.ID, "engine", 0)
	return q
}

func (q *queue) Advance() (time.Duration, bool) {
	if len(q.items) == 0 {
		return 0, false
	}
	left := q.ready - q.sheet.When()
	if left < 0 {
		left = 0
	}
	return left, true
}

func (q *queue) Output() []engine.Message {
	if len(q.items) == 0 {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(q.items[0])}}
}

func (q *queue) Internal() {
	if len(q.items) == 0 {
		return
	}
	q.items = q.items[1:]
	q.sheet.SetQueued(q.id, len(q.items))
	if len(q.items) > 0 {
		q.ready = q.sheet.When() + scaled(q.sheet, q.hold)
	}
	q.draw()
}

func (q *queue) External(_ time.Duration, msg engine.Message) {
	if q.cap > 0 && len(q.items) >= q.cap {
		q.sheet.AddFlow(q.id, 0, 0, 1)
		return
	}
	q.items = append(q.items, bodyID(msg, q.id))
	if len(q.items) == 1 {
		q.ready = q.sheet.When() + scaled(q.sheet, q.hold)
	}
	q.sheet.SetQueued(q.id, len(q.items))
	q.draw()
}

func (q *queue) draw() {
	x, y, z, yaw, length := spanOf(q.el)
	for i, id := range q.items {
		along := -length/2 - float64(i)*0.45
		px, py, pz := place(x, y+0.78, z, yaw, along, 0)
		q.sheet.Begin(id, "box", "wait", specOf(px, py, pz, px, py, pz, yaw, 0, q.hold.Seconds()))
	}
}

type sink struct {
	sheet domain.Board
	id    string
	el    domain.Block
	token string
}

func (s *sink) Advance() (time.Duration, bool) {
	if s.token == "" {
		return 0, false
	}
	return scaled(s.sheet, 300*time.Millisecond), true
}
func (s *sink) Output() []engine.Message { return nil }
func (s *sink) Internal() {
	if s.token != "" {
		s.sheet.Hide(s.token)
		s.sheet.AddFlow(s.id, 0, 1, 0)
		s.token = ""
	}
}
func (s *sink) External(_ time.Duration, msg engine.Message) {
	s.token = bodyID(msg, s.id)
	showStay(s.sheet, s.token, s.el, "leave", 0.3)
}

type delayAtom struct {
	delay time.Duration
	busy  bool
	open  bool
	ride  bool
	sheet domain.Board
	id    string
	body  string
	el    domain.Block
	x, y  float64
}

func newDelay(sheet domain.Board, el domain.Block) *delayAtom {
	d := Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 1))
	ride := false
	switch canon(el.Type) {
	case "Enter", "Exit":
		d = 0
	case "PlainTransfer", "MoveTo":
		_, _, _, _, length := spanOf(el)
		d = time.Duration(length / 1 * float64(time.Second))
		ride = true
	case "Hold":
		d = 0
	}
	open := true
	if canon(el.Type) == "Hold" {
		v, _ := Eval(Text(el.Props["condition"]), Env{})
		open = v > 0
	}
	return &delayAtom{delay: d, open: open, ride: ride, sheet: sheet, id: el.ID, el: el, x: el.X, y: el.Y}
}

func (h *delayAtom) Advance() (time.Duration, bool) {
	if h.busy {
		return scaled(h.sheet, h.delay), true
	}
	return 0, false
}

func (h *delayAtom) Output() []engine.Message {
	return []engine.Message{{Port: "out", Payload: []byte(h.body)}}
}

func (h *delayAtom) Internal() {
	h.busy = false
	h.sheet.Leave(h.id)
}

func (h *delayAtom) External(_ time.Duration, msg engine.Message) {
	if !h.open || h.busy {
		h.sheet.AddFlow(h.id, 0, 0, 1)
		return
	}
	h.busy = true
	h.body = bodyID(msg, h.id)
	h.sheet.Enter(h.id, h.x, h.y)
	if h.ride {
		showRide(h.sheet, h.body, h.el, h.delay.Seconds())
		return
	}
	showStay(h.sheet, h.body, h.el, "work", h.delay.Seconds())
}

type pick struct {
	sheet domain.Board
	id    string
	el    domain.Block
	text  string
	body  string
	ready bool
	port  string
}

func newPick(sheet domain.Board, el domain.Block) *pick {
	return &pick{sheet: sheet, id: el.ID, el: el, text: Text(el.Props["condition"]), port: "out"}
}

func (p *pick) Advance() (time.Duration, bool) {
	if p.ready {
		return scaled(p.sheet, 400*time.Millisecond), true
	}
	return 0, false
}

func (p *pick) Output() []engine.Message {
	return []engine.Message{{Port: p.port, Payload: []byte(p.body)}}
}

func (p *pick) Internal() {
	p.ready = false
	p.sheet.AddFlow(p.id, 0, 1, 0)
}

func (p *pick) External(_ time.Duration, msg engine.Message) {
	p.ready = true
	p.body = bodyID(msg, p.id)
	v, err := Eval(p.text, Env{})
	p.port = "out"
	side := 0.0
	if err == nil && v > 0 {
		p.port = "out1"
		side = 0.55
	}
	x, y, z, yaw, length := spanOf(p.el)
	ax, ay, az := place(x, y+0.78, z, yaw, 0, 0)
	bx, by, bz := place(x, y+0.78, z, yaw, length/2, side)
	p.sheet.Begin(p.body, "box", "leave", specOf(ax, ay, az, bx, by, bz, yaw, 1, 0.4))
}
