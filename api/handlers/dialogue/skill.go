package dialogue

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"strconv"
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
	page := in.Page
	if page == "" {
		page = "Main"
	}
	text.WriteString("当前打开的是 " + page + "。没有点名页面时，agent 填这个名字。点名了别的页面，agent 填那个名字。\n")
	for _, item := range in.Agents {
		writeAgent(&text, item)
	}
	if in.Memory != "" {
		text.WriteString(in.Memory)
	}
	text.WriteString(in.Text)
	return text.String()
}

func num(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}

func writeAgent(text *strings.Builder, item agent) {
	text.WriteString(item.Name + "：")
	if len(item.Elements) == 0 {
		text.WriteString("还没有模块。\n")
		return
	}
	for i, element := range item.Elements {
		if i > 0 {
			text.WriteString("、")
		}
		text.WriteString(element.Name + "（" + element.Type)
		if element.W > 0 && element.H > 0 {
			text.WriteString(" " + strconv.Itoa(int(element.W)) + "×" + strconv.Itoa(int(element.H)))
		}
		text.WriteString(" 平面 " + strconv.Itoa(int(element.X)) + "," + strconv.Itoa(int(element.Y)))
		if element.Pose != nil {
			pose := element.Pose
			text.WriteString(" 三维 " + num(pose.X) + "," + num(pose.Y) + "," + num(pose.Z) + " 朝向 " + num(pose.Yaw) + " 长 " + num(pose.Len))
		}
		text.WriteString("）")
	}
	text.WriteString("。\n")
	if len(item.Links) == 0 {
		return
	}
	text.WriteString("连线：")
	for i, link := range item.Links {
		if i > 0 {
			text.WriteString("、")
		}
		text.WriteString(link.From)
		if link.FromPort > 0 {
			text.WriteString(" 出口" + strconv.Itoa(link.FromPort+1))
		}
		text.WriteString("→" + link.To)
	}
	text.WriteString("。\n")
}
