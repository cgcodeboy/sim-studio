package dialogue

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	setPattern    = regexp.MustCompile(`把\s*(\S+?)\s*的\s*(容量|时间|延迟|到达)\s*(?:改成|设成|设为)\s*(\S+)`)
	linkPattern   = regexp.MustCompile(`(?:把\s*)?(\S+?)\s*连到\s*(\S+)`)
	removePattern = regexp.MustCompile(`(?:删除|去掉)\s*(\S+)`)
	agentPattern  = regexp.MustCompile(`在\s*(\S+?)\s*(?:上|里|中)`)
	numberPattern = regexp.MustCompile(`容量\s*(\d+)`)
	delayPattern  = regexp.MustCompile(`(?:延迟|时间)\s*([0-9]+(?:\.[0-9]+)?\s*(?:秒|分钟|小时)?)`)
)

func fromRules(in Input, known map[string]bool) Task {
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return clarify("先写要做的事。")
	}
	if loosePorts(text) {
		return finish(Task{Intent: "tie", Agent: pickAgent(in, text)})
	}
	if types := repeatStations(text); len(types) > 0 {
		return finish(Task{Intent: "place", Agent: pickAgent(in, text), Types: types, Props: propsOf(text), Chain: strings.Contains(text, "连"), Run: wantsRun(text)})
	}
	if task, ok := editTask(text, in); ok {
		return task
	}
	if strings.Contains(text, "暂停") {
		return finish(Task{Intent: "pause"})
	}
	if strings.Contains(text, "停止") && !strings.Contains(text, "停止时间") {
		return finish(Task{Intent: "stop"})
	}
	found := mention(text, active(known))
	if len(familiesOf(found)) > 1 {
		return clarify("这句话跨了两个领域。先说要加哪一种模块？")
	}
	if len(found) > 0 && wantsPlace(text) {
		return placeTask(in, text, found)
	}
	if wantsRun(text) && len(found) == 0 {
		return finish(Task{Intent: "run"})
	}
	return clarify("")
}

func editTask(text string, in Input) (Task, bool) {
	if match := setPattern.FindStringSubmatch(text); match != nil {
		return finish(Task{
			Intent:  "set",
			Agent:   pickAgent(in, text),
			Element: clean(match[1]),
			Key:     propKey(match[2]),
			Value:   propValue(match[2], clean(match[3])),
		}), true
	}
	if match := linkPattern.FindStringSubmatch(text); match != nil && !wantsPlace(text) {
		return finish(Task{Intent: "link", Agent: pickAgent(in, text), From: clean(match[1]), To: clean(match[2])}), true
	}
	if match := removePattern.FindStringSubmatch(text); match != nil {
		return finish(Task{Intent: "remove", Agent: pickAgent(in, text), Element: clean(match[1])}), true
	}
	return Task{}, false
}

func placeTask(in Input, text string, found []block) Task {
	types := make([]string, 0, len(found))
	for _, item := range found {
		types = append(types, item.display)
	}
	task := Task{Intent: "place", Agent: pickAgent(in, text), Types: types, Props: propsOf(text), Chain: len(types) > 1 && strings.Contains(text, "连"), Run: wantsRun(text)}
	return finish(task)
}

func propsOf(text string) map[string]any {
	props := map[string]any{}
	if match := numberPattern.FindStringSubmatch(text); match != nil {
		number, _ := strconv.Atoi(match[1])
		props["capacity"] = number
	}
	if match := delayPattern.FindStringSubmatch(text); match != nil {
		props["delay"] = strings.TrimSpace(match[1])
	}
	if len(props) == 0 {
		return nil
	}
	return props
}

func familiesOf(found []block) map[string]bool {
	out := map[string]bool{}
	for _, item := range found {
		out[item.family] = true
	}
	return out
}

func loosePorts(text string) bool {
	return strings.Contains(text, "没有接上") || (strings.Contains(text, "输出口") && strings.Contains(text, "接"))
}

func wantsPlace(text string) bool {
	return strings.Contains(text, "加") || strings.Contains(text, "放") || strings.Contains(text, "摆")
}

func wantsRun(text string) bool {
	return strings.Contains(text, "运行") || strings.Contains(text, "跑起来")
}

func pickAgent(in Input, text string) string {
	if match := agentPattern.FindStringSubmatch(text); match != nil {
		name := clean(match[1])
		for _, item := range in.Agents {
			if strings.EqualFold(item.Name, name) {
				return item.Name
			}
		}
	}
	if in.Page != "" {
		return in.Page
	}
	if len(in.Agents) > 0 {
		return in.Agents[0].Name
	}
	return "Main"
}

func propKey(label string) string {
	switch label {
	case "容量":
		return "capacity"
	case "到达":
		return "rate"
	default:
		return "delay"
	}
}

func propValue(label, raw string) any {
	if label != "容量" {
		return raw
	}
	number, err := strconv.Atoi(raw)
	if err != nil {
		return raw
	}
	return number
}

func clean(value string) string {
	return strings.Trim(value, "，。,.的 ")
}
