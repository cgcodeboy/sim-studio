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

func TestSessionsBelongToTheProject(t *testing.T) {
	svc := New(repo.New())
	ctx := context.Background()
	first, err := svc.Create(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Create(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.Create(ctx, "p2")
	if err != nil {
		t.Fatal(err)
	}
	if first.Title != "会话 1" || second.Title != "会话 2" || other.Project != "p2" {
		t.Fatalf("%+v %+v %+v", first, second, other)
	}
	_, err = svc.Commit(ctx, domain.Commit{Project: "p1", Session: first.ID, Turn: domain.Turn{User: "洗瓶要 8 分钟", Reply: "已改。", Status: "已写入"}})
	if err != nil {
		t.Fatal(err)
	}
	view, err := svc.Recall(ctx, "p1", second.ID)
	if err != nil || len(view.Recent) != 0 {
		t.Fatalf("%+v %v", view, err)
	}
	opened, err := svc.Open(ctx, first.ID)
	if err != nil || len(opened.Turns) != 1 || opened.Title != "洗瓶要 8 分钟" {
		t.Fatalf("%+v %v", opened, err)
	}
	renamed, err := svc.Rename(ctx, first.ID, "粉针")
	if err != nil || !renamed.Named || renamed.Title != "粉针" {
		t.Fatalf("%+v", renamed)
	}
	if _, err = svc.Commit(ctx, domain.Commit{Project: "p1", Session: first.ID, Turn: domain.Turn{User: "再加一道灯检", Status: "未写入"}}); err != nil {
		t.Fatal(err)
	}
	kept, err := svc.Open(ctx, first.ID)
	if err != nil || kept.Title != "粉针" || len(kept.Turns) != 2 {
		t.Fatalf("%+v", kept)
	}
	cards, err := svc.List(ctx, "p1")
	if err != nil || len(cards) != 2 {
		t.Fatalf("%+v %v", cards, err)
	}
	seen := map[string]bool{}
	for i, card := range cards {
		seen[card.ID] = true
		if i > 0 && card.Updated > cards[i-1].Updated {
			t.Fatalf("order %+v", cards)
		}
	}
	if !seen[first.ID] || !seen[second.ID] {
		t.Fatalf("%+v", cards)
	}
	if err = svc.Delete(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Open(ctx, second.ID); err != domain.ErrNotFound {
		t.Fatal(err)
	}
}
