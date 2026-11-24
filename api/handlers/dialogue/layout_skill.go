package dialogue

import (
	"context"
	"math"
	"sort"
	"strings"
)

func layoutFromSkill(ctx context.Context, call skillCall, gap int, extra []string) (Task, bool) {
	if len(roster(openAgent(call.in), extra)) >= 80 {
		return fitted(call.in, gap, extra), true
	}
	return reviseLayout(ctx, call, gap, extra)
}

func askLayout(ctx context.Context, call skillCall, user string, extra []string) (Task, string) {
	task, ok := askSpecialist(ctx, call, "layout", user)
	if !ok {
		return Task{}, "布局专员没有返回"
	}
	task.Places = bindPlaces(call.in, task.Places, extra)
	if why := rejectPlaces(call.in, task, extra); why != "" && !softOverlap(why) {
		return task, why
	}
	return task, ""
}

func softOverlap(why string) bool {
	return why == "三维占地重叠" || why == "平面卡片重叠" || why == "模块堆在同一处"
}

func settle(in Input, spots []Place, extra []string) []Place {
	opened := copyPlaces(spots)
	members := roster(openAgent(in), extra)
	for pass := 0; pass < 8; pass++ {
		if !nudgeOnce(opened, members) {
			break
		}
	}
	return opened
}

func copyPlaces(spots []Place) []Place {
	out := make([]Place, len(spots))
	for i, spot := range spots {
		out[i] = spot
		if spot.Pose != nil {
			pose := *spot.Pose
			out[i].Pose = &pose
		}
	}
	return out
}

func nudgeOnce(out []Place, members map[string]member) bool {
	moved := false
	for i := range out {
		for j := 0; j < i; j++ {
			if pushPose(out, i, j) || pushCard(out, members, i, j) {
				moved = true
			}
		}
	}
	return moved
}

func pushPose(out []Place, i, j int) bool {
	if !crowded(out[i], out[j]) {
		return false
	}
	ax0, ax1, az0, az1 := footprint(out[i].Pose)
	bx0, bx1, bz0, bz1 := footprint(out[j].Pose)
	ox := math.Min(ax1, bx1) - math.Max(ax0, bx0)
	oz := math.Min(az1, bz1) - math.Max(az0, bz0)
	if ox <= oz {
		out[i].Pose.X = round3(out[i].Pose.X + pushSign(out[i].Pose.X, out[j].Pose.X)*(ox+0.2))
		return true
	}
	out[i].Pose.Z = round3(out[i].Pose.Z + pushSign(out[i].Pose.Z, out[j].Pose.Z)*(oz+0.2))
	return true
}

func pushCard(out []Place, members map[string]member, i, j int) bool {
	lw, lh := cardSize(out[i], members)
	rw, rh := cardSize(out[j], members)
	ox := math.Min(out[i].X+lw, out[j].X+rw) - math.Max(out[i].X, out[j].X)
	oy := math.Min(out[i].Y+lh, out[j].Y+rh) - math.Max(out[i].Y, out[j].Y)
	if ox <= 1 || oy <= 1 {
		return false
	}
	if ox <= oy {
		out[i].X = snap(out[i].X + pushSign(out[i].X, out[j].X)*(ox+48))
		return true
	}
	out[i].Y = snap(out[i].Y + pushSign(out[i].Y, out[j].Y)*(oy+48))
	return true
}

func pushSign(a, b float64) float64 {
	if a < b {
		return -1
	}
	return 1
}

func crowded(a, b Place) bool {
	if a.Pose == nil || b.Pose == nil || beltMeets(a, b) {
		return false
	}
	return boxesHit(a.Pose, b.Pose)
}

func boxesHit(a, b *PoseSpot) bool {
	ax0, ax1, az0, az1 := footprint(a)
	bx0, bx1, bz0, bz1 := footprint(b)
	return ax0 < bx1-0.05 && bx0 < ax1-0.05 && az0 < bz1-0.05 && bz0 < az1-0.05
}

func take(members map[string]member, seen map[string]bool, name string) (member, string, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	if item, ok := members[key]; ok {
		if seen[key] {
			return member{}, key, false
		}
		return item, key, true
	}
	foundKey := ""
	var found member
	for candidate, item := range members {
		if seen[candidate] || !sameKind(item.kind, name) {
			continue
		}
		if foundKey != "" {
			return member{}, key, false
		}
		found, foundKey = item, candidate
	}
	if foundKey == "" {
		return member{}, key, false
	}
	return found, foundKey, true
}

func sameKind(kind, name string) bool {
	return strings.EqualFold(kindOf(name), kind)
}

func kindOf(name string) string {
	best, bestN := "", 0
	for _, item := range blocks {
		terms := append([]string{item.display}, item.aliases...)
		for _, term := range terms {
			if term == "" {
				continue
			}
			if strings.EqualFold(term, name) || strings.Contains(strings.ToLower(name), strings.ToLower(term)) {
				if len(term) > bestN {
					best, bestN = item.display, len(term)
				}
			}
		}
	}
	return best
}

func bindPlaces(in Input, spots []Place, extra []string) []Place {
	members := roster(openAgent(in), extra)
	seen := map[string]bool{}
	var out, pending []Place
	for _, spot := range spots {
		if item, key, ok := take(members, seen, spot.Element); ok {
			spot.Element = item.name
			seen[key] = true
			out = append(out, spot)
			continue
		}
		pending = append(pending, spot)
	}
	out = append(out, assignKind(members, seen, pending)...)
	for key, item := range members {
		if seen[key] {
			continue
		}
		out = append(out, Place{Element: item.name, X: item.x, Y: item.y, Pose: clonePose(item.pose)})
		seen[key] = true
	}
	return out
}

func assignKind(members map[string]member, seen map[string]bool, pending []Place) []Place {
	byKind := map[string][]string{}
	for key, item := range members {
		if !seen[key] {
			byKind[item.kind] = append(byKind[item.kind], key)
		}
	}
	buckets := map[string][]Place{}
	for _, spot := range pending {
		if kind := kindOf(spot.Element); kind != "" {
			buckets[kind] = append(buckets[kind], spot)
		}
	}
	var out []Place
	for kind, spots := range buckets {
		keys := byKind[kind]
		sort.Strings(keys)
		sort.Slice(spots, func(i, j int) bool { return spots[i].Element < spots[j].Element })
		n := len(spots)
		if len(keys) < n {
			n = len(keys)
		}
		for i := 0; i < n; i++ {
			item := members[keys[i]]
			spots[i].Element = item.name
			seen[keys[i]] = true
			out = append(out, spots[i])
		}
	}
	return out
}

func clonePose(pose *PoseSpot) *PoseSpot {
	if pose == nil {
		return &PoseSpot{Len: 1.8}
	}
	next := *pose
	if next.Len <= 0 {
		next.Len = 1.8
	}
	return &next
}

func separate(spots []Place, members map[string]member) []Place {
	out := make([]Place, len(spots))
	for i, spot := range spots {
		out[i] = spot
		if spot.Pose != nil {
			pose := *spot.Pose
			out[i].Pose = &pose
		}
	}
	spread(out, 1.2, 0.2, func(i int) float64 { return out[i].Pose.Z }, func(i int) (float64, float64) {
		x0, x1, _, _ := footprint(out[i].Pose)
		return x0, x1
	}, func(i int, delta float64) { out[i].Pose.X += delta })
	spread(out, 40, 48, func(i int) float64 { return out[i].Y }, func(i int) (float64, float64) {
		w, _ := cardSize(out[i], members)
		return out[i].X, out[i].X + w
	}, func(i int, delta float64) { out[i].X += delta })
	return out
}

func spread(out []Place, band, pad float64, key func(int) float64, span func(int) (float64, float64), move func(int, float64)) {
	used := make([]bool, len(out))
	for i := range out {
		if used[i] || out[i].Pose == nil {
			continue
		}
		group := []int{}
		for j := range out {
			if out[j].Pose == nil || used[j] || math.Abs(key(i)-key(j)) > band {
				continue
			}
			used[j] = true
			group = append(group, j)
		}
		sort.Slice(group, func(a, b int) bool {
			as, _ := span(group[a])
			bs, _ := span(group[b])
			return as < bs
		})
		_, cursor := span(group[0])
		cursor += pad
		for _, id := range group[1:] {
			start, end := span(id)
			if start < cursor {
				move(id, cursor-start)
				start, end = span(id)
			}
			cursor = end + pad
		}
	}
}

func rejectPlaces(in Input, task Task, extra []string) string {
	members := roster(openAgent(in), extra)
	if len(members) == 0 {
		return "页面上还没有模块"
	}
	seen := map[string]bool{}
	var spots []Place
	for _, spot := range task.Places {
		item, key, ok := take(members, seen, spot.Element)
		if !ok {
			return spot.Element + " 不能归位"
		}
		seen[key] = true
		if spot.Pose == nil || spot.Pose.Len <= 0 || !rightAngle(spot.Pose.Yaw) {
			return item.name + " 缺少可用的三维姿态"
		}
		spots = append(spots, spot)
	}
	for key, item := range members {
		if !seen[key] {
			return item.name + " 没有归位"
		}
	}
	if piled(spots) {
		return "模块堆在同一处"
	}
	if cardsOverlap(spots, members) {
		return "平面卡片重叠"
	}
	if footprintsOverlap(spots) {
		return "三维占地重叠"
	}
	return ""
}

func rightAngle(yaw float64) bool {
	turn := math.Mod(math.Abs(yaw), math.Pi/2)
	if turn > math.Pi/4 {
		turn = math.Pi/2 - turn
	}
	return turn < 0.08
}

func piled(spots []Place) bool {
	if len(spots) < 2 {
		return false
	}
	first := spots[0].Pose
	for _, spot := range spots[1:] {
		if math.Abs(spot.Pose.X-first.X) > 0.5 || math.Abs(spot.Pose.Z-first.Z) > 0.5 {
			return false
		}
	}
	return true
}

func cardsOverlap(spots []Place, members map[string]member) bool {
	for i, left := range spots {
		lw, lh := cardSize(left, members)
		for _, right := range spots[i+1:] {
			rw, rh := cardSize(right, members)
			if left.X < right.X+rw-1 && right.X < left.X+lw-1 && left.Y < right.Y+rh-1 && right.Y < left.Y+lh-1 {
				return true
			}
		}
	}
	return false
}

func cardSize(spot Place, members map[string]member) (float64, float64) {
	w, h := card(members[strings.ToLower(spot.Element)])
	if spot.W > 0 {
		w = spot.W
	}
	return w, h
}

func footprintsOverlap(spots []Place) bool {
	for i, left := range spots {
		for _, right := range spots[i+1:] {
			if crowded(left, right) {
				return true
			}
		}
	}
	return false
}

func beltMeets(left, right Place) bool {
	return isSpan(kindOf(left.Element)) != isSpan(kindOf(right.Element))
}

func footprint(pose *PoseSpot) (float64, float64, float64, float64) {
	hx, hz := pose.Len/2, 0.5
	if math.Abs(math.Sin(pose.Yaw)) > 0.5 {
		hx, hz = 0.5, pose.Len/2
	}
	return pose.X - hx, pose.X + hx, pose.Z - hz, pose.Z + hz
}
