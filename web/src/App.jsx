import { useEffect } from "react";
import { experimentSpec } from "./catalog.js";
import Canvas from "./canvas.jsx";
import { Overlays, Problems, ProjectTree, Properties, StatusBar, Toolbar } from "./panels.jsx";
import RunView from "./run.jsx";
import { useStore } from "./store.jsx";

export default function App() {
  const { state, dispatch, project } = useStore();
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
  const columns = [
    state.ui.panels.tree ? "240px" : null,
    "minmax(0,1fr)",
    state.ui.panels.props ? "280px" : null
  ].filter(Boolean).join(" ");
  const page = project?.experiments.find((item) => item.id === state.ui.pageId);
  const editing = state.ui.mode === "edit";
  return (
    <div className="app">
      <Toolbar />
      <div className="work" style={{ gridTemplateColumns: columns }}>
        {state.ui.panels.tree && <ProjectTree />}
        <main className="center">
          {project && <Tabs />}
          {editing ? (page ? <ExperimentSheet /> : <Canvas />) : <RunView />}
          <Problems />
        </main>
        {state.ui.panels.props && <Properties />}
      </div>
      <StatusBar />
      <Overlays />
    </div>
  );
}

function Tabs() {
  const { state, dispatch, project } = useStore();
  return (
    <div className="tabs">
      {state.ui.tabs.map((id) => {
        const agent = project.agents.find((item) => item.id === id);
        const exp = project.experiments.find((item) => item.id === id);
        const label = agent?.name || exp?.name || id;
        return (
          <button type="button" key={id} aria-selected={state.ui.pageId === id} onClick={() => dispatch({ type: "tab", id })}>
            {label}
            <i onClick={(event) => { event.stopPropagation(); dispatch({ type: "closeTab", id }); }}>×</i>
          </button>
        );
      })}
    </div>
  );
}

function ExperimentSheet() {
  const { experiment } = useStore();
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
            <tr><td>顶层智能体</td><td>Main</td></tr>
            {spec.fields.map((field) => <tr key={field.key}><td>{field.label}</td><td>{String(experiment.props[field.key] ?? "")}</td></tr>)}
          </tbody></table>
        </div>
      </div>
    </div>
  );
}
