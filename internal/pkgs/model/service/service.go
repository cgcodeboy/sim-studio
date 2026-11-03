package service

import (
	"context"
	"errors"
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
)

type Service struct {
	projects domain.Repository
}

func New(projects domain.Repository) *Service {
	return &Service{projects: projects}
}

func (s *Service) List(ctx context.Context) ([]domain.Summary, error) {
	return s.projects.List(ctx)
}

func (s *Service) Read(ctx context.Context, id string) (domain.Project, error) {
	p, err := s.projects.Read(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Project{}, &domain.Error{Code: domain.CodeNotFound}
	}
	return p, err
}

func (s *Service) Put(ctx context.Context, expected int64, p domain.Project) (domain.Project, error) {
	domain.Normalize(&p)
	if err := domain.Validate(p); err != nil {
		return domain.Project{}, err
	}
	if expected == 0 {
		p.Revision = 1
		p.UpdatedAt = time.Now().UTC()
		err := s.projects.Insert(ctx, p)
		if errors.Is(err, domain.ErrDuplicate) {
			cur, readErr := s.Read(ctx, p.ID)
			if readErr != nil {
				return domain.Project{}, readErr
			}
			return domain.Project{}, &domain.Error{Code: domain.CodeConflict, Project: &cur}
		}
		if err != nil {
			return domain.Project{}, err
		}
		return p, nil
	}
	p.Revision = expected + 1
	p.UpdatedAt = time.Now().UTC()
	matched, err := s.projects.Replace(ctx, expected, p)
	if err != nil {
		return domain.Project{}, err
	}
	if matched {
		return p, nil
	}
	cur, readErr := s.Read(ctx, p.ID)
	if readErr != nil {
		return domain.Project{}, readErr
	}
	return domain.Project{}, &domain.Error{Code: domain.CodeConflict, Project: &cur}
}

func (s *Service) IDs(ctx context.Context) ([]string, error) {
	return s.projects.IDs(ctx)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	deleted, err := s.projects.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return &domain.Error{Code: domain.CodeNotFound}
	}
	return nil
}
