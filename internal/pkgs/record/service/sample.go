package service

import (
	"context"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// Sample 读取活动运行的内存帧。已经结束的运行才读记录里的最终样本。
func (m *Records) Sample(recordID string) (domain.Result, error) {
	m.mu.Lock()
	cur := m.current
	remembered, ok := m.remember[recordID]
	m.mu.Unlock()
	if cur != nil && cur.record.ID == recordID && cur.session != nil {
		return domain.Result{RecordID: recordID, Status: cur.record.Status, Sample: cur.session.Frame()}, nil
	}
	if ok {
		return remembered, nil
	}
	record, err := m.records.Read(context.Background(), recordID)
	if err != nil {
		return domain.Result{}, &domain.Error{Code: domain.CodeNotFound}
	}
	sample := domain.Sample{Source: record.Source, Movers: []domain.Mover{}}
	if record.Sample != nil {
		sample = *record.Sample
	}
	return domain.Result{RecordID: record.ID, Status: record.Status, Sample: sample}, nil
}
