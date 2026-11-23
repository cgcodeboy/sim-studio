package dialogue

import (
	"context"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/cgcodeboy/sim-studio/api/handlers/respond"
)

type conflictBody struct {
	Input
	Places []Place `json:"places"`
}

func (h *Handler) Conflicts(w http.ResponseWriter, r *http.Request) {
	var body conflictBody
	if err := respond.Decode(w, r, &body); err != nil {
		return
	}
	respond.WriteJSON(w, http.StatusOK, map[string]any{"conflicts": conflicts(body.Input, body.Places, nil)})
}

func reviseLayout(ctx context.Context, call skillCall, gap int, extra []string) (Task, bool) {
	user := layoutUser(call.in, plan{Gap: gap}, extra)
	task, why := askLayout(ctx, call, user, extra)
	if why != "" {
		log.Printf("dialogue: layout %s", why)
		task, why = askLayout(ctx, call, user+"\n上次不合格："+why+"。只输出 JSON。", extra)
	}
	if why != "" || len(task.Places) == 0 {
		return fitted(call.in, gap, extra), true
	}
	for round := 0; round < 2; round++ {
		notes := conflicts(call.in, task.Places, extra)
		if len(notes) == 0 {
			break
		}
		log.Printf("dialogue: conflict %s", strings.Join(notes, "；"))
		next, again := askLayout(ctx, call, user+conflictAsk(notes), extra)
		if again != "" || len(next.Places) == 0 {
			break
		}
		task = next
	}
	return doneLayout(call.in, task, gap, extra), true
}

func fitted(in Input, gap int, extra []string) Task {
	spots := fitScene(in, extra)
	if len(spots) == 0 {
		return refuseTask("页面上还没有可排的模块。")
	}
	return doneLayout(in, Task{Places: spots, Reply: fitReply(in, extra)}, gap, extra)
}

func doneLayout(in Input, task Task, gap int, extra []string) Task {
	page := in.Page
	if page == "" {
		page = "Main"
	}
	notes := conflicts(in, task.Places, extra)
	task.Intent = "layout"
	task.Agent = page
	task.Gap = layoutGap(gap)
	task.Reply = withNotes(task.Reply, notes)
	return finish(task)
}

func conflictAsk(notes []string) string {
	return "\n冲突检测：\n" + strings.Join(notes, "\n") + "\n按这些冲突改位置、朝向、高度和长度。没有点名的模块保持不动。只输出 JSON。"
}

func withNotes(reply string, notes []string) string {
	reply = strings.TrimSpace(reply)
	if len(notes) == 0 {
		if reply == "" {
			return "已按冲突检测调过布局。"
		}
		return reply
	}
	tail := "仍有冲突：" + strings.Join(notes, "；") + "。"
	if reply == "" {
		return tail
	}
	return reply + tail
}

func conflicts(in Input, spots []Place, extra []string) []string {
	members := roster(openAgent(in), extra)
	at := map[string]Place{}
	for _, spot := range spots {
		at[strings.ToLower(spot.Element)] = spot
	}
	notes := overlapNotes(spots)
	notes = append(notes, cardNotes(spots, members)...)
	if page := openAgent(in); page != nil {
		notes = append(notes, gapNotes(page.Links, members, at)...)
		notes = append(notes, blockNotes(page.Links, members, at)...)
	}
	if len(notes) > 12 {
		return notes[:12]
	}
	return notes
}

func overlapNotes(spots []Place) []string {
	var notes []string
	for i, left := range spots {
		for _, right := range spots[i+1:] {
			if left.Pose == nil || right.Pose == nil || !crowded(left, right) {
				continue
			}
			notes = append(notes, left.Element+" 与 "+right.Element+" 三维占地重叠")
		}
	}
	return notes
}

func cardNotes(spots []Place, members map[string]member) []string {
	var notes []string
	for i, left := range spots {
		lw, lh := cardSize(left, members)
		for _, right := range spots[i+1:] {
			rw, rh := cardSize(right, members)
			if left.X < right.X+rw-1 && right.X < left.X+lw-1 && left.Y < right.Y+rh-1 && right.Y < left.Y+lh-1 {
				notes = append(notes, left.Element+" 与 "+right.Element+" 平面卡片重叠")
			}
		}
	}
	return notes
}

func gapNotes(links []pageLink, members map[string]member, at map[string]Place) []string {
	var notes []string
	for _, link := range links {
		if note := linkNote(link, members, at); note != "" {
			notes = append(notes, note)
		}
	}
	return notes
}

func linkNote(link pageLink, members map[string]member, at map[string]Place) string {
	from := members[strings.ToLower(link.From)]
	to := members[strings.ToLower(link.To)]
	left, right := at[strings.ToLower(link.From)], at[strings.ToLower(link.To)]
	if from.kind == "" || to.kind == "" || left.Pose == nil || right.Pose == nil {
		return ""
	}
	if stationKind(from.kind) && isSpan(to.kind) {
		return apart(from.name, to.name, faceOut(left.Pose, from.kind), spanEnd(right.Pose, true))
	}
	if isSpan(from.kind) && stationKind(to.kind) {
		return apart(from.name, to.name, spanEnd(left.Pose, false), faceIn(right.Pose, to.kind))
	}
	if stationKind(from.kind) && stationKind(to.kind) {
		return apart(from.name, to.name, faceOut(left.Pose, from.kind), faceIn(right.Pose, to.kind))
	}
	return ""
}

func apart(from, to string, a, b point) string {
	dist := math.Hypot(a.x-b.x, a.z-b.z)
	if dist <= 1 {
		return ""
	}
	return from + " 到 " + to + " 端面相距 " + num(dist) + " 米，没有接上"
}

func blockNotes(links []pageLink, members map[string]member, at map[string]Place) []string {
	var notes []string
	for key, item := range members {
		if !logicKind(item.kind) {
			continue
		}
		up, down := logicHosts(item.name, links, members)
		if up == "" || down == "" {
			continue
		}
		if blocksPath(at[key], at[strings.ToLower(up)], at[strings.ToLower(down)], members) {
			notes = append(notes, item.name+" 挡在 "+up+" 与 "+down+" 的连线上")
		}
	}
	return notes
}

func blocksPath(logic, up, down Place, members map[string]member) bool {
	if logic.Pose == nil || up.Pose == nil || down.Pose == nil {
		return false
	}
	a := faceOut(up.Pose, members[strings.ToLower(up.Element)].kind)
	b := faceIn(down.Pose, members[strings.ToLower(down.Element)].kind)
	return nearSegment(point{logic.Pose.X, logic.Pose.Z}, a, b)
}

func logicHosts(name string, links []pageLink, members map[string]member) (string, string) {
	var up, down string
	for _, link := range links {
		if strings.EqualFold(link.To, name) {
			if host := solidName(link.From, members); host != "" {
				up = host
			}
		}
		if strings.EqualFold(link.From, name) {
			if host := solidName(link.To, members); host != "" {
				down = host
			}
		}
	}
	return up, down
}

func solidName(name string, members map[string]member) string {
	item := members[strings.ToLower(name)]
	if stationKind(item.kind) {
		return item.name
	}
	return ""
}

func stationKind(kind string) bool {
	return kind != "" && !isSpan(kind) && !logicKind(kind)
}

type point struct{ x, z float64 }

func faceOut(pose *PoseSpot, kind string) point {
	_, front, _ := measure(kind)
	sx, sz := stepOf(pose.Yaw)
	return point{pose.X + sx*front, pose.Z + sz*front}
}

func faceIn(pose *PoseSpot, kind string) point {
	back, _, _ := measure(kind)
	sx, sz := stepOf(pose.Yaw)
	return point{pose.X - sx*back, pose.Z - sz*back}
}

func spanEnd(pose *PoseSpot, upstream bool) point {
	sx, sz := stepOf(pose.Yaw)
	sign := 1.0
	if upstream {
		sign = -1
	}
	return point{pose.X + sign*sx*pose.Len/2, pose.Z + sign*sz*pose.Len/2}
}

func nearSegment(p, a, b point) bool {
	dx, dz := b.x-a.x, b.z-a.z
	span := dx*dx + dz*dz
	if span < 0.01 {
		return math.Hypot(p.x-a.x, p.z-a.z) < 0.8
	}
	t := ((p.x-a.x)*dx + (p.z-a.z)*dz) / span
	if t < 0 || t > 1 {
		return false
	}
	px, pz := a.x+t*dx, a.z+t*dz
	return math.Hypot(p.x-px, p.z-pz) < 0.8
}
