package scenario

func juiceCampusScene() scene {
	return scene{name: "果汁厂从建到跑", level: "复杂", turns: []round{
		{say: "Main 上先做一条果汁灌装线，按顺序连起来：来料，排队，占用清洗工，清洗，释放，灌装，检验，合格下线，最后放上清洗工。", want: expect{
			written: true, has: []string{"Source", "Queue", "Seize", "Delay", "Release", "Select Output", "Sink", "Resource Pool"},
			path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "同样的灌装线再做 80 条。每条都是来料、排队、占用、清洗、释放、灌装、检验、合格下线，清洗工放在最后。每条自己连，不要串成一条长线。", want: expect{
			written: true, atLeast: map[string]int{"Source": 80}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "再做 40 条同样的灌装线，模块顺序和上一条一样，每条自己连。", want: expect{
			written: true, atLeast: map[string]int{"Source": 100}, minCount: 300, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "再做 12 套一样的搬运。每套车队、入库、出库自己连上，别和瓶子串。", want: expect{
			written: true, atLeast: map[string]int{"TransporterFleet": 12, "Store": 12},
			path: [][2]string{{"Store", "Retrieve"}}, cut: [][2]string{{"Source", "Store"}}, minCount: 300,
		}},
		{say: "再做 10 条一样的药液线。每条从源头进储罐，过阀门，再排掉，自己连上，别碰瓶子和入库。", want: expect{
			written: true, atLeast: map[string]int{"FluidSource": 10, "Tank": 10},
			path: [][2]string{{"FluidSource", "FluidDispose"}}, cut: [][2]string{{"Source", "Tank"}, {"Tank", "Store"}},
		}},
		{say: "再做 10 条一样的行人通道。每条是出现、走到服务点、等待、接受服务、离开，自己连上，别进瓶子的下线。", want: expect{
			written: true, atLeast: map[string]int{"PedSource": 10},
			path: [][2]string{{"PedSource", "PedSink"}}, cut: [][2]string{{"PedSource", "Sink"}, {"Source", "PedSink"}},
		}},
		{say: "门口再做 8 条一样的车道。每条发车、开到卸货点、离开，自己连上，别串到瓶子上。", want: expect{
			written: true, atLeast: map[string]int{"CarSource": 8},
			path: [][2]string{{"CarSource", "CarDispose"}}, cut: [][2]string{{"Source", "CarDispose"}},
		}},
		{say: "外侧再做 6 条一样的铁路。每条出发、开到站台、离开，自己连上，别和车道串，也别串瓶子。", want: expect{
			written: true, atLeast: map[string]int{"TrainSource": 6},
			path: [][2]string{{"TrainSource", "TrainDispose"}},
			cut:  [][2]string{{"Source", "TrainDispose"}, {"CarSource", "TrainDispose"}},
		}},
		{say: "清洗改成 8 分钟。已经做好的线都留着。", want: expect{
			written: true, holds: "8", atLeast: map[string]int{"Source": 100}, minCount: 300,
		}},
		{say: "排队容量改成 15。别删模块。", want: expect{
			written: true, holds: "15", minCount: 300,
		}},
		{say: "按整场业务优化布局。物料、搬运、药液、行人、车道、铁路分成不同的排，不要叠在同一条横线上。", want: expect{
			written: true, layout: true, minCount: 300, minPlaces: 300,
			row: [][2]string{{"Source", "Tank"}, {"Source", "Store"}, {"Source", "PedSource"}, {"Source", "CarSource"}, {"Source", "TrainSource"}},
		}},
		{say: "整座厂开始运行。", want: expect{written: true, intent: "run", minCount: 300}},
	}}
}

func partsCampusScene() scene {
	return scene{name: "零件园从建到跑", level: "复杂", turns: []round{
		{say: "你先别动手，我还没说零件园怎么建。", want: expect{open: true, ban: []string{"做不到"}}},
		{say: "做一条零件线，按顺序连起来：来料，排队，用延迟加工，检验，合格下线。", want: expect{
			written: true, has: []string{"Source", "Queue", "Delay", "Select Output", "Sink"},
			path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "同样的零件线再做 80 条。每条都是来料、排队、加工、检验、合格下线，每条自己连，不要串成一条长线。", want: expect{
			written: true, atLeast: map[string]int{"Source": 80}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "再做 40 条同样的零件线，模块顺序和上一条一样，每条自己连。", want: expect{
			written: true, atLeast: map[string]int{"Source": 100}, minCount: 300,
		}},
		{say: "再做 12 套一样的搬运。每套车队、入库、出库自己连上，别串进零件线。", want: expect{
			written: true, atLeast: map[string]int{"TransporterFleet": 12, "Store": 12},
			path: [][2]string{{"Store", "Retrieve"}}, cut: [][2]string{{"Source", "Store"}}, minCount: 300,
		}},
		{say: "门口再做 8 条一样的车道。每条发车、开到卸货点、离开，自己连上，别串零件。", want: expect{
			written: true, atLeast: map[string]int{"CarSource": 8},
			path: [][2]string{{"CarSource", "CarDispose"}}, cut: [][2]string{{"Source", "CarDispose"}},
		}},
		{say: "外侧再做 6 条一样的铁路。每条出发、开到站台、离开，自己连上。别和车道串，也别串零件。", want: expect{
			written: true, atLeast: map[string]int{"TrainSource": 6},
			path: [][2]string{{"TrainSource", "TrainDispose"}},
			cut:  [][2]string{{"Source", "TrainDispose"}, {"CarSource", "TrainDispose"}},
		}},
		{say: "加工时间改成 15 分钟。已经做好的线都留着。", want: expect{
			written: true, holds: "15", atLeast: map[string]int{"Source": 100}, minCount: 300,
		}},
		{say: "排队容量改成 20。别删模块。", want: expect{
			written: true, holds: "20", minCount: 300,
		}},
		{say: "确认零件线、搬运、车道和铁路都还在。", want: expect{
			minCount: 300, atLeast: map[string]int{"Source": 100, "Store": 12, "CarSource": 8, "TrainSource": 6},
		}},
		{say: "按现在这座零件园优化布局。加工、搬运、车道、铁路分成不同的排。合格和不合格不要摆在同一排。", want: expect{
			written: true, layout: true, minCount: 300, minPlaces: 300, spread: "Sink",
			row: [][2]string{{"Source", "Store"}, {"Source", "CarSource"}, {"Source", "TrainSource"}},
		}},
		{say: "整座零件园开始运行。", want: expect{written: true, intent: "run", minCount: 300}},
	}}
}
