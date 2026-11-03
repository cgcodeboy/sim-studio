// Package simulation 是仿真这门业务。它组合模型、元件库、各求解领域和运行记录。
package simulation

import (
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	"github.com/cgcodeboy/sim-studio/internal/app"
)

type Handler struct {
	app *app.App
}

func New(application *app.App) *Handler {
	return &Handler{app: application}
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RequestID    string `json:"requestId"`
		ProjectID    string `json:"projectId"`
		ExperimentID string `json:"experimentId"`
		Speed        string `json:"speed"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	project, err := h.app.Model.Read(r.Context(), body.ProjectID)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	items, err := h.app.Library.List(r.Context())
	if err != nil {
		respond.Fail(w, err)
		return
	}
	session, err := assemble(h.app, project, body.ExperimentID, items)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	result, err := h.app.Record.Start(r.Context(), body.RequestID, body.ProjectID, body.ExperimentID, body.Speed, session)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Sample(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.Record.Sample(r.PathValue("id"))
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) { h.act(w, r, "pause") }

func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	h.tune(r)
	h.act(w, r, "resume")
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) { h.act(w, r, "stop") }

func (h *Handler) Speed(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RequestID string `json:"requestId"`
		Speed     string `json:"speed"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	result, err := h.app.Record.Speed(r.Context(), body.RequestID, r.PathValue("id"), body.Speed)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) tune(r *http.Request) {
	id := r.PathValue("id")
	params := h.app.Record.Params(id)
	projectID := h.app.Record.ProjectID(id)
	if params == nil || projectID == "" {
		return
	}
	project, err := h.app.Model.Read(r.Context(), projectID)
	if err != nil {
		return
	}
	for _, agent := range project.Agents {
		for _, el := range agent.Elements {
			switch canon(el.Type) {
			case "PMLSettings":
				h.app.Process.Open(nil, params).Tune(flowBlock(el))
			case "MHLSettings":
				h.app.Handling.Open(nil, params).Tune(haulBlock(el))
			case "PedSettings":
				h.app.Pedestrian.Open(nil, params).Tune(pedBlock(el))
			}
		}
	}
}

func (h *Handler) act(w http.ResponseWriter, r *http.Request, op string) {
	var body struct {
		RequestID string `json:"requestId"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	var err error
	var result any
	switch op {
	case "pause":
		result, err = h.app.Record.Pause(r.Context(), body.RequestID, r.PathValue("id"))
	case "resume":
		result, err = h.app.Record.Resume(r.Context(), body.RequestID, r.PathValue("id"))
	default:
		result, err = h.app.Record.Stop(r.Context(), body.RequestID, r.PathValue("id"))
	}
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, result)
}
