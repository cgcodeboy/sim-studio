package api

import (
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/dialogue"
	"github.com/cgcodeboy/sim-studio/api/handlers/library"
	"github.com/cgcodeboy/sim-studio/api/handlers/project"
	"github.com/cgcodeboy/sim-studio/api/handlers/simulation"
)

func (s *server) routes() http.Handler {
	catalog := library.New(s.app.Library)
	turns := dialogue.New(s.app.Library, s.app.Conversation)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/library", catalog.List)
	mux.HandleFunc("POST /api/dialogue/turns", turns.Turn)
	mux.HandleFunc("GET /api/dialogue/sessions", turns.ListSessions)
	mux.HandleFunc("POST /api/dialogue/sessions", turns.CreateSession)
	mux.HandleFunc("GET /api/dialogue/sessions/{id}", turns.ReadSession)
	mux.HandleFunc("PUT /api/dialogue/sessions/{id}", turns.RenameSession)
	mux.HandleFunc("DELETE /api/dialogue/sessions/{id}", turns.DeleteSession)
	if s.app.Model == nil || s.app.Record == nil {
		return mux
	}
	projects := project.New(s.app.Model, s.app.Record)
	simulations := simulation.New(s.app)
	mux.HandleFunc("GET /api/projects", projects.List)
	mux.HandleFunc("GET /api/projects/{id}", projects.Read)
	mux.HandleFunc("PUT /api/projects/{id}", projects.Put)
	mux.HandleFunc("DELETE /api/projects/{id}", projects.Delete)
	mux.HandleFunc("POST /api/records", simulations.Start)
	mux.HandleFunc("GET /api/records/{id}", simulations.Sample)
	mux.HandleFunc("POST /api/records/{id}/pause", simulations.Pause)
	mux.HandleFunc("POST /api/records/{id}/resume", simulations.Resume)
	mux.HandleFunc("POST /api/records/{id}/stop", simulations.Stop)
	mux.HandleFunc("POST /api/records/{id}/speed", simulations.Speed)
	return mux
}
