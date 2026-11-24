package dialogue

import (
	"math"
	"sort"
	"strings"
)

type yard struct {
	links   []pageLink
	members map[string]member
	at      map[string]*Place
	slot    map[string]int
}

func weld(in Input, spots []Place, extra []string) []Place {
	page := openAgent(in)
	if page == nil || len(spots) == 0 {
		return spots
	}
	out := copyPlaces(spots)
	work := &yard{links: page.Links, members: roster(page, extra)}
	for pass := 0; pass < 4; pass++ {
		work.at = placeIndex(out)
		work.slot = map[string]int{}
		if !work.pass() {
			break
		}
	}
	return out
}

func placeIndex(spots []Place) map[string]*Place {
	at := map[string]*Place{}
	for i := range spots {
		at[strings.ToLower(spots[i].Element)] = &spots[i]
	}
	return at
}

func (y *yard) pass() bool {
	moved := false
	for _, key := range y.logicKeys() {
		if y.park(key) {
			moved = true
		}
	}
	for _, key := range y.logicKeys() {
		if y.bridge(y.members[key].name) {
			moved = true
		}
	}
	for _, link := range y.links {
		if y.pull(link) {
			moved = true
		}
	}
	return moved
}

func (y *yard) logicKeys() []string {
	var keys []string
	for key, item := range y.members {
		if logicKind(item.kind) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func (y *yard) bridge(name string) bool {
	up, down := logicHosts(name, y.links, y.members)
	left, right := y.at[strings.ToLower(up)], y.at[strings.ToLower(down)]
	if left == nil || right == nil || left.Pose == nil || right.Pose == nil {
		return false
	}
	exit := faceOut(left.Pose, y.members[strings.ToLower(up)].kind)
	inlet := faceIn(right.Pose, y.members[strings.ToLower(down)].kind)
	return join(right.Pose, left.Pose, exit, inlet)
}

func (y *yard) pull(link pageLink) bool {
	from := y.members[strings.ToLower(link.From)]
	to := y.members[strings.ToLower(link.To)]
	left, right := y.at[strings.ToLower(link.From)], y.at[strings.ToLower(link.To)]
	if !stationKind(from.kind) || !stationKind(to.kind) || left == nil || right == nil || left.Pose == nil || right.Pose == nil {
		return false
	}
	exit := exitPoint(left.Pose, from.kind, link.FromPort)
	return join(right.Pose, left.Pose, exit, faceIn(right.Pose, to.kind))
}

func (y *yard) park(key string) bool {
	item := y.members[key]
	upName, downName := logicHosts(item.name, y.links, y.members)
	logic := y.at[key]
	up, down := y.at[strings.ToLower(upName)], y.at[strings.ToLower(downName)]
	if logic == nil || up == nil || down == nil || logic.Pose == nil || up.Pose == nil || down.Pose == nil {
		return false
	}
	if !blocksPath(*logic, *up, *down, y.members) {
		return false
	}
	n := y.slot[up.Element]
	y.slot[up.Element] = n + 1
	_, _, wide := measure(y.members[strings.ToLower(up.Element)].kind)
	sx, sz := stepOf(up.Pose.Yaw - math.Pi/2)
	dist := wide + 1.2 + float64(n)*1.8
	logic.Pose.X = round3(up.Pose.X + sx*dist)
	logic.Pose.Z = round3(up.Pose.Z + sz*dist)
	logic.Pose.Yaw = round3(up.Pose.Yaw - math.Pi/2)
	return true
}

func join(down, up *PoseSpot, exit, inlet point) bool {
	dist := math.Hypot(inlet.x-exit.x, inlet.z-exit.z)
	moved := false
	if dist > 1 {
		slide(down, exit, inlet)
		moved = true
	}
	dir := axisOf(down.X-up.X, down.Z-up.Z)
	if boxesHit(up, down) && pushClear(down, up, dir) {
		return true
	}
	return moved
}

func axisOf(dx, dz float64) point {
	if math.Abs(dx) >= math.Abs(dz) {
		if dx < 0 {
			return point{-1, 0}
		}
		return point{1, 0}
	}
	if dz < 0 {
		return point{0, -1}
	}
	return point{0, 1}
}

func slide(pose *PoseSpot, exit, inlet point) {
	dx, dz := inlet.x-exit.x, inlet.z-exit.z
	dist := math.Hypot(dx, dz)
	if dist <= 0.15 {
		return
	}
	scale := (dist - 0.15) / dist
	pose.X = round3(pose.X - dx*scale)
	pose.Z = round3(pose.Z - dz*scale)
}

func pushClear(pose, block *PoseSpot, dir point) bool {
	ax0, ax1, az0, az1 := footprint(block)
	bx0, bx1, bz0, bz1 := footprint(pose)
	overlap := ax1 - bx0
	if dir.x < 0 {
		overlap = bx1 - ax0
	}
	if math.Abs(dir.z) > 0.5 {
		overlap = az1 - bz0
		if dir.z < 0 {
			overlap = bz1 - az0
		}
	}
	if overlap <= 0 {
		return false
	}
	pose.X = round3(pose.X + dir.x*(overlap+0.08))
	pose.Z = round3(pose.Z + dir.z*(overlap+0.08))
	return true
}

func exitPoint(pose *PoseSpot, kind string, port int) point {
	if port <= 0 {
		return faceOut(pose, kind)
	}
	_, _, wide := measure(kind)
	sx, sz := exitStep(pose, port)
	return point{pose.X + sx*wide, pose.Z + sz*wide}
}

func exitStep(pose *PoseSpot, port int) (float64, float64) {
	if port <= 0 {
		return stepOf(pose.Yaw)
	}
	sign := 1.0
	if port%2 == 0 {
		sign = -1
	}
	return stepOf(pose.Yaw + sign*math.Pi/2)
}
