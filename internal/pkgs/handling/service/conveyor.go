package service

import (
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

func PlaceConveyor(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Conveyor", "Convey", "ConveyorEnter":
		return true, w.Add(el.ID, newBelt(w.Board(), el))
	default:
		return false, nil
	}
}

type belt struct {
	sheet domain.Board
	id    string
	el    domain.Block
	slots []string
	step  time.Duration
	ready time.Duration
	seen  []string
}

func newBelt(sheet domain.Board, el domain.Block) *belt {
	_, _, _, _, length := spanOf(el)
	n := int(length / 0.55)
	if n < 1 {
		n = 1
	}
	return &belt{sheet: sheet, id: el.ID, el: el, slots: make([]string, n), step: 1100 * time.Millisecond}
}

func (b *belt) occupied() bool {
	for _, id := range b.slots {
		if id != "" {
			return true
		}
	}
	return false
}

func (b *belt) Advance() (time.Duration, bool) {
	if !b.occupied() {
		return 0, false
	}
	left := b.ready - b.sheet.When()
	if left < 0 {
		left = 0
	}
	return left, true
}

func (b *belt) Output() []engine.Message {
	if b.slots[0] == "" {
		return nil
	}
	return []engine.Message{{Port: "out", Payload: []byte(b.slots[0])}}
}

func (b *belt) Internal() {
	if len(b.slots) == 0 {
		return
	}
	prev := append([]string(nil), b.slots...)
	b.slots[0] = ""
	next := append([]string(nil), b.slots...)
	for i := 0; i < len(b.slots)-1; i++ {
		if b.slots[i] == "" && b.slots[i+1] != "" {
			next[i] = b.slots[i+1]
			next[i+1] = ""
		}
	}
	b.slots = next
	b.seen = prev
	if b.occupied() {
		b.ready = b.sheet.When() + scaled(b.sheet, b.step)
	}
	b.draw()
}

func (b *belt) External(_ time.Duration, msg engine.Message) {
	last := len(b.slots) - 1
	if b.slots[last] != "" {
		b.sheet.AddFlow(b.id, 0, 0, 1)
		return
	}
	b.slots[last] = bodyID(msg, b.id)
	if b.ready == 0 || b.ready < b.sheet.When() {
		b.ready = b.sheet.When() + scaled(b.sheet, b.step)
	}
	b.draw()
}

func (b *belt) draw() {
	x, y, z, yaw, length := spanOf(b.el)
	for i, id := range b.slots {
		if id == "" {
			continue
		}
		along := length/2 - (float64(i)+0.5)*0.55
		px, py, pz := place(x, y+0.78, z, yaw, along, 0)
		from := i
		if i < len(b.seen) {
			for j, prev := range b.seen {
				if prev == id {
					from = j
				}
			}
		}
		fx := length/2 - (float64(from)+0.5)*0.55
		ax, ay, az := place(x, y+0.78, z, yaw, fx, 0)
		b.sheet.Begin(id, "box", "ride", specOf(ax, ay, az, px, py, pz, yaw, 0.5, b.step.Seconds()))
	}
}
