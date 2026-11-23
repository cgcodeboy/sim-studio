package dialogue

import (
	"context"
	"strings"
)

func executeAll(ctx context.Context, call skillCall, planned plan) (Task, bool) {
	var actions []Task
	var batch []planStep
	flush := func() bool {
		if len(batch) == 0 {
			return true
		}
		task, ok := executePlan(ctx, call, plan{Act: "build", Steps: batch, Accept: planned.Accept})
		batch = nil
		if !ok {
			return false
		}
		actions = append(actions, task)
		return true
	}
	for _, step := range planned.Steps {
		if isChange(step.Intent) {
			task, ok := directAction(step, call.in.Page, planned.Gap)
			if !ok {
				task, ok = changeAction(ctx, call, step)
			}
			if !ok || !flush() {
				return Task{}, false
			}
			actions = append(actions, task)
			continue
		}
		if step.Intent == "layout" {
			if !flush() {
				return Task{}, false
			}
			task, ok := layPage(ctx, call, planned, placedTypes(actions))
			if !ok {
				return Task{}, false
			}
			actions = append(actions, task)
			continue
		}
		batch = append(batch, step)
	}
	if !flush() {
		return Task{}, false
	}
	return pack(actions, planned.Say), len(actions) > 0
}

func isChange(intent string) bool {
	switch intent {
	case "set", "link", "remove", "tie":
		return true
	default:
		return false
	}
}

func directAction(step planStep, page string, gap int) (Task, bool) {
	if step.Intent == "layout" {
		return finish(Task{Intent: "layout", Agent: page, Gap: layoutGap(gap)}), true
	}
	task, ok := taskFromStep(step, page)
	if !ok {
		return Task{}, false
	}
	return finish(task), true
}

func changeAction(ctx context.Context, call skillCall, step planStep) (Task, bool) {
	switch step.Intent {
	case "set", "link", "remove", "tie":
	default:
		return Task{}, false
	}
	task, ok := askSpecialist(ctx, call, "change", stepUser(call.in, step))
	if !ok || task.Status != "已写入" || task.Intent == "place" {
		return Task{}, false
	}
	return task, true
}

func pack(actions []Task, say string) Task {
	if len(actions) == 0 {
		return Task{}
	}
	head := actions[0]
	if len(actions) > 1 {
		head.More = actions[1:]
	}
	if strings.TrimSpace(say) != "" {
		head.Reply = strings.TrimSpace(say)
	}
	return head
}

func hasOtherWork(planned plan) bool {
	for _, step := range planned.Steps {
		if len(step.Types) > 0 {
			return true
		}
		switch step.Intent {
		case "", "layout", "run", "pause", "stop":
		default:
			return true
		}
	}
	return false
}

func refuseTask(reason string) Task {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "这一轮没有形成可执行的改动。"
	}
	if !strings.HasPrefix(reason, "做不到") {
		reason = "做不到。" + reason
	}
	return clarify(reason)
}
