package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

func (s *Service) List(ctx context.Context, project string) ([]domain.Card, error) {
	items, err := s.repo.List(ctx, project)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Updated > items[j].Updated })
	cards := make([]domain.Card, 0, len(items))
	for _, item := range items {
		cards = append(cards, cardOf(item))
	}
	return cards, nil
}

func (s *Service) Create(ctx context.Context, project string) (domain.Session, error) {
	if strings.TrimSpace(project) == "" {
		return domain.Session{}, domain.ErrInvalid
	}
	existing, err := s.repo.List(ctx, project)
	if err != nil {
		return domain.Session{}, err
	}
	session := domain.Session{
		ID:      newID(),
		Project: project,
		Title:   nextTitle(existing),
		Turns:   []domain.Turn{},
		Updated: now(),
	}
	if err := s.repo.Save(ctx, session); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

func (s *Service) Open(ctx context.Context, id string) (domain.Session, error) {
	stored, err := s.repo.Load(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}
	if stored.Project == "" {
		return domain.Session{}, domain.ErrNotFound
	}
	if stored.Turns == nil {
		stored.Turns = []domain.Turn{}
	}
	return stored, nil
}

func (s *Service) Rename(ctx context.Context, id, title string) (domain.Session, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Session{}, domain.ErrInvalid
	}
	stored, err := s.Open(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}
	stored.Title = title
	stored.Named = true
	stored.Updated = now()
	if err := s.repo.Save(ctx, stored); err != nil {
		return domain.Session{}, err
	}
	return stored, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.Open(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func cardOf(session domain.Session) domain.Card {
	return domain.Card{ID: session.ID, Project: session.Project, Title: session.Title, Named: session.Named, Updated: session.Updated}
}

func nextTitle(items []domain.Session) string {
	used := map[string]bool{}
	for _, item := range items {
		used[item.Title] = true
	}
	for n := 1; ; n++ {
		title := "会话 " + strconv.Itoa(n)
		if !used[title] {
			return title
		}
	}
}

func newID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func now() int64 { return time.Now().UnixMilli() }
