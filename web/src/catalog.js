const text = (key, label, def = "") => ({ key, label, type: "text", def });
const num = (key, label, def) => ({ key, label, type: "number", def });
const sel = (key, label, options, def) => ({ key, label, type: "select", options, def: def ?? options[0] });
const check = (key, label, def = false) => ({ key, label, type: "check", def });
const area = (key, label, def = "") => ({ key, label, type: "area", def });

const ENTRIES = new Set(["Source", "Enter", "PedSource", "PedEnter", "TrainSource", "TrainEnter", "CarSource", "CarEnter", "FluidSource", "ConveyorEnter"]);
const EXITS = new Set(["Sink", "Exit", "PedSink", "PedExit", "TrainDispose", "TrainExit", "CarDispose", "FluidDispose", "ConveyorExit"]);

export function isEntry(type) { return ENTRIES.has(type); }
export function isExit(type) { return EXITS.has(type); }

function autoPorts(type) {
  if (isEntry(type)) return { inn: 0, out: 1 };
  if (isExit(type)) return { inn: 1, out: 0 };
  if (type === "Select Output") return { inn: 1, out: 2 };
  if (type === "Select Output 5") return { inn: 1, out: 5 };
  return { inn: 1, out: 1 };
}

function flow(type, fields = [], ports) {
  return { type, kind: "flow", look: "flow", fields, ports: ports || autoPorts(type), w: type === "Select Output 5" ? 118 : 104, h: type === "Select Output 5" ? 92 : 52 };
}
function resource(type, fields = []) {
  return { type, kind: "resource", look: "resource", fields, ports: { inn: 0, out: 0 }, w: 112, h: 52 };
}
function mark(type, fields = [], look = "mark", w = 132, h = 68) {
  return { type, kind: "markup", look, fields, ports: { inn: 0, out: 0 }, w, h };
}

const rate = text("rate", "到达", "每分钟 1 个");
const delay = text("delay", "时间", "1 分钟");
const cap = num("capacity", "容量", 15);

export const LIBRARIES = [
  {
    id: "pml", name: "流程建模库",
    groups: [
      { name: "模块", items: [
        flow("Source", [sel("arrival", "到达方式", ["速率", "间隔", "时刻表"]), rate, text("limit", "到达上限", "无限")]),
        flow("Sink"),
        flow("Delay", [delay, cap]),
        flow("Queue", [cap, sel("policy", "排队规则", ["FIFO", "LIFO", "优先级"])]),
        flow("Select Output", [text("condition", "条件", "true")]),
        flow("Select Output 5", [text("condition", "条件", "0")]),
        flow("Hold", [text("condition", "放行条件", "true")]),
        flow("Match"),
        flow("Split", [num("count", "份数", 2)]),
        flow("Combine"),
        flow("Assembler", [num("quantity", "装配数", 2), delay]),
        flow("Move To", [text("destination", "目的地", "")]),
        flow("Seize", [text("resource", "资源", "resourcePool")]),
        flow("Release", [text("resource", "资源", "resourcePool")]),
        flow("Service", [text("resource", "资源", "resourcePool"), delay, cap]),
        flow("Batch", [num("size", "批量", 5)]),
        flow("Enter"),
        flow("Exit")
      ]},
      { name: "资源", items: [resource("Resource Pool", [num("capacity", "容量", 1)])] }
    ]
  },
  {
    id: "mh", name: "物料搬运库",
    groups: [
      { name: "模块", items: [
        flow("Convey", [sel("route", "路由", ["自动", "指定路径"]), text("destination", "目的地", "")]),
        flow("ConveyorEnter"),
        flow("ConveyorExit"),
        flow("ProcessByRobot", [delay, text("robot", "机器人", "robot")]),
        flow("SeizeRobot", [text("robot", "机器人", "robot")]),
        flow("ReleaseRobot", [text("robot", "机器人", "robot")]),
        flow("MoveByCrane", [text("crane", "起重机", "crane"), text("destination", "目的地", "")]),
        flow("MoveByTransporter", [text("fleet", "车队", "fleet"), text("destination", "目的地", "")]),
        flow("Store", [text("storage", "货位", "storage")]),
        flow("Retrieve", [text("storage", "货位", "storage")])
      ]},
      { name: "资源", items: [resource("TransporterFleet", [num("capacity", "车辆数", 2), num("speed", "速度", 1)])] },
      { name: "空间标记", items: [
        mark("Conveyor", [num("speed", "速度", 1)], "conveyor", 220, 48),
        mark("Station", [delay], "station"),
        mark("Turntable"),
        mark("Jib Crane"),
        mark("Robot"),
        mark("Storage", [num("capacity", "货位数", 20)], "storage")
      ]}
    ]
  },
  {
    id: "ped", name: "行人库",
    groups: [
      { name: "模块", items: [
        flow("PedSource", [rate]),
        flow("PedGoTo", [text("target", "目标", "targetLine")]),
        flow("PedWait", [delay]),
        flow("PedService", [delay, cap]),
        flow("PedEnter"),
        flow("PedExit"),
        flow("PedSink")
      ]},
      { name: "空间标记", items: [
        mark("Wall", [], "wall", 180, 18),
        mark("Target Line", [], "line", 140, 18),
        mark("Pathway", [], "path", 180, 28),
        mark("Area", [cap], "area")
      ]}
    ]
  },
  {
    id: "rail", name: "轨道库",
    groups: [
      { name: "模块", items: [
        flow("TrainSource", [rate]),
        flow("TrainEnter"),
        flow("TrainMoveTo", [text("destination", "目的地", "")]),
        flow("TrainExit"),
        flow("TrainDispose")
      ]},
      { name: "空间标记", items: [
        mark("Railway Track", [], "rail", 220, 28),
        mark("Point Node")
      ]}
    ]
  },
  {
    id: "road", name: "道路交通库",
    groups: [
      { name: "模块", items: [
        flow("CarSource", [rate]),
        flow("CarEnter"),
        flow("CarMoveTo", [text("destination", "目的地", "")]),
        flow("CarDispose")
      ]},
      { name: "空间标记", items: [
        mark("Road", [num("lanes", "车道", 2)], "road", 220, 36),
        mark("Intersection"),
        mark("Stop Line", [], "line", 120, 18),
        mark("Parking Lot", [cap], "storage")
      ]}
    ]
  },
  {
    id: "fluid", name: "流体库",
    groups: [
      { name: "模块", items: [
        flow("FluidSource", [num("rate", "流量", 10)]),
        flow("Tank", [num("capacity", "容量", 100), num("initial", "初始", 20)]),
        flow("Valve", [num("rate", "最大流量", 10)]),
        flow("FluidDispose")
      ]},
      { name: "空间标记", items: [
        mark("Pipe", [num("rate", "流量上限", 20)], "pipe", 180, 22),
        mark("Storage Tank", [num("capacity", "容量", 100)], "tank")
      ]}
    ]
  },
  {
    id: "space", name: "空间标记",
    groups: [
      { name: "GIS", items: [
        mark("GIS Map", [sel("tiles", "底图", ["示意道路", "Shapefile"]), { key: "query", label: "搜索", type: "search", def: "" }], "map", 420, 240),
        mark("GIS Point", [text("label", "地点", "地点")], "pin", 88, 48),
        mark("GIS Route", [text("from", "起点", "仓库"), text("to", "终点", "客户")], "route", 200, 48),
        mark("GIS Region", [text("name", "区域", "区域")], "area")
      ]},
      { name: "网络", items: [
        mark("Point Node"),
        mark("Path", [], "path", 180, 28),
        mark("Rectangular Node", [num("capacity", "容量", 1)])
      ]}
    ]
  },
  {
    id: "present", name: "演示",
    groups: [
      { name: "平面", items: [
        mark("Rectangle", [text("color", "颜色", "#dbe4ee")], "rect"),
        mark("Oval", [text("color", "颜色", "#dbe4ee")], "oval"),
        mark("Line", [], "line", 160, 18),
        mark("Text", [text("text", "文字", "文字")], "text", 100, 32),
        mark("Image", [text("title", "标题", "图片")], "rect")
      ]},
      { name: "三维", items: [
        mark("3D Window", [text("camera", "摄像机", "camera"), check("follow", "跟随摄像机", true)], "win3d", 240, 140),
        mark("3D Object", [text("object", "对象", "box")], "box3d"),
        mark("Camera", [sel("mode", "投影", ["透视", "正交"])]),
        mark("Light", [num("intensity", "强度", 1)])
      ]}
    ]
  },
  {
    id: "analysis", name: "分析",
    groups: [
      { name: "图表", items: [
        mark("Bar Chart", [text("value", "取值", "已离开")], "bars", 180, 110),
        mark("Time Plot", [text("value", "取值", "队列")], "plot", 200, 110),
        mark("Histogram", [text("value", "取值", "停留时间")], "bars", 180, 110),
        mark("Pie Chart", [text("value", "取值", "状态")], "pie", 140, 110)
      ]},
      { name: "数据", items: [
        mark("Data Set", [text("value", "取值", "")], "data", 140, 48),
        mark("Statistics", [text("value", "取值", "停留时间")], "data", 140, 48)
      ]}
    ]
  },
  {
    id: "controls", name: "控件",
    groups: [
      { name: "控件", items: [
        mark("Button", [text("label", "标题", "按钮")], "button", 96, 32),
        mark("Slider", [num("min", "最小", 1), num("max", "最大", 30), num("value", "当前", 6)], "slider", 180, 52),
        mark("Edit Box", [text("value", "内容", "")], "edit", 140, 36),
        mark("Check Box", [text("label", "标题", "选项"), check("on", "勾选", false)], "check", 120, 28),
        mark("Combo Box", [text("items", "选项", "甲,乙,丙"), text("value", "当前", "甲")], "edit", 140, 36)
      ]}
    ]
  },
  {
    id: "sd", name: "系统动力学",
    groups: [
      { name: "存量流量", items: [
        mark("Stock", [num("initial", "初值", 100)], "stock"),
        mark("Flow", [text("rate", "速率", "10")], "sdflow", 120, 36),
        mark("Dynamic Variable", [text("formula", "公式", "stock")]),
        mark("Parameter", [num("value", "值", 1)]),
        mark("Link", [], "line", 120, 18)
      ]}
    ]
  },
  {
    id: "state", name: "状态图",
    groups: [
      { name: "状态", items: [
        mark("State", [text("action", "进入时", "")], "state"),
        mark("Transition", [text("condition", "条件", "timeout")], "line", 140, 28),
        mark("Branch", [text("condition", "条件", "true")], "branch", 64, 64),
        mark("Final State", [], "final", 48, 48)
      ]}
    ]
  }
];

export const EXPERIMENTS = [
  { type: "simulation", name: "仿真", fields: [text("stop", "停止时间", "100"), sel("seed", "随机性", ["随机", "固定种子"]), text("seedValue", "种子", "1"), num("windowW", "窗口宽", 1000), num("windowH", "窗口高", 600)] },
  { type: "variation", name: "参数扫描", fields: [text("parameter", "参数", "到达"), num("min", "最小", 1), num("max", "最大", 10), num("step", "步长", 1), num("replications", "重复次数", 3), text("stop", "停止时间", "100")] },
  { type: "optimization", name: "优化", fields: [sel("engine", "引擎", ["Genetic", "OptQuest"]), text("objective", "目标", "已离开"), num("iterations", "迭代", 40), text("parameter", "参数", "到达"), num("min", "最小", 1), num("max", "最大", 20)] },
  { type: "compare", name: "对比运行", fields: [text("left", "方案甲", "到达 4"), text("right", "方案乙", "到达 8"), text("stop", "停止时间", "100")] },
  { type: "montecarlo", name: "蒙特卡洛", fields: [num("replications", "重复次数", 30), text("stop", "停止时间", "100"), text("output", "输出", "已离开")] },
  { type: "sensitivity", name: "敏感性", fields: [text("parameter", "参数", "到达"), num("min", "下界", 1), num("max", "上界", 12), text("output", "输出", "已离开"), text("stop", "停止时间", "100")] },
  { type: "calibration", name: "校准", fields: [text("target", "目标值", "80"), num("iterations", "迭代", 30), text("parameter", "参数", "到达")] },
  { type: "rl", name: "强化学习", fields: [area("observation", "Observation", ""), area("action", "Action", "speed"), area("configuration", "Configuration", "seed"), num("episodes", "回合", 40)] },
  { type: "custom", name: "自定义", fields: [area("note", "说明", ""), text("stop", "停止时间", "100")] }
];

export function findType(type) {
  for (const lib of LIBRARIES) {
    for (const group of lib.groups) {
      const item = group.items.find((entry) => entry.type === type);
      if (item) return { ...item, lib: lib.id, libName: lib.name };
    }
  }
  return null;
}

export function experimentSpec(type) {
  return EXPERIMENTS.find((item) => item.type === type) || EXPERIMENTS[0];
}

export function defaultProps(fields) {
  const props = {};
  for (const field of fields || []) {
    if (field.type === "search") props.pins = [];
    props[field.key] = field.def;
  }
  return props;
}
