package service

import (
	"strconv"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
)

func PlaceBatch(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "Split", "Combine", "Batch", "Unbatch", "Pickup", "Dropoff", "Assembler":
		return true, w.Add(el.ID, newBatch(w.Board(), el))
	default:
		return false, nil
	}
}

type batchAtom struct {
	kind  string
	need  int
	got   int
	delay time.Duration
	busy  bool
	sheet domain.Board
	id    string
	el    domain.Block
	body  string
	seq   int
}

func newBatch(sheet domain.Board, el domain.Block) *batchAtom {
	need := int(Number(el.Props["size"], Number(el.Props["count"], Number(el.Props["quantity"], 2))))
	if need < 1 {
		need = 1
	}
	return &batchAtom{
		kind: canon(el.Type), need: need, sheet: sheet, id: el.ID, el: el,
		delay: Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 0)),
	}
}

func (b *batchAtom) Advance() (time.Duration, bool) {
	if b.busy {
		return scaled(b.sheet, b.delay), true
	}
	return 0, false
}

func (b *batchAtom) Output() []engine.Message {
	n := 1
	if b.kind == "Split" || b.kind == "Unbatch" {
		n = b.need
	}
	msgs := make([]engine.Message, n)
	for i := range msgs {
		b.seq++
		id := b.body
		if n > 1 {
			id = b.id + "#" + strconv.Itoa(b.seq)
		}
		msgs[i] = engine.Message{Port: "out", Payload: []byte(id)}
	}
	return msgs
}

func (b *batchAtom) Internal() {
	b.busy = false
	b.got = 0
	b.sheet.AddFlow(b.id, 0, 1, 0)
}

func (b *batchAtom) External(_ time.Duration, msg engine.Message) {
	need := b.need
	if b.kind == "Split" || b.kind == "Dropoff" || b.kind == "Unbatch" {
		need = 1
	}
	b.got++
	b.body = bodyID(msg, b.id)
	b.sheet.SetSeries(b.id, "engine", float64(b.got))
	showStay(b.sheet, b.body, b.el, "wait", 0)
	if b.got >= need {
		b.busy = true
		showStay(b.sheet, b.body, b.el, "work", b.delay.Seconds())
	}
}
