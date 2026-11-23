package dialogue

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

const planRounds = 3

type planStep struct {
	Domain  string   `json:"domain"`
	Intent  string   `json:"intent"`
	Work    string   `json:"work"`
	Types   []string `json:"types"`
	Element string   `json:"element"`
	Key     string   `json:"key"`
	Value   any      `json:"value"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	Chain   bool     `json:"chain"`
	Run     bool     `json:"run"`
}

type plan struct {
	Act    string     `json:"act"`
	Say    string     `json:"say"`
	Steps  []planStep `json:"steps"`
	Accept string     `json:"accept"`
	Gap    int        `json:"gap,omitempty"`
}

type verdict struct {
	OK   bool     `json:"ok"`
	Gaps []string `json:"gaps"`
}

func runLoop(ctx context.Context, call skillCall) (Task, bool) {
	var gaps string
	var last Task
	var has bool
	for round := 1; round <= planRounds; round++ {
		planned, ok := makePlan(ctx, call, gaps, last)
		if !ok {
			continue
		}
		if planned.Act == "layout" && !hasOtherWork(planned) {
			task, laid := layPage(ctx, call, planned, nil)
			if laid {
				return task, true
			}
			gaps = "这不是工序缺口。请再输出 act 为 layout，不要新增模块。由布局专员给出互不重叠的平面位置和三维姿态。"
			continue
		}
		if task, done := readyTask(planned, call.in.Page); done {
			return task, true
		}
		if len(planned.Steps) == 0 {
			gaps = "没有动作。缺决定用 guide，只问一个问题。做不到用 refuse，say 以做不到开头并写理由。"
			continue
		}
		task, ok := executeAll(ctx, call, planned)
		if !ok {
			gaps = "动作没有做完。缺的决定用 guide 问，不要只写说明。"
			continue
		}
		last, has = task, true
		if len(task.More) > 0 || task.Intent != "place" || task.Status != "已写入" {
			return task, true
		}
		passed, next := review(ctx, call, planned, task)
		if passed {
			log.Printf("dialogue: check round %d ok", round)
			return task, true
		}
		log.Printf("dialogue: check round %d %s", round, clip(next))
		gaps = next
	}
	if !has {
		return refuseTask("这一轮没有形成可执行的改动。"), true
	}
	return last, true
}

func makePlan(ctx context.Context, call skillCall, gaps string, last Task) (plan, bool) {
	prompt, ok := skillPrompt("planner", call.known)
	if !ok {
		return plan{}, false
	}
	content, ok := complete(ctx, call.key, prompt, planUser(call.in, gaps, last), 4096)
	if !ok {
		return plan{}, false
	}
	planned := parsePlan(content)
	log.Printf("dialogue: plan steps %d types %d", len(planned.Steps), planTypes(planned))
	return planned, true
}

func executePlan(ctx context.Context, call skillCall, planned plan) (Task, bool) {
	var types []string
	var placed Task
	var hasPlace bool
	for _, step := range planned.Steps {
		if !knownSpecialist(step.Domain) {
			continue
		}
		task, ok := askSpecialist(ctx, call, step.Domain, stepUser(call.in, step))
		if !ok {
			continue
		}
		if task.Intent == "place" && task.Status == "已写入" {
			types = append(types, task.Types...)
			if !hasPlace {
				placed = task
			}
			placed.Chain = placed.Chain || task.Chain || step.Chain
			placed.Run = placed.Run || task.Run || step.Run
			hasPlace = true
			continue
		}
		if !hasPlace && task.Status == "已写入" {
			return task, true
		}
	}
	if !hasPlace {
		return Task{}, false
	}
	placed.Intent = "place"
	placed.Types = types
	return finish(placed), true
}

func review(ctx context.Context, call skillCall, planned plan, task Task) (bool, string) {
	prompt, ok := skillPrompt("checker", nil)
	if !ok {
		return false, "检查员没有就绪"
	}
	content, ok := complete(ctx, call.key, prompt, checkUser(call.in, planned.Accept, task), 512)
	if !ok {
		return false, "检查没有返回结果"
	}
	result := parseVerdict(content)
	if result.OK {
		return true, ""
	}
	gaps := strings.Join(result.Gaps, "；")
	if gaps == "" {
		gaps = "结果还没有覆盖需求"
	}
	return false, gaps
}

func planTypes(planned plan) int {
	n := 0
	for _, step := range planned.Steps {
		n += len(step.Types)
	}
	return n
}

func parsePlan(content string) plan {
	var planned plan
	if err := json.Unmarshal([]byte(extractJSON(content)), &planned); err != nil {
		return plan{}
	}
	planned.Act = strings.ToLower(strings.TrimSpace(planned.Act))
	var steps []planStep
	for _, step := range planned.Steps {
		step.Domain = strings.ToLower(strings.TrimSpace(step.Domain))
		step.Intent = strings.ToLower(strings.TrimSpace(step.Intent))
		if (planned.Act == "" || planned.Act == "build") && step.Domain != "" && !knownSpecialist(step.Domain) {
			continue
		}
		steps = append(steps, step)
	}
	planned.Steps = steps
	return planned
}

func parseVerdict(content string) verdict {
	var result verdict
	if err := json.Unmarshal([]byte(extractJSON(content)), &result); err != nil {
		return verdict{}
	}
	return result
}

func planUser(in Input, gaps string, last Task) string {
	text := skillUser(in) + "\n只输出 JSON。"
	if gaps == "" {
		return text
	}
	text += "\n上一轮做出的模块：" + strings.Join(last.Types, "、")
	text += "\n检查没有通过：" + gaps
	return text + "\n针对缺的部分重新制定计划。"
}

func stepUser(in Input, step planStep) string {
	text := skillUser(in) + "\n按下面的计划执行这一步，不要向用户追问模块类型。\n"
	text += "要做的事：" + step.Work + "\n"
	if len(step.Types) > 0 {
		text += "计划中的模块顺序：" + strings.Join(step.Types, "、") + "\n"
	}
	if step.Chain {
		text += "chain 为 true。\n"
	}
	if step.Run {
		text += "run 为 true。\n"
	}
	if step.Intent != "" {
		text += "intent 用 " + step.Intent + "。\n"
	}
	return text
}

func checkUser(in Input, accept string, task Task) string {
	return "需求：" + in.Text + "\n完成标准：" + accept + "\n实际模块：" + strings.Join(task.Types, "、") + "\nchain：" + strconv.FormatBool(task.Chain) + "\n只输出 JSON。"
}

func readyTask(planned plan, page string) (Task, bool) {
	switch planned.Act {
	case "guide":
		return clarify(spoken(planned.Say, "你已经说了一部分。还缺哪一个决定？")), true
	case "refuse":
		return refuseTask(planned.Say), true
	case "query":
		return clarify(spoken(planned.Say, "这句没读成可执行的计划。")), true
	case "control":
		if hasOtherWork(planned) {
			return Task{}, false
		}
		intent := "run"
		if len(planned.Steps) > 0 && controlIntent(planned.Steps[0].Intent) {
			intent = planned.Steps[0].Intent
		}
		return finish(Task{Intent: intent, Agent: page}), true
	default:
		return Task{}, false
	}
}

func taskFromStep(step planStep, page string) (Task, bool) {
	task := Task{Intent: step.Intent, Agent: page, Element: step.Element, Key: step.Key, Value: step.Value, From: step.From, To: step.To}
	switch step.Intent {
	case "set":
		if step.Element == "" || step.Key == "" || step.Value == nil {
			return Task{}, false
		}
	case "link":
		if step.From == "" || step.To == "" {
			return Task{}, false
		}
	case "remove":
		if step.Element == "" {
			return Task{}, false
		}
	case "tie":
	default:
		return Task{}, false
	}
	return task, true
}

func controlIntent(intent string) bool {
	return intent == "run" || intent == "pause" || intent == "stop"
}

func layoutGap(gap int) int {
	if gap < 4 {
		return 8
	}
	if gap > 40 {
		return 40
	}
	return gap
}

func spoken(say, fallback string) string {
	say = strings.TrimSpace(say)
	if say == "" {
		return fallback
	}
	return say
}
