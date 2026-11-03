package domain

import (
	"errors"
	"time"
)

// ErrInvalid 表示这个元件不能按本领域的规则装上。
var ErrInvalid = errors.New("invalid")

// Block 是调用方抄进来的一个元件。本领域不引用模型。
type Block struct {
	ID, Type   string
	X, Y, W, H float64
	Props      map[string]any
}

// Link 是抄进来的一根连线。
type Link struct{ From, To string }

// Agent 是抄进来的智能体。
type Agent struct {
	Elements []Block
	Links    []Link
}

// Board 是样本板。由运行领域的样本实现，本领域只定义需要的方法。
type Board interface {
	AddFlow(id string, arrived, exited, overflow float64)
	SetQueued(id string, n int)
	Enter(id string, x, y float64)
	Leave(id string)
	SetSeries(name, source string, value float64)
	SetMover(id string, x, y, heading float64)
	Spots() []struct {
		ID string
		X  float64
		Y  float64
	}
}

// Units 是资源或货位的剩余数量。
type Units struct{ Free int }

// Gate 是限制区的容量和占用。
type Gate struct {
	Cap  int
	Used int
}

// Turnout 是一副道岔的占用。
type Turnout struct{ Busy bool }

// Repository 是本领域这一次装载的状态。service 只依赖这个接口。
type Repository interface {
	Board() Board
	Add(id string, model any) error
	SetBind(func(string, any) error)
	Params() map[string]float64
	EnsurePool(name string, free int) *Units
	EnsureArea(name string, cap int) *Gate
	EnsureSwitch(name string) *Turnout
	AppendStep(func(time.Duration))
	Steps() []func(time.Duration)
	MarkShell()
	Shell() bool
	MarkDynamics()
	Dynamics() bool
	Agent() Agent
	SetAgent(Agent)
}

// Service 是本领域对外的入口。handler 只调用它。
type Service interface {
	Place(block Block) (bool, error)
	Tune(block Block)
	Steps() []func(time.Duration)
	Shell() bool
	Dynamics() bool
	SetBind(func(string, any) error)
	SetAgent(Agent)
}

// Opener 为一次运行打开一个带新仓储的服务。
type Opener interface {
	Open(board Board, params map[string]float64) Service
}
