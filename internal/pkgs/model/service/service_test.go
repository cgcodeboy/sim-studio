package service

import (
	"testing"

	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	docrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/model/repo"
)

func sampleProject() model.Project {
	return model.Project{
		ID: "p1", Name: "物流", Unit: "分钟",
		Agents: []model.Agent{{
			ID: "main", Name: "Main", Space: "连续", W: 800, H: 600,
			Elements: []model.Element{
				{ID: "a", Name: "source", Type: "Source"},
				{ID: "b", Name: "queue", Type: "Queue"},
			},
			Links: []model.Link{{ID: "l", From: "a", FromPort: 0, To: "b", ToPort: 0}},
		}},
		Experiments: []model.Experiment{{ID: "sim", Name: "Simulation", Type: "simulation"}},
	}
}

func TestRevision(t *testing.T) {
	svc := New(docrepo.NewMem())
	p, err := svc.Put(t.Context(), 0, sampleProject())
	if err != nil || p.Revision != 1 {
		t.Fatal(err, p.Revision)
	}
	p.Name = "改名"
	next, err := svc.Put(t.Context(), 1, p)
	if err != nil || next.Revision != 2 {
		t.Fatal(err, next.Revision)
	}
	_, err = svc.Put(t.Context(), 1, p)
	if model.AsDoc(err) == nil || model.AsDoc(err).Code != model.CodeConflict || model.AsDoc(err).Project.Revision != 2 {
		t.Fatal(err)
	}
}
