package repo

import (
	"time"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
)

// Mem 实现本领域的仓储。一次运行一份，不进数据库。
type Mem struct {
	board    domain.Board
	params   map[string]float64
	pools    map[string]*domain.Units
	areas    map[string]*domain.Gate
	switches map[string]*domain.Turnout
	steps    []func(time.Duration)
	shell    bool
	dynamics bool
	agent    domain.Agent
	bind     func(string, any) error
}

// New 准备一份空仓储。
func New(board domain.Board, params map[string]float64) *Mem {
	if params == nil {
		params = map[string]float64{}
	}
	return &Mem{
		board: board, params: params,
		pools: map[string]*domain.Units{}, areas: map[string]*domain.Gate{},
		switches: map[string]*domain.Turnout{},
	}
}

func (m *Mem) Board() domain.Board                { return m.board }
func (m *Mem) Params() map[string]float64         { return m.params }
func (m *Mem) Steps() []func(time.Duration)       { return m.steps }
func (m *Mem) Shell() bool                        { return m.shell }
func (m *Mem) Dynamics() bool                     { return m.dynamics }
func (m *Mem) Agent() domain.Agent                { return m.agent }
func (m *Mem) SetAgent(agent domain.Agent)        { m.agent = agent }
func (m *Mem) SetBind(fn func(string, any) error) { m.bind = fn }
func (m *Mem) MarkShell()                         { m.shell = true }
func (m *Mem) MarkDynamics()                      { m.dynamics = true }
func (m *Mem) AppendStep(fn func(time.Duration))  { m.steps = append(m.steps, fn) }

func (m *Mem) Add(id string, model any) error {
	if m.bind == nil {
		return domain.ErrInvalid
	}
	return m.bind(id, model)
}

func (m *Mem) EnsurePool(name string, free int) *domain.Units {
	if m.pools[name] == nil {
		m.pools[name] = &domain.Units{Free: free}
	}
	return m.pools[name]
}

func (m *Mem) EnsureArea(name string, cap int) *domain.Gate {
	if m.areas[name] == nil {
		m.areas[name] = &domain.Gate{Cap: cap}
	}
	return m.areas[name]
}

func (m *Mem) EnsureSwitch(name string) *domain.Turnout {
	if m.switches[name] == nil {
		m.switches[name] = &domain.Turnout{}
	}
	return m.switches[name]
}
