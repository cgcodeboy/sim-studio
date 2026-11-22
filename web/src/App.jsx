import { useEffect } from "react";
import { experimentSpec } from "./catalog.js";
import { projectAgent } from "./model.js";
import Canvas from "./canvas.jsx";
import { rightMode, useDock } from "./dock.js";
import { Icon } from "./icons.jsx";
import { Overlays, Problems, ProjectTree, Properties, Runs, SessionPanel, StatusBar, Toolbar } from "./panels.jsx";
import RunView from "./run.jsx";
import { useStore } from "./store.jsx";

function clamp(min, max, value) {
  return Math.min(max, Math.max(min, value));
}

function track(event, move) {
  event.preventDefault();
  const go = (ev) => move(ev);
  const up = () => {
    window.removeEventListener("pointermove", go);
    window.removeEventListener("pointerup", up);
  };
  window.addEventListener("pointermove", go);
  window.addEventListener("pointerup", up);
}

export default function App() {
  const { state, dispatch, project } = useStore();
  const [dock, setDock] = useDock();
  useEffect(() => {
    let fromMenu = false;
    const mark = () => { fromMenu = true; setTimeout(() => { fromMenu = false; }, 0); };
    const close = () => { if (fromMenu) return; dispatch({ type: "closePop" }); };
    window.addEventListener("contextmenu", mark, true);
    window.addEventListener("click", close);
    return () => {
      window.removeEventListener("contextmenu", mark, true);
      window.removeEventListener("click", close);
    };
  }, [dispatch]);
  useEffect(() => {
    const onKey = (event) => {
      const tag = document.activeElement?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (event.key === "[") setDock({ left: dock.left === "shut" ? "pages" : "shut" });
      if (event.key === "]") setDock({ right: rightMode(dock) === "props" ? "shut" : "props" });
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [dock, setDock]);
  const right = rightMode(dock);
  const columns = [
    dock.left === "shut" ? "22px" : (dock.leftW || 268) + "px",
    "minmax(0,1fr)",
    right === "shut" ? "22px" : (dock.rightW || 320) + "px"
  ].join(" ");
  const page = project?.experiments.find((item) => item.id === state.ui.pageId);
  const editing = state.ui.mode === "edit";
  const mapOn = dock.map !== false;
  return (
    <div className="app">
      <Toolbar />
      <div className="work" style={{ gridTemplateColumns: columns }}>
        <LeftDock dock={dock} setDock={setDock} />
        <main className="center">
          <div className="map-seam">
            <button type="button" className="edge-toggle flat" aria-label={mapOn ? "隐藏地图" : "显示地图"} onClick={() => setDock({ map: !mapOn })}>
              <Icon name={mapOn ? "chevronUp" : "chevronDown"} size={14} />
            </button>
          </div>
          {mapOn && (editing ? (page ? <ExperimentSheet /> : <Canvas />) : <RunView />)}
          <Problems />
          <SessionDock dock={dock} setDock={setDock} />
        </main>
        <RightDock dock={dock} setDock={setDock} />
      </div>
      <StatusBar />
      <Overlays />
    </div>
  );
}

function LeftDock({ dock, setDock }) {
  const { dispatch, project } = useStore();
  const drag = (event) => {
    const origin = event.clientX;
    const base = dock.leftW || 268;
    track(event, (ev) => setDock({ leftW: clamp(200, 520, base + ev.clientX - origin) }));
  };
  const projectMenu = (event) => {
    if (!project) return;
    event.preventDefault();
    dispatch({ type: "context", context: { x: event.clientX, y: event.clientY, kind: "project", id: project.id, projectId: project.id, pageId: "main", selected: { kind: "project", id: project.id } } });
  };
  if (dock.left === "shut") {
    return (
      <aside className="edge-rail">
        <button type="button" className="edge-toggle" aria-label="显示工程" onClick={() => setDock({ left: "pages" })}><Icon name="chevronRight" size={14} /></button>
      </aside>
    );
  }
  return (
    <aside className="dock">
      <div className="cardhead" onContextMenu={projectMenu}>
        <div><span className="badge-ico folder"><Icon name="folder" /></span><b>{project?.name || "工程"}</b></div>
        {project && <div className="head-actions"><button type="button" aria-label="新建页面" onClick={() => dispatch({ type: "dialog", dialog: { kind: "agent" } })}>＋</button></div>}
      </div>
      <ProjectTree />
      <button type="button" className="edge-grip v pull-left" aria-label="拖拽左侧宽度" onPointerDown={drag} />
      <button type="button" className="edge-toggle in-right" aria-label="隐藏工程" onClick={() => setDock({ left: "shut" })}><Icon name="chevronLeft" size={14} /></button>
    </aside>
  );
}

function SessionDock({ dock, setDock }) {
  const height = dock.talkH || 220;
  const drag = (event) => {
    event.preventDefault();
    const origin = event.clientY;
    const base = height;
    const move = (ev) => setDock({ talkH: Math.min(520, Math.max(120, base + origin - ev.clientY)) });
    const up = () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  };
  if (dock.bottom === false) {
    return <button type="button" className="session-rail" onClick={() => setDock({ bottom: true })}>会话</button>;
  }
  return (
    <div className="session-dock" style={{ height }}>
      <button type="button" className="session-grip" aria-label="拖拽会话高度" onPointerDown={drag} />
      <SessionPanel onClose={() => setDock({ bottom: false })} />
    </div>
  );
}

function RightDock({ dock, setDock }) {
  const { dispatch, project } = useStore();
  const right = rightMode(dock);
  const pick = (mode) => setDock({ right: right === mode ? "shut" : mode });
  const switches = (
    <>
      <button type="button" className={"edge-toggle" + (right === "props" ? " on" : "")} aria-label={right === "props" ? "隐藏属性" : "显示属性"} onClick={() => pick("props")}><Icon name="sliders" size={14} /></button>
      <button type="button" className={"edge-toggle" + (right === "runs" ? " on" : "")} aria-label={right === "runs" ? "隐藏仿真" : "显示仿真"} onClick={() => pick("runs")}><Icon name="play" size={14} /></button>
    </>
  );
  if (right === "shut") return <aside className="edge-rail end">{switches}</aside>;
  const dragWidth = (event) => {
    const origin = event.clientX;
    const base = dock.rightW || 320;
    track(event, (ev) => setDock({ rightW: clamp(260, 560, base + origin - ev.clientX) }));
  };
  return (
    <aside className="dock">
      <button type="button" className="edge-grip v pull-right" aria-label="拖拽右侧宽度" onPointerDown={dragWidth} />
      <div className="edge-stack on-right">{switches}</div>
      {right === "props" ? <Properties /> : <>
        <div className="cardhead">
          <b><span className="badge-ico sim"><Icon name="play" /></span>仿真</b>
          <div className="head-actions">
            <button type="button" disabled={!project} onClick={() => dispatch({ type: "dialog", dialog: { kind: "experiment" } })}>新建</button>
          </div>
        </div>
        <Runs />
      </>}
    </aside>
  );
}

function ExperimentSheet() {
  const { project, experiment } = useStore();
  if (!experiment) return null;
  const spec = experimentSpec(experiment.type);
  return (
    <div className="sheet grid">
      <div className="ruler"><span>0</span><span>100</span><span>200</span><span>300</span></div>
      <div className="frame" style={{ width: 520, height: 320, margin: 18 }}>
        <div className="study">
          <b>{experiment.name}</b>
          <table><tbody>
            <tr><td>类型</td><td>{spec.name}</td></tr>
            <tr><td>页面</td><td>{projectAgent(project, experiment.page || "main")?.name || "Main"}</td></tr>
            {spec.fields.map((field) => <tr key={field.key}><td>{field.label}</td><td>{String(experiment.props[field.key] ?? "")}</td></tr>)}
          </tbody></table>
        </div>
      </div>
    </div>
  );
}
