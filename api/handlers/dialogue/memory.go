package dialogue

import (
	"context"
	"log"
	"os"
	"strings"

	convdom "github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

func (h *Handler) recall(ctx context.Context, in Input) string {
	if h.memory == nil || in.Session == "" {
		return ""
	}
	view, err := h.memory.Recall(ctx, in.Project, in.Session)
	if err != nil {
		log.Printf("dialogue: memory %v", err)
		return ""
	}
	return memoryText(view)
}

func (h *Handler) remember(ctx context.Context, in Input, task Task) {
	if h.memory == nil {
		return
	}
	job, err := h.memory.Commit(ctx, convdom.Commit{
		Project: in.Project,
		Session: in.Session,
		Turn:    convdom.Turn{User: in.Text, Reply: task.Reply, Intent: task.Intent, Status: task.Status},
	})
	if err != nil {
		log.Printf("dialogue: memory %v", err)
		return
	}
	if len(job.Turns) == 0 {
		return
	}
	key := os.Getenv("DEEPSEEK_API_KEY")
	summary, ok := summarize(ctx, key, job)
	if !ok {
		return
	}
	if err := h.memory.SaveSummary(ctx, in.Session, summary, job.Covered); err != nil {
		log.Printf("dialogue: memory %v", err)
	}
}

func memoryText(view convdom.View) string {
	var text strings.Builder
	if view.Summary != "" {
		text.WriteString("更早的会话摘要：" + view.Summary + "\n")
	}
	if len(view.Recent) > 0 {
		text.WriteString("最近的对话：\n")
		for _, turn := range view.Recent {
			text.WriteString("用户：" + turn.User + "\n")
			text.WriteString("结果：" + turn.Reply + "\n")
		}
	}
	if len(view.Facts) > 0 {
		text.WriteString("这个工程还成立的事实：\n")
		for _, fact := range view.Facts {
			text.WriteString(fact + "\n")
		}
	}
	return text.String()
}

func summarize(ctx context.Context, key string, job convdom.Job) (string, bool) {
	if key == "" {
		return "", false
	}
	prompt, ok := skillPrompt("summary", nil)
	if !ok {
		return "", false
	}
	content, ok := complete(ctx, key, prompt, summaryUser(job), 512)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(content), content != ""
}

func summaryUser(job convdom.Job) string {
	var text strings.Builder
	if job.Prior != "" {
		text.WriteString("已有摘要：" + job.Prior + "\n")
	}
	for _, turn := range job.Turns {
		text.WriteString("用户：" + turn.User + "\n")
		text.WriteString("结果：" + turn.Reply + "\n")
	}
	text.WriteString("只输出合并后的摘要。")
	return text.String()
}
