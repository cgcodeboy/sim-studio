package scenario

import "testing"

func TestSimpleBottling(t *testing.T)       { play(t, simpleScenes()[0]) }
func TestSimpleQueueEdit(t *testing.T)      { play(t, simpleScenes()[1]) }
func TestSimpleAskBeforeBuild(t *testing.T) { play(t, simpleScenes()[2]) }

func simpleScenes() []scene {
	return []scene{bottling(), queueEdit(), askBeforeBuild()}
}

func bottling() scene {
	return scene{name: "灌装中间加队列", level: "简单", turns: []round{
		{say: "帮我在 Main 上做灌装。瓶子从来料进来，清洗，然后下线，三段连上。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, minLinks: 2,
			path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "清洗改成 6 分钟。", want: expect{written: true, holds: "6"}},
		{say: "来料和清洗之间加一个队列，容量 15。瓶子要先排队，再去清洗。", want: expect{
			written: true, has: []string{"Queue"}, holds: "15",
			path: [][2]string{{"Source", "Queue"}, {"Queue", "Delay"}},
		}},
		{say: "可以了，开始运行。", want: expect{written: true, intent: "run"}},
	}}
}

func queueEdit() scene {
	return scene{name: "删掉排队两边接上", level: "简单", turns: []round{
		{say: "零件从来料进，先排队，再加工，最后下线。按这个顺序连好。", want: expect{
			written: true, has: []string{"Source", "Queue", "Delay", "Sink"},
			path: [][2]string{{"Source", "Queue"}, {"Queue", "Delay"}, {"Delay", "Sink"}},
		}},
		{say: "排队这步不要了，删掉。", want: expect{
			written: true, absent: []string{"Queue"}, has: []string{"Source", "Delay", "Sink"},
		}},
		{say: "来料改成直接进加工，加工后面还是要接着下线。", want: expect{
			path: [][2]string{{"Source", "Delay"}, {"Delay", "Sink"}},
		}},
		{say: "加工时间改成 4 分钟。", want: expect{written: true, holds: "4"}},
	}}
}

func askBeforeBuild() scene {
	return scene{name: "先问清再补队列", level: "简单", turns: []round{
		{say: "你看着做一条线吧。", want: expect{open: true, ban: []string{"做不到"}}},
		{say: "做注射液。来料，用延迟做灭菌，然后下线，连起来。", want: expect{
			written: true, has: []string{"Source", "Delay", "Sink"}, path: [][2]string{{"Source", "Sink"}},
		}},
		{say: "灭菌改成 12 分钟。下线前面再加一个队列，容量 5，灭菌完先进队列，再下线。", want: expect{
			written: true, has: []string{"Queue"}, holds: "12",
			path: [][2]string{{"Delay", "Queue"}, {"Queue", "Sink"}},
		}},
		{say: "先别运行。", want: expect{forbid: "run", holds: "12"}},
	}}
}
