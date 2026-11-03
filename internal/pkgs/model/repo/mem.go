package repo

import (
	"context"
	"sync"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
)

type Mem struct {
	mu       sync.Mutex
	projects map[string]domain.Project
}

func NewMem() *Mem {
	return &Mem{projects: map[string]domain.Project{}}
}

func (m *Mem) List(_ context.Context) ([]domain.Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Summary, 0, len(m.projects))
	for _, p := range m.projects {
		out = append(out, domain.Summary{ID: p.ID, Name: p.Name})
	}
	return out, nil
}

func (m *Mem) IDs(_ context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.projects))
	for id := range m.projects {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *Mem) Read(_ context.Context, id string) (domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.projects[id]
	if !ok {
		return domain.Project{}, domain.ErrNotFound
	}
	domain.Normalize(&p)
	return p, nil
}

func (m *Mem) Insert(_ context.Context, p domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[p.ID]; ok {
		return domain.ErrDuplicate
	}
	m.projects[p.ID] = p
	return nil
}

func (m *Mem) Replace(_ context.Context, expected int64, p domain.Project) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.projects[p.ID]
	if !ok || cur.Revision != expected {
		return false, nil
	}
	m.projects[p.ID] = p
	return true, nil
}

func (m *Mem) Delete(_ context.Context, id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[id]; !ok {
		return false, nil
	}
	delete(m.projects, id)
	return true, nil
}
