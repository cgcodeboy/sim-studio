package dialogue

import "strings"

type block struct {
	display string
	family  string
	aliases []string
}

func (b block) canon() string {
	return strings.ReplaceAll(b.display, " ", "")
}

func (b block) terms() []string {
	out := []string{b.display}
	return append(out, b.aliases...)
}

var blocks = []block{
	{"Select Output 5", "process", nil},
	{"Select Output", "process", nil},
	{"Resource Pool", "process", []string{"资源池"}},
	{"Source", "process", nil},
	{"Sink", "process", []string{"吸收"}},
	{"Delay", "process", []string{"延迟"}},
	{"Queue", "process", []string{"队列"}},
	{"Hold", "process", nil},
	{"Enter", "process", nil},
	{"Exit", "process", nil},
	{"Split", "process", nil},
	{"Combine", "process", nil},
	{"Assembler", "process", nil},
	{"Move To", "process", nil},
	{"Seize", "process", []string{"占用"}},
	{"Release", "process", []string{"释放"}},
	{"Service", "process", []string{"服务"}},
	{"Batch", "process", nil},
	{"Convey", "handling", nil},
	{"Conveyor", "handling", []string{"传送带"}},
	{"MoveByTransporter", "handling", nil},
	{"TransporterFleet", "handling", []string{"车队"}},
	{"Store", "handling", []string{"入库"}},
	{"Retrieve", "handling", []string{"出库"}},
	{"PedSource", "pedestrian", []string{"行人源"}},
	{"PedGoTo", "pedestrian", nil},
	{"PedWait", "pedestrian", nil},
	{"PedService", "pedestrian", nil},
	{"PedEnter", "pedestrian", nil},
	{"PedExit", "pedestrian", nil},
	{"PedSink", "pedestrian", nil},
	{"CarSource", "road", nil},
	{"CarMoveTo", "road", nil},
	{"CarDispose", "road", nil},
	{"TrainSource", "rail", nil},
	{"TrainMoveTo", "rail", nil},
	{"TrainDispose", "rail", nil},
	{"FluidSource", "fluid", nil},
	{"Tank", "fluid", []string{"储罐"}},
	{"Valve", "fluid", []string{"阀门"}},
	{"FluidDispose", "fluid", nil},
	{"Stock", "dynamics", []string{"存量"}},
	{"Flow", "dynamics", nil},
	{"Dynamic Variable", "dynamics", nil},
	{"Parameter", "dynamics", []string{"参数"}},
	{"Camera", "camera", nil},
	{"Light", "camera", nil},
}

func active(known map[string]bool) []block {
	if known == nil {
		return blocks
	}
	var out []block
	for _, item := range blocks {
		if known[item.canon()] {
			out = append(out, item)
		}
	}
	return out
}

type hit struct {
	at int
	n  int
	b  block
}

func mention(text string, list []block) []block {
	var found []hit
	for _, item := range list {
		for _, term := range item.terms() {
			if at, ok := findTerm(text, term); ok {
				found = append(found, hit{at: at, n: len(term), b: item})
				break
			}
		}
	}
	sortHits(found)
	return takeHits(text, found)
}

func findTerm(text, term string) (int, bool) {
	lower := strings.ToLower(text)
	term = strings.ToLower(term)
	from := 0
	for from < len(lower) {
		at := strings.Index(lower[from:], term)
		if at < 0 {
			return 0, false
		}
		at += from
		if bounded(lower, at, at+len(term)) {
			return at, true
		}
		from = at + len(term)
	}
	return 0, false
}

func bounded(text string, start, end int) bool {
	if start > 0 && word(text[start-1]) {
		return false
	}
	if end < len(text) && word(text[end]) {
		return false
	}
	return true
}

func word(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func sortHits(found []hit) {
	for i := 1; i < len(found); i++ {
		item := found[i]
		j := i
		for j > 0 && lessHit(item, found[j-1]) {
			found[j] = found[j-1]
			j--
		}
		found[j] = item
	}
}

func lessHit(a, b hit) bool {
	if a.n != b.n {
		return a.n > b.n
	}
	return a.at < b.at
}

func takeHits(text string, found []hit) []block {
	used := make([]bool, len(text))
	chosen := make([]hit, 0, len(found))
	seen := map[string]bool{}
	for _, item := range found {
		end := item.at + item.n
		if overlap(used, item.at, end) || seen[item.b.display] {
			continue
		}
		for i := item.at; i < end && i < len(used); i++ {
			used[i] = true
		}
		seen[item.b.display] = true
		chosen = append(chosen, item)
	}
	sortByPlace(chosen)
	out := make([]block, 0, len(chosen))
	for _, item := range chosen {
		out = append(out, item.b)
	}
	return out
}

func sortByPlace(found []hit) {
	for i := 1; i < len(found); i++ {
		item := found[i]
		j := i
		for j > 0 && found[j-1].at > item.at {
			found[j] = found[j-1]
			j--
		}
		found[j] = item
	}
}

func overlap(used []bool, start, end int) bool {
	for i := start; i < end && i < len(used); i++ {
		if used[i] {
			return true
		}
	}
	return false
}
