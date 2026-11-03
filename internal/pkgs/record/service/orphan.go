package service

import (
	"context"
	"errors"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// DropOrphans 删掉工程已经不存在的运行记录。工程标识由调用方传入。
func (m *Records) DropOrphans(ctx context.Context, projectIDs []string) error {
	return m.records.DeleteOrphans(ctx, projectIDs)
}

// CountActive 返回该工程仍占着活动槽的记录数。
func (m *Records) CountActive(ctx context.Context, projectID string) (int64, error) {
	return m.records.CountActive(ctx, projectID)
}

// DeleteByProject 在该工程没有活动运行时删除它的记录。
func (m *Records) DeleteByProject(ctx context.Context, projectID string) error {
	n, err := m.records.CountActive(ctx, projectID)
	if err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrActive
	}
	err = m.records.DeleteByProject(ctx, projectID)
	if errors.Is(err, domain.ErrActive) {
		return err
	}
	return err
}
