package simulation

import (
	"strings"
	"time"

	"github.com/cgcodeboy/des-engine/engine"
	"github.com/cgcodeboy/sim-studio/internal/app"
	behavior "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	dynamics "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/domain"
	experiment "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
	fluid "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
	handling "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
	library "github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	pedestrian "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
	process "github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	rail "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	record "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
	road "github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
)

type tables struct {
	flow  process.Service
	haul  handling.Service
	ped   pedestrian.Service
	road  road.Service
	rail  rail.Service
	fluid fluid.Service
	dyn   dynamics.Service
	act   behavior.Service
	exp   experiment.Service
}

func assemble(application *app.App, p model.Project, experimentID string, types []library.Type) (*studio, error) {
	agent, exp, err := locate(p, experimentID)
	if err != nil {
		return nil, err
	}
	if err := refuse(p); err != nil {
		return nil, err
	}
	sheet := record.NewSheet(p.Unit)
	params := map[string]float64{}
	st := &studio{
		sheet: sheet, params: params, source: "engine",
		horizon: behavior.Horizon(p.Unit, exp.Props["stop"]),
	}
	rewards := rounds(exp)
	if len(rewards) > 0 {
		sheet.SetReward(rewards)
		st.source = "experiment"
	}
	eng, coord, err := newEngine(st.horizon)
	if err != nil {
		return nil, err
	}
	added := map[string]bool{}
	loaded := newTables(application, sheet, params, agent, bindOf(coord, added))
	if err := loaded.placeAll(agent, kinds(types), sheet); err != nil {
		return nil, err
	}
	if err := linkAll(coord, agent, added); err != nil {
		return nil, err
	}
	st.eng = eng
	if len(added) == 0 {
		st.eng = nil
	}
	st.finish(loaded, len(added))
	return st, nil
}

func rounds(exp model.Experiment) []float64 {
	min := behavior.Number(exp.Props["min"], 1)
	max := behavior.Number(exp.Props["max"], 2)
	step := behavior.Number(exp.Props["step"], 1)
	if exp.Type == "rl" {
		max = behavior.Number(exp.Props["episodes"], float64(experiment.RoundLimit))
		return experiment.Rounds("rl", 0, max, 1)
	}
	return experiment.Rounds(exp.Type, min, max, step)
}

func newEngine(horizon time.Duration) (*engine.Engine, *engine.Coordinator, error) {
	eng, err := engine.NewEngine(
		engine.WithRealtime(false),
		engine.WithHorizon(engine.Time(horizon/time.Millisecond)),
		engine.WithExitWhenIdle(true),
	)
	if err != nil {
		return nil, nil, unavailable()
	}
	coord, err := engine.NewCoordinator(eng)
	if err != nil {
		return nil, nil, unavailable()
	}
	return eng, coord, nil
}

func bindOf(coord *engine.Coordinator, added map[string]bool) func(string, any) error {
	return func(id string, raw any) error {
		atom, ok := raw.(engine.Atomic)
		if !ok || coord == nil {
			return invalid()
		}
		if err := coord.Add(id, atom); err != nil {
			return invalid()
		}
		added[id] = true
		return nil
	}
}

func newTables(application *app.App, sheet *record.Sheet, params map[string]float64, agent model.Agent, bind func(string, any) error) tables {
	t := tables{
		flow: application.Process.Open(sheet, params), haul: application.Handling.Open(sheet, params),
		ped: application.Pedestrian.Open(sheet, params), road: application.Road.Open(sheet, params),
		rail: application.Rail.Open(sheet, params), fluid: application.Fluid.Open(sheet, params),
		dyn: application.Dynamics.Open(sheet, params), act: application.Behavior.Open(sheet, params),
		exp: application.Experiment.Open(sheet, params),
	}
	t.flow.SetBind(bind)
	t.haul.SetBind(bind)
	t.ped.SetBind(bind)
	t.road.SetBind(bind)
	t.rail.SetBind(bind)
	t.fluid.SetBind(bind)
	t.dyn.SetBind(bind)
	t.act.SetBind(bind)
	t.exp.SetBind(bind)
	t.dyn.SetAgent(dynAgent(agent))
	t.act.SetAgent(actAgent(agent))
	return t
}

func (t tables) placeAll(agent model.Agent, kind map[string]string, sheet *record.Sheet) error {
	for _, el := range agent.Elements {
		switch kind[canon(el.Type)] {
		case "":
			return invalid()
		case "refuse":
			return invalid()
		case "camera":
			sheet.SetCamera(record.Camera{X: el.X, Y: el.Y, Target: behavior.Text(el.Props["target"])})
			continue
		case "geometry", "skip":
			continue
		}
		ok, err := t.place(el)
		if err != nil || !ok {
			return invalid()
		}
	}
	return nil
}

func (t tables) place(el model.Element) (bool, error) {
	steps := []func() (bool, error){
		func() (bool, error) { return t.flow.Place(flowBlock(el)) },
		func() (bool, error) { return t.haul.Place(haulBlock(el)) },
		func() (bool, error) { return t.ped.Place(pedBlock(el)) },
		func() (bool, error) { return t.road.Place(roadBlock(el)) },
		func() (bool, error) { return t.rail.Place(railBlock(el)) },
		func() (bool, error) { return t.fluid.Place(fluidBlock(el)) },
		func() (bool, error) { return t.dyn.Place(dynBlock(el)) },
		func() (bool, error) { return t.act.Place(actBlock(el)) },
		func() (bool, error) { return t.exp.Place(expBlock(el)) },
	}
	for _, step := range steps {
		ok, err := step()
		if ok || err != nil {
			return ok, err
		}
	}
	return false, nil
}

func (s *studio) finish(t tables, added int) {
	parts := []struct {
		steps    []func(time.Duration)
		shell    bool
		dynamics bool
	}{
		{t.flow.Steps(), t.flow.Shell(), t.flow.Dynamics()},
		{t.haul.Steps(), t.haul.Shell(), t.haul.Dynamics()},
		{t.ped.Steps(), t.ped.Shell(), t.ped.Dynamics()},
		{t.road.Steps(), t.road.Shell(), t.road.Dynamics()},
		{t.rail.Steps(), t.rail.Shell(), t.rail.Dynamics()},
		{t.fluid.Steps(), t.fluid.Shell(), t.fluid.Dynamics()},
		{t.dyn.Steps(), t.dyn.Shell(), t.dyn.Dynamics()},
		{t.act.Steps(), t.act.Shell(), t.act.Dynamics()},
		{t.exp.Steps(), t.exp.Shell(), t.exp.Dynamics()},
	}
	for _, part := range parts {
		s.steps = append(s.steps, part.steps...)
		if part.shell {
			s.shell = true
		}
		if part.dynamics && added == 0 && s.source != "experiment" {
			s.source = "dynamics"
		}
	}
}

func locate(p model.Project, experimentID string) (model.Agent, model.Experiment, error) {
	var agent model.Agent
	var exp model.Experiment
	foundA, foundE := false, false
	for _, item := range p.Agents {
		if item.ID == "main" {
			agent, foundA = item, true
		}
	}
	for _, item := range p.Experiments {
		if item.ID == experimentID {
			exp, foundE = item, true
		}
	}
	if !foundA || !foundE {
		return model.Agent{}, model.Experiment{}, invalid()
	}
	return agent, exp, nil
}

func kinds(types []library.Type) map[string]string {
	out := map[string]string{}
	for _, item := range types {
		out[canon(item.Name)] = item.Kind
	}
	return out
}

func linkAll(coord *engine.Coordinator, agent model.Agent, added map[string]bool) error {
	for _, link := range agent.Links {
		if link.From == link.To || !added[link.From] || !added[link.To] {
			continue
		}
		err := coord.Link(link.From, portName("out", link.FromPort), link.To, portName("in", link.ToPort))
		if err != nil {
			return invalid()
		}
	}
	return nil
}

func canon(name string) string { return strings.ReplaceAll(name, " ", "") }

func portName(prefix string, n int) string {
	if n <= 0 {
		return prefix
	}
	return prefix + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func dynAgent(agent model.Agent) dynamics.Agent {
	out := dynamics.Agent{}
	for _, el := range agent.Elements {
		out.Elements = append(out.Elements, dynBlock(el))
	}
	for _, link := range agent.Links {
		out.Links = append(out.Links, dynamics.Link{From: link.From, To: link.To})
	}
	return out
}

func actAgent(agent model.Agent) behavior.Agent {
	out := behavior.Agent{}
	for _, el := range agent.Elements {
		out.Elements = append(out.Elements, actBlock(el))
	}
	for _, link := range agent.Links {
		out.Links = append(out.Links, behavior.Link{From: link.From, To: link.To})
	}
	return out
}
