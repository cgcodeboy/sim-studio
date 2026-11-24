package simulation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	record "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

type preview struct {
	id     string
	st     *studio
	cancel context.CancelFunc
	status string
}

func (h *Handler) Open(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Project      model.Project `json:"project"`
		ExperimentID string        `json:"experimentId"`
		Speed        string        `json:"speed"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	markLive(&body.Project, body.ExperimentID)
	items, err := h.app.Library.List(r.Context())
	if err != nil {
		respond.Fail(w, err)
		return
	}
	session, err := assemble(h.app, body.Project, body.ExperimentID, items)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	session.sheet.SetPace(paceOf(body.Speed))
	session.ticking = time.Now()
	run := h.start(session)
	respond.WriteJSON(w, http.StatusOK, resultOf(run))
}

func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	run := h.one(r.PathValue("id"))
	if run == nil {
		respond.Fail(w, &model.Error{Code: model.CodeNotFound})
		return
	}
	respond.WriteJSON(w, http.StatusOK, resultOf(run))
}

func (h *Handler) Hold(w http.ResponseWriter, r *http.Request) {
	h.play(w, r, "pause")
}

func (h *Handler) Continue(w http.ResponseWriter, r *http.Request) {
	h.play(w, r, "resume")
}

func (h *Handler) End(w http.ResponseWriter, r *http.Request) {
	h.play(w, r, "stop")
}

func (h *Handler) Pace(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Speed string `json:"speed"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	run := h.one(r.PathValue("id"))
	if run == nil {
		respond.Fail(w, &model.Error{Code: model.CodeNotFound})
		return
	}
	run.st.sheet.SetPace(paceOf(body.Speed))
	respond.WriteJSON(w, http.StatusOK, resultOf(run))
}

func (h *Handler) play(w http.ResponseWriter, r *http.Request, op string) {
	run := h.one(r.PathValue("id"))
	if run == nil {
		respond.Fail(w, &model.Error{Code: model.CodeNotFound})
		return
	}
	switch op {
	case "pause":
		run.st.HoldClock()
		_ = run.st.Pause()
		run.status = "paused"
	case "resume":
		run.st.ReleaseClock()
		_ = run.st.Resume()
		run.status = "running"
	default:
		_ = run.st.Stop()
		run.cancel()
		run.status = "interrupted"
	}
	respond.WriteJSON(w, http.StatusOK, resultOf(run))
}

func (h *Handler) start(session *studio) *preview {
	h.mu.Lock()
	for _, old := range h.runs {
		_ = old.st.Stop()
		old.cancel()
	}
	h.runs = map[string]*preview{}
	ctx, cancel := context.WithCancel(context.Background())
	run := &preview{id: "live-" + newID(), st: session, cancel: cancel, status: "running"}
	h.runs[run.id] = run
	h.mu.Unlock()
	if err := session.Start(ctx); err != nil {
		run.status = "interrupted"
		return run
	}
	go watch(run)
	return run
}

func watch(run *preview) {
	done := make(chan struct{})
	go func() {
		run.st.Wait()
		close(done)
	}()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			if run.status == "running" {
				run.status = "finished"
			}
			run.st.Catchup()
			return
		case <-tick.C:
			if run.status == "running" {
				run.st.Catchup()
			}
		}
	}
}

func (h *Handler) one(id string) *preview {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.runs[id]
}

func resultOf(run *preview) record.Result {
	return record.Result{RecordID: run.id, Status: run.status, Sample: run.st.Frame()}
}

func markLive(p *model.Project, experimentID string) {
	for i := range p.Experiments {
		if p.Experiments[i].ID != experimentID {
			continue
		}
		if p.Experiments[i].Props == nil {
			p.Experiments[i].Props = map[string]any{}
		}
		p.Experiments[i].Props["live"] = "1"
	}
}

func paceOf(name string) float64 {
	switch name {
	case "x10":
		return 10
	case "asap":
		return 60
	default:
		return 1
	}
}

func newID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
