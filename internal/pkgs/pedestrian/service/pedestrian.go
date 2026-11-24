package service

import (
	"strconv"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
)

func PlacePed(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "PedSource", "PedEnter":
		return true, w.Add(el.ID, newWalker(w.Board(), el, "source"))
	case "PedSink", "PedExit":
		return true, w.Add(el.ID, newWalker(w.Board(), el, "sink"))
	case "PedSelectOutput":
		return true, w.Add(el.ID, newWalker(w.Board(), el, "pick"))
	case "PedGroupAssemble", "PedGroupDisassemble":
		return true, w.Add(el.ID, newWalker(w.Board(), el, "group"))
	case "PedGoTo", "PedWait", "PedService":
		return true, w.Add(el.ID, newWalker(w.Board(), el, canon(el.Type)))
	default:
		return false, nil
	}
}

type walker struct {
	sheet domain.Board
	el    domain.Block
	mode  string
	body  string
	born  int
	left  int
	got   int
	need  int
	busy  bool
	wait  bool
	port  string
	delay time.Duration
}

func newWalker(sheet domain.Board, el domain.Block, mode string) *walker {
	delay := Lasting(Text(el.Props["delay"]), Number(el.Props["delay"], 1))
	if mode == "PedGoTo" || mode == "source" || mode == "sink" || mode == "pick" {
		_, _, _, _, length := spanOf(el)
		delay = time.Duration(length / 1.3 * float64(time.Second))
	}
	need := int(Number(el.Props["size"], Number(el.Props["count"], 2)))
	if need < 1 {
		need = 1
	}
	return &walker{
		sheet: sheet, el: el, mode: mode, left: Limit(el.Props["limit"]),
		need: need, delay: delay, port: "out",
	}
}

func (w *walker) Advance() (time.Duration, bool) {
	if w.mode == "source" {
		if w.left == 0 {
			return 0, false
		}
		return scaled(w.sheet, w.gap()), true
	}
	if !w.busy {
		return 0, false
	}
	if w.wait {
		return scaled(w.sheet, 200*time.Millisecond), true
	}
	return scaled(w.sheet, w.delay), true
}

func (w *walker) gap() time.Duration {
	text := Text(w.el.Props["rate"])
	if text == "" && Number(w.el.Props["rate"], 0) == 0 {
		return time.Second
	}
	return Every(text, Number(w.el.Props["rate"], 1))
}

func (w *walker) Output() []engine.Message {
	if w.mode == "sink" || w.wait {
		return nil
	}
	if w.mode == "source" {
		w.born++
		w.body = w.el.ID + "#" + strconv.Itoa(w.born)
	}
	if w.mode == "PedGroupDisassemble" {
		return w.copies()
	}
	if w.body == "" {
		return nil
	}
	return []engine.Message{{Port: w.port, Payload: []byte(w.body)}}
}

func (w *walker) copies() []engine.Message {
	msgs := make([]engine.Message, w.need)
	for i := range msgs {
		w.born++
		id := w.el.ID + "#" + strconv.Itoa(w.born)
		msgs[i] = engine.Message{Port: "out", Payload: []byte(id)}
	}
	return msgs
}

func (w *walker) Internal() {
	if w.mode == "source" {
		if w.left > 0 {
			w.left--
		}
		w.sheet.AddFlow(w.el.ID, 1, 0, 0)
		w.body = ""
		return
	}
	if w.wait {
		if w.ahead() {
			return
		}
		w.wait = false
		w.show()
		return
	}
	if w.mode == "sink" && w.body != "" {
		w.sheet.Hide(w.body)
		w.sheet.AddFlow(w.el.ID, 0, 1, 0)
	}
	w.busy = false
	w.body = ""
	w.got = 0
}

func (w *walker) External(_ time.Duration, msg engine.Message) {
	if w.mode == "source" {
		return
	}
	w.body = bodyID(msg, w.el.ID)
	w.busy = true
	if w.mode == "PedGroupAssemble" {
		w.got++
		if w.got < w.need {
			w.stand(0)
			w.busy = false
			return
		}
	}
	if w.mode == "pick" {
		w.choose()
	}
	if w.mode == "PedGoTo" && w.ahead() {
		w.wait = true
		w.stand(0.2)
		return
	}
	w.show()
}

func (w *walker) choose() {
	v, err := Eval(Text(w.el.Props["condition"]), Env{})
	w.port = "out"
	if err == nil && v > 0 {
		w.port = "out1"
	}
}

func (w *walker) ahead() bool {
	x, _, z, yaw, length := spanOf(w.el)
	px, _, pz := place(x, 0, z, yaw, length/2, 0)
	return blocked(w.sheet, w.body, px, pz)
}

func (w *walker) stand(seconds float64) {
	x, _, z, yaw, _ := spanOf(w.el)
	w.sheet.Begin(w.body, "person", "stand", specOf(x, 0, z, x, 0, z, yaw, 0, seconds))
}

func (w *walker) show() {
	x, _, z, yaw, length := spanOf(w.el)
	ax, _, az := place(x, 0, z, yaw, -length/2, 0)
	bx, _, bz := place(x, 0, z, yaw, length/2, 0)
	action := "walk"
	seconds := w.delay.Seconds()
	if w.mode != "PedGoTo" {
		action = "stand"
		ax, az, bx, bz = x, z, x, z
	}
	if w.mode == "sink" {
		seconds = 0.3
		w.delay = 300 * time.Millisecond
	}
	w.sheet.Begin(w.body, "person", action, specOf(ax, 0, az, bx, 0, bz, yaw, 1.3, seconds))
}
