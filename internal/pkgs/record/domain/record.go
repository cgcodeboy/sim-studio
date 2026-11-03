// Package domain 守住一条运行记录的状态迁移。
// 活动标记只在 running 或 paused 时为 1。离开这两个状态时必须清掉。
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// ErrHeld 表示已经处于这次命令的目标状态，调用方不再写入。
var ErrHeld = errors.New("held")

const (
	CodeInvalid     = "invalid"
	CodeNotFound    = "notfound"
	CodeConflict    = "conflict"
	CodeRejected    = "rejected"
	CodeUnavailable = "unavailable"
)

// Error 是运行记录自己的错误。它不引用模型。
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// As 取出运行记录的错误。
func As(err error) *Error {
	var re *Error
	if errors.As(err, &re) {
		return re
	}
	return nil
}

// NewID 生成一条运行记录的标识。
func NewID(prefix string) string {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	return prefix + hex.EncodeToString(buf[:])
}

// Record 是一条运行记录。版本从 1 起。
type Record struct {
	ID               string    `json:"id" bson:"_id"`
	RequestID        string    `json:"requestId" bson:"requestId"`
	ProjectID        string    `json:"projectId" bson:"projectId"`
	ExperimentID     string    `json:"experimentId" bson:"experimentId"`
	Status           string    `json:"status" bson:"status"`
	Speed            string    `json:"speed" bson:"speed"`
	Source           string    `json:"source" bson:"source"`
	Revision         int64     `json:"revision" bson:"revision"`
	Active           *int32    `json:"active,omitempty" bson:"active,omitempty"`
	ControlRequestID string    `json:"controlRequestId,omitempty" bson:"controlRequestId,omitempty"`
	ControlOp        string    `json:"controlOp,omitempty" bson:"controlOp,omitempty"`
	Sample           *Sample   `json:"sample,omitempty" bson:"sample,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updatedAt"`
}

// MarkActive 返回占用活动槽的标记。用 int32，避免和索引里的整数类型对不上。
func MarkActive() *int32 {
	n := int32(1)
	return &n
}

// Occupied 表示这条记录正占着唯一的活动槽。
func (r Record) Occupied() bool {
	return r.Active != nil && *r.Active == 1
}

// Advance 只改内存里的下一次状态，不访问数据库。
func (r *Record) Advance(op string) error {
	switch r.Status {
	case "running":
		return advanceRunning(r, op)
	case "paused":
		return advancePaused(r, op)
	default:
		return &Error{Code: CodeRejected}
	}
}

func advanceRunning(r *Record, op string) error {
	switch op {
	case "pause":
		r.Status = "paused"
	case "speed":
	case "stop":
		r.Status = "interrupted"
		r.Active = nil
	case "finish":
		r.Status = "finished"
		r.Active = nil
	default:
		return &Error{Code: CodeRejected}
	}
	return nil
}

func advancePaused(r *Record, op string) error {
	switch op {
	case "pause":
		return ErrHeld
	case "resume", "speed":
		if op == "resume" {
			r.Status = "running"
		}
	case "stop":
		r.Status = "interrupted"
		r.Active = nil
	case "finish":
		r.Status = "finished"
		r.Active = nil
	default:
		return &Error{Code: CodeRejected}
	}
	return nil
}
