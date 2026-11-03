package app

import (
	behaviordom "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	behaviorrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/repo"
	behaviorsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/service"
	dynamicsdom "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/domain"
	dynamicsrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/repo"
	dynamicssvc "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/service"
	experimentdom "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
	experimentrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/repo"
	experimentsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/service"
	fluiddom "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
	fluidrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/repo"
	fluidsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/service"
	handlingdom "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
	handlingrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/repo"
	handlingsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/service"
	pedestriandom "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
	pedestrianrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/repo"
	pedestriansvc "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/service"
	processdom "github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	processrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/process/repo"
	processsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/process/service"
	raildom "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	railrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/repo"
	railsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/service"
	roaddom "github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
	roadrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/road/repo"
	roadsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/road/service"
)

type processOpener struct{}
type handlingOpener struct{}
type pedestrianOpener struct{}
type roadOpener struct{}
type railOpener struct{}
type fluidOpener struct{}
type dynamicsOpener struct{}
type behaviorOpener struct{}
type experimentOpener struct{}

func (processOpener) Open(board processdom.Board, params map[string]float64) processdom.Service {
	return processsvc.New(processrepo.New(board, params))
}
func (handlingOpener) Open(board handlingdom.Board, params map[string]float64) handlingdom.Service {
	return handlingsvc.New(handlingrepo.New(board, params))
}
func (pedestrianOpener) Open(board pedestriandom.Board, params map[string]float64) pedestriandom.Service {
	return pedestriansvc.New(pedestrianrepo.New(board, params))
}
func (roadOpener) Open(board roaddom.Board, params map[string]float64) roaddom.Service {
	return roadsvc.New(roadrepo.New(board, params))
}
func (railOpener) Open(board raildom.Board, params map[string]float64) raildom.Service {
	return railsvc.New(railrepo.New(board, params))
}
func (fluidOpener) Open(board fluiddom.Board, params map[string]float64) fluiddom.Service {
	return fluidsvc.New(fluidrepo.New(board, params))
}
func (dynamicsOpener) Open(board dynamicsdom.Board, params map[string]float64) dynamicsdom.Service {
	return dynamicssvc.New(dynamicsrepo.New(board, params))
}
func (behaviorOpener) Open(board behaviordom.Board, params map[string]float64) behaviordom.Service {
	return behaviorsvc.New(behaviorrepo.New(board, params))
}
func (experimentOpener) Open(board experimentdom.Board, params map[string]float64) experimentdom.Service {
	return experimentsvc.New(experimentrepo.New(board, params))
}

func solvers() (processdom.Opener, handlingdom.Opener, pedestriandom.Opener, roaddom.Opener, raildom.Opener, fluiddom.Opener, dynamicsdom.Opener, behaviordom.Opener, experimentdom.Opener) {
	return processOpener{}, handlingOpener{}, pedestrianOpener{}, roadOpener{}, railOpener{}, fluidOpener{}, dynamicsOpener{}, behaviorOpener{}, experimentOpener{}
}
