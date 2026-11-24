package dialogue

import (
	"fmt"
	"math"
	"testing"
)

func TestFitCampusPlacesEveryModule(t *testing.T) {
	in := campusPage()
	page := in.Agents[0]
	if len(page.Elements) < 300 {
		t.Fatalf("modules %d", len(page.Elements))
	}
	spots := weld(in, fitScene(in, nil), nil)
	if len(spots) != len(page.Elements) {
		t.Fatalf("places %d modules %d", len(spots), len(page.Elements))
	}
	if pairs := campusOverlapNames(spots); len(pairs) > 0 {
		t.Fatalf("overlaps %v", pairs)
	}
	at := map[string]PoseSpot{}
	for _, spot := range spots {
		if spot.Pose == nil {
			t.Fatal(spot.Element)
		}
		at[spot.Element] = *spot.Pose
	}
	src, queue := at["l0src"], at["l0q"]
	front, back := faceOut(&src, "Source"), faceIn(&queue, "Queue")
	gap := math.Hypot(front.x-back.x, front.z-back.z)
	if gap > 1 {
		t.Fatalf("line gap %v", gap)
	}
	if math.Abs(at["f0tank"].Z-at["l0src"].Z) < 2 || math.Abs(at["w0store"].Z-at["l0src"].Z) < 2 {
		t.Fatalf("zones src %v tank %v store %v", at["l0src"].Z, at["f0tank"].Z, at["w0store"].Z)
	}
}

func campusOverlapNames(spots []Place) []string {
	var names []string
	for i := range spots {
		for j := i + 1; j < len(spots); j++ {
			if crowded(spots[i], spots[j]) {
				names = append(names, spots[i].Element+"×"+spots[j].Element)
			}
		}
	}
	return names
}

type campus struct {
	els   []element
	links []pageLink
}

func campusPage() Input {
	yard := &campus{}
	for i := 0; i < 24; i++ {
		yard.line(fmt.Sprintf("l%d", i))
	}
	yard.shared()
	return Input{Page: "Main", Agents: []agent{{Name: "Main", Elements: yard.els, Links: yard.links}}}
}

func (c *campus) line(prefix string) {
	seq := []struct{ name, typ string }{
		{"src", "Source"}, {"q", "Queue"}, {"sz", "Seize"}, {"wash", "Delay"},
		{"rel", "Release"}, {"conveyor", "Conveyor"}, {"fill", "Delay"},
		{"conveyor2", "Conveyor"}, {"div", "Select Output"}, {"ok", "Sink"},
	}
	var prev string
	for _, step := range seq {
		name := prefix + step.name
		c.put(name, step.typ)
		if prev != "" {
			c.tie(prev, name, 0)
		}
		prev = name
	}
	c.put(prefix+"bad", "Sink")
	c.tie(prefix+"div", prefix+"bad", 1)
	c.put(prefix+"pool", "Resource Pool")
}

func (c *campus) shared() {
	for i := 0; i < 6; i++ {
		p := fmt.Sprintf("w%d", i)
		c.put(p+"fleet", "TransporterFleet")
		c.put(p+"store", "Store")
		c.put(p+"get", "Retrieve")
		c.tie(p+"fleet", p+"store", 0)
		c.tie(p+"store", p+"get", 0)
	}
	for i := 0; i < 4; i++ {
		p := fmt.Sprintf("f%d", i)
		c.put(p+"src", "FluidSource")
		c.put(p+"tank", "Tank")
		c.put(p+"valve", "Valve")
		c.put(p+"out", "FluidDispose")
		c.tie(p+"src", p+"tank", 0)
		c.tie(p+"tank", p+"valve", 0)
		c.tie(p+"valve", p+"out", 0)
	}
	c.peds()
	c.traffic()
}

func (c *campus) peds() {
	for i := 0; i < 4; i++ {
		p := fmt.Sprintf("p%d", i)
		seq := []struct{ name, typ string }{
			{"src", "PedSource"}, {"go", "PedGoTo"}, {"wait", "PedWait"}, {"svc", "PedService"}, {"end", "PedSink"},
		}
		var prev string
		for _, step := range seq {
			name := p + step.name
			c.put(name, step.typ)
			if prev != "" {
				c.tie(prev, name, 0)
			}
			prev = name
		}
	}
}

func (c *campus) traffic() {
	for i := 0; i < 3; i++ {
		p := fmt.Sprintf("c%d", i)
		c.put(p+"src", "CarSource")
		c.put(p+"go", "CarMoveTo")
		c.put(p+"end", "CarDispose")
		c.tie(p+"src", p+"go", 0)
		c.tie(p+"go", p+"end", 0)
	}
	for i := 0; i < 2; i++ {
		p := fmt.Sprintf("r%d", i)
		c.put(p+"src", "TrainSource")
		c.put(p+"go", "TrainMoveTo")
		c.put(p+"end", "TrainDispose")
		c.tie(p+"src", p+"go", 0)
		c.tie(p+"go", p+"end", 0)
	}
}

func (c *campus) put(name, typ string) {
	c.els = append(c.els, element{Name: name, Type: typ, W: 156, H: 84})
}

func (c *campus) tie(from, to string, port int) {
	c.links = append(c.links, pageLink{From: from, To: to, FromPort: port})
}
