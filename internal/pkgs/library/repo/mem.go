package repo

import (
	"context"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
)

// Mem 是本程序内置的类型目录。
type Mem struct{}

// New 返回内置目录。
func New() *Mem { return &Mem{} }

// List 返回类型和它所属的求解种类。
func (Mem) List(context.Context) ([]domain.Type, error) {
	var out []domain.Type
	for kind, names := range groups {
		for _, name := range names {
			out = append(out, domain.Type{Name: name, Kind: kind})
		}
	}
	return out, nil
}

var groups = map[string][]string{
	"process":    {"Source", "Sink", "Delay", "Queue", "Hold", "Wait", "Enter", "Exit", "PlainTransfer", "MoveTo", "SelectOutput", "SelectOutput5", "SelectOutputIn", "SelectOutputOut", "Split", "Combine", "Batch", "Unbatch", "Pickup", "Dropoff", "Assembler", "TimeMeasureStart", "TimeMeasureEnd", "RestrictedAreaStart", "RestrictedAreaEnd", "ResourcePool", "Seize", "Release", "Service", "ResourceSendTo", "ResourceTaskStart", "ResourceTaskEnd", "ResourceAttach", "ResourceDetach", "PMLSettings"},
	"handling":   {"Conveyor", "Convey", "ConveyorEnter", "TransporterFleet", "TransporterControl", "SeizeTransporter", "ReleaseTransporter", "MoveByTransporter", "RackSystem", "RackStore", "RackPick", "Store", "Retrieve", "MHLSettings"},
	"pedestrian": {"PedSource", "PedEnter", "PedGoTo", "PedWait", "PedService", "PedSelectOutput", "PedExit", "PedSink", "PedGroupAssemble", "PedGroupDisassemble", "PedSettings"},
	"road":       {"CarSource", "CarEnter", "CarMoveTo", "CarExit", "CarDispose"},
	"rail":       {"TrainSource", "TrainMoveTo", "TrainDispose"},
	"fluid":      {"FluidSource", "Tank", "Valve", "ProcessTank", "FluidSplit", "FluidSelectOutput", "FluidDispose"},
	"dynamics":   {"Stock", "Flow", "DynamicVariable", "Parameter"},
	"behavior":   {"Statechart", "State", "Event", "Schedule", "TableFunction", "Table", "InputTable", "Connection", "AgentLink"},
	"experiment": {"Population"},
	"camera":     {"Camera", "Light"},
	"geometry":   {"ConveyorPathDescriptor", "PositionOnConveyorDescriptor", "ConveyorSimpleStationDescriptor", "ConveyorTransferTableDescriptor", "ConveyorTurntableDescriptor", "ConveyorTurnStationDescriptor", "PedAreaDescriptor", "GISRoute", "Path"},
	"refuse":     {"GISMap"},
	"skip":       {"Sound", "Text", "Rectangle", "Oval", "Chart", "Control", "Button", "Slider", "Link", "Wall", "TargetLine", "Pathway", "Area", "Road", "Intersection", "StopLine", "ParkingLot", "RailwayTrack", "Pipe", "Station", "Turntable", "GISPoint", "GISRegion"},
}
