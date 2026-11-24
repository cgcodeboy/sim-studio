package dialogue

import "testing"

func TestAskedCopiesReadsOneNumber(t *testing.T) {
	if askedCopies("同样的灌装线再做 80 条，每条自己连。") != 80 {
		t.Fatal(askedCopies("同样的灌装线再做 80 条，每条自己连。"))
	}
	if askedCopies("再做 12 套一样的搬运。") != 12 {
		t.Fatal("套")
	}
	if askedCopies("再做 3 条车道和 2 条铁路。") != 1 {
		t.Fatal("两个数字不能猜")
	}
	if askedCopies("清洗改成 8 分钟。") != 1 {
		t.Fatal("分钟")
	}
}

func TestFoldCopiesKeepsOnePattern(t *testing.T) {
	steps := foldCopies([]planStep{
		{Intent: "place", Types: []string{"Source", "Delay", "Sink"}},
		{Intent: "place", Types: []string{"Source", "Delay", "Sink"}},
		{Intent: "layout"},
	}, "再做 40 条")
	if len(steps) != 1 || steps[0].Lines != 40 {
		t.Fatalf("%+v", steps)
	}
}

func TestRepeatedPlaceCopiesThePattern(t *testing.T) {
	task, ok := repeatedPlace(planStep{Intent: "place", Types: []string{"Source", "Delay", "Sink"}, Lines: 80}, "Main", nil)
	if !ok || !task.Chain || len(task.Types) != 3 || len(task.More) != 79 {
		t.Fatalf("ok %v types %v more %d", ok, task.Types, len(task.More))
	}
	if task.More[0].Types[0] != "Source" || !task.More[78].Chain {
		t.Fatal(task.More[0])
	}
}

func TestPackKeepsLineCopies(t *testing.T) {
	head := pack([]Task{{
		Intent: "place", Status: "已写入",
		More: []Task{{Intent: "place"}, {Intent: "place"}},
	}, {Intent: "layout", Status: "已写入"}}, "先铺开再摆。")
	if len(head.More) != 3 || head.More[2].Intent != "layout" || head.Reply != "先铺开再摆。" {
		t.Fatalf("%+v", head)
	}
}
