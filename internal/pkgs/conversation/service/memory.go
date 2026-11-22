// Package service 实现会话的读写。它不调用模型，也不读取工程正文。
package service

import (
	"context"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

const (
	recentTurns  = 6
	summaryAfter = 12
	maxFacts     = 30
	shownFacts   = 20
)

// Service 实现会话入口。
type Service struct{ repo domain.Repository }

func New(repo domain.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Recall(ctx context.Context, project, session string) (domain.View, error) {
	if session == "" {
		return domain.View{}, nil
	}
	stored, err := s.repo.Load(ctx, session)
	if err != nil {
		return domain.View{}, err
	}
	facts, err := s.repo.Facts(ctx, project)
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{Recent: lastTurns(stored.Turns, recentTurns), Summary: stored.Summary, Facts: factText(facts, shownFacts)}, nil
}

func (s *Service) Commit(ctx context.Context, in domain.Commit) (domain.Job, error) {
	if in.Session == "" || in.Project == "" {
		return domain.Job{}, nil
	}
	stored, err := s.repo.Load(ctx, in.Session)
	if err != nil {
		return domain.Job{}, err
	}
	stored.ID = in.Session
	stored.Project = in.Project
	first := len(stored.Turns) == 0
	stored.Turns = append(stored.Turns, in.Turn)
	if !stored.Named && first {
		if title := clipText(in.Turn.User, 18); title != "" {
			stored.Title = title
		}
	}
	if stored.Title == "" {
		stored.Title = "会话 1"
	}
	stored.Updated = now()
	if err := s.repo.Save(ctx, stored); err != nil {
		return domain.Job{}, err
	}
	if err := s.keepFact(ctx, in); err != nil {
		return domain.Job{}, err
	}
	return summaryJob(stored), nil
}

func (s *Service) SaveSummary(ctx context.Context, session, summary string, covered int) error {
	if session == "" {
		return nil
	}
	stored, err := s.repo.Load(ctx, session)
	if err != nil {
		return err
	}
	stored.ID = session
	stored.Summary = summary
	stored.Covered = covered
	return s.repo.Save(ctx, stored)
}

func (s *Service) keepFact(ctx context.Context, in domain.Commit) error {
	text := factOf(in.Turn)
	if text == "" {
		return nil
	}
	facts, err := s.repo.Facts(ctx, in.Project)
	if err != nil {
		return err
	}
	facts = append(facts, domain.Fact{Text: text})
	if len(facts) > maxFacts {
		facts = facts[len(facts)-maxFacts:]
	}
	return s.repo.ReplaceFacts(ctx, in.Project, facts)
}

func summaryJob(stored domain.Session) domain.Job {
	end := len(stored.Turns) - recentTurns
	if len(stored.Turns) <= summaryAfter || stored.Covered >= end {
		return domain.Job{}
	}
	return domain.Job{Turns: append([]domain.Turn(nil), stored.Turns[stored.Covered:end]...), Prior: stored.Summary, Covered: end}
}

func lastTurns(turns []domain.Turn, n int) []domain.Turn {
	if len(turns) <= n {
		return append([]domain.Turn(nil), turns...)
	}
	return append([]domain.Turn(nil), turns[len(turns)-n:]...)
}

func factText(facts []domain.Fact, n int) []string {
	if len(facts) > n {
		facts = facts[len(facts)-n:]
	}
	out := make([]string, 0, len(facts))
	for _, fact := range facts {
		out = append(out, fact.Text)
	}
	return out
}

func factOf(turn domain.Turn) string {
	if turn.Status != "已写入" || (turn.Intent != "place" && turn.Intent != "set") {
		return ""
	}
	if turn.Intent == "set" {
		return turn.Reply
	}
	return clipText(turn.User, 80) + " → " + turn.Reply
}

func clipText(text string, n int) string {
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[:n])
}
