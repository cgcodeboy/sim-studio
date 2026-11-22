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
	return runLoop(ctx, skillCall{key: key, known: known, in: in})
}

func askSpecialist(ctx context.Context, call skillCall, name, user string) (Task, bool) {
	prompt, ok := skillPrompt(name, call.known)
	if !ok {
		return Task{}, false
	}
	content, ok := complete(ctx, call.key, prompt, user, 4096)
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
	text := strings.ReplaceAll(string(raw), "{{TYPES}}", typeList(name, known))
	return strings.ReplaceAll(text, "{{CATALOG}}", catalogByFamily(known)), true
}

func catalogByFamily(known map[string]bool) string {
	var order []string
	seen := map[string]bool{}
	grouped := map[string][]string{}
	for _, item := range active(known) {
		if !seen[item.family] {
			seen[item.family] = true
			order = append(order, item.family)
		}
		grouped[item.family] = append(grouped[item.family], item.display)
	}
	var lines []string
	for _, family := range order {
		lines = append(lines, family+"："+strings.Join(grouped[family], "、"))
	}
	return strings.Join(lines, "\n")
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
	if in.Memory != "" {
		text.WriteString(in.Memory)
	}
	text.WriteString(in.Text)
	return text.String()
}
