package service

import (
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
)

// Service 实现本领域的入口，只通过仓储接口改状态。
type Service struct{ repo domain.Repository }

// New 把仓储交给服务。仓储由 app 创建，service 不引用 repo 包。
func New(repo domain.Repository) *Service { return &Service{repo: repo} }

// Place 按本领域的规则装一个元件。不是本领域的元件返回 false。
func (s *Service) Place(block domain.Block) (bool, error) {
	steps := []func(domain.Repository, domain.Block) (bool, error){
		PlacePopulation,
	}
	for _, step := range steps {
		ok, err := step(s.repo, block)
		if ok || err != nil {
			return ok, err
		}
	}
	return false, nil
}

// Tune 把设置抄进仓储里的参数表。
func (s *Service) Tune(block domain.Block) { _ = block }

func (s *Service) Steps() []func(time.Duration)       { return s.repo.Steps() }
func (s *Service) Shell() bool                        { return s.repo.Shell() }
func (s *Service) Dynamics() bool                     { return s.repo.Dynamics() }
func (s *Service) SetBind(fn func(string, any) error) { s.repo.SetBind(fn) }
func (s *Service) SetAgent(agent domain.Agent)        { s.repo.SetAgent(agent) }
