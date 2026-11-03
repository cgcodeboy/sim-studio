package api

import (
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/library"
	"github.com/cgcodeboy/sim-studio/api/handlers/project"
	"github.com/cgcodeboy/sim-studio/api/handlers/simulation"
)

func (s *server) routes() http.Handler {
	catalog := library.New(s.app.Library)
	projects := project.New(s.app.Model, s.app.Record)
	simulations := simulation.New(s.app)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/library", catalog.List)
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
