import { experimentSpec } from "./catalog.js";
import Canvas from "./canvas.jsx";
import { useStore } from "./store.jsx";

function clock(time, unit) {
  if (unit === "秒") return time.toFixed(1) + " 秒";
  const whole = Math.floor(time);
  const frac = Math.floor((time - whole) * 60);
  return String(whole).padStart(2, "0") + ":" + String(frac).padStart(2, "0");
}

function Controls() {
  const { state, dispatch, project } = useStore();
  const time = state.runtime?.time || 0;
  return (
    <div className="ctrl">
      <button type="button" aria-label="继续" onClick={() => dispatch({ type: "resume" })}>▶</button>
      <button type="button" aria-label="暂停" onClick={() => dispatch({ type: "pause" })}>❚❚</button>
      <button type="button" aria-label="停止" onClick={() => dispatch({ type: "stop" })}>■</button>
      <span>模型时间 <b>{clock(time, project?.unit)}</b> {project?.unit}</span>
      <label>速度
        <select aria-label="速度" value={state.ui.speed} onChange={(event) => dispatch({ type: "speed", value: Number(event.target.value) })}>
          <option value={1}>1×</option>
          <option value={10}>10×</option>
          <option value={0}>尽快</option>
        </select>
      </label>
      <span>{state.ui.play ? "运行中" : "已暂停"}</span>
    </div>
  );
}

function Scene3D() {
  const { state } = useStore();
  const movers = state.runtime?.movers || [];
  return (
    <div className="scene">
      <div className="stage3d">
        {movers.map((mover) => (
          <div key={mover.id} className="cube here" style={{ left: Math.max(8, mover.x / 4), top: Math.max(8, mover.y / 4) }}>{mover.id}</div>
        ))}
      </div>
    </div>
  );
}

function MapRun() {
  const { agent, state } = useStore();
  const map = agent?.elements.find((item) => item.type === "GIS Map");
  const route = agent?.elements.find((item) => item.type === "GIS Route");
  const pins = map?.props.pins || [];
  const phase = (state.runtime?.tokenPhase || 0) % 1;
  const x = 70 + phase * 270;
  const y = 150 - Math.sin(phase * Math.PI) * 80;
  return (
    <div className="maprun">
      <svg viewBox="0 0 420 240">
        <rect width="420" height="240" fill="#d7e8cf" />
        <path d="M20 70 H400 M20 150 H400 M120 10 V230 M260 10 V230" stroke="#f8fafc" strokeWidth="12" />
        <path d="M70 150 C 140 80, 240 190, 340 70" fill="none" stroke="#1d4ed8" strokeWidth="3" />
        <text x="48" y="146" fontSize="12">{route?.props.from || "起点"}</text>
        <text x="348" y="64" fontSize="12">{route?.props.to || "终点"}</text>
        {pins.map((pin, i) => <g key={pin}><circle cx={90 + i * 60} cy={50 + (i % 2) * 30} r="5" fill="#dc2626" /><text x={98 + i * 60} y={54 + (i % 2) * 30} fontSize="12">{pin}</text></g>)}
        <circle cx={x} cy={y} r="6" fill="#111827" />
      </svg>
    </div>
  );
}

function FlowRun() {
  const { state, dispatch, project, agent } = useStore();
  const hasMap = agent?.space === "GIS" || agent?.elements.some((item) => item.type === "GIS Map");
  const rt = state.runtime || {};
  return (
    <div className="modelwin">
      <h3>
        Simulation — {project.name}
        <button type="button" aria-pressed={state.ui.view === "2d"} onClick={() => dispatch({ type: "view", value: "2d" })}>2D</button>
        <button type="button" aria-pressed={state.ui.view === "3d"} onClick={() => dispatch({ type: "view", value: "3d" })}>3D</button>
        {hasMap && <button type="button" aria-pressed={state.ui.view === "map"} onClick={() => dispatch({ type: "view", value: "map" })}>地图</button>}
      </h3>
      <div className="stage">
        {state.ui.view === "3d" ? <Scene3D /> : state.ui.view === "map" ? <MapRun /> : <Canvas running />}
      </div>
      <div className="metrics">
        <span>已到达 {Math.floor(rt.arrived || 0)}</span>
        <span>队列 {Math.ceil(rt.queued || 0)}</span>
        <span>处理中 {Math.ceil(rt.inProcess || 0)}</span>
        <span>已离开 {Math.floor(rt.exited || 0)}</span>
        <span>队列外等待 {Math.ceil(rt.overflow || 0)}</span>
        <span>来源 {rt.source || "—"}</span>
      </div>
      {rt.overflow > 0 && <p className="warn">队列已满，多出来的在外面等。运行继续。</p>}
      <Controls />
    </div>
  );
}

function Study() {
  const { state, experiment } = useStore();
  const spec = experimentSpec(experiment.type);
  const rt = state.runtime || {};
  let body = null;
  if (experiment.type === "variation") {
    body = <table><tbody><tr><th>{experiment.props.parameter}</th><th>输出</th></tr>{(rt.rows || []).map((row) => <tr key={row.value}><td>{row.value}</td><td>{row.output}</td></tr>)}</tbody></table>;
  } else if (experiment.type === "optimization") {
    body = <div><div>引擎 {experiment.props.engine} · 目标 {experiment.props.objective}</div><div>迭代 {rt.rows?.[0]?.iteration || 0} / {experiment.props.iterations}</div><div>当前最优 {rt.best ?? "—"}</div><div className="barline"><i style={{ width: (rt.progress || 0) + "%" }} /></div></div>;
  } else if (experiment.type === "compare") {
    body = <table><tbody><tr><th>{experiment.props.left}</th><th>{experiment.props.right}</th></tr><tr><td>{rt.left || 0}</td><td>{rt.right || 0}</td></tr></tbody></table>;
  } else if (experiment.type === "montecarlo") {
    const max = Math.max(1, ...(rt.bins || [1]));
    body = <div>{(rt.bins || []).map((bin, i) => <div key={i} className="hbar" style={{ width: (bin / max) * 100 + "%" }} />)}<div>{rt.rows?.[0] ? rt.rows[0].done + " / " + rt.rows[0].total : ""}</div></div>;
  } else if (experiment.type === "sensitivity") {
    body = <div>{(rt.rows || []).map((row) => <div key={row.name}>{row.name}<div className="hbar" style={{ width: row.high + "%" }} /></div>)}</div>;
  } else if (experiment.type === "calibration") {
    body = <div>误差 {rt.error}<div className="barline"><i style={{ width: Math.max(2, 100 - rt.error * 4) + "%" }} /></div>拟合 {rt.best}</div>;
  } else if (experiment.type === "rl") {
    const vals = rt.reward || [];
    const max = Math.max(1, ...vals, 1);
    const d = vals.map((v, i) => `${(i / Math.max(1, vals.length - 1)) * 280 + 10},${120 - (v / max) * 100}`).join(" ");
    body = <div><div>回合 {vals.length} / {experiment.props.episodes}</div><svg viewBox="0 0 300 130" width="300" height="130">{d && <polyline points={d} fill="none" stroke="#6d5efc" strokeWidth="2" />}</svg><div>Observation {experiment.props.observation}</div><div>Action {experiment.props.action}</div></div>;
  } else {
    body = <div className="log">{(rt.log || []).join("\n")}</div>;
  }
  return (
    <div className="modelwin">
      <h3>{spec.name} — {experiment.name}</h3>
      <div className="study">{body}</div>
      <Controls />
    </div>
  );
}

export default function RunView() {
  const { experiment } = useStore();
  if (experiment && experiment.type !== "simulation") return <Study />;
  return <FlowRun />;
}
