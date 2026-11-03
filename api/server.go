package api

import (
	"net/http"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	"github.com/cgcodeboy/sim-studio/internal/app"
)

type server struct {
	app *app.App
}

func Listen(application *app.App, addr string) error {
	srv := &server{app: application}
	return http.ListenAndServe(addr, srv.routes())
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	respond.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
