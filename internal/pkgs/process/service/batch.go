package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
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
}

func newBatch(sheet domain.Board, el domain.Block) *batchAtom {
	need := int(Number(el.Props["size"], Number(el.Props["count"], Number(el.Props["quantity"], 2))))
	if need < 1 {
		need = 1
	}
	return &batchAtom{
		kind: canon(el.Type), need: need, sheet: sheet, id: el.ID,
		delay: Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 0)),
	}
}

func (b *batchAtom) Advance() (time.Duration, bool) {
	if b.busy {
		return b.delay, true
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
		msgs[i] = engine.Message{Port: "out", Payload: []byte{1}}
	}
	return msgs
}

func (b *batchAtom) Internal() {
	b.busy = false
	b.got = 0
	b.sheet.AddFlow(b.id, 0, 1, 0)
}

func (b *batchAtom) External(time.Duration, engine.Message) {
	need := b.need
	if b.kind == "Split" || b.kind == "Dropoff" || b.kind == "Unbatch" {
		need = 1
	}
	b.got++
	b.sheet.SetSeries(b.id, "engine", float64(b.got))
	if b.got >= need {
		b.busy = true
	}
}
