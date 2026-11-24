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

func TestRealtimeQueueShowsBody(t *testing.T) {
	application := app.Wire()
	items, err := application.Library.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	session, err := assemble(application, motionProject(), "sim", items)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	sample := session.Frame()
	_ = session.Stop()
	if len(sample.Movers) == 0 {
		t.Fatalf("no mover, arrived %v exited %v", sample.Arrived, sample.Exited)
	}
	if sample.Movers[0].Kind != "box" {
		t.Fatalf("kind %s", sample.Movers[0].Kind)
	}
	if sample.Movers[0].FromX == sample.Movers[0].ToX && sample.Movers[0].Action == "ride" {
		t.Fatalf("ride has no length: %+v", sample.Movers[0])
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

func motionProject() model.Project {
	pose := func(x float64) *model.Pose { return &model.Pose{X: x, Z: 0, Len: 2} }
	return model.Project{
		ID: "p1", Name: "动", Unit: "秒",
		Agents: []model.Agent{{
			ID: "main", Name: "Main", Space: "连续", W: 800, H: 600,
			Elements: []model.Element{
				{ID: "src", Type: "Source", Pose: pose(0), Props: map[string]any{"rate": "每秒 5 个", "limit": "4"}},
				{ID: "q", Type: "Queue", Pose: pose(3), Props: map[string]any{"capacity": float64(5)}},
				{ID: "d", Type: "Delay", Pose: pose(6), Props: map[string]any{"delay": "2 秒"}},
				{ID: "belt", Type: "Conveyor", Pose: pose(10), Props: map[string]any{}},
				{ID: "sink", Type: "Sink", Pose: pose(14)},
			},
			Links: []model.Link{
				{ID: "l1", From: "src", To: "q"},
				{ID: "l2", From: "q", To: "d"},
				{ID: "l3", From: "d", To: "belt"},
				{ID: "l4", From: "belt", To: "sink"},
			},
		}},
		Experiments: []model.Experiment{{ID: "sim", Type: "simulation", Props: map[string]any{"stop": "8", "live": "1"}}},
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
