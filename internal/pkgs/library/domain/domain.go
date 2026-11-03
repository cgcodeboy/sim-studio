// Package domain 是元件库的领域接口。它不引用其他领域。
package domain

import "context"

// Type 是一条类型目录。求解规则不在这里。
type Type struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// Repository 保存类型目录。
type Repository interface {
	List(ctx context.Context) ([]Type, error)
}

// Service 是元件库的入口。
type Service interface {
	List(ctx context.Context) ([]Type, error)
}
