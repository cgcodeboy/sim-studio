package service

import "context"

// Sweep 在开始接请求之前把活动运行改成 interrupted。不删除记录。
func (m *Records) Sweep(ctx context.Context) error {
	return m.records.InterruptActive(ctx)
}
