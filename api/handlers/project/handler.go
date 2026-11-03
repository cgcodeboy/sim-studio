// Package project 是工程这个业务的 handler。保存和删除时查看运行记录，模型本身不查运行。
package project

import (
	"errors"
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	record "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

type Handler struct {
	documents model.Service
	records   record.Service
}

func New(documents model.Service, records record.Service) *Handler {
	return &Handler{documents: documents, records: records}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.documents.List(r.Context())
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	p, err := h.documents.Read(r.Context(), r.PathValue("id"))
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision int64         `json:"revision"`
		Project  model.Project `json:"project"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	body.Project.ID = r.PathValue("id")
	if body.Revision != 0 {
		n, err := h.records.CountActive(r.Context(), body.Project.ID)
		if err != nil {
			respond.Fail(w, err)
			return
		}
		if n > 0 {
			respond.Fail(w, &model.Error{Code: model.CodeRejected})
			return
		}
	}
	p, err := h.documents.Put(r.Context(), body.Revision, body.Project)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n, err := h.records.CountActive(r.Context(), id)
	if err != nil {
		respond.Fail(w, err)
		return
	}
	if n > 0 {
		respond.Fail(w, &model.Error{Code: model.CodeRejected})
		return
	}
	if err := h.documents.Delete(r.Context(), id); err != nil {
		respond.Fail(w, err)
		return
	}
	if err := h.records.DeleteByProject(r.Context(), id); err != nil {
		if errors.Is(err, record.ErrActive) {
			respond.Fail(w, &model.Error{Code: model.CodeRejected})
			return
		}
		respond.Fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
