package dialogue

import (
	"strings"
	"testing"
)

func TestWeldClosesGapsAndParksSeize(t *testing.T) {
	in := Input{Page: "Main", Agents: []agent{{
		Name: "Main",
		Links: []pageLink{
			{From: "queue", To: "seize"},
			{From: "seize", To: "delay"},
			{From: "delay", To: "sink"},
			{From: "check", To: "bad", FromPort: 1},
		},
		Elements: []element{
			{Name: "queue", Type: "Queue", W: 156, H: 84},
			{Name: "seize", Type: "Seize", W: 156, H: 84},
			{Name: "delay", Type: "Delay", W: 156, H: 84},
			{Name: "sink", Type: "Sink", W: 156, H: 84},
			{Name: "check", Type: "Select Output", W: 156, H: 84},
			{Name: "bad", Type: "Sink", W: 156, H: 84},
		},
	}}}
	spots := []Place{
		{Element: "queue", X: 40, Y: 80, Pose: &PoseSpot{X: 1.2, Z: 0, Yaw: 0, Len: 2.2}},
		{Element: "seize", X: 200, Y: 80, Pose: &PoseSpot{X: 4, Z: 0, Yaw: 0, Len: 1.6}},
		{Element: "delay", X: 360, Y: 80, Pose: &PoseSpot{X: 8, Z: 0, Yaw: 0, Len: 2.3}},
		{Element: "sink", X: 560, Y: 80, Pose: &PoseSpot{X: 12, Z: 0, Yaw: 0, Len: 1.12}},
		{Element: "check", X: 40, Y: 300, Pose: &PoseSpot{X: 5, Z: 6, Yaw: 0, Len: 2.2}},
		{Element: "bad", X: 240, Y: 300, Pose: &PoseSpot{X: 5, Z: 10, Yaw: 1.5708, Len: 1.12}},
	}
	task := doneLayout(in, Task{Reply: "物料从左到右，不合格分向一侧。", Places: spots}, 0, nil)
	if strings.Contains(task.Reply, "仍有冲突") || strings.Contains(task.Reply, "没有接上") {
		t.Fatal(task.Reply)
	}
	notes := strings.Join(conflicts(in, task.Places, nil), " ")
	if strings.Contains(notes, "没有接上") || strings.Contains(notes, "挡在") || strings.Contains(notes, "重叠") {
		t.Fatal(notes)
	}
	at := map[string]PoseSpot{}
	for _, spot := range task.Places {
		at[spot.Element] = *spot.Pose
	}
	if at["seize"].Z > -0.8 || at["bad"].Z < at["check"].Z+0.8 {
		t.Fatalf("poses %+v", at)
	}
}
