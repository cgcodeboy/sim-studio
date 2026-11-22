package dialogue

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	convdom "github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	project := strings.TrimSpace(r.URL.Query().Get("project"))
	if project == "" {
		respond.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid"})
		return
	}
	if h.memory == nil {
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
		return
	}
	cards, err := h.memory.List(r.Context(), project)
	if err != nil {
		writeSessionErr(w, err)
		return
	}
	if cards == nil {
		cards = []convdom.Card{}
	}
	respond.WriteJSON(w, http.StatusOK, cards)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Project string `json:"project"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	if h.memory == nil {
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
		return
	}
	session, err := h.memory.Create(r.Context(), body.Project)
	if err != nil {
		writeSessionErr(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, session)
}

func (h *Handler) ReadSession(w http.ResponseWriter, r *http.Request) {
	if h.memory == nil {
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
		return
	}
	session, err := h.memory.Open(r.Context(), r.PathValue("id"))
	if err != nil {
		writeSessionErr(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, session)
}

func (h *Handler) RenameSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	if h.memory == nil {
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
		return
	}
	session, err := h.memory.Rename(r.Context(), r.PathValue("id"), body.Title)
	if err != nil {
		writeSessionErr(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, session)
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if h.memory == nil {
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
		return
	}
	if err := h.memory.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeSessionErr(w, err)
		return
	}
	respond.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func writeSessionErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, convdom.ErrNotFound):
		respond.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "notfound"})
	case errors.Is(err, convdom.ErrInvalid):
		respond.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid"})
	default:
		respond.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "unavailable"})
	}
}
