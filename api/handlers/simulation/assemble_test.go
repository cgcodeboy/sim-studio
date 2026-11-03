package simulation

import (
	"testing"
	"time"

	"github.com/cgcodeboy/sim-studio/internal/app"
	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
)

func TestQueueOverflowDoesNotStop(t *testing.T) {
	application := app.Wire()
	items, err := application.Library.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	session, err := assemble(application, lineProject(), "sim", items)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	var sample = session.Frame()
	for time.Now().Before(deadline) {
		sample = session.Frame()
		if sample.Overflow > 0 && sample.Exited > 0 {
			break
		}
		session.Advance()
		time.Sleep(20 * time.Millisecond)
	}
	_ = session.Stop()
	if sample.Overflow == 0 {
		t.Fatalf("overflow = 0, sample %+v", sample)
	}
	if sample.Source != "engine" {
		t.Fatalf("source %s", sample.Source)
	}
}

func TestMixedRunKeepsStockSource(t *testing.T) {
	application := app.Wire()
	items, _ := application.Library.List(t.Context())
	session, err := assemble(application, mixedProject(), "sim", items)
	if err != nil {
		t.Fatal(err)
	}
	session.Advance()
	sample := session.Frame()
	if sample.Source != "engine" {
		t.Fatalf("source %s", sample.Source)
	}
	var stock string
	for _, series := range sample.Series {
		if series.Name == "s" {
			stock = series.Source
		}
	}
	if stock != "dynamics" {
		t.Fatalf("stock source %s, series %+v", stock, sample.Series)
	}
}

func lineProject() model.Project {
	return model.Project{
		ID: "p1", Name: "线", Unit: "秒",
		Agents: []model.Agent{{
			ID: "main", Name: "Main", Space: "连续", W: 800, H: 600,
			Elements: []model.Element{
				{ID: "src", Type: "Source", Props: map[string]any{"rate": "每秒 20 个", "limit": "30"}},
				{ID: "q", Type: "Queue", X: 100, Y: 40, Props: map[string]any{"capacity": float64(1)}},
				{ID: "d", Type: "Delay", X: 200, Y: 40, Props: map[string]any{"delay": "1 秒"}},
				{ID: "sink", Type: "Sink"},
			},
			Links: []model.Link{
				{ID: "l1", From: "src", To: "q"},
				{ID: "l2", From: "q", To: "d"},
				{ID: "l3", From: "d", To: "sink"},
			},
		}},
		Experiments: []model.Experiment{{ID: "sim", Type: "simulation", Props: map[string]any{"stop": "3"}}},
	}
}

func mixedProject() model.Project {
	return model.Project{
		ID: "p1", Name: "混", Unit: "分钟",
		Agents: []model.Agent{{
			ID: "main", Name: "Main", Space: "连续", W: 800, H: 600,
			Elements: []model.Element{
				{ID: "q", Type: "Queue"},
				{ID: "s", Type: "Stock", Props: map[string]any{"initial": float64(1)}},
			},
		}},
		Experiments: []model.Experiment{{ID: "sim", Type: "simulation"}},
	}
}
