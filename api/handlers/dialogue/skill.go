package dialogue

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"strings"
)

//go:embed skills/*.txt
var skillFiles embed.FS

var specialists = []string{
	"process", "handling", "pedestrian", "road", "rail", "fluid", "dynamics",
	"behavior", "experiment", "library", "project", "simulation",
}

type skillCall struct {
	key   string
	known map[string]bool
	in    Input
}

func askSkills(ctx context.Context, key string, known map[string]bool, in Input) (Task, bool) {
	names, ok := routeTo(ctx, key, in.Text)
	if !ok {
		return Task{}, false
	}
	if len(names) == 0 {
		return clarify("这句话需要先指定一个领域，工程没有改。"), true
	}
	return mergeSkills(ctx, skillCall{key: key, known: known, in: in}, names)
}

func routeTo(ctx context.Context, key, text string) ([]string, bool) {
	prompt, ok := skillPrompt("router", nil)
	if !ok {
		return nil, false
	}
	content, ok := complete(ctx, key, prompt, text+"\n按系统要求只输出 JSON。", 256)
	if !ok {
		return nil, false
	}
	names := specialistNames(content)
	if len(names) == 0 {
		log.Printf("dialogue: route empty %s", clip(content))
		return nil, true
	}
	log.Printf("dialogue: route %s", strings.Join(names, ","))
	return names, true
}

func mergeSkills(ctx context.Context, call skillCall, names []string) (Task, bool) {
	var types []string
	var placed Task
	var hasPlace bool
	for _, name := range names {
		task, ok := askSpecialist(ctx, call, name)
		if !ok {
			continue
		}
		if task.Intent == "place" && task.Status == "已写入" {
			types = append(types, task.Types...)
			if !hasPlace {
				placed = task
			}
			placed.Chain = placed.Chain || task.Chain
			placed.Run = placed.Run || task.Run
			hasPlace = true
			continue
		}
		if !hasPlace && task.Status == "已写入" {
			return task, true
		}
	}
	if !hasPlace {
		return clarify("模型没有给出可执行的结果，工程没有改。"), true
	}
	placed.Intent = "place"
	placed.Types = types
	return finish(placed), true
}

func askSpecialist(ctx context.Context, call skillCall, name string) (Task, bool) {
	prompt, ok := skillPrompt(name, call.known)
	if !ok {
		return Task{}, false
	}
	content, ok := complete(ctx, call.key, prompt, skillUser(call.in), 4096)
	if !ok {
		return Task{}, false
	}
	log.Printf("dialogue: skill %s", name)
	return acceptModel(content, call.known)
}

func skillPrompt(name string, known map[string]bool) (string, bool) {
	raw, err := skillFiles.ReadFile("skills/" + name + ".txt")
	if err != nil {
		return "", false
	}
	return strings.ReplaceAll(string(raw), "{{TYPES}}", typeList(name, known)), true
}

func typeList(name string, known map[string]bool) string {
	var names []string
	for _, item := range active(known) {
		if name == "library" || item.family == name {
			names = append(names, item.display)
		}
	}
	return strings.Join(names, "、")
}

func specialistNames(content string) []string {
	var routed struct {
		Specialist  string   `json:"specialist"`
		Specialists []string `json:"specialists"`
	}
	if err := json.Unmarshal([]byte(extractJSON(content)), &routed); err != nil {
		return nil
	}
	raw := routed.Specialists
	if len(raw) == 0 && routed.Specialist != "" {
		raw = strings.Split(routed.Specialist, ",")
	}
	var names []string
	seen := map[string]bool{}
	for _, item := range raw {
		name := strings.ToLower(strings.TrimSpace(item))
		if name == "" || name == "clarify" || seen[name] || !knownSpecialist(name) {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func knownSpecialist(name string) bool {
	for _, item := range specialists {
		if item == name {
			return true
		}
	}
	return false
}

func clip(content string) string {
	content = strings.TrimSpace(content)
	if len(content) > 180 {
		return content[:180]
	}
	return content
}

func skillUser(in Input) string {
	var text strings.Builder
	text.WriteString("页面名是 " + in.Page + "。agent 填这个名字，不要写「当前页」。\n")
	if len(in.Agents) > 0 {
		text.WriteString("已有智能体：")
		for i, item := range in.Agents {
			if i > 0 {
				text.WriteString("，")
			}
			text.WriteString(item.Name)
		}
		text.WriteString("。\n")
	}
	text.WriteString(in.Text)
	return text.String()
}
