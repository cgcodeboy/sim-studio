package service

import (
	"context"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
)

// Service 实现元件库入口。类型从仓储读，不写死在服务里。
type Service struct{ types domain.Repository }

// New 接入仓储。
func New(types domain.Repository) *Service { return &Service{types: types} }

// List 返回类型目录。
func (s *Service) List(ctx context.Context) ([]domain.Type, error) {
	return s.types.List(ctx)
}
