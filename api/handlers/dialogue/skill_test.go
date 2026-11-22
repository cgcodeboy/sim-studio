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
