package dialogue

import (
	"fmt"
	"strings"
)

// Task 是一轮对话交给页面执行的一件事。未写入时页面不改工程。
type Task struct {
	Intent  string         `json:"intent"`
	Reply   string         `json:"reply"`
	Status  string         `json:"status"`
	Agent   string         `json:"agent,omitempty"`
	Types   []string       `json:"types,omitempty"`
	Props   map[string]any `json:"props,omitempty"`
	Element string         `json:"element,omitempty"`
	Key     string         `json:"key,omitempty"`
	Value   any            `json:"value,omitempty"`
	From    string         `json:"from,omitempty"`
	To      string         `json:"to,omitempty"`
	Chain   bool           `json:"chain,omitempty"`
	Run     bool           `json:"run,omitempty"`
	Gap     int            `json:"gap,omitempty"`
	More    []Task         `json:"more,omitempty"`
}

type element struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type agent struct {
	Name     string    `json:"name"`
	Elements []element `json:"elements"`
}

// Input 是这一轮的原话和当前工程的轮廓。
type Input struct {
	Text    string  `json:"text"`
	Page    string  `json:"page"`
	Project string  `json:"project,omitempty"`
	Session string  `json:"session,omitempty"`
	Agents  []agent `json:"agents"`
	Memory  string  `json:"-"`
}

func finish(task Task) Task {
	switch task.Intent {
	case "place":
		task.Status = "已写入"
		task.Reply = "在 " + task.Agent + " 加了 " + placedNames(task.Types) + "。"
		task.Reply = withProp(task.Reply, task.Props)
		if task.Run {
			task.Reply = strings.TrimSuffix(task.Reply, "。") + "，接着运行。"
		}
	case "set":
		task.Status = "已写入"
		task.Reply = "已把 " + task.Element + " 的 " + propLabel(task.Key) + " 改成 " + fmt.Sprint(task.Value) + "。"
	case "link":
		task.Status = "已写入"
		task.Reply = "已把 " + task.From + " 连到 " + task.To + "。"
	case "remove":
		task.Status = "已写入"
		task.Reply = "已删除 " + task.Element + "。"
	case "run":
		task.Status = "已写入"
		task.Reply = "开始运行。"
	case "pause":
		task.Status = "已写入"
		task.Reply = "暂停运行。"
	case "stop":
		task.Status = "已写入"
		task.Reply = "停止运行。"
	case "tie":
		task.Status = "已写入"
		task.Reply = "已把空着的输出口接到 Sink。"
	case "layout":
		task.Status = "已写入"
		if strings.TrimSpace(task.Reply) == "" {
			task.Reply = "已把模块按流程排开。"
		}
	default:
		task.Intent = "clarify"
		task.Status = "未写入"
		if strings.TrimSpace(task.Reply) == "" {
			task.Reply = "这句话对不上一个领域。要加模块、改属性、连线，还是运行？"
		}
	}
	return task
}

func placedNames(types []string) string {
	if len(types) > 12 {
		return fmt.Sprintf("%d 个模块", len(types))
	}
	return strings.Join(types, "、")
}

func withProp(reply string, props map[string]any) string {
	if len(props) == 0 {
		return reply
	}
	base := strings.TrimSuffix(reply, "。")
	if value, ok := props["capacity"]; ok {
		base += "，容量 " + fmt.Sprint(value)
	}
	if value, ok := props["delay"]; ok && fmt.Sprint(value) != "" {
		base += "，时间 " + fmt.Sprint(value)
	}
	return base + "。"
}

func propLabel(key string) string {
	switch key {
	case "capacity":
		return "容量"
	case "delay":
		return "时间"
	case "rate":
		return "到达"
	default:
		return key
	}
}

func clarify(reply string) Task {
	return finish(Task{Intent: "clarify", Reply: reply})
}
