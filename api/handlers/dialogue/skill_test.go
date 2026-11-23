package dialogue

import (
	"context"
	"math"
	"strings"
	"testing"
)

func TestDecideWithoutModelDoesNotUseRules(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	text := "在 Main 上建一条线。每道工序各放一个 Queue 和一个 Delay：来料、清洗、加工。然后运行。"
	task := New(nil, nil).decide(context.Background(), Input{Text: text, Page: "Main"})
	if task.Status != "未写入" || task.Intent != "clarify" || len(task.Types) != 0 {
		t.Fatalf("%+v", task)
	}
}

func TestRouteKeepsEveryDomain(t *testing.T) {
	names := specialistNames(`{"specialists":["process","fluid","handling","clarify"]}`)
	if len(names) != 3 || names[0] != "process" || names[1] != "fluid" || names[2] != "handling" {
		t.Fatalf("%v", names)
	}
}

func TestNormalizeSkipsOtherFamily(t *testing.T) {
	task := normalize(Task{Intent: "place", Agent: "Main", Types: []string{"Source", "Delay", "Tank", "Sink"}}, nil)
	if task.Status != "已写入" || len(task.Types) != 3 || task.Types[2] != "Sink" {
		t.Fatalf("%+v", task)
	}
}

func TestPlanDropsUnknownDomain(t *testing.T) {
	planned := parsePlan(`{"steps":[{"domain":"process","work":"洗瓶"},{"domain":"clarify","work":"不行"}],"accept":"有洗瓶"}`)
	if len(planned.Steps) != 1 || planned.Steps[0].Domain != "process" || planned.Accept != "有洗瓶" {
		t.Fatalf("%+v", planned)
	}
}

func TestVerdictReadsGaps(t *testing.T) {
	result := parseVerdict(`{"ok":false,"gaps":["没有冻干"]}`)
	if result.OK || len(result.Gaps) != 1 || result.Gaps[0] != "没有冻干" {
		t.Fatalf("%+v", result)
	}
}

func TestQueryAnswersWithoutBuilding(t *testing.T) {
	task, ok := readyTask(plan{Act: "query", Say: "这条线先洗瓶再灌装。"}, "Main")
	if !ok || task.Status != "未写入" || task.Reply != "这条线先洗瓶再灌装。" {
		t.Fatalf("%+v", task)
	}
}

func TestFitJoinsFacesAndSeparatesFluid(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{
		Name: "Main",
		Links: []pageLink{
			{From: "source", To: "belt"},
			{From: "belt", To: "delay"},
			{From: "delay", To: "check"},
			{From: "check", To: "sink"},
			{From: "check", To: "bad", FromPort: 1},
			{From: "check", To: "seize"},
			{From: "seize", To: "sink"},
			{From: "pump", To: "tank"},
		},
		Elements: []element{
			{Name: "source", Type: "Source", W: 156, H: 84},
			{Name: "belt", Type: "Conveyor", W: 220, H: 48},
			{Name: "delay", Type: "Delay", W: 156, H: 84},
			{Name: "check", Type: "Select Output", W: 156, H: 84},
			{Name: "sink", Type: "Sink", W: 156, H: 84},
			{Name: "bad", Type: "Sink", W: 156, H: 84},
			{Name: "seize", Type: "Seize", W: 156, H: 84},
			{Name: "pump", Type: "FluidSource", W: 156, H: 84},
			{Name: "tank", Type: "Tank", W: 156, H: 84},
		},
	}}}
	spots := fitScene(in, nil)
	at := map[string]PoseSpot{}
	for _, spot := range spots {
		at[spot.Element] = *spot.Pose
	}
	if len(at) != 9 {
		t.Fatalf("count %d", len(at))
	}
	sourceFront := at["source"].X + 1.09
	beltStart := at["belt"].X - at["belt"].Len/2
	if math.Abs(sourceFront-beltStart) > 0.05 || math.Abs(at["belt"].Z) > 0.01 || at["delay"].Z != 0 {
		t.Fatalf("join %+v", at)
	}
	if at["bad"].Z < 1 || at["seize"].Z > -0.8 || math.Abs(at["tank"].Z-at["source"].Z) < 2 {
		t.Fatalf("split %+v", at)
	}
}

func TestConflictsNamesAGapAndAnOverlap(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{
		Name:  "Main",
		Links: []pageLink{{From: "source", To: "delay"}},
		Elements: []element{
			{Name: "source", Type: "Source", W: 156, H: 84},
			{Name: "delay", Type: "Delay", W: 156, H: 84},
		},
	}}}
	far := []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 1.2, Z: 0, Len: 1.45}},
		{Element: "delay", X: 250, Y: 80, Pose: &PoseSpot{X: 12, Z: 0, Len: 2.3}},
	}
	if !strings.Contains(strings.Join(conflicts(in, far, nil), " "), "没有接上") {
		t.Fatal(conflicts(in, far, nil))
	}
	piled := []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0, Len: 1.45}},
		{Element: "delay", X: 40, Y: 80, Pose: &PoseSpot{X: 0.1, Z: 0, Len: 2.3}},
	}
	if !strings.Contains(strings.Join(conflicts(in, piled, nil), " "), "重叠") {
		t.Fatal(conflicts(in, piled, nil))
	}
}

func TestSettleOpensAStackedLayout(t *testing.T) {
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0, Len: 1.8}},
		{Element: "delay", X: 40, Y: 80, Pose: &PoseSpot{X: 0.1, Z: 0, Len: 2.8}},
		{Element: "sink", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0.1, Len: 1.6}},
	}}
	task.Places = settle(sampleLine(), task.Places, nil)
	if why := rejectPlaces(sampleLine(), task, nil); why != "" {
		t.Fatal(why)
	}
}

func TestRejectsAStackedLayout(t *testing.T) {
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0, Len: 1.8}},
		{Element: "delay", X: 40, Y: 80, Pose: &PoseSpot{X: 0.1, Z: 0, Len: 2.8}},
		{Element: "sink", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0.1, Len: 1.6}},
	}}
	if rejectPlaces(sampleLine(), task, nil) == "" {
		t.Fatal("stacked layout accepted")
	}
}

func TestAcceptsTheSceneTheSkillDrew(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{Name: "Main", Elements: []element{
		{Name: "source", Type: "Source", W: 156, H: 84},
		{Name: "check", Type: "Select Output", W: 156, H: 84},
		{Name: "ok", Type: "Sink", W: 156, H: 84},
		{Name: "bad", Type: "Sink", W: 156, H: 84},
		{Name: "tank", Type: "Tank", W: 156, H: 84},
	}}}}
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 1.2, Z: 0, Yaw: 0, Len: 1.8}},
		{Element: "check", X: 250, Y: 80, Pose: &PoseSpot{X: 5, Z: 0, Yaw: 0, Len: 2.2}},
		{Element: "ok", X: 460, Y: 80, Pose: &PoseSpot{X: 8.5, Z: 0, Yaw: 0, Len: 1.6}},
		{Element: "bad", X: 250, Y: 220, Pose: &PoseSpot{X: 5, Z: 3.2, Yaw: 1.5708, Len: 1.6}},
		{Element: "tank", X: 40, Y: 360, Pose: &PoseSpot{X: 1.2, Z: 8, Yaw: 0, Len: 2}},
	}}
	if why := rejectPlaces(in, task, nil); why != "" {
		t.Fatal(why)
	}
}

func TestBindsNumberedConveyors(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{Name: "Main", Elements: []element{
		{Name: "source", Type: "Source", X: 40, Y: 200, W: 156, H: 84, Pose: &PoseSpot{X: 1.2, Len: 1.8}},
		{Name: "conveyor", Type: "Conveyor", X: 40, Y: 400, W: 220, H: 48, Pose: &PoseSpot{X: 2, Z: 2, Len: 4}},
		{Name: "conveyor2", Type: "Conveyor", X: 400, Y: 400, W: 220, H: 48, Pose: &PoseSpot{X: 8, Z: 2, Len: 4}},
	}}}}
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 1.2, Z: 0, Yaw: 0, Len: 1.8}},
		{Element: "传送带1", X: 40, Y: 320, Pose: &PoseSpot{X: 2, Z: 2, Yaw: 0, Len: 6}},
		{Element: "传送带2", X: 400, Y: 320, Pose: &PoseSpot{X: 8, Z: 2, Yaw: 0, Len: 6}},
	}}
	task.Places = bindPlaces(in, task.Places, nil)
	task.Places = separate(task.Places, roster(openAgent(in), nil))
	if why := rejectPlaces(in, task, nil); why != "" {
		t.Fatal(why)
	}
}

func TestBindsConveyorByAlias(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{Name: "Main", Elements: []element{
		{Name: "belt", Type: "Conveyor", W: 220, H: 48},
	}}}}
	task := Task{Places: []Place{{Element: "传送带", X: 40, Y: 80, Pose: &PoseSpot{X: 4, Z: 1.6, Yaw: 0, Len: 9}}}}
	task.Places = bindPlaces(in, task.Places, nil)
	if why := rejectPlaces(in, task, nil); why != "" || task.Places[0].Element != "belt" {
		t.Fatal(why, task.Places[0].Element)
	}
}

func TestSeparateOpensATightLine(t *testing.T) {
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 0, Z: 0, Len: 1.8}},
		{Element: "delay", X: 40, Y: 80, Pose: &PoseSpot{X: 1, Z: 0, Len: 2.8}},
		{Element: "sink", X: 40, Y: 80, Pose: &PoseSpot{X: 2, Z: 0, Len: 1.6}},
	}}
	task.Places = separate(task.Places, roster(openAgent(sampleLine()), nil))
	if why := rejectPlaces(sampleLine(), task, nil); why != "" {
		t.Fatal(why)
	}
}

func TestAcceptsASeparatedLayout(t *testing.T) {
	task := Task{Places: []Place{
		{Element: "source", X: 40, Y: 80, Pose: &PoseSpot{X: 1.2, Y: 0, Z: 0, Yaw: 0, Len: 1.8}},
		{Element: "delay", X: 244, Y: 80, Pose: &PoseSpot{X: 3.7, Y: 0, Z: 0, Yaw: 0, Len: 2.8}},
		{Element: "sink", X: 448, Y: 80, Pose: &PoseSpot{X: 6.4, Y: 0, Z: 1.6, Yaw: 0, Len: 1.6}},
	}}
	if why := rejectPlaces(sampleLine(), task, nil); why != "" {
		t.Fatal(why)
	}
}

func sampleLine() Input {
	return Input{Page: "Main", Agents: []agent{{
		Name:  "Main",
		Links: []pageLink{{From: "source", To: "delay"}, {From: "delay", To: "sink"}},
		Elements: []element{
			{Name: "source", Type: "Source", W: 156, H: 84},
			{Name: "delay", Type: "Delay", W: 156, H: 84},
			{Name: "sink", Type: "Sink", W: 156, H: 84},
		},
	}}}
}

func TestJoinAbutsTheLine(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{Name: "Main", Links: []pageLink{{From: "source", To: "sink"}}, Elements: []element{
		{Name: "source", Type: "Source", W: 156, H: 84},
		{Name: "belt", Type: "Conveyor", W: 220, H: 48},
		{Name: "sink", Type: "Sink", W: 156, H: 84},
	}}}}
	places := map[string]Place{}
	for _, spot := range joinLinks(in, nil, 8) {
		places[spot.Element] = spot
	}
	source, sink, belt := places["source"].Pose, places["sink"].Pose, places["belt"].Pose
	gap := (sink.X - sink.Len/2) - (source.X + source.Len/2)
	if math.Abs(gap-0.2) > 0.02 {
		t.Fatalf("gap %v", gap)
	}
	run := (sink.X + sink.Len/2) - (source.X - source.Len/2)
	if math.Abs(belt.Len-run) > 0.02 || belt.Z < 1 {
		t.Fatalf("belt %+v run %v", belt, run)
	}
}

func TestGuideAsksOneQuestion(t *testing.T) {
	task, ok := readyTask(plan{Act: "guide", Say: "你说想做个工厂。先定产品：药、食品，还是零件？"}, "Main")
	if !ok || task.Status != "未写入" || task.Reply != "你说想做个工厂。先定产品：药、食品，还是零件？" {
		t.Fatalf("%+v", task)
	}
}

func TestRefuseNamesTheReason(t *testing.T) {
	task, ok := readyTask(plan{Act: "refuse", Say: "页面上没有冻干柜。"}, "Main")
	if !ok || task.Status != "未写入" || task.Reply != "做不到。页面上没有冻干柜。" {
		t.Fatalf("%+v", task)
	}
}

func TestPackKeepsTheFollowingAction(t *testing.T) {
	head := pack([]Task{{Intent: "set", Status: "已写入"}, {Intent: "layout", Status: "已写入"}}, "先改时间，再收紧。")
	if head.Intent != "set" || len(head.More) != 1 || head.More[0].Intent != "layout" || head.Reply != "先改时间，再收紧。" {
		t.Fatalf("%+v", head)
	}
}

func TestChangeStepSetsTheNamedElement(t *testing.T) {
	task, ok := taskFromStep(planStep{Intent: "set", Element: "delay", Key: "delay", Value: "8分钟"}, "Main")
	if !ok || task.Intent != "set" || task.Element != "delay" || task.Value != "8分钟" {
		t.Fatalf("%+v", task)
	}
}

func TestSkillsAreLoaded(t *testing.T) {
	names := append([]string{"router", "planner", "checker", "layout"}, specialists...)
	for _, name := range names {
		text, ok := skillPrompt(name, nil)
		if !ok || !strings.Contains(text, "JSON") {
			t.Fatalf("skill %s", name)
		}
	}
	process, ok := skillPrompt("process", nil)
	if !ok || !strings.Contains(process, "Select Output") || !strings.Contains(process, "tie") {
		t.Fatal("process skill")
	}
}
