package dialogue

import (
	"context"
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

func TestLayoutArrangesTheOpenPage(t *testing.T) {
	task, ok := readyTask(plan{Act: "layout"}, "Main")
	if !ok || task.Intent != "layout" || task.Status != "已写入" || task.Agent != "Main" || task.Gap != 8 {
		t.Fatalf("%+v", task)
	}
}

func TestTightLayoutKeepsTheRequestedGap(t *testing.T) {
	task, ok := readyTask(plan{Act: "layout", Gap: 4}, "Main")
	if !ok || task.Gap != 4 {
		t.Fatalf("%+v", task)
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
	names := append([]string{"router", "planner", "checker"}, specialists...)
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
