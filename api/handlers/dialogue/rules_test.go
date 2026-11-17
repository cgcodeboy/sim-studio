package dialogue

import "testing"

func TestPlaceDelay(t *testing.T) {
	task := fromRules(Input{Text: "在 Main 上加一个 Delay", Page: "Main"}, nil)
	if task.Status != "已写入" || task.Intent != "place" || len(task.Types) != 1 || task.Types[0] != "Delay" {
		t.Fatalf("task %+v", task)
	}
	if task.Agent != "Main" {
		t.Fatalf("agent %s", task.Agent)
	}
}

func TestPlaceQueueCapacity(t *testing.T) {
	task := fromRules(Input{Text: "加一个队列，容量 20", Page: "Main"}, nil)
	if task.Intent != "place" || task.Types[0] != "Queue" || task.Props["capacity"] != 20 {
		t.Fatalf("task %+v", task)
	}
}

func TestCrossDomainDoesNotWrite(t *testing.T) {
	task := fromRules(Input{Text: "加一个 Delay 再加一个行人源", Page: "Main"}, nil)
	if task.Status != "未写入" || task.Intent != "clarify" {
		t.Fatalf("task %+v", task)
	}
}

func TestVagueDoesNotWrite(t *testing.T) {
	task := fromRules(Input{Text: "优化一下", Page: "Main"}, nil)
	if task.Status != "未写入" {
		t.Fatalf("task %+v", task)
	}
}

func TestCurrentPageIsNotAnAgentName(t *testing.T) {
	in := Input{Text: "加一个 Delay", Page: "Main", Agents: []agent{{Name: "Main"}}}
	task := withPage(Task{Intent: "place", Agent: "当前页", Types: []string{"Delay"}}, in)
	if task.Agent != "Main" {
		t.Fatalf("agent %s", task.Agent)
	}
}

func TestRepeatStations(t *testing.T) {
	text := "在 Main 上建一条线，模块按顺序连起来，然后运行。Source 进料。每道工序各放一个 Queue 和一个 Delay：来料、清洗、加工。Queue 容量 30，Delay 时间 2 分钟。Sink 离开。"
	task := fromRules(Input{Text: text, Page: "Main"}, nil)
	if task.Status != "已写入" || !task.Chain || !task.Run {
		t.Fatalf("task %+v", task)
	}
	if task.Types[0] != "Source" || task.Types[len(task.Types)-1] != "Sink" {
		t.Fatalf("types %v", task.Types)
	}
	seen := map[string]bool{}
	for _, name := range task.Types {
		seen[name] = true
	}
	for _, name := range []string{"Seize", "Service", "Delay"} {
		if !seen[name] {
			t.Fatalf("missing %s in %v", name, task.Types)
		}
	}
}

func TestLooseOutputPort(t *testing.T) {
	task := fromRules(Input{Text: "selectoutput 的输出口 2没有接上", Page: "Main"}, nil)
	if task.Intent != "tie" || task.Status != "已写入" || task.Agent != "Main" {
		t.Fatalf("%+v", task)
	}
}

func TestSetCapacity(t *testing.T) {
	task := fromRules(Input{Text: "把 delay 的容量改成 8", Page: "Main"}, nil)
	if task.Intent != "set" || task.Key != "capacity" || task.Value != 8 {
		t.Fatalf("task %+v", task)
	}
}
