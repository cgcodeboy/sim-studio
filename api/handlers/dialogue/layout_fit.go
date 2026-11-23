package dialogue

import (
	"math"
	"strconv"
	"strings"
)

type fitEdge struct{ to, via string }

type fitter struct {
	members map[string]member
	links   []pageLink
	names   []string
	zone    map[string]string
	here    map[string]bool
	seen    map[string]bool
	stack   map[string]bool
	spots   map[string]Place
	cardX   float64
	cardY   float64
	maxZ    float64
}

func fitScene(in Input, extra []string) []Place {
	page := openAgent(in)
	members := roster(page, extra)
	if len(members) == 0 {
		return nil
	}
	fit := &fitter{
		members: members, names: fitNames(page, members, extra),
		zone: map[string]string{}, here: map[string]bool{}, seen: map[string]bool{},
		stack: map[string]bool{}, spots: map[string]Place{}, cardY: 80,
	}
	if page != nil {
		fit.links = page.Links
	}
	for _, name := range fit.names {
		key := strings.ToLower(name)
		fit.zone[key] = businessOf(fit.members[key].kind)
	}
	for _, name := range fit.names {
		key := strings.ToLower(name)
		if isSpan(fit.members[key].kind) {
			fit.zone[key] = fit.spanZone(name)
		}
	}
	z := 0.0
	for _, biz := range []string{"物料", "搬运", "流体", "行人", "道路", "铁路", "其他"} {
		group := fit.pick(biz)
		if len(group) == 0 {
			continue
		}
		z = fit.lay(group, z)
		fit.cardY += 220
	}
	var out []Place
	for _, name := range fit.names {
		if spot, ok := fit.spots[strings.ToLower(name)]; ok {
			out = append(out, spot)
		}
	}
	return out
}

func fitReply(in Input, extra []string) string {
	page := openAgent(in)
	members := roster(page, extra)
	var order []string
	zones := map[string][]string{}
	for _, name := range fitNames(page, members, extra) {
		item := members[strings.ToLower(name)]
		if isSpan(item.kind) || logicKind(item.kind) {
			continue
		}
		biz := businessOf(item.kind)
		if _, ok := zones[biz]; !ok {
			order = append(order, biz)
		}
		zones[biz] = append(zones[biz], name)
	}
	if len(order) == 0 {
		return "已按连线和占地接好。"
	}
	var bits []string
	for _, biz := range order {
		bits = append(bits, biz+"："+strings.Join(zones[biz], " → "))
	}
	return "已按连线和占地接好。" + strings.Join(bits, "。") + "。"
}

func fitNames(page *agent, members map[string]member, extra []string) []string {
	seen := map[string]bool{}
	var names []string
	for _, name := range orderPage(page) {
		key := strings.ToLower(name)
		if _, ok := members[key]; ok && !seen[key] {
			names = append(names, name)
			seen[key] = true
		}
	}
	for i := range extra {
		name := "新" + strconv.Itoa(i+1)
		key := strings.ToLower(name)
		if _, ok := members[key]; ok && !seen[key] {
			names = append(names, name)
			seen[key] = true
		}
	}
	return names
}

func (f *fitter) pick(biz string) []string {
	var group []string
	for _, name := range f.names {
		if f.zone[strings.ToLower(name)] == biz {
			group = append(group, name)
		}
	}
	return group
}

func (f *fitter) spanZone(name string) string {
	for _, link := range f.links {
		other := ""
		switch {
		case strings.EqualFold(link.From, name):
			other = link.To
		case strings.EqualFold(link.To, name):
			other = link.From
		}
		if other == "" {
			continue
		}
		item := f.members[strings.ToLower(other)]
		if item.kind == "" || isSpan(item.kind) || logicKind(item.kind) {
			continue
		}
		return businessOf(item.kind)
	}
	return f.zone[strings.ToLower(name)]
}

func (f *fitter) lay(group []string, z float64) float64 {
	f.cardX = 40
	f.maxZ = z
	f.here = map[string]bool{}
	for _, name := range group {
		f.here[strings.ToLower(name)] = true
	}
	f.dockLinked(group)
	var stations []string
	for _, name := range group {
		kind := f.kind(name)
		if logicKind(kind) || isSpan(kind) || f.seen[strings.ToLower(name)] {
			continue
		}
		stations = append(stations, name)
	}
	cursorZ := z
	for i, root := range f.roots(stations) {
		if i > 0 {
			cursorZ = f.maxZ + 2.8
		}
		back, _, _ := measure(f.kind(root))
		f.walk(root, 1.2+back, cursorZ, 0)
	}
	f.dockLogic(group)
	f.leftovers(group, f.maxZ+2.2)
	return f.maxZ + 3.2
}

func (f *fitter) roots(stations []string) []string {
	incoming := map[string]bool{}
	for _, name := range stations {
		for _, edge := range f.outs(name) {
			incoming[strings.ToLower(edge.to)] = true
		}
	}
	var roots []string
	for _, name := range stations {
		if !incoming[strings.ToLower(name)] {
			roots = append(roots, name)
		}
	}
	if len(roots) == 0 && len(stations) > 0 {
		return []string{stations[0]}
	}
	return roots
}

func (f *fitter) walk(name string, x, z, yaw float64) {
	if f.seen[strings.ToLower(name)] {
		return
	}
	back, front, wide := measure(f.kind(name))
	f.drop(name, x, z, yaw, back+front)
	sx, sz := stepOf(yaw)
	frontX, frontZ := x+sx*front, z+sz*front
	for i, edge := range dedupe(f.outs(name)) {
		head, ox, oz := yaw, frontX, frontZ
		if i > 0 {
			head = yaw + math.Pi/2*branchSign(i)
			bx, bz := stepOf(head)
			ox, oz = x+bx*(wide+0.5), z+bz*(wide+0.5)
		}
		f.stepTo(edge, ox, oz, head)
	}
}

func (f *fitter) stepTo(edge fitEdge, x, z, yaw float64) {
	sx, sz := stepOf(yaw)
	if edge.via != "" && !f.seen[strings.ToLower(edge.via)] {
		f.drop(edge.via, x+sx*1.3, z+sz*1.3, yaw, 2.6)
		x += sx * 2.6
		z += sz * 2.6
	} else {
		x += sx * 0.2
		z += sz * 0.2
	}
	back, _, _ := measure(f.kind(edge.to))
	f.walk(edge.to, x+sx*back, z+sz*back, yaw)
}

func (f *fitter) leftovers(group []string, z float64) {
	cursor := 1.2
	for _, name := range group {
		if f.seen[strings.ToLower(name)] {
			continue
		}
		back, front, _ := measure(f.kind(name))
		f.drop(name, cursor+back, z, 0, back+front)
		cursor += back + front + 0.8
	}
}

func (f *fitter) drop(name string, x, z, yaw, length float64) {
	key := strings.ToLower(name)
	if f.seen[key] {
		return
	}
	item := f.members[key]
	half := halfOf(item.kind)
	f.spots[key] = Place{Element: item.name, X: snap(f.cardX), Y: snap(f.cardY), Pose: &PoseSpot{
		X: round3(x), Y: craneLift(item.kind), Z: round3(z), Yaw: round3(yaw), Len: round3(length),
	}}
	w, _ := card(item)
	f.cardX += w + 48
	f.seen[key] = true
	if z+half > f.maxZ {
		f.maxZ = z + half
	}
}

func halfOf(kind string) float64 {
	back, front, wide := measure(kind)
	half := back
	if front > half {
		half = front
	}
	if wide > half {
		half = wide
	}
	return half
}

func (f *fitter) kind(name string) string {
	return f.members[strings.ToLower(name)].kind
}

func dedupe(edges []fitEdge) []fitEdge {
	seen := map[string]bool{}
	var out []fitEdge
	for _, edge := range edges {
		key := strings.ToLower(edge.to)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, edge)
	}
	return out
}

func branchSign(index int) float64 {
	if index%2 == 1 {
		return 1
	}
	return -1
}

func stepOf(yaw float64) (float64, float64) {
	yaw = math.Mod(yaw, 2*math.Pi)
	if yaw < 0 {
		yaw += 2 * math.Pi
	}
	switch int(math.Round(yaw/(math.Pi/2))) % 4 {
	case 0:
		return 1, 0
	case 1:
		return 0, 1
	case 2:
		return -1, 0
	default:
		return 0, -1
	}
}

func logicKind(kind string) bool {
	switch kind {
	case "Seize", "Release", "Resource Pool":
		return true
	default:
		return false
	}
}

func craneLift(kind string) float64 {
	if kind == "Jib Crane" || kind == "MoveByCrane" {
		return 4
	}
	return 0
}

func measure(kind string) (float64, float64, float64) {
	switch kind {
	case "Source", "Enter", "ConveyorEnter":
		return 0.36, 1.09, 0.4
	case "Delay", "Service", "Station", "ProcessByRobot", "Assembler", "PedService":
		return 1.60, 0.68, 0.45
	case "Sink", "Exit", "ConveyorExit":
		return 0.56, 0.56, 0.45
	case "Select Output", "Select Output 5", "Split", "Branch", "Queue", "Hold":
		return 1.10, 1.10, 0.4
	case "Tank", "Storage Tank":
		return 0.8, 0.8, 0.8
	case "CarSource", "CarMoveTo", "CarDispose", "CarEnter":
		return 2.1, 2.1, 0.9
	case "TrainSource", "TrainMoveTo", "TrainDispose", "TrainEnter", "TrainExit":
		return 1.8, 1.8, 0.7
	case "PedSource", "PedGoTo", "PedWait", "PedSink", "PedEnter", "PedExit":
		return 0.4, 0.4, 0.4
	case "MoveByTransporter", "TransporterFleet":
		return 0.8, 0.8, 0.55
	default:
		if isSpan(kind) {
			return 1.3, 1.3, 0.32
		}
		return 0.8, 0.8, 0.5
	}
}
