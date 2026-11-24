package scenario

import (
	"math"
	"strconv"
	"strings"
)

const meter = 0.012

type el struct {
	id, name, typ string
	x, y, w, h    float64
	props         map[string]any
	pose          *spotPose
}

type edge struct {
	from, to string
	port     int
}

type page struct {
	id, name string
	els      []*el
	links    []edge
}

type board struct {
	id, name string
	page     *page
}

func newBoard(id string) *board {
	return &board{id: id, name: id, page: &page{id: "main", name: "Main"}}
}

func (b *board) outline(text, session string) turnBody {
	items := make([]elemBody, 0, len(b.page.els))
	for _, item := range b.page.els {
		items = append(items, elemBody{
			Name: item.name, Type: item.typ, X: item.x, Y: item.y, W: item.w, H: item.h, Pose: item.pose,
		})
	}
	links := make([]linkBody, 0, len(b.page.links))
	for _, link := range b.page.links {
		from, to := b.page.byID(link.from), b.page.byID(link.to)
		if from == nil || to == nil {
			continue
		}
		links = append(links, linkBody{From: from.name, To: to.name, FromPort: link.port})
	}
	return turnBody{
		Text: text, Page: "Main", Project: b.id, Session: session,
		Agents: []pageBody{{Name: "Main", Elements: items, Links: links}},
	}
}

func (b *board) apply(task task) {
	var created []string
	b.applyOne(task, &created)
	for _, extra := range task.More {
		if extra.Status == "" {
			extra.Status = "已写入"
		}
		b.applyOne(extra, &created)
	}
}

func (b *board) applyOne(task task, created *[]string) {
	if task.Status != "已写入" {
		return
	}
	switch task.Intent {
	case "place":
		b.place(task, created)
	case "layout":
		b.arrange(task, *created)
	case "tie":
		b.page.tie(nil)
	case "set":
		if item := b.page.find(task.Element); item != nil {
			item.props[task.Key] = task.Value
		}
	case "link":
		b.join(task.From, task.To)
	case "remove":
		if item := b.page.find(task.Element); item != nil {
			b.page.drop(item.id)
		}
	}
}

func (b *board) place(task task, created *[]string) {
	var ids []string
	for _, typ := range task.Types {
		item := b.page.add(typ, 0, 0)
		for key, value := range task.Props {
			item.props[key] = value
		}
		ids = append(ids, item.id)
		*created = append(*created, item.id)
	}
	if task.Chain && len(ids) > 1 {
		for i := 0; i < len(ids)-1; i++ {
			b.page.links = append(b.page.links, edge{from: ids[i], to: ids[i+1], port: 0})
		}
		b.page.tie(ids)
	}
}

func (b *board) arrange(task task, created []string) {
	seen := map[string]bool{}
	for _, spot := range task.Places {
		item := b.page.resolve(spot.Element, created)
		if item == nil || spot.Pose == nil || seen[item.id] {
			continue
		}
		seen[item.id] = true
		item.x, item.y = snap(spot.X), snap(spot.Y)
		if spot.W > 0 {
			item.w = snap(spot.W)
		}
		item.pose = spot.Pose
	}
}

func (b *board) join(from, to string) {
	left, right := b.page.find(from), b.page.find(to)
	if left == nil || right == nil || left.id == right.id {
		return
	}
	b.page.links = append(b.page.links, edge{from: left.id, to: right.id})
}

func (p *page) add(typ string, x, y float64) *el {
	if x == 0 && y == 0 {
		n := len(p.els)
		x = 36 + float64(n%5)*128
		y = 64 + float64(n/5)*88
	}
	w, h := 156.0, 84.0
	if typ == "Conveyor" {
		w, h = 220, 48
	}
	item := &el{
		id: "n" + strconv.Itoa(len(p.els)+1), name: uniqueName(p, slug(typ)), typ: typ,
		x: x, y: y, w: w, h: h, props: map[string]any{},
	}
	item.pose = &spotPose{
		X: round3((item.x + item.w/2) * meter), Z: round3((item.y + item.h/2) * meter),
	}
	p.els = append(p.els, item)
	return item
}

func (p *page) tie(only []string) {
	want := map[string]bool{}
	for _, id := range only {
		want[id] = true
	}
	var ids []string
	for _, item := range p.els {
		if only == nil || want[item.id] {
			ids = append(ids, item.id)
		}
	}
	for _, id := range ids {
		item := p.byID(id)
		for port := 1; port < outs(item.typ); port++ {
			if p.taken(item.id, port) {
				continue
			}
			sink := p.add("Sink", item.x+150, item.y+80)
			p.links = append(p.links, edge{from: item.id, to: sink.id, port: port})
		}
	}
}

func (p *page) drop(id string) {
	var kept []*el
	for _, item := range p.els {
		if item.id != id {
			kept = append(kept, item)
		}
	}
	p.els = kept
	var links []edge
	for _, link := range p.links {
		if link.from != id && link.to != id {
			links = append(links, link)
		}
	}
	p.links = links
}

func (p *page) find(name string) *el {
	needle := strings.ToLower(strings.TrimSpace(name))
	flat := strings.ReplaceAll(needle, " ", "")
	for _, item := range p.els {
		if strings.ToLower(item.name) == needle || strings.EqualFold(item.typ, name) || slug(item.typ) == flat {
			return item
		}
	}
	return nil
}

func (p *page) resolve(name string, created []string) *el {
	if n, ok := freshIndex(name); ok && n <= len(created) {
		return p.byID(created[n-1])
	}
	return p.find(name)
}

func (p *page) byID(id string) *el {
	for _, item := range p.els {
		if item.id == id {
			return item
		}
	}
	return nil
}

func (p *page) taken(id string, port int) bool {
	for _, link := range p.links {
		if link.from == id && link.port == port {
			return true
		}
	}
	return false
}

func freshIndex(name string) (int, bool) {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "新") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(name, "新"))
	return n, err == nil && n > 0
}

func uniqueName(p *page, base string) string {
	used := map[string]bool{}
	for _, item := range p.els {
		used[item.name] = true
	}
	if !used[base] {
		return base
	}
	for n := 2; ; n++ {
		name := base + strconv.Itoa(n)
		if !used[name] {
			return name
		}
	}
}

func slug(typ string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(typ) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "item"
	}
	return b.String()
}

func outs(typ string) int {
	switch typ {
	case "Select Output":
		return 2
	case "Select Output 5":
		return 5
	case "Resource Pool", "TransporterFleet", "Conveyor", "Camera", "Light":
		return 0
	}
	if strings.HasSuffix(typ, "Sink") || strings.HasSuffix(typ, "Dispose") || strings.HasSuffix(typ, "Exit") {
		return 0
	}
	return 1
}

func snap(value float64) float64 { return math.Round(value/10) * 10 }

func round3(value float64) float64 { return math.Round(value*1000) / 1000 }
