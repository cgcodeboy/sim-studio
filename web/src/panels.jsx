import { useState } from "react";
import { EXPERIMENTS, LIBRARIES, experimentSpec, findType } from "./catalog.js";
import { collectProblems, projectAgent } from "./model.js";
import { useStore } from "./store.jsx";

function Menu({ id, label, children }) {
  const { state, dispatch } = useStore();
  const open = state.ui.menu === id;
  return (
    <div className="menu">
      <button type="button" onClick={(event) => { event.stopPropagation(); dispatch({ type: "menu", id }); }}>{label}</button>
      {open && <div className="menu-pop" onClick={(event) => event.stopPropagation()}>{children}</div>}
    </div>
  );
}

export function Toolbar() {
  const { state, dispatch, project } = useStore();
  const panels = state.ui.panels;
  return (
    <header className="toolbar">
      <i className="mark" />
      <b>Sim Studio</b>
      <Menu id="file" label="工程">
        <button type="button" onClick={() => dispatch({ type: "dialog", dialog: { kind: "project" } })}>新建工程</button>
        <button type="button" onClick={() => dispatch({ type: "dialog", dialog: { kind: "open" } })}>打开…</button>
        <button type="button" onClick={() => dispatch({ type: "save" })}>保存</button>
        <button type="button" onClick={() => project && dispatch({ type: "dialog", dialog: { kind: "saveAs", name: project.name + " 副本" } })}>另存为…</button>
        <button type="button" onClick={() => dispatch({ type: "closeProject" })}>关闭工程</button>
      </Menu>
      <Menu id="edit" label="编辑">
        <button type="button" onClick={() => dispatch({ type: "undo" })}>撤销</button>
        <button type="button" onClick={() => dispatch({ type: "redo" })}>重做</button>
        <button type="button" onClick={() => dispatch({ type: "cut" })}>剪切</button>
        <button type="button" onClick={() => dispatch({ type: "copy" })}>复制</button>
        <button type="button" onClick={() => dispatch({ type: "paste" })}>粘贴</button>
        <button type="button" onClick={() => dispatch({ type: "remove" })}>删除</button>
      </Menu>
      <Menu id="view" label="视图">
        {[["tree", "工程"], ["props", "属性"], ["problems", "问题"], ["grid", "网格"]].map(([key, name]) => (
          <button type="button" key={key} onClick={() => dispatch({ type: "panel", key })}>{panels[key] ? "✓ " : ""}{name}</button>
        ))}
      </Menu>
      <Menu id="help" label="帮助">
        <button type="button" onClick={() => dispatch({ type: "help", value: true })}>快捷键</button>
      </Menu>
      <span className="crumb">{project ? project.name + " / " + (state.ui.pageId === "main" ? "Main" : pageTitle(project, state.ui.pageId)) : ""}</span>
      <span className="spacer" />
      <button type="button" onClick={() => dispatch({ type: "undo" })}>撤销</button>
      <button type="button" className="runbtn" onClick={() => dispatch({ type: state.ui.play ? "pause" : state.ui.mode === "pause" ? "resume" : "run" })}>
        {state.ui.play ? "暂停" : state.ui.mode === "pause" ? "继续" : "运行"}
      </button>
      {state.ui.mode !== "edit" && <button type="button" onClick={() => dispatch({ type: "stop" })}>停止</button>}
    </header>
  );
}

function pageTitle(project, id) {
  return project.agents.find((item) => item.id === id)?.name || project.experiments.find((item) => item.id === id)?.name || "";
}

function listedProjects(state) {
  const map = new Map();
  Object.values(state.saved).forEach((project) => map.set(project.id, project));
  Object.values(state.session).forEach((project) => map.set(project.id, project));
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name, "zh"));
}

export function ProjectTree() {
  const { state, dispatch } = useStore();
  const [collapsed, setCollapsed] = useState(() => new Set());
  const projects = listedProjects(state);
  const selected = state.ui.selected;
  const mark = (projectId, kind, id) => projectId === state.openId && selected?.kind === kind && selected.id === id ? "true" : undefined;
  const openBranch = (id) => setCollapsed((prev) => {
    if (!prev.has(id)) return prev;
    const next = new Set(prev);
    next.delete(id);
    return next;
  });
  const toggle = (id, opened) => setCollapsed((prev) => {
    const next = new Set(prev);
    if (opened) next.add(id);
    else next.delete(id);
    return next;
  });
  const focus = (projectId, pageId, kind, id) => dispatch({ type: "focusProject", id: projectId, pageId, selected: { kind, id } });
  const menu = (event, context) => {
    event.preventDefault();
    event.stopPropagation();
    if (context.projectId) focus(context.projectId, context.pageId, context.selected.kind, context.selected.id);
    dispatch({ type: "context", context: { x: event.clientX, y: event.clientY, ...context } });
  };
  return (
    <aside className="tree" onContextMenu={(event) => { if (event.target.closest("button")) return; menu(event, { kind: "tree" }); }}>
      {projects.map((project) => {
        const opened = !collapsed.has(project.id);
        return (
          <div key={project.id}>
            <div className="prow">
              <button type="button" className="twist" aria-label={opened ? "折叠" : "展开"} onClick={() => toggle(project.id, opened)}>{opened ? "▾" : "▸"}</button>
              <button type="button" className={"node root" + (project.id === state.openId ? " on" : "")} aria-current={project.id === state.openId ? "true" : undefined}
                onClick={() => { openBranch(project.id); focus(project.id, "main", "project", project.id); }}
                onContextMenu={(event) => menu(event, { kind: "project", id: project.id, projectId: project.id, pageId: "main", selected: { kind: "project", id: project.id } })}>
                {project.name}
              </button>
            </div>
            {opened && <>
              <div className="prow">
                <button type="button" className="node group" onContextMenu={(event) => menu(event, { kind: "agents", id: project.id, projectId: project.id, pageId: "main", selected: { kind: "project", id: project.id } })}>智能体</button>
                <button type="button" className="mini" aria-label="新建智能体" onClick={() => { focus(project.id, "main", "project", project.id); dispatch({ type: "dialog", dialog: { kind: "agent" } }); }}>＋</button>
              </div>
              {project.agents.map((agent) => (
                <div key={agent.id}>
                  <div className="prow">
                    <button type="button" className="node agent" aria-current={mark(project.id, "agent", agent.id)}
                      onClick={() => focus(project.id, agent.id, "agent", agent.id)}
                      onContextMenu={(event) => menu(event, { kind: "agent", id: agent.id, projectId: project.id, pageId: agent.id, selected: { kind: "agent", id: agent.id } })}>
                      {agent.name}
                    </button>
                    <button type="button" className="mini" aria-label={"往 " + agent.name + " 加模块"} onClick={() => { focus(project.id, agent.id, "agent", agent.id); dispatch({ type: "dialog", dialog: { kind: "module", agentId: agent.id } }); }}>＋</button>
                  </div>
                  {agent.elements.map((element) => (
                    <button type="button" key={element.id} className="node child" aria-current={mark(project.id, "element", element.id)}
                      onClick={() => focus(project.id, agent.id, "element", element.id)}
                      onContextMenu={(event) => menu(event, { kind: "element", id: element.id, projectId: project.id, pageId: agent.id, selected: { kind: "element", id: element.id } })}>
                      {element.name}
                    </button>
                  ))}
                </div>
              ))}
              <div className="prow">
                <button type="button" className="node group" onContextMenu={(event) => menu(event, { kind: "experiments", id: project.id, projectId: project.id, pageId: "main", selected: { kind: "project", id: project.id } })}>实验</button>
                <button type="button" className="mini" aria-label="新建实验" onClick={() => { focus(project.id, "main", "project", project.id); dispatch({ type: "dialog", dialog: { kind: "experiment" } }); }}>＋</button>
              </div>
              {project.experiments.map((item) => {
                const spec = experimentSpec(item.type);
                return (
                  <button type="button" key={item.id} className="node agent" aria-current={mark(project.id, "experiment", item.id)}
                    onClick={() => focus(project.id, item.id, "experiment", item.id)}
                    onContextMenu={(event) => menu(event, { kind: "experiment", id: item.id, projectId: project.id, pageId: item.id, selected: { kind: "experiment", id: item.id } })}>
                    {item.name}{spec.name !== item.name && <i className="meta">{spec.name}</i>}
                  </button>
                );
              })}
            </>}
          </div>
        );
      })}
    </aside>
  );
}

function Editor({ field, value, onChange }) {
  const { dispatch } = useStore();
  const focus = () => dispatch({ type: "checkpoint" });
  if (field.type === "select") {
    return <select aria-label={field.label} value={value ?? field.def} onChange={(event) => onChange(event.target.value, true)}>{field.options.map((option) => <option key={option}>{option}</option>)}</select>;
  }
  if (field.type === "check") return <input aria-label={field.label} type="checkbox" checked={!!value} onChange={(event) => onChange(event.target.checked, true)} />;
  if (field.type === "area") return <textarea aria-label={field.label} value={value ?? ""} onFocus={focus} onChange={(event) => onChange(event.target.value, false)} />;
  return <input aria-label={field.label} type={field.type === "number" ? "number" : "text"} value={value ?? ""} onFocus={focus} onChange={(event) => onChange(field.type === "number" ? Number(event.target.value) : event.target.value, false)} />;
}

export function Properties() {
  const { state, dispatch, project, agent } = useStore();
  const selected = state.ui.selected;
  let body = null;
  if (project && selected?.kind === "project") {
    body = section("基本", <>
      <label className="field"><span>名称</span><Editor field={{ type: "text", label: "名称" }} value={project.name} onChange={(value, history) => dispatch({ type: "rename", name: value, history, target: { kind: "project", id: project.id } })} /></label>
      <label className="field"><span>时间单位</span><Editor field={{ type: "select", label: "时间单位", options: ["秒", "分钟", "小时"] }} value={project.unit} onChange={(value) => dispatch({ type: "patchProject", key: "unit", value })} /></label>
    </>);
  } else if (project && selected?.kind === "agent") {
    const item = projectAgent(project, selected.id);
    if (item) body = <>
      {section("基本", <>
        <label className="field"><span>名称</span><Editor field={{ type: "text", label: "名称" }} value={item.name} onChange={(value, history) => dispatch({ type: "rename", name: value, history, target: { kind: "agent", id: item.id } })} /></label>
        <label className="field"><span>时间单位</span><Editor field={{ type: "select", label: "时间单位", options: ["秒", "分钟", "小时"] }} value={project.unit} onChange={(value) => dispatch({ type: "patchProject", key: "unit", value })} /></label>
      </>)}
      {section("空间", <>
        <label className="field"><span>空间</span><Editor field={{ type: "select", label: "空间", options: ["连续", "离散", "GIS"] }} value={item.space} onChange={(value) => dispatch({ type: "patchAgent", agentId: item.id, key: "space", value })} /></label>
        <label className="field"><span>宽度</span><Editor field={{ type: "number", label: "宽度" }} value={item.w} onChange={(value, history) => dispatch({ type: "patchAgent", agentId: item.id, key: "w", value, history })} /></label>
        <label className="field"><span>高度</span><Editor field={{ type: "number", label: "高度" }} value={item.h} onChange={(value, history) => dispatch({ type: "patchAgent", agentId: item.id, key: "h", value, history })} /></label>
      </>)}
    </>;
  } else if (project && selected?.kind === "experiment") {
    const item = project.experiments.find((exp) => exp.id === selected.id);
    const spec = item && experimentSpec(item.type);
    if (item) body = <>
      {section("常规", <>
        <label className="field"><span>名称</span><Editor field={{ type: "text", label: "名称" }} value={item.name} onChange={(value, history) => dispatch({ type: "rename", name: value, history, target: { kind: "experiment", id: item.id } })} /></label>
        <label className="field"><span>类型</span><input aria-label="类型" value={spec.name} readOnly /></label>
        <label className="field"><span>顶层</span><input aria-label="顶层" value="Main" readOnly /></label>
      </>)}
      {section("参数", spec.fields.map((field) => fieldRow(field, item.props[field.key], (value, history) => dispatch({ type: "patchExp", id: item.id, key: field.key, value, history }))))}
    </>;
  } else if (project && selected?.kind === "element") {
    const element = agent?.elements.find((item) => item.id === selected.id);
    const spec = element && findType(element.type);
    if (element && spec) body = <>
      {section("基本", <>
        <label className="field"><span>类型</span><input aria-label="类型" value={element.type} readOnly /></label>
        <label className="field"><span>名称</span><Editor field={{ type: "text", label: "名称" }} value={element.name} onChange={(value, history) => dispatch({ type: "rename", name: value, history, target: { kind: "element", id: element.id } })} /></label>
        <label className="field"><span>库</span><input aria-label="库" value={spec.libName} readOnly /></label>
      </>)}
      {spec.fields.length > 0 && section("参数", spec.fields.map((field) => field.type === "search"
        ? <div key={field.key}>
            <label className="field"><span>{field.label}</span><Editor field={field} value={element.props.query || ""} onChange={(value, history) => dispatch({ type: "setProp", id: element.id, key: "query", value, history })} /></label>
            <button type="button" onClick={() => dispatch({ type: "addPin", id: element.id, label: element.props.query || "" })}>标到图上</button>
            <div>{(element.props.pins || []).join("、")}</div>
          </div>
        : fieldRow(field, element.props[field.key], (value, history) => dispatch({ type: "setProp", id: element.id, key: field.key, value, history }))))}
      {section("位置", <>
        <label className="field"><span>X</span><Editor field={{ type: "number", label: "X" }} value={element.x} onChange={(value, history) => dispatch({ type: "move", id: element.id, x: value, y: element.y, history })} /></label>
        <label className="field"><span>Y</span><Editor field={{ type: "number", label: "Y" }} value={element.y} onChange={(value, history) => dispatch({ type: "move", id: element.id, x: element.x, y: value, history })} /></label>
      </>)}
    </>;
  } else if (project && selected?.kind === "link") {
    body = section("连线", <button type="button" onClick={() => dispatch({ type: "removeLink", id: selected.id })}>删除这条连线</button>);
  }
  return <aside className="props"><h2>属性</h2>{body}</aside>;
}

function section(title, children) {
  return <section className="psec"><h3>{title}</h3><div className="bd">{children}</div></section>;
}
function fieldRow(field, value, onChange) {
  if (field.type === "check") return <label key={field.key} className="check"><Editor field={field} value={value} onChange={onChange} />{field.label}</label>;
  return <label key={field.key} className="field"><span>{field.label}</span><Editor field={field} value={value} onChange={onChange} /></label>;
}

export function Problems() {
  const { state, project, agent, experiment } = useStore();
  if (!state.ui.panels.problems) return null;
  const list = state.ui.showErrors ? collectProblems(project, experiment ? projectAgent(project, "main") : agent, experiment?.type === "rl" ? experiment : null) : [];
  if (!list.length) return null;
  return (
    <div className="problems">
      <table><tbody>
        <tr><th>问题</th><th>位置</th></tr>
        {list.map((item, index) => <tr key={index}><td>{item.text}</td><td>{item.where}</td></tr>)}
      </tbody></table>
    </div>
  );
}

export function StatusBar() {
  const { state, project } = useStore();
  const time = state.runtime ? state.runtime.time.toFixed(1) : "";
  return <footer className="status">{state.ui.notice || (project ? project.name + "    时间：" + project.unit + (time ? "    " + time : "") : "")}</footer>;
}

export function Overlays() {
  const { state, dispatch } = useStore();
  return <>
    {state.ui.dialog && <Dialogs />}
    {state.ui.help && <Help />}
    {state.ui.context && <ContextMenu />}
  </>;
}

function Dialogs() {
  const { state, dispatch } = useStore();
  const dialog = state.ui.dialog;
  const close = () => dispatch({ type: "dialog", dialog: null });
  if (dialog.kind === "project") return <NameUnit title="新建工程" onClose={close} onSubmit={(name, unit) => dispatch({ type: "createProject", name, unit })} />;
  if (dialog.kind === "agent") return <NameUnit title="新建智能体" unit={false} onClose={close} onSubmit={(name) => dispatch({ type: "createAgent", name })} />;
  if (dialog.kind === "experiment") return <ExperimentDialog onClose={close} />;
  if (dialog.kind === "rename") return <NameUnit title="重命名" unit={false} initial={dialog.name} onClose={close} onSubmit={(name) => dispatch({ type: "rename", name, target: dialog.target })} />;
  if (dialog.kind === "saveAs") return <NameUnit title="另存为" unit={false} initial={dialog.name} onClose={close} onSubmit={(name) => dispatch({ type: "saveAs", name })} />;
  if (dialog.kind === "open") return <OpenDialog onClose={close} />;
  if (dialog.kind === "module") return <ModuleDialog agentId={dialog.agentId} onClose={close} />;
  return null;
}

function ModuleDialog({ agentId, onClose }) {
  const { dispatch } = useStore();
  const [libId, setLibId] = useState(LIBRARIES[0].id);
  const lib = LIBRARIES.find((item) => item.id === libId) || LIBRARIES[0];
  return (
    <div className="modal" onMouseDown={onClose}>
      <div className="dialog picker" onMouseDown={(event) => event.stopPropagation()}>
        <h3>添加模块</h3>
        <div className="picker-body">
          <div>
            {LIBRARIES.map((item) => (
              <button type="button" key={item.id} className={item.id === lib.id ? "on" : ""} onClick={() => setLibId(item.id)}>{item.name}</button>
            ))}
          </div>
          <div>
            {lib.groups.map((group) => (
              <div key={group.name}>
                <h4>{group.name}</h4>
                {group.items.map((item) => (
                  <button type="button" key={item.type} onClick={() => dispatch({ type: "addElement", elementType: item.type, agentId })}>{item.type}</button>
                ))}
              </div>
            ))}
          </div>
        </div>
        <div className="row"><button type="button" onClick={onClose}>取消</button></div>
      </div>
    </div>
  );
}

function NameUnit({ title, onClose, onSubmit, unit = true, initial = "" }) {
  const [name, setName] = useState(initial);
  const [time, setTime] = useState("分钟");
  const { dispatch } = useStore();
  return (
    <div className="modal" onMouseDown={onClose}>
      <form className="dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); if (!name.trim()) { dispatch({ type: "notice", text: "先写名称。" }); return; } onSubmit(name, time); }}>
        <h3>{title}</h3>
        <label className="field"><span>名称</span><input aria-label="名称" value={name} onChange={(event) => setName(event.target.value)} autoFocus /></label>
        {unit && <label className="field"><span>时间单位</span><select aria-label="时间单位" value={time} onChange={(event) => setTime(event.target.value)}><option>秒</option><option>分钟</option><option>小时</option></select></label>}
        <div className="row"><button type="button" onClick={onClose}>取消</button><button type="submit">确定</button></div>
      </form>
    </div>
  );
}

function ExperimentDialog({ onClose }) {
  const { dispatch } = useStore();
  const [name, setName] = useState("参数扫描");
  const [expType, setExpType] = useState("variation");
  return (
    <div className="modal" onMouseDown={onClose}>
      <form className="dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); if (!name.trim()) return; dispatch({ type: "createExperiment", name, expType }); }}>
        <h3>新建实验</h3>
        <label className="field"><span>名称</span><input aria-label="名称" value={name} onChange={(event) => setName(event.target.value)} autoFocus /></label>
        <label className="field"><span>类型</span>
          <select aria-label="类型" value={expType} onChange={(event) => { setExpType(event.target.value); setName(experimentSpec(event.target.value).name); }}>
            {EXPERIMENTS.map((item) => <option key={item.type} value={item.type}>{item.name}</option>)}
          </select>
        </label>
        <div className="row"><button type="button" onClick={onClose}>取消</button><button type="submit">创建</button></div>
      </form>
    </div>
  );
}

function OpenDialog({ onClose }) {
  const { state, dispatch } = useStore();
  const map = { ...state.saved };
  Object.entries(state.session).forEach(([id, project]) => { map[id] = project; });
  const rows = Object.values(map);
  return (
    <div className="modal" onMouseDown={onClose}>
      <div className="dialog" onMouseDown={(event) => event.stopPropagation()}>
        <h3>打开工程</h3>
        {rows.length === 0 && <p>还没有可打开的工程。</p>}
        {rows.map((project) => (
          <div key={project.id} className="open-row">
            <button type="button" onClick={() => dispatch({ type: "openProject", id: project.id })}>{project.name}</button>
            <button type="button" onClick={() => dispatch({ type: "deleteProject", id: project.id })}>删除</button>
          </div>
        ))}
        <div className="row"><button type="button" onClick={onClose}>关闭</button></div>
      </div>
    </div>
  );
}

function Help() {
  const { dispatch } = useStore();
  return (
    <div className="modal" onMouseDown={() => dispatch({ type: "help", value: false })}>
      <div className="dialog" onMouseDown={(event) => event.stopPropagation()}>
        <h3>快捷键</h3>
        <p>工程上右键可以新建智能体、实验、运行和删除。智能体行上的 ＋ 往画布加一个模块。</p>
        <p>Main 是画布。Simulation 是默认实验，右键运行。</p>
        <p>从一个端口拖到另一个端口来连线。同一对模块可以有多根线，点中后 Delete 或右键删除。</p>
        <p>Delete 删除，⌘Z 撤销，⌘⇧Z 重做，⌘C / ⌘V 复制粘贴，⌘S 保存。</p>
        <div className="row"><button type="button" onClick={() => dispatch({ type: "help", value: false })}>关闭</button></div>
      </div>
    </div>
  );
}

function projectFrom(state, menu) {
  return state.session[menu.projectId] || state.saved[menu.projectId] || state.session[state.openId] || null;
}

function itemName(project, menu) {
  if (!project) return "";
  if (menu.kind === "project") return project.name;
  if (menu.kind === "agent") return project.agents.find((item) => item.id === menu.id)?.name || "";
  if (menu.kind === "experiment") return project.experiments.find((item) => item.id === menu.id)?.name || "";
  for (const agent of project.agents) {
    const element = agent.elements.find((item) => item.id === menu.id);
    if (element) return element.name;
  }
  return "";
}

function ContextMenu() {
  const { state, dispatch } = useStore();
  const menu = state.ui.context;
  const project = projectFrom(state, menu);
  const ask = (dialog) => dispatch({ type: "dialog", dialog });
  const runAt = (pageId, kind) => {
    dispatch({ type: "focusProject", id: project.id, pageId, selected: { kind, id: pageId } });
    dispatch({ type: "run" });
  };
  const buttons = [];
  const push = (label, run) => buttons.push(<button type="button" key={label} onClick={run}>{label}</button>);
  if (menu.kind === "tree") push("新建工程", () => ask({ kind: "project" }));
  if (menu.kind === "project" && project) {
    const exp = project.experiments.find((item) => item.type === "simulation") || project.experiments[0];
    push("打开", () => dispatch({ type: "focusProject", id: project.id, pageId: "main", selected: { kind: "agent", id: "main" } }));
    push("新建智能体", () => ask({ kind: "agent" }));
    push("新建实验", () => ask({ kind: "experiment" }));
    push("添加模块", () => ask({ kind: "module", agentId: "main" }));
    if (exp) push("运行", () => runAt(exp.id, "experiment"));
    push("重命名", () => ask({ kind: "rename", name: project.name, target: { kind: "project", id: project.id } }));
    push("保存", () => dispatch({ type: "save" }));
    push("关闭", () => dispatch({ type: "closeProject", id: project.id }));
    push("删除", () => dispatch({ type: "deleteProject", id: project.id }));
  }
  if (menu.kind === "agents") push("新建智能体", () => ask({ kind: "agent" }));
  if (menu.kind === "experiments") push("新建实验", () => ask({ kind: "experiment" }));
  if (menu.kind === "agent" && project) {
    push("打开", () => dispatch({ type: "focusProject", id: project.id, pageId: menu.id, selected: { kind: "agent", id: menu.id } }));
    push("添加模块", () => ask({ kind: "module", agentId: menu.id }));
    push("运行", () => runAt(menu.id, "agent"));
    push("重命名", () => ask({ kind: "rename", name: itemName(project, menu), target: { kind: "agent", id: menu.id } }));
    push("粘贴", () => dispatch({ type: "paste" }));
    push("删除", () => dispatch({ type: "remove" }));
  }
  if (menu.kind === "experiment" && project) {
    push("打开", () => dispatch({ type: "focusProject", id: project.id, pageId: menu.id, selected: { kind: "experiment", id: menu.id } }));
    push("运行", () => runAt(menu.id, "experiment"));
    push("重命名", () => ask({ kind: "rename", name: itemName(project, menu), target: { kind: "experiment", id: menu.id } }));
    push("删除", () => dispatch({ type: "remove" }));
  }
  if (menu.kind === "element" && project) {
    push("重命名", () => ask({ kind: "rename", name: itemName(project, menu), target: { kind: "element", id: menu.id } }));
    push("复制", () => dispatch({ type: "copy" }));
    push("删除", () => dispatch({ type: "remove" }));
  }
  if (menu.kind === "canvas" && project) {
    push("添加模块", () => ask({ kind: "module", agentId: menu.id }));
    push("粘贴", () => dispatch({ type: "paste" }));
    push("运行", () => runAt(menu.id, "agent"));
  }
  if (menu.kind === "link") push("删除连线", () => dispatch({ type: "remove" }));
  return (
    <div className="context" style={{ left: menu.x, top: menu.y }} onClick={(event) => event.stopPropagation()} onContextMenu={(event) => event.preventDefault()}>
      {buttons}
    </div>
  );
}
