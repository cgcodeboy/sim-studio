package dialogue

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func complete(ctx context.Context, key, system, user string, maxTokens int) (string, bool) {
	body, err := json.Marshal(chatBody(system, user, maxTokens))
	if err != nil {
		return "", false
	}
	limit := 20 * time.Second
	if maxTokens > 512 {
		limit = 90 * time.Second
	}
	call, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	request, err := http.NewRequestWithContext(call, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", false
	}
	request.Header.Set("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("dialogue: deepseek %v", err)
		return "", false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 300))
		log.Printf("dialogue: deepseek status %d %s", response.StatusCode, strings.TrimSpace(string(detail)))
		return "", false
	}
	log.Printf("dialogue: deepseek")
	var decoded chatReply
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil || len(decoded.Choices) == 0 {
		return "", false
	}
	return decoded.Choices[0].Message.Content, true
}

func acceptModel(content string, known map[string]bool) (Task, bool) {
	var task Task
	if err := json.Unmarshal([]byte(extractJSON(content)), &task); err != nil {
		return Task{}, false
	}
	return normalize(task, known), true
}

func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}

func normalize(task Task, known map[string]bool) Task {
	if task.Intent != "place" {
		return finish(task)
	}
	byCanon := map[string]block{}
	for _, item := range active(known) {
		byCanon[item.canon()] = item
		byCanon[item.display] = item
	}
	var types []string
	var family string
	for _, name := range task.Types {
		item, ok := byCanon[strings.ReplaceAll(name, " ", "")]
		if !ok {
			item, ok = byCanon[name]
		}
		if !ok {
			continue
		}
		if family == "" {
			family = item.family
		}
		if family != item.family {
			continue
		}
		types = append(types, item.display)
	}
	if len(types) == 0 {
		return clarify("")
	}
	task.Types = types
	return finish(task)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
	Thinking    struct {
		Type string `json:"type"`
	} `json:"thinking"`
}

type chatReply struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func chatBody(system, user string, maxTokens int) chatRequest {
	var request chatRequest
	request.Model = "deepseek-flash"
	request.Temperature = 0
	request.MaxTokens = maxTokens
	request.Thinking.Type = "disabled"
	request.Messages = []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
	return request
}
