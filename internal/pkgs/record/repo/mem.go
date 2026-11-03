package repo

import (
	"context"
	"sync"
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

type Mem struct {
	mu      sync.Mutex
	records map[string]domain.Record
}

func NewMem() *Mem {
	return &Mem{records: map[string]domain.Record{}}
}

func (m *Mem) Insert(_ context.Context, record domain.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.records {
		if existing.RequestID == record.RequestID {
			return domain.ErrDuplicate
		}
	}
	if record.Occupied() {
		for _, existing := range m.records {
			if existing.Occupied() {
				return domain.ErrActive
			}
		}
	}
	m.records[record.ID] = record
	return nil
}

func (m *Mem) FindByRequestID(_ context.Context, requestID string) (domain.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, record := range m.records {
		if record.RequestID == requestID {
			return record, nil
		}
	}
	return domain.Record{}, domain.ErrNotFound
}

func (m *Mem) Read(_ context.Context, id string) (domain.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.records[id]
	if !ok {
		return domain.Record{}, domain.ErrNotFound
	}
	return record, nil
}

func (m *Mem) MatchRevision(_ context.Context, id string, revision int64, next domain.Record) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.records[id]
	if !ok {
		return 0, domain.ErrNotFound
	}
	if record.Revision != revision {
		return 0, nil
	}
	next.Revision = revision + 1
	next.UpdatedAt = time.Now().UTC()
	if next.Status == "finished" || next.Status == "interrupted" {
		next.Active = nil
	}
	m.records[id] = next
	return 1, nil
}

func (m *Mem) CountActive(_ context.Context, projectID string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, record := range m.records {
		if record.ProjectID == projectID && (record.Status == "running" || record.Status == "paused") {
			n++
		}
	}
	return n, nil
}

func (m *Mem) DeleteByProject(_ context.Context, projectID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, record := range m.records {
		if record.ProjectID == projectID && record.Occupied() {
			return domain.ErrActive
		}
	}
	for id, record := range m.records {
		if record.ProjectID == projectID {
			delete(m.records, id)
		}
	}
	return nil
}

func (m *Mem) InterruptActive(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, record := range m.records {
		if record.Status == "running" || record.Status == "paused" {
			record.Status = "interrupted"
			record.Active = nil
			record.UpdatedAt = time.Now().UTC()
			m.records[id] = record
		}
	}
	return nil
}

func (m *Mem) DeleteOrphans(_ context.Context, projectIDs []string) error {
	keep := map[string]struct{}{}
	for _, id := range projectIDs {
		keep[id] = struct{}{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, record := range m.records {
		if _, ok := keep[record.ProjectID]; !ok {
			delete(m.records, id)
		}
	}
	return nil
}
