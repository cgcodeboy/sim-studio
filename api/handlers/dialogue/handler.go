// Package dialogue 是对话这门业务。它在返回前完成计划、执行和检查。
package dialogue

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
	librarydom "github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
)

type Handler struct {
	catalog librarydom.Service
}

func New(catalog librarydom.Service) *Handler {
	return &Handler{catalog: catalog}
}

func (h *Handler) Turn(w http.ResponseWriter, r *http.Request) {
	var in Input
	if err := respond.Decode(w, r, &in); err != nil {
		return
	}
	respond.WriteJSON(w, http.StatusOK, h.decide(r.Context(), in))
}

func (h *Handler) decide(ctx context.Context, in Input) Task {
	key := os.Getenv("DEEPSEEK_API_KEY")
	if key == "" {
		return clarify("对话模型没有配置，工程没有改。")
	}
	task, ok := askSkills(ctx, key, h.names(ctx), in)
	if !ok {
		return clarify("模型没有给出可执行的结果，工程没有改。")
	}
	return withPage(task, in)
}

func withPage(task Task, in Input) Task {
	if task.Intent == "clarify" || task.Intent == "run" || task.Intent == "pause" || task.Intent == "stop" || knownAgent(in, task.Agent) {
		return task
	}
	task.Agent = pickAgent(in, in.Text)
	return finish(task)
}

func knownAgent(in Input, name string) bool {
	for _, item := range in.Agents {
		if strings.EqualFold(item.Name, name) {
			return true
		}
	}
	return false
}

func (h *Handler) names(ctx context.Context) map[string]bool {
	if h.catalog == nil {
		return nil
	}
	items, err := h.catalog.List(ctx)
	if err != nil {
		return nil
	}
	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[item.Name] = true
	}
	return out
}
