package service

import (
	"context"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// Finish 把这条运行收成 finished 或 interrupted。不新开一条记录。
func (m *Records) Finish(ctx context.Context, cur *live, status string) {
	if cur == nil {
		return
	}
	cur.once.Do(func() {
		m.writeEnd(ctx, cur, status)
	})
}

func (m *Records) writeEnd(ctx context.Context, cur *live, status string) {
	sample := cur.session.Frame()
	var last domain.Record
	for attempt := 0; attempt <= retryOnce; attempt++ {
		record, err := m.records.Read(ctx, cur.record.ID)
		if err != nil {
			return
		}
		next := record
		if status == "interrupted" {
			_ = next.Advance("stop")
		} else if err := next.Advance("finish"); err != nil {
			m.finishLive(record, sample)
			return
		}
		next.Sample = &sample
		matched, err := m.records.MatchRevision(ctx, record.ID, record.Revision, next)
		if err != nil {
			return
		}
		if matched == 1 {
			last = next
			break
		}
	}
	if last.ID == "" {
		return
	}
	m.finishLive(last, sample)
}

func (m *Records) finishLive(record domain.Record, sample domain.Sample) {
	result := domain.Result{RecordID: record.ID, Status: record.Status, Sample: sample}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.remember[record.RequestID] = result
	m.remember[record.ID] = result
	if m.current != nil && m.current.record.ID == record.ID {
		m.current.record.Status = record.Status
		m.current = nil
	}
}
