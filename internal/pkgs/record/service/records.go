// Package service 只写运行记录。它不读取模型，也不装求解规则。
package service

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

const retryOnce = 1

// Records 同时只盯着一条活动运行。
type Records struct {
	mu       sync.Mutex
	records  domain.Repository
	current  *live
	remember map[string]domain.Result
}

type live struct {
	record  domain.Record
	session domain.Session
	ctx     context.Context
	cancel  context.CancelFunc
	user    atomic.Bool
	idle    atomic.Bool
	speed   string
	once    sync.Once
}

// New 只接入运行记录。
func New(records domain.Repository) *Records {
	return &Records{records: records, remember: map[string]domain.Result{}}
}
