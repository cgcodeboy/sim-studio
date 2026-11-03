package simulation

import (
	behavior "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	dynamics "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/domain"
	experiment "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
	fluid "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
	handling "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	pedestrian "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
	process "github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	rail "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	road "github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
)

func flowBlock(el model.Element) process.Block  { return process.Block(copyBlock(el)) }
func haulBlock(el model.Element) handling.Block { return handling.Block(copyBlock(el)) }
func pedBlock(el model.Element) pedestrian.Block {
	return pedestrian.Block(copyBlock(el))
}
func roadBlock(el model.Element) road.Block   { return road.Block(copyBlock(el)) }
func railBlock(el model.Element) rail.Block   { return rail.Block(copyBlock(el)) }
func fluidBlock(el model.Element) fluid.Block { return fluid.Block(copyBlock(el)) }
func dynBlock(el model.Element) dynamics.Block {
	return dynamics.Block(copyBlock(el))
}
func actBlock(el model.Element) behavior.Block { return behavior.Block(copyBlock(el)) }
func expBlock(el model.Element) experiment.Block {
	return experiment.Block(copyBlock(el))
}

type copied struct {
	ID, Type   string
	X, Y, W, H float64
	Props      map[string]any
}

func copyBlock(el model.Element) copied {
	return copied{ID: el.ID, Type: el.Type, X: el.X, Y: el.Y, W: el.W, H: el.H, Props: el.Props}
}
