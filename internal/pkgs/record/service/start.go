package service

import (
	"context"
	"errors"
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// firstFrameBudget 是本程序分到的第一帧预算。
const firstFrameBudget = 350 * time.Millisecond

// Start 在求解已经装好之后插入唯一的一条活动记录。编译失败由调用方处理，这里不插入。
func (m *Records) Start(ctx context.Context, requestID, projectID, experimentID, speed string, session domain.Session) (domain.Result, error) {
	if hit, ok := m.recalled(requestID); ok {
		return hit, nil
	}
	if requestID == "" || projectID == "" || experimentID == "" || session == nil || !knownSpeed(speed) {
		return domain.Result{}, &domain.Error{Code: domain.CodeInvalid}
	}
	_ = firstFrameBudget
	runCtx, stop := context.WithCancel(context.Background())
	if err := session.Start(runCtx); err != nil {
		stop()
		return domain.Result{}, &domain.Error{Code: domain.CodeUnavailable}
	}
	record, created, err := m.persist(ctx, requestID, projectID, experimentID, speed, session.Source())
	if err != nil || !created {
		_ = session.Stop()
		stop()
		if err != nil {
			return domain.Result{}, err
		}
		return resultOf(record, session.Frame()), nil
	}
	m.arm(record, session, runCtx, stop, speed)
	return domain.Result{RecordID: record.ID, Status: record.Status, Sample: session.Frame()}, nil
}

func (m *Records) persist(ctx context.Context, requestID, projectID, experimentID, speed, source string) (domain.Record, bool, error) {
	record := domain.Record{
		ID: domain.NewID("r"), RequestID: requestID, ProjectID: projectID,
		ExperimentID: experimentID, Status: "running", Speed: speed, Source: source,
		Revision: 1, Active: domain.MarkActive(), UpdatedAt: time.Now().UTC(),
	}
	err := m.records.Insert(ctx, record)
	if errors.Is(err, domain.ErrDuplicate) {
		existing, readErr := m.records.FindByRequestID(ctx, requestID)
		return existing, false, readErr
	}
	if errors.Is(err, domain.ErrActive) {
		return domain.Record{}, false, &domain.Error{Code: domain.CodeRejected}
	}
	if err != nil {
		return domain.Record{}, false, &domain.Error{Code: domain.CodeUnavailable}
	}
	return record, true, nil
}

func (m *Records) arm(record domain.Record, session domain.Session, ctx context.Context, stop context.CancelFunc, speed string) {
	cur := &live{record: record, session: session, ctx: ctx, cancel: stop, speed: speed}
	if !session.HasEngine() {
		cur.idle.Store(true)
	}
	m.mu.Lock()
	m.current = cur
	m.remember[record.RequestID] = domain.Result{RecordID: record.ID, Status: "running", Sample: session.Frame()}
	m.mu.Unlock()
	go m.watch(cur)
}

func (m *Records) recalled(requestID string) (domain.Result, bool) {
	if requestID == "" {
		return domain.Result{}, false
	}
	if existing, err := m.records.FindByRequestID(context.Background(), requestID); err == nil {
		sample := domain.Sample{Source: existing.Source, Movers: []domain.Mover{}}
		if existing.Sample != nil {
			sample = *existing.Sample
		}
		return domain.Result{RecordID: existing.ID, Status: existing.Status, Sample: sample}, true
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	hit, ok := m.remember[requestID]
	return hit, ok
}

func knownSpeed(speed string) bool {
	return speed == "x1" || speed == "x10" || speed == "asap"
}

func resultOf(record domain.Record, fresh domain.Sample) domain.Result {
	if record.Sample != nil {
		fresh = *record.Sample
	}
	return domain.Result{RecordID: record.ID, Status: record.Status, Sample: fresh}
}
