// Package library 是元件库这个业务的 handler。
package library

import (
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	librarydom "github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
)

type Handler struct {
	catalog librarydom.Service
}

func New(catalog librarydom.Service) *Handler {
	return &Handler{catalog: catalog}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.catalog.List(r.Context())
	if err != nil {
		respond.Fail(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, items)
}
