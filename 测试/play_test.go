package scenario

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func play(t *testing.T, sc scene) {
	t.Helper()
	requireAPI(t)
	project := fmt.Sprintf("%s-%d", sc.name, time.Now().UnixNano())
	status, payload := call(http.MethodPost, "/api/dialogue/sessions", map[string]string{"project": project})
	session := mustDecode[sessionCard](t, status, payload, http.StatusOK)
	if session.ID == "" {
		t.Fatal("会话没有编号")
	}
	desk := newBoard(project)
	if sc.build != nil {
		sc.build(desk)
	}
	fmt.Printf("\n== %s · %s ==\n", sc.level, sc.name)
	for i, step := range sc.turns {
		before := desk.count()
		status, payload = call(http.MethodPost, "/api/dialogue/turns", desk.outline(step.say, session.ID))
		task := mustDecode[task](t, status, payload, http.StatusOK)
		fmt.Printf("第%d轮 我说：%s\n返回：%s %s %s\n", i+1, step.say, task.Status, task.Intent, oneLine(task.Reply))
		desk.apply(task)
		fmt.Printf("现场：%s\n", desk.brief())
		step.want.verify(t, task, desk, before)
	}
}

func oneLine(text string) string {
	text = stringsCut(text)
	if text == "" {
		return "（没有回复）"
	}
	return text
}

func stringsCut(text string) string {
	out := make([]rune, 0, len(text))
	for _, r := range text {
		if r == '\n' || r == '\r' {
			r = ' '
		}
		out = append(out, r)
		if len(out) == 180 {
			return string(out) + "…"
		}
	}
	return string(out)
}
