package repo

import (
	"context"
	"sync"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

// Memory 把会话留在进程里。重启之后这些记录不再存在。
type Memory struct {
	mu       sync.Mutex
	sessions map[string]domain.Session
	facts    map[string][]domain.Fact
}

func New() *Memory {
	return &Memory{sessions: map[string]domain.Session{}, facts: map[string][]domain.Fact{}}
}

func (r *Memory) Load(_ context.Context, id string) (domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[id]
	if !ok {
		return domain.Session{ID: id}, nil
	}
	return cloneSession(session), nil
}

func (r *Memory) Save(_ context.Context, session domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = cloneSession(session)
	return nil
}

func (r *Memory) Facts(_ context.Context, project string) ([]domain.Fact, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.Fact(nil), r.facts[project]...), nil
}

func (r *Memory) ReplaceFacts(_ context.Context, project string, facts []domain.Fact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.facts[project] = append([]domain.Fact(nil), facts...)
	return nil
}

func cloneSession(session domain.Session) domain.Session {
	session.Turns = append([]domain.Turn(nil), session.Turns...)
	return session
}
