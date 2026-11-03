package service

import (
	"context"
	"errors"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// Pause 把活动运行改成 paused，并在处理函数之外暂停引擎。
func (m *Records) Pause(ctx context.Context, requestID, recordID string) (domain.Result, error) {
	return m.control(ctx, requestID, recordID, "pause", "")
}

// Resume 继续这条运行。参数由调用方先写进 domain.Session。
func (m *Records) Resume(ctx context.Context, requestID, recordID string) (domain.Result, error) {
	return m.control(ctx, requestID, recordID, "resume", "")
}

// Stop 结束这条运行并去掉活动标记。
func (m *Records) Stop(ctx context.Context, requestID, recordID string) (domain.Result, error) {
	return m.control(ctx, requestID, recordID, "stop", "")
}

// Speed 只改速度，不改状态。
func (m *Records) Speed(ctx context.Context, requestID, recordID, speed string) (domain.Result, error) {
	if !knownSpeed(speed) {
		return domain.Result{}, &domain.Error{Code: domain.CodeInvalid}
	}
	return m.control(ctx, requestID, recordID, "speed", speed)
}

// ProjectID 返回活动运行所属的工程。调用方用它去读模型。
func (m *Records) ProjectID(recordID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || m.current.record.ID != recordID {
		return ""
	}
	return m.current.record.ProjectID
}

// Params 返回活动运行的参数表，供调用方在继续之前写入设置。
func (m *Records) Params(recordID string) map[string]float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || m.current.record.ID != recordID || m.current.session == nil {
		return nil
	}
	return m.current.session.Params()
}

func (m *Records) control(ctx context.Context, requestID, recordID, op, speed string) (domain.Result, error) {
	if requestID == "" || recordID == "" {
		return domain.Result{}, &domain.Error{Code: domain.CodeInvalid}
	}
	var last error
	for attempt := 0; attempt <= retryOnce; attempt++ {
		result, done, err := m.controlOnce(ctx, requestID, recordID, op, speed)
		if done || err == nil {
			return result, err
		}
		last = err
		if domain.As(err) != nil && domain.As(err).Code != domain.CodeConflict {
			return domain.Result{}, err
		}
	}
	if last == nil {
		last = &domain.Error{Code: domain.CodeConflict}
	}
	return domain.Result{}, last
}

func (m *Records) controlOnce(ctx context.Context, requestID, recordID, op, speed string) (domain.Result, bool, error) {
	record, err := m.records.Read(ctx, recordID)
	if err != nil {
		return domain.Result{}, true, &domain.Error{Code: domain.CodeNotFound}
	}
	if record.ControlRequestID == requestID {
		return resultOf(record, domain.Sample{Movers: []domain.Mover{}}), true, nil
	}
	next := record
	if speed != "" {
		next.Speed = speed
	}
	if err := next.Advance(op); err != nil {
		if errors.Is(err, domain.ErrHeld) {
			return resultOf(record, m.frame(recordID)), true, nil
		}
		return domain.Result{}, true, err
	}
	next.ControlRequestID = requestID
	next.ControlOp = op
	if op == "stop" || op == "finish" {
		sample := m.frame(recordID)
		next.Sample = &sample
	}
	matched, err := m.records.MatchRevision(ctx, record.ID, record.Revision, next)
	if err != nil {
		return domain.Result{}, true, &domain.Error{Code: domain.CodeUnavailable}
	}
	if matched == 0 {
		return domain.Result{}, false, &domain.Error{Code: domain.CodeConflict}
	}
	m.note(recordID, next)
	m.Pace(recordID, op, next.Speed)
	if op == "stop" {
		m.finishLive(next, *next.Sample)
	}
	return resultOf(next, m.frame(recordID)), true, nil
}

func (m *Records) note(recordID string, next domain.Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current != nil && m.current.record.ID == recordID {
		m.current.record.Status = next.Status
		m.current.speed = next.Speed
	}
}

func (m *Records) frame(recordID string) domain.Sample {
	m.mu.Lock()
	cur := m.current
	m.mu.Unlock()
	if cur != nil && cur.record.ID == recordID && cur.session != nil {
		return cur.session.Frame()
	}
	return domain.Sample{Movers: []domain.Mover{}}
}
