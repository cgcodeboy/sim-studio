package dialogue

import (
	"regexp"
	"strings"
)

var stationList = regexp.MustCompile(`(?s)每道工序各放一个\s*Queue\s*和一个\s*Delay[：:]\s*(.+?)(?:Queue\s*容量|$)`)

func repeatStations(text string) []string {
	match := stationList.FindStringSubmatch(text)
	if match == nil {
		return nil
	}
	var types []string
	if strings.Contains(text, "Source") {
		types = append(types, "Source")
	}
	for _, part := range strings.FieldsFunc(match[1], splitStation) {
		name := strings.TrimSpace(part)
		if name == "" || blockWord(name) {
			continue
		}
		types = append(types, stationRecipe(name)...)
	}
	if strings.Contains(text, "Sink") {
		types = append(types, "Sink")
	}
	if len(types) < 4 {
		return nil
	}
	return types
}

func splitStation(r rune) bool {
	switch r {
	case '、', '，', ',', '。', '；', '\n', '\r':
		return true
	default:
		return false
	}
}

func stationRecipe(name string) []string {
	switch {
	case containsAny(name, "检", "测", "探伤", "点数", "称重"):
		return []string{"Queue", "Select Output"}
	case containsAny(name, "压装", "装配"):
		return []string{"Queue", "Assembler"}
	case containsAny(name, "配组", "装袋", "封箱", "码垛"):
		return []string{"Queue", "Batch"}
	case containsAny(name, "卸", "上料"):
		return []string{"Queue", "Move To"}
	case containsAny(name, "待运", "固化"):
		return []string{"Hold"}
	case containsAny(name, "清洗", "加工"):
		return []string{"Queue", "Service"}
	default:
		return []string{"Seize", "Delay", "Release"}
	}
}

func containsAny(name string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(name, part) {
			return true
		}
	}
	return false
}

func blockWord(name string) bool {
	switch strings.ToLower(name) {
	case "queue", "delay", "source", "sink", "seize", "release":
		return true
	default:
		return strings.Contains(name, "容量") || strings.Contains(name, "分钟")
	}
}
