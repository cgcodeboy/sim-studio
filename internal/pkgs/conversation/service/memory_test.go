package service

import (
	"context"
	"testing"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/repo"
)

func TestRecallKeepsTheLatestTurns(t *testing.T) {
	svc := New(repo.New())
	for i := 0; i < 7; i++ {
		_, err := svc.Commit(context.Background(), domain.Commit{Project: "p", Session: "s", Turn: domain.Turn{User: string(rune('a' + i)), Reply: "好", Status: "未写入"}})
		if err != nil {
			t.Fatal(err)
		}
	}
	view, err := svc.Recall(context.Background(), "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Recent) != 6 || view.Recent[0].User != "b" || view.Summary != "" {
		t.Fatalf("%+v", view)
	}
}

func TestPlaceBecomesAFactAndTheOldestDrops(t *testing.T) {
	svc := New(repo.New())
	for i := 0; i < 31; i++ {
		_, err := svc.Commit(context.Background(), domain.Commit{Project: "p", Session: "s", Turn: domain.Turn{User: "粉针线", Reply: "加了模块", Intent: "place", Status: "已写入"}})
		if err != nil {
			t.Fatal(err)
		}
	}
	view, err := svc.Recall(context.Background(), "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Facts) != 20 {
		t.Fatalf("facts %d", len(view.Facts))
	}
}

func TestSummaryStartsAfterTwelveTurns(t *testing.T) {
	svc := New(repo.New())
	var job domain.Job
	for i := 0; i < 13; i++ {
		next, err := svc.Commit(context.Background(), domain.Commit{Project: "p", Session: "s", Turn: domain.Turn{User: "一句", Status: "未写入"}})
		if err != nil {
			t.Fatal(err)
		}
		job = next
	}
	if len(job.Turns) != 7 || job.Covered != 7 {
		t.Fatalf("%+v", job)
	}
}
