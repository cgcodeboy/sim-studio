package domain

import (
	"context"
	"errors"
)

var (
	ErrDuplicate = errors.New("duplicate")
	ErrNotFound  = errors.New("notfound")
)

// Repository 是工程的仓储。接口属于模型领域，实现在 repo。
type Repository interface {
	List(ctx context.Context) ([]Summary, error)
	IDs(ctx context.Context) ([]string, error)
	Read(ctx context.Context, id string) (Project, error)
	Insert(ctx context.Context, p Project) error
	Replace(ctx context.Context, expected int64, p Project) (bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// Service 是模型领域的入口。它不认识运行记录。
type Service interface {
	List(ctx context.Context) ([]Summary, error)
	Read(ctx context.Context, id string) (Project, error)
	Put(ctx context.Context, expected int64, p Project) (Project, error)
	IDs(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, id string) error
}
