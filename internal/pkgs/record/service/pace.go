package service

import "time"

// Pace 在处理函数之外拨动引擎。asap 不因为墙钟而暂停。
func (m *Records) Pace(recordID, op, speed string) {
	m.mu.Lock()
	cur := m.current
	m.mu.Unlock()
	if cur == nil || cur.record.ID != recordID || cur.session == nil {
		return
	}
	if op == "pause" || op == "stop" {
		cur.user.Store(true)
	}
	if op == "resume" {
		cur.user.Store(false)
	}
	if !cur.session.HasEngine() {
		if op == "stop" && cur.cancel != nil {
			cur.cancel()
		}
		return
	}
	switch op {
	case "pause":
		_ = cur.session.Pause()
	case "resume":
		_ = cur.session.Resume()
	case "stop":
		_ = cur.session.Stop()
		if cur.cancel != nil {
			cur.cancel()
		}
	default:
		_ = speed
	}
}

func (m *Records) watch(cur *live) {
	if cur.session.HasEngine() {
		go func() {
			cur.session.Wait()
			cur.idle.Store(true)
		}()
	}
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if cur.ctx.Err() != nil {
			return
		}
		if cur.idle.Load() && m.shellDone(cur) {
			m.Finish(cur.ctx, cur, "finished")
			return
		}
		select {
		case <-cur.ctx.Done():
			return
		case <-tick.C:
			if cur.user.Load() || !cur.session.Shell() {
				continue
			}
			cur.session.Advance()
		}
	}
}

func (m *Records) shellDone(cur *live) bool {
	if !cur.session.Shell() {
		return true
	}
	return cur.session.Done()
}
