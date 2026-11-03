package service

import (
	"context"
	"testing"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
	recordrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/record/repo"
)

type quiet struct{ source string }

func (q quiet) Source() string { return q.source }
func (q quiet) Frame() domain.Sample {
	return domain.Sample{Source: q.source, Movers: []domain.Mover{}}
}
func (q quiet) Params() map[string]float64  { return map[string]float64{} }
func (q quiet) HasEngine() bool             { return false }
func (q quiet) Start(context.Context) error { return nil }
func (q quiet) Pause() error                { return nil }
func (q quiet) Resume() error               { return nil }
func (q quiet) Stop() error                 { return nil }
func (q quiet) Wait()                       {}
func (q quiet) Shell() bool                 { return false }
func (q quiet) Advance()                    {}
func (q quiet) Done() bool                  { return true }

func TestDuplicateRequest(t *testing.T) {
	mgr := New(recordrepo.NewMem())
	first, err := mgr.Start(t.Context(), "req-1", "p1", "sim", "asap", quiet{source: "engine"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := mgr.Start(t.Context(), "req-1", "p1", "sim", "asap", quiet{source: "engine"})
	if err != nil || again.RecordID != first.RecordID {
		t.Fatal(err, again.RecordID, first.RecordID)
	}
}
