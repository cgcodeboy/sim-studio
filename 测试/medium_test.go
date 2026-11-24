package scenario

import "testing"

func TestMediumInspection(t *testing.T) { play(t, mediumScenes()[0]) }
func TestMediumConveyor(t *testing.T)   { play(t, mediumScenes()[1]) }
func TestMediumStorage(t *testing.T)    { play(t, mediumScenes()[2]) }
func TestMediumFluid(t *testing.T)      { play(t, mediumScenes()[3]) }
func TestMediumRewire(t *testing.T)     { play(t, mediumScenes()[4]) }

func mediumScenes() []scene {
	return []scene{inspection(), conveyorLine(), storageLine(), fluidLine(), rewire()}
}

func inspection() scene {
	return scene{name: "两路下线再改时间", level: "中等", turns: []round{
		{say: "瓶子先排队，再检验。合格下线，不合格也要单独下线，都连上。", want: expect{
			written: true, has: []string{"Source", "Queue", "Select Output", "Sink"},
			path: [][2]string{{"Source", "Select Output"}}, atLeast: map[string]int{"Sink": 2},
		}},
		{say: "检验前面那个队列，容量改成 25。", want: expect{written: true, holds: "25"}},
		{say: "看一下不合格是不是单独一个口。要是还和合格共用，把另一个口接到新的下线。", want: expect{
			atLeast: map[string]int{"Sink": 2}, ban: []string{"做不到"},
		}},
		{say: "合格和不合格分成两排摆，别在同一条横线上。按这条检验线优化布局。", want: expect{
			written: true, layout: true, spread: "Sink",
		}},
		{say: "开始运行。", want: expect{written: true, intent: "run"}},
		{say: "先暂停，我看眼两排下线。", want: expect{written: true, intent: "pause"}},
	}}
}

func conveyorLine() scene {
	return scene{name: "传送带卡进主线", level: "中等", turns: []round{
		{say: "来料、清洗、下线，先连成一条。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "清洗改成 7 分钟。", want: expect{written: true, holds: "7"}},
		{say: "加一条传送带，卡在来料和清洗中间。来料不要再直接连清洗，先上传送带，传送带再进清洗。", want: expect{
			written: true, has: []string{"Conveyor"},
			hop:   [][2]string{{"Source", "Conveyor"}, {"Conveyor", "Delay"}},
			noHop: [][2]string{{"Source", "Delay"}},
		}},
		{say: "清洗后面还得接着下线，这条别断。", want: expect{
			path: [][2]string{{"Delay", "Sink"}}, hop: [][2]string{{"Conveyor", "Delay"}},
		}},
		{say: "整条物料线按流向优化布局。", want: expect{written: true, layout: true, holds: "7"}},
	}}
}

func storageLine() scene {
	return scene{name: "搬运和药液都别串线", level: "中等", turns: []round{
		{say: "先做瓶子。来料，洗一下，下线，连起来。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "旁边做一套搬运，别和瓶子连上。车队、入库、出库自己连起来。", want: expect{
			written: true, has: []string{"TransporterFleet", "Store", "Retrieve"},
			path: [][2]string{{"Store", "Retrieve"}}, cut: [][2]string{{"Source", "Store"}},
		}},
		{say: "再做一条药液，别碰瓶子，也别碰入库。从源头进储罐，过阀门，然后排掉，自己连上。", want: expect{
			written: true, has: []string{"FluidSource", "Tank", "Valve", "FluidDispose"},
			path: [][2]string{{"FluidSource", "FluidDispose"}},
			cut:  [][2]string{{"Source", "Tank"}, {"Tank", "Store"}, {"FluidSource", "Store"}},
		}},
		{say: "车队改成 5 辆，瓶子的清洗改成 3 分钟。药液那条别动。", want: expect{
			written: true, holds: "5", path: [][2]string{{"FluidSource", "FluidDispose"}},
			cut: [][2]string{{"Source", "Tank"}},
		}},
		{say: "清洗是 3 分钟这事确认一下，车队那 5 辆也还在。", want: expect{
			holds: "3", path: [][2]string{{"Store", "Retrieve"}, {"FluidSource", "Valve"}},
		}},
		{say: "瓶子、药液、入库分成不同的排，不要摆在同一条横线上。优化布局。", want: expect{
			written: true, layout: true, row: [][2]string{{"Source", "Tank"}, {"Source", "Store"}},
		}},
	}}
}

func fluidLine() scene {
	return scene{name: "行人单独绕开瓶子", level: "中等", turns: []round{
		{say: "先把瓶子做了：来料、清洗、下线，连上。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "厂区边上加人。行人出现，走到服务点，等一下，接受服务，然后离开。这条自己连上，别进瓶子的来料，也别进瓶子的下线。", want: expect{
			written: true, has: []string{"PedSource", "PedGoTo", "PedWait", "PedService", "PedSink"},
			path: [][2]string{{"PedSource", "PedSink"}},
			cut:  [][2]string{{"PedSource", "Sink"}, {"Source", "PedSink"}},
		}},
		{say: "行人服务改成 2 分钟。瓶子那条别改。", want: expect{
			written: true, holds: "2", path: [][2]string{{"Source", "Sink"}, {"PedSource", "PedSink"}},
		}},
		{say: "再看一眼，行人有没有被连到来料或者瓶子的下线。有的话拆掉。", want: expect{
			cut: [][2]string{{"PedSource", "Sink"}, {"Source", "PedSink"}, {"PedService", "Sink"}},
		}},
		{say: "瓶子和行人分成两排，优化布局。", want: expect{
			written: true, layout: true, row: [][2]string{{"Source", "PedSource"}},
		}},
	}}
}

func rewire() scene {
	return scene{name: "时间改完又改回", level: "中等", turns: []round{
		{say: "来料、加工、下线连上。加工用延迟。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, path: [][2]string{{"Source", "Delay"}, {"Delay", "Sink"}},
		}},
		{say: "加工改成 10 分钟。", want: expect{written: true, holds: "10"}},
		{say: "不对，太慢了，改回 4 分钟。", want: expect{written: true, holds: "4"}},
		{say: "下线前加检验。合格走原来的下线，不合格另接一条下线。加工还是 4 分钟，别改掉。", want: expect{
			written: true, has: []string{"Select Output"}, atLeast: map[string]int{"Sink": 2},
			path: [][2]string{{"Delay", "Select Output"}}, holds: "4",
		}},
		{say: "两条下线不要摆在同一排，优化布局。", want: expect{
			written: true, layout: true, spread: "Sink", holds: "4",
		}},
	}}
}
