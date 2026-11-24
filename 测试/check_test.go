package scenario

import (
	"fmt"
	"strings"
)

type expect struct {
	written   bool
	open      bool
	intent    string
	forbid    string
	has       []string
	absent    []string
	atLeast   map[string]int
	atMost    map[string]int
	minLinks  int
	path      [][2]string
	cut       [][2]string
	hop       [][2]string
	noHop     [][2]string
	row       [][2]string
	spread    string
	holds     string
	minCount  int
	minPlaces int
	layout    bool
	ban       []string
}

type scene struct {
	name  string
	level string
	build func(*board)
	turns []round
}

type round struct {
	say  string
	want expect
}

func (e expect) verify(t fataler, task task, b *board, before int) {
	if e.written && task.Status != "已写入" {
		t.Fatalf("没有写入。%s\n%s", task.Reply, b.summary())
	}
	if e.open && (task.Status != "未写入" || b.count() != before) {
		t.Fatalf("这轮不该改工程。状态 %s，模块 %d→%d。%s", task.Status, before, b.count(), task.Reply)
	}
	for _, word := range e.ban {
		if strings.Contains(task.Reply, word) {
			t.Fatalf("回复里不该有「%s」：%s", word, task.Reply)
		}
	}
	if e.intent != "" && !intentIs(task, e.intent) {
		t.Fatalf("意图是 %s，期望 %s。%s", task.Intent, e.intent, task.Reply)
	}
	if e.forbid != "" && intentIs(task, e.forbid) {
		t.Fatalf("不该出现意图 %s。%s", e.forbid, task.Reply)
	}
	for _, typ := range e.has {
		if b.countType(typ) < 1 {
			t.Fatalf("缺少 %s。%s", typ, b.summary())
		}
	}
	for _, typ := range e.absent {
		if b.countType(typ) > 0 {
			t.Fatalf("还留着 %s。%s", typ, b.summary())
		}
	}
	for typ, need := range e.atLeast {
		if b.countType(typ) < need {
			t.Fatalf("%s 至少要 %d 个，现在 %d。%s", typ, need, b.countType(typ), b.summary())
		}
	}
	if b.linkCount() < e.minLinks {
		t.Fatalf("连线 %d 条，至少要 %d。%s", b.linkCount(), e.minLinks, b.summary())
	}
	for _, pair := range e.path {
		if !b.reaches(pair[0], pair[1]) {
			t.Fatalf("%s 到不了 %s。%s", pair[0], pair[1], b.summary())
		}
	}
	for _, pair := range e.cut {
		if b.reaches(pair[0], pair[1]) {
			t.Fatalf("%s 不该连到 %s。%s", pair[0], pair[1], b.summary())
		}
	}
	e.verifyLinks(t, b)
	if e.minCount > 0 && b.count() < e.minCount {
		t.Fatalf("模块只有 %d", b.count())
	}
	if e.minPlaces > 0 && placesOf(task) < e.minPlaces {
		t.Fatalf("位置只有 %d", placesOf(task))
	}
	if e.holds != "" && !b.propHas(e.holds) {
		t.Fatalf("属性里没有「%s」。%s", e.holds, b.summary())
	}
	if e.layout {
		e.checkLayout(t, task)
	}
}

func (e expect) verifyLinks(t fataler, b *board) {
	for typ, limit := range e.atMost {
		if b.countType(typ) > limit {
			t.Fatalf("%s 最多 %d 个，现在 %d。%s", typ, limit, b.countType(typ), b.summary())
		}
	}
	for _, pair := range e.hop {
		if !b.hops(pair[0], pair[1]) {
			t.Fatalf("%s 没有直接连到 %s。%s", pair[0], pair[1], b.summary())
		}
	}
	for _, pair := range e.noHop {
		if b.hops(pair[0], pair[1]) {
			t.Fatalf("%s 不该直接连到 %s。%s", pair[0], pair[1], b.summary())
		}
	}
	for _, pair := range e.row {
		if !b.rowsApart(pair[0], pair[1]) {
			t.Fatalf("%s 和 %s 还在同一排。%s", pair[0], pair[1], b.summary())
		}
	}
	if e.spread != "" && !b.spreadZ(e.spread) {
		t.Fatalf("%s 没有分成两排。%s", e.spread, b.summary())
	}
}

func (e expect) checkLayout(t fataler, task task) {
	if !intentIs(task, "layout") {
		t.Fatalf("没有做布局。意图 %s。%s", task.Intent, task.Reply)
	}
	if placesOf(task) == 0 {
		t.Fatalf("布局没有给出位置。%s", task.Reply)
	}
	for _, word := range []string{"仍有冲突", "做不到", "没有接上"} {
		if strings.Contains(task.Reply, word) {
			t.Fatalf("布局把问题退回来了：%s", task.Reply)
		}
	}
}

func intentIs(task task, intent string) bool {
	if task.Intent == intent {
		return true
	}
	for _, extra := range task.More {
		if intentIs(extra, intent) {
			return true
		}
	}
	return false
}

func placesOf(task task) int {
	n := len(task.Places)
	for _, extra := range task.More {
		n += placesOf(extra)
	}
	return n
}

func (b *board) count() int { return len(b.page.els) }

func (b *board) linkCount() int { return len(b.page.links) }

func (b *board) countType(typ string) int {
	n := 0
	for _, item := range b.page.els {
		if item.typ == typ {
			n++
		}
	}
	return n
}

func (b *board) propHas(text string) bool {
	for _, item := range b.page.els {
		for _, value := range item.props {
			if strings.Contains(fmt.Sprint(value), text) {
				return true
			}
		}
	}
	return false
}

func (b *board) reaches(from, to string) bool {
	goal := map[string]bool{}
	for _, id := range b.idsOf(to) {
		goal[id] = true
	}
	seen := map[string]bool{}
	queue := b.idsOf(from)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		for _, link := range b.page.links {
			if link.from != id {
				continue
			}
			if goal[link.to] {
				return true
			}
			queue = append(queue, link.to)
		}
	}
	return false
}

func (b *board) hops(from, to string) bool {
	fromID := map[string]bool{}
	toID := map[string]bool{}
	for _, id := range b.idsOf(from) {
		fromID[id] = true
	}
	for _, id := range b.idsOf(to) {
		toID[id] = true
	}
	for _, link := range b.page.links {
		if fromID[link.from] && toID[link.to] {
			return true
		}
	}
	return false
}

func (b *board) rowsApart(left, right string) bool {
	a0, a1, aok := b.zRange(left)
	b0, b1, bok := b.zRange(right)
	if !aok || !bok {
		return false
	}
	return a1 < b0-0.8 || b1 < a0-0.8
}

func (b *board) spreadZ(typ string) bool {
	lo, hi, ok := b.zRange(typ)
	return ok && hi-lo >= 1.2 && b.countType(typ) >= 2
}

func (b *board) zRange(typ string) (float64, float64, bool) {
	var lo, hi float64
	ok := false
	for _, item := range b.page.els {
		if item.typ != typ || item.pose == nil {
			continue
		}
		z := item.pose.Z
		if !ok {
			lo, hi, ok = z, z, true
			continue
		}
		if z < lo {
			lo = z
		}
		if z > hi {
			hi = z
		}
	}
	return lo, hi, ok
}

func (b *board) idsOf(typ string) []string {
	var ids []string
	for _, item := range b.page.els {
		if item.typ == typ {
			ids = append(ids, item.id)
		}
	}
	return ids
}

func (b *board) brief() string {
	if b.count() > 40 {
		return fmt.Sprintf("%d 个模块，%d 条连线", b.count(), b.linkCount())
	}
	return b.summary()
}

func (b *board) summary() string {
	var parts []string
	for _, item := range b.page.els {
		text := item.name + "(" + item.typ + ")"
		if len(item.props) > 0 {
			text += fmt.Sprint(item.props)
		}
		parts = append(parts, text)
	}
	var links []string
	for _, link := range b.page.links {
		from, to := b.page.byID(link.from), b.page.byID(link.to)
		if from == nil || to == nil {
			continue
		}
		links = append(links, from.name+"→"+to.name)
	}
	return "模块：" + strings.Join(parts, "，") + "。连线：" + strings.Join(links, "，")
}
