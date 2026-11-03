package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
)

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
