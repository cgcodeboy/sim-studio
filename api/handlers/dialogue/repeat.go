package dialogue

import (
	"fmt"
	"regexp"
	"strconv"
)

var copyAsk = regexp.MustCompile(`(\d+)\s*(?:条|套)`)

func askedCopies(text string) int {
	found := copyAsk.FindAllStringSubmatch(text, -1)
	if len(found) != 1 {
		return 1
	}
	n, err := strconv.Atoi(found[0][1])
	if err != nil || n < 2 {
		return 1
	}
	if n > 80 {
		return 80
	}
	return n
}

func foldCopies(steps []planStep, text string) []planStep {
	n := askedCopies(text)
	if n < 2 {
		return steps
	}
	var out []planStep
	folded := false
	for _, step := range steps {
		if step.Intent == "layout" {
			continue
		}
		if !copyStep(step) {
			out = append(out, step)
			continue
		}
		if folded {
			continue
		}
		step.Lines = n
		out = append(out, step)
		folded = true
	}
	return out
}

func copyStep(step planStep) bool {
	if isChange(step.Intent) || step.Intent == "layout" || controlIntent(step.Intent) {
		return false
	}
	return len(step.Types) > 0
}

func repeatedPlace(step planStep, page string, known map[string]bool) (Task, bool) {
	task := normalize(Task{Intent: "place", Agent: page, Types: step.Types, Chain: true}, known)
	if task.Status != "已写入" || step.Lines < 2 {
		return Task{}, false
	}
	return expandCopies(task, step.Lines), true
}

func widen(task Task, text string) Task {
	if len(task.More) > 0 {
		return task
	}
	return expandCopies(task, askedCopies(text))
}

func expandCopies(task Task, n int) Task {
	pattern := onePattern(task.Types, n)
	if n < 2 || task.Intent != "place" || task.Status != "已写入" || len(pattern) == 0 {
		return task
	}
	task.Types = pattern
	task.Chain = true
	copies := make([]Task, 0, n-1+len(task.More))
	for i := 1; i < n; i++ {
		copies = append(copies, Task{
			Intent: "place", Status: "已写入", Agent: task.Agent, Chain: true,
			Types: append([]string(nil), pattern...), Props: task.Props,
		})
	}
	task.More = append(copies, task.More...)
	task.Reply = fmt.Sprintf("按这条做了 %d 条，每条 %d 个模块。", n, len(pattern))
	return task
}

func onePattern(types []string, n int) []string {
	if n < 2 || len(types) == 0 {
		return nil
	}
	if len(types) <= 16 {
		return append([]string(nil), types...)
	}
	if len(types)%n != 0 {
		return nil
	}
	size := len(types) / n
	if size == 0 || size > 16 {
		return nil
	}
	return append([]string(nil), types[:size]...)
}
