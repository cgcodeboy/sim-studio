package scenario

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSceneCatalog(t *testing.T) {
	all := append(append(simpleScenes(), mediumScenes()...), complexScenes()...)
	level := map[string]int{}
	seen := map[string]bool{}
	for _, sc := range all {
		if seen[sc.name] || len(sc.turns) < 3 || len(sc.turns) > 16 {
			t.Fatalf("%s 轮数 %d", sc.name, len(sc.turns))
		}
		if sc.level == "复杂" && (len(sc.turns) < 10 || sc.build != nil) {
			t.Fatalf("%s 要从空页面聊满 10 轮，不能事先摆好", sc.name)
		}
		seen[sc.name] = true
		level[sc.level]++
		for _, step := range sc.turns {
			if strings.TrimSpace(step.say) == "" {
				t.Fatalf("%s 有空话", sc.name)
			}
		}
	}
	if len(all) < 10 || level["简单"] < 3 || level["中等"] < 5 || level["复杂"] < 2 {
		t.Fatalf("场景不够：%v", level)
	}
}

func TestHealth(t *testing.T) {
	requireAPI(t)
}

func TestLibraryListsCoreModules(t *testing.T) {
	requireAPI(t)
	status, payload := call(http.MethodGet, "/api/library", nil)
	items := mustDecode[[]struct {
		Name string `json:"name"`
		Kind string `json:"kind"`
	}](t, status, payload, http.StatusOK)
	have := map[string]string{}
	for _, item := range items {
		have[item.Name] = item.Kind
	}
	need := map[string]string{
		"Source": "process", "Delay": "process", "Queue": "process", "Sink": "process",
		"SelectOutput": "process", "Seize": "process", "Release": "process", "ResourcePool": "process",
		"Conveyor": "handling", "TransporterFleet": "handling", "Store": "handling", "Retrieve": "handling",
		"PedSource": "pedestrian", "PedSink": "pedestrian",
		"CarSource": "road", "CarDispose": "road",
		"TrainSource": "rail", "TrainDispose": "rail",
		"FluidSource": "fluid", "Tank": "fluid", "Valve": "fluid", "FluidDispose": "fluid",
	}
	for name, kind := range need {
		if have[name] != kind {
			t.Fatalf("目录里 %s 不是 %s", name, kind)
		}
	}
}

func TestSessionLifecycle(t *testing.T) {
	requireAPI(t)
	project := "basic-session"
	status, payload := call(http.MethodGet, "/api/dialogue/sessions", nil)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("缺工程编号应拒绝，得到 %d %s", status, trim(payload))
	}
	status, payload = call(http.MethodPost, "/api/dialogue/sessions", map[string]string{"project": project})
	created := mustDecode[sessionCard](t, status, payload, http.StatusOK)
	status, payload = call(http.MethodGet, "/api/dialogue/sessions/"+created.ID, nil)
	opened := mustDecode[struct {
		ID    string `json:"id"`
		Turns []any  `json:"turns"`
	}](t, status, payload, http.StatusOK)
	if opened.ID != created.ID || opened.Turns == nil {
		t.Fatalf("读回的会话不对 %+v", opened)
	}
	status, payload = call(http.MethodPut, "/api/dialogue/sessions/"+created.ID, map[string]string{"title": "灌装讨论"})
	renamed := mustDecode[sessionCard](t, status, payload, http.StatusOK)
	if renamed.Title != "灌装讨论" {
		t.Fatalf("改名失败 %+v", renamed)
	}
	status, payload = call(http.MethodGet, "/api/dialogue/sessions?project="+project, nil)
	cards := mustDecode[[]sessionCard](t, status, payload, http.StatusOK)
	found := false
	for _, card := range cards {
		if card.ID == created.ID && card.Title == "灌装讨论" {
			found = true
		}
	}
	if !found {
		t.Fatalf("列表里没有这场会话")
	}
	status, payload = call(http.MethodDelete, "/api/dialogue/sessions/"+created.ID, nil)
	if status != http.StatusOK {
		t.Fatalf("删除失败 %d %s", status, trim(payload))
	}
	status, _ = call(http.MethodGet, "/api/dialogue/sessions/"+created.ID, nil)
	if status != http.StatusNotFound {
		t.Fatalf("删掉之后还能读到 %d", status)
	}
}

func TestTurnRejectsBrokenJSON(t *testing.T) {
	requireAPI(t)
	reqStatus, payload := postRaw("/api/dialogue/turns", "{")
	if reqStatus != http.StatusUnprocessableEntity || !strings.Contains(string(payload), "invalid") {
		t.Fatalf("坏报文应拒绝，得到 %d %s", reqStatus, trim(payload))
	}
}

func TestProjectAndRecordStayClosedWithoutDatabase(t *testing.T) {
	requireAPI(t)
	for _, path := range []string{"/api/projects", "/api/records"} {
		method := http.MethodGet
		if path == "/api/records" {
			method = http.MethodPost
		}
		status, _ := call(method, path, map[string]any{})
		if status != http.StatusNotFound {
			t.Fatalf("%s %s 在数据库没连上时应是 404，得到 %d", method, path, status)
		}
	}
}

func postRaw(path, raw string) (int, []byte) {
	req, err := http.NewRequest(http.MethodPost, apiRoot()+path, strings.NewReader(raw))
	if err != nil {
		return 0, []byte(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer res.Body.Close()
	payload, _ := io.ReadAll(res.Body)
	return res.StatusCode, payload
}
