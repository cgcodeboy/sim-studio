package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type turnBody struct {
	Text    string     `json:"text"`
	Page    string     `json:"page"`
	Project string     `json:"project"`
	Session string     `json:"session"`
	Agents  []pageBody `json:"agents"`
}

type pageBody struct {
	Name     string     `json:"name"`
	Elements []elemBody `json:"elements"`
	Links    []linkBody `json:"links,omitempty"`
}

type elemBody struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	X    float64   `json:"x,omitempty"`
	Y    float64   `json:"y,omitempty"`
	W    float64   `json:"w,omitempty"`
	H    float64   `json:"h,omitempty"`
	Pose *spotPose `json:"pose,omitempty"`
}

type linkBody struct {
	From     string `json:"from"`
	To       string `json:"to"`
	FromPort int    `json:"fromPort,omitempty"`
}

type spotPose struct {
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Z   float64 `json:"z"`
	Yaw float64 `json:"yaw"`
	Len float64 `json:"len,omitempty"`
}

type task struct {
	Intent  string         `json:"intent"`
	Reply   string         `json:"reply"`
	Status  string         `json:"status"`
	Agent   string         `json:"agent"`
	Types   []string       `json:"types"`
	Props   map[string]any `json:"props"`
	Element string         `json:"element"`
	Key     string         `json:"key"`
	Value   any            `json:"value"`
	From    string         `json:"from"`
	To      string         `json:"to"`
	Chain   bool           `json:"chain"`
	Places  []placeBody    `json:"places"`
	More    []task         `json:"more"`
}

type placeBody struct {
	Element string    `json:"element"`
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
	W       float64   `json:"w"`
	Pose    *spotPose `json:"pose"`
}

type sessionCard struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Title   string `json:"title"`
}

func apiRoot() string {
	if value := strings.TrimSpace(os.Getenv("SIM_API")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "http://127.0.0.1:8080"
}

var httpClient = &http.Client{Timeout: 3 * time.Minute}

func call(method, path string, body any) (int, []byte) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, []byte(err.Error())
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, apiRoot()+path, reader)
	if err != nil {
		return 0, []byte(err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer res.Body.Close()
	payload, _ := io.ReadAll(res.Body)
	return res.StatusCode, payload
}

type fataler interface{ Fatalf(string, ...any) }

func mustDecode[T any](t fataler, status int, payload []byte, want int) T {
	if status != want {
		t.Fatalf("HTTP %d，期望 %d：%s", status, want, trim(payload))
	}
	var dest T
	if err := json.Unmarshal(payload, &dest); err != nil {
		t.Fatalf("解析失败：%s", trim(payload))
	}
	return dest
}

func trim(payload []byte) string {
	text := strings.TrimSpace(string(payload))
	if len(text) > 500 {
		return text[:500]
	}
	return text
}

var live struct {
	once sync.Once
	ok   bool
	why  string
}

func requireAPI(t fataler) {
	live.once.Do(func() {
		status, payload := call(http.MethodGet, "/api/health", nil)
		if status == http.StatusOK && strings.Contains(string(payload), "ok") {
			live.ok = true
			return
		}
		live.why = fmt.Sprintf("对话服务没连上：%d %s", status, trim(payload))
	})
	if !live.ok {
		t.Fatalf("%s", live.why)
	}
}
