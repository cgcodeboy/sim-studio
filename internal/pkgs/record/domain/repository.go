package domain

import (
	"context"
	"errors"
)

var (
	ErrDuplicate = errors.New("duplicate")
	ErrNotFound  = errors.New("notfound")
	ErrActive    = errors.New("active")
)

// Session 是一次已经装好的求解。运行记录只通过这个接口拨动它。
type Session interface {
	Source() string
	Frame() Sample
	Params() map[string]float64
	HasEngine() bool
	Start(ctx context.Context) error
	Pause() error
	Resume() error
	Stop() error
	Wait()
	Shell() bool
	Advance()
	Done() bool
}

// Repository 是运行记录的仓储。
type Repository interface {
	Insert(ctx context.Context, record Record) error
	FindByRequestID(ctx context.Context, requestID string) (Record, error)
	Read(ctx context.Context, id string) (Record, error)
	MatchRevision(ctx context.Context, id string, revision int64, next Record) (int64, error)
	CountActive(ctx context.Context, projectID string) (int64, error)
	DeleteByProject(ctx context.Context, projectID string) error
	InterruptActive(ctx context.Context) error
	DeleteOrphans(ctx context.Context, projectIDs []string) error
}

// Result 是交给接入的一次结果。
type Result struct {
	RecordID string `json:"recordId"`
	Status   string `json:"status"`
	Sample   Sample `json:"sample"`
}

// Service 是运行记录的入口。它不装求解规则。
type Service interface {
	Start(ctx context.Context, requestID, projectID, experimentID, speed string, session Session) (Result, error)
	Pause(ctx context.Context, requestID, recordID string) (Result, error)
	Resume(ctx context.Context, requestID, recordID string) (Result, error)
	Stop(ctx context.Context, requestID, recordID string) (Result, error)
	Speed(ctx context.Context, requestID, recordID, speed string) (Result, error)
	Sample(recordID string) (Result, error)
	ProjectID(recordID string) string
	Params(recordID string) map[string]float64
	CountActive(ctx context.Context, projectID string) (int64, error)
	DeleteByProject(ctx context.Context, projectID string) error
	Sweep(ctx context.Context) error
	DropOrphans(ctx context.Context, projectIDs []string) error
}
