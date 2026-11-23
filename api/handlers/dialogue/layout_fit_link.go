package dialogue

import (
	"math"
	"sort"
	"strings"
)

func (f *fitter) outs(name string) []fitEdge {
	var edges []fitEdge
	for _, link := range f.leaving(name) {
		edges = append(edges, f.follow(link.To, "")...)
	}
	return edges
}

func (f *fitter) follow(name, via string) []fitEdge {
	key := strings.ToLower(name)
	item, ok := f.members[key]
	if !ok || f.stack[key] {
		return nil
	}
	if logicKind(item.kind) || isSpan(item.kind) {
		f.stack[key] = true
		next := via
		if isSpan(item.kind) && next == "" {
			next = item.name
		}
		var edges []fitEdge
		for _, link := range f.leaving(name) {
			edges = append(edges, f.follow(link.To, next)...)
		}
		delete(f.stack, key)
		return edges
	}
	if !f.here[key] {
		return nil
	}
	return []fitEdge{{to: item.name, via: via}}
}

func (f *fitter) leaving(name string) []pageLink {
	var raw []pageLink
	for _, link := range f.links {
		if strings.EqualFold(link.From, name) {
			raw = append(raw, link)
		}
	}
	sort.Slice(raw, func(i, j int) bool {
		if raw[i].FromPort != raw[j].FromPort {
			return raw[i].FromPort < raw[j].FromPort
		}
		return raw[i].To < raw[j].To
	})
	return raw
}

func (f *fitter) dockLinked(group []string) {
	slot := map[string]int{}
	for _, name := range group {
		kind := f.kind(name)
		if f.seen[strings.ToLower(name)] || logicKind(kind) || isSpan(kind) {
			continue
		}
		f.dockAt(name, slot, -1)
	}
}

func (f *fitter) dockLogic(group []string) {
	slot := map[string]int{}
	for _, name := range group {
		if f.seen[strings.ToLower(name)] || !logicKind(f.kind(name)) {
			continue
		}
		f.dockAt(name, slot, -1)
	}
}

func (f *fitter) dockAt(name string, slot map[string]int, sign float64) {
	hostName := f.anchor(name)
	host, ok := f.spots[strings.ToLower(hostName)]
	if !ok || host.Pose == nil {
		return
	}
	n := slot[strings.ToLower(host.Element)]
	slot[strings.ToLower(host.Element)] = n + 1
	_, _, wide := measure(f.kind(host.Element))
	sx, sz := stepOf(host.Pose.Yaw + sign*math.Pi/2)
	dist := wide + 1 + float64(n)*1.8
	back, front, _ := measure(f.kind(name))
	f.drop(name, host.Pose.X+sx*dist, host.Pose.Z+sz*dist, host.Pose.Yaw+sign*math.Pi/2, back+front)
}

func (f *fitter) anchor(name string) string {
	for _, link := range f.links {
		if strings.EqualFold(link.To, name) {
			if host := f.physical(link.From); host != "" {
				return host
			}
		}
	}
	for _, link := range f.links {
		if strings.EqualFold(link.From, name) {
			if host := f.physical(link.To); host != "" {
				return host
			}
		}
	}
	return ""
}

func (f *fitter) physical(name string) string {
	item, ok := f.members[strings.ToLower(name)]
	if !ok || logicKind(item.kind) || isSpan(item.kind) {
		return ""
	}
	if _, placed := f.spots[strings.ToLower(item.name)]; placed {
		return item.name
	}
	return ""
}
