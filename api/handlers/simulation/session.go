package simulation

import (
	"context"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	record "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

// studio 是一次运行业务装出来的求解。运行记录只看见 Session。
type studio struct {
	sheet   *record.Sheet
	eng     *engine.Engine
	params  map[string]float64
	steps   []func(time.Duration)
	shell   bool
	now     time.Duration
	horizon time.Duration
	source  string
}

func (s *studio) Source() string             { return s.source }
func (s *studio) Frame() record.Sample       { return s.sheet.Frame(s.now) }
func (s *studio) Params() map[string]float64 { return s.params }
func (s *studio) HasEngine() bool            { return s.eng != nil }
func (s *studio) Shell() bool                { return s.shell }
func (s *studio) Done() bool                 { return !s.shell || s.now >= s.horizon }

func (s *studio) Start(ctx context.Context) error {
	if s.eng == nil {
		return nil
	}
	return s.eng.Start(ctx)
}

func (s *studio) Pause() error {
	if s.eng == nil {
		return nil
	}
	return s.eng.Pause()
}

func (s *studio) Resume() error {
	if s.eng == nil {
		return nil
	}
	return s.eng.Resume()
}

func (s *studio) Stop() error {
	if s.eng == nil {
		return nil
	}
	return s.eng.Stop()
}

func (s *studio) Wait() {
	if s.eng == nil {
		return
	}
	_ = s.eng.Wait()
}

func (s *studio) Advance() {
	if s.shell {
		s.now += time.Second
	}
	s.sheet.SetWhen(s.now)
	for _, step := range s.steps {
		step(s.now)
	}
}
