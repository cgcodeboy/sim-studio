package dialogue

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
)

type member struct {
	name string
	kind string
	w    float64
	h    float64
	x    float64
	y    float64
	pose *PoseSpot
}

func layPage(ctx context.Context, call skillCall, planned plan, extra []string) (Task, bool) {
	return layoutFromSkill(ctx, call, planned.Gap, extra)
}

func wantsLayout(text string) bool {
	for _, word := range []string{"布局", "优化", "排布", "摆放", "排整齐", "排一下", "重排", "断开", "衔接", "位置乱", "乱放"} {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func layoutUser(in Input, planned plan, extra []string) string {
	text := skillUser(in) + "\n只排当前打开的页面。平面间距 " + strconv.Itoa(layoutGap(planned.Gap)) + " 像素。\n"
	if len(extra) > 0 {
		text += "本轮按顺序新增、还没有名字："
		for i, kind := range extra {
			if i > 0 {
				text += "，"
			}
			text += "新" + strconv.Itoa(i+1) + "=" + kind
		}
		text += "。\n"
	}
	return text + businesses(in, extra) + "先看整场业务，再为每个模块定平面 x、y 和三维的位置、朝向、高度、长度。不要套固定形状。places 里每个模块都要有平面 x、y 和完整 pose。只输出 JSON。"
}

func businesses(in Input, extra []string) string {
	var parts []string
	if page := openAgent(in); page != nil {
		for _, item := range page.Elements {
			parts = append(parts, item.Name+"="+businessOf(item.Type))
		}
	}
	for i, kind := range extra {
		parts = append(parts, "新"+strconv.Itoa(i+1)+"="+businessOf(kind))
	}
	if len(parts) == 0 {
		return ""
	}
	return "业务归属：" + strings.Join(parts, "，") + "。\n"
}

func joinScheme(in Input, task Task, extra []string, gap int) ([]Place, string) {
	members := roster(openAgent(in), extra)
	seen := map[string]bool{}
	var line []string
	for _, name := range task.Line {
		item, key, ok := take(members, seen, name)
		if !ok {
			return nil, name + " 不能放进主干"
		}
		seen[key] = true
		line = append(line, item.name)
	}
	if len(line) == 0 {
		return nil, "没有主干"
	}
	for _, dock := range task.Side {
		if _, key, ok := take(members, seen, dock.Element); !ok || !onLine(line, dock.Host) {
			return nil, dock.Element + " 没有靠上主干"
		} else {
			seen[key] = true
		}
	}
	for _, bridge := range task.Span {
		if _, key, ok := take(members, seen, bridge.Element); !ok || !onLine(line, bridge.From) || !onLine(line, bridge.To) {
			return nil, bridge.Element + " 的两端不在主干上"
		} else {
			seen[key] = true
		}
	}
	for key, item := range members {
		if !seen[key] {
			return nil, item.name + " 没有归位"
		}
	}
	return build(members, line, task.Side, task.Span, gap), ""
}

func joinLinks(in Input, extra []string, gap int) []Place {
	page := openAgent(in)
	members := roster(page, extra)
	var line []string
	for _, name := range orderPage(page) {
		if !isSpan(members[strings.ToLower(name)].kind) {
			line = append(line, name)
		}
	}
	for i := range extra {
		name := "新" + strconv.Itoa(i+1)
		if !isSpan(members[strings.ToLower(name)].kind) {
			line = append(line, name)
		}
	}
	if len(line) == 0 {
		for _, item := range orderPage(page) {
			line = append(line, item)
		}
	}
	if len(line) == 0 {
		return nil
	}
	var spans []Bridge
	for _, item := range members {
		if isSpan(item.kind) {
			spans = append(spans, Bridge{Element: item.name, From: line[0], To: line[len(line)-1]})
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].Element < spans[j].Element })
	return build(members, line, nil, spans, gap)
}

func build(members map[string]member, line []string, side []Dock, span []Bridge, gap int) []Place {
	places := along(members, line)
	laySpans(places, members, span)
	laySides(places, members, side, belt(members))
	paint(places, members, line, side, span, gap)
	var out []Place
	for _, spot := range places {
		out = append(out, spot)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Element < out[j].Element })
	return out
}

func along(members map[string]member, line []string) map[string]Place {
	places := map[string]Place{}
	cursor := 1.2
	for i, name := range line {
		item := members[strings.ToLower(name)]
		length := meters(item.kind)
		places[strings.ToLower(name)] = Place{Element: item.name, Pose: &PoseSpot{X: round3(cursor + length/2), Len: length}}
		cursor += length
		if i+1 < len(line) {
			cursor += 0.2
		}
	}
	return places
}

func laySpans(places map[string]Place, members map[string]member, spans []Bridge) {
	for i, bridge := range spans {
		from, to := places[strings.ToLower(bridge.From)].Pose, places[strings.ToLower(bridge.To)].Pose
		if from == nil || to == nil {
			continue
		}
		start, end := from.X-from.Len/2, to.X+to.Len/2
		if end < start {
			start, end = end, start
		}
		item := members[strings.ToLower(bridge.Element)]
		places[strings.ToLower(bridge.Element)] = Place{Element: item.name, Pose: &PoseSpot{
			X: round3((start + end) / 2), Z: 1.2 + float64(i)*1.2, Len: round3(end - start),
		}}
	}
}

func laySides(places map[string]Place, members map[string]member, side []Dock, raised bool) {
	for i, dock := range side {
		host := places[strings.ToLower(dock.Host)].Pose
		x := 1.2
		if host != nil {
			x = host.X
		}
		item := members[strings.ToLower(dock.Element)]
		y := 0.0
		if raised && (item.kind == "Batch" || item.kind == "3D Object") {
			y = 0.82
		}
		places[strings.ToLower(dock.Element)] = Place{Element: item.name, Pose: &PoseSpot{
			X: x, Y: y, Z: -1.8 - float64(i)*1.4, Len: meters(item.kind),
		}}
	}
}

func paint(places map[string]Place, members map[string]member, line []string, side []Dock, span []Bridge, gap int) {
	step := float64(layoutGap(gap))
	x, bottom := 40.0, 80.0
	for _, name := range line {
		item := members[strings.ToLower(name)]
		w, h := card(item)
		spot := places[strings.ToLower(name)]
		spot.X, spot.Y = snap(x), 80
		places[strings.ToLower(name)] = spot
		x += w + step
		if 80+h > bottom {
			bottom = 80 + h
		}
	}
	y := bottom + step
	sx := 40.0
	for _, dock := range side {
		item := members[strings.ToLower(dock.Element)]
		w, _ := card(item)
		spot := places[strings.ToLower(dock.Element)]
		spot.X, spot.Y = snap(sx), snap(y)
		places[strings.ToLower(dock.Element)] = spot
		sx += w + step
	}
	if len(side) > 0 {
		y += 84 + step
	}
	for _, bridge := range span {
		from, to := places[strings.ToLower(bridge.From)], places[strings.ToLower(bridge.To)]
		item := members[strings.ToLower(bridge.Element)]
		_, height := card(item)
		toWidth, _ := card(members[strings.ToLower(bridge.To)])
		width := to.X + toWidth - from.X
		if width < height {
			width = 160
		}
		spot := places[strings.ToLower(bridge.Element)]
		spot.X, spot.Y, spot.W = from.X, snap(y), snap(width)
		places[strings.ToLower(bridge.Element)] = spot
	}
}

func card(item member) (float64, float64) {
	w, h := item.w, item.h
	if w < 40 {
		w = 156
	}
	if h < 20 {
		h = 84
	}
	return w, h
}

func roster(page *agent, extra []string) map[string]member {
	out := map[string]member{}
	if page != nil {
		for _, item := range page.Elements {
			out[strings.ToLower(item.Name)] = member{item.Name, item.Type, item.W, item.H, item.X, item.Y, item.Pose}
		}
	}
	for i, kind := range extra {
		name := "新" + strconv.Itoa(i+1)
		out[strings.ToLower(name)] = member{name: name, kind: kind, w: 156, h: 84}
	}
	return out
}

func onLine(line []string, name string) bool {
	for _, item := range line {
		if strings.EqualFold(item, name) {
			return true
		}
	}
	return false
}

func orderPage(page *agent) []string {
	if page == nil {
		return nil
	}
	depth := map[string]int{}
	var walk func(string, map[string]bool) int
	walk = func(name string, stack map[string]bool) int {
		key := strings.ToLower(name)
		if value, ok := depth[key]; ok {
			return value
		}
		if stack[key] {
			return 0
		}
		stack[key] = true
		best := 0
		for _, link := range page.Links {
			if !strings.EqualFold(link.To, name) {
				continue
			}
			if next := walk(link.From, stack) + 1; next > best {
				best = next
			}
		}
		delete(stack, key)
		depth[key] = best
		return best
	}
	names := make([]string, 0, len(page.Elements))
	for _, item := range page.Elements {
		walk(item.Name, map[string]bool{})
		names = append(names, item.Name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		return depth[strings.ToLower(names[i])] < depth[strings.ToLower(names[j])]
	})
	return names
}

func belt(members map[string]member) bool {
	for _, item := range members {
		if item.kind == "Conveyor" || item.kind == "Convey" {
			return true
		}
	}
	return false
}

func isSpan(kind string) bool {
	switch kind {
	case "Conveyor", "Road", "Railway Track", "Pipe", "Wall", "Pathway", "Target Line", "Path", "Line", "Link", "Transition", "GIS Route", "Flow", "Stop Line":
		return true
	default:
		return false
	}
}

func meters(kind string) float64 {
	switch kind {
	case "Delay", "Service", "Station", "Assembler", "PedService", "ProcessByRobot", "Robot", "SeizeRobot", "ReleaseRobot", "Jib Crane", "MoveByCrane", "Storage", "Store", "Retrieve", "Turntable":
		return 2.8
	case "Queue", "Hold", "Convey", "Select Output", "Select Output 5", "Split", "Combine", "Match", "Branch":
		return 2.2
	case "Sink", "Exit", "ConveyorExit":
		return 1.6
	default:
		if strings.HasPrefix(kind, "Train") {
			return 7.2
		}
		if strings.HasPrefix(kind, "Car") {
			return 4.6
		}
		return 1.8
	}
}

func snap(value float64) float64 {
	return math.Round(value/10) * 10
}

func round3(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func openAgent(in Input) *agent {
	page := in.Page
	if page == "" {
		page = "Main"
	}
	for i := range in.Agents {
		if strings.EqualFold(in.Agents[i].Name, page) {
			return &in.Agents[i]
		}
	}
	if len(in.Agents) == 0 {
		return nil
	}
	return &in.Agents[0]
}

func placedTypes(actions []Task) []string {
	var types []string
	for _, task := range actions {
		if task.Intent == "place" {
			types = append(types, task.Types...)
		}
	}
	return types
}
