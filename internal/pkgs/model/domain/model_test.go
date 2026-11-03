package domain

import "testing"

func sampleProject() Project {
	return Project{
		ID: "p1", Name: "物流", Unit: "分钟",
		Agents: []Agent{{
			ID: "main", Name: "Main", Space: "连续", W: 800, H: 600,
			Elements: []Element{
				{ID: "a", Name: "source", Type: "Source"},
				{ID: "b", Name: "queue", Type: "Queue"},
			},
			Links: []Link{{ID: "l", From: "a", FromPort: 0, To: "b", ToPort: 0}},
		}},
		Experiments: []Experiment{{ID: "sim", Name: "Simulation", Type: "simulation"}},
	}
}

func TestRejectsBrokenLink(t *testing.T) {
	p := sampleProject()
	p.Agents[0].Links[0].To = "missing"
	if err := Validate(p); AsDoc(err) == nil || AsDoc(err).Code != CodeInvalid {
		t.Fatalf("got %v", err)
	}
}
