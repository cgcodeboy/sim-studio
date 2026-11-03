import { useState } from "react";
import { findType } from "./catalog.js";
import { collectProblems } from "./model.js";
import { activeStateId, stockLevel } from "./sim.js";
import { useStore } from "./store.jsx";

function wirePath(a, b, slot) {
  const lift = slot * 22;
  const dx = Math.max(36, Math.abs(b.x - a.x) * 0.45);
  const dir = b.x >= a.x ? 1 : -1;
  return `M ${a.x} ${a.y} C ${a.x + dx * dir} ${a.y + lift}, ${b.x - dx * dir} ${b.y + lift}, ${b.x} ${b.y}`;
}

function portPoint(element, dir, index) {
  const spec = findType(element.type);
  const count = Math.max(1, dir === "in" ? spec.ports.inn : spec.ports.out);
  return {
    x: dir === "in" ? element.x : element.x + element.w,
    y: element.y + ((index + 1) * element.h) / (count + 1)
  };
}

function Chart({ element, runtime }) {
  const series = runtime?.series?.length ? runtime.series : [{ out: 4, q: 2 }, { out: 8, q: 5 }, { out: 6, q: 3 }, { out: 11, q: 4 }, { out: 9, q: 6 }];
  const key = element.type === "Time Plot" || element.props.value === "队列" ? "q" : "out";
  const vals = series.map((row) => Number(row[key]) || 0);
  const max = Math.max(1, ...vals);
  if (element.type === "Time Plot") {
    const d = vals.map((v, i) => `${(i / Math.max(1, vals.length - 1)) * 180 + 8},${96 - (v / max) * 76}`).join(" ");
    return <svg viewBox="0 0 196 104" width="100%" height="100%"><polyline points={d} fill="none" stroke="#6d5efc" strokeWidth="2" /></svg>;
  }
  if (element.type === "Pie Chart") {
    return <svg viewBox="0 0 100 100" width="70" height="70"><path d="M50 50 L50 8 A42 42 0 0 1 88 68 Z" fill="#6d5efc" /><path d="M50 50 L88 68 A42 42 0 1 1 50 8 Z" fill="#c4b5fd" /></svg>;
  }
  return <svg viewBox="0 0 180 100" width="100%" height="100%">{vals.slice(-8).map((v, i) => <rect key={i} x={8 + i * 22} y={92 - (v / max) * 74} width="14" height={(v / max) * 74} fill="#6d5efc" />)}</svg>;
}

function MapFace({ element }) {
  const pins = element.props.pins || [];
  return (
    <svg viewBox="0 0 420 240" width="100%" height="100%">
      <rect width="420" height="240" fill="#d7e8cf" />
      <path d="M20 70 H400 M20 150 H400 M120 10 V230 M260 10 V230" stroke="#f8fafc" strokeWidth="10" />
      <path d="M70 150 C 140 80, 240 190, 340 70" fill="none" stroke="#1d4ed8" strokeWidth="3" strokeDasharray="6 4" />
      {pins.map((pin, i) => <g key={pin + i}><circle cx={80 + (i * 70) % 280} cy={60 + (i % 3) * 50} r="5" fill="#dc2626" /><text x={90 + (i * 70) % 280} y={64 + (i % 3) * 50} fontSize="12">{pin}</text></g>)}
    </svg>
  );
}

function Face({ element, runtime, running, dispatch }) {
  const look = findType(element.type)?.look || "mark";
  if (look === "map") return <div className="face map"><MapFace element={element} /></div>;
  if (look === "bars" || look === "plot" || look === "pie") return <div className={"face " + look}><Chart element={element} runtime={runtime} /></div>;
  if (element.type === "Slider") {
    return (
      <label className="face slider" onPointerDown={(event) => event.stopPropagation()}>
        到达 {element.props.value}
        <input type="range" min={element.props.min} max={element.props.max} value={element.props.value} aria-label="到达"
          onPointerDown={() => dispatch({ type: "checkpoint" })}
          onChange={(event) => dispatch({ type: "setProp", id: element.id, key: "value", value: Number(event.target.value), history: false })} />
      </label>
    );
  }
  if (element.type === "Button") {
    return <button type="button" className="face button" onPointerDown={(event) => event.stopPropagation()} onClick={() => running && dispatch({ type: "notice", text: element.props.label + " 已按下" })}>{element.props.label || "按钮"}</button>;
  }
  if (element.type === "Check Box") {
    return <label className="face check" onPointerDown={(event) => event.stopPropagation()}><input type="checkbox" checked={!!element.props.on} onChange={(event) => dispatch({ type: "setProp", id: element.id, key: "on", value: event.target.checked })} />{element.props.label}</label>;
  }
  if (element.type === "Edit Box" || element.type === "Combo Box") return <div className="face edit">{element.props.value || element.type}</div>;
  if (element.type === "Text") return <div className="face text">{element.props.text}</div>;
  if (element.type === "Stock") {
    const level = runtime ? stockLevel({ elements: [element, { type: "Flow", props: { rate: 10 } }] }, runtime.time) : element.props.initial;
    return <div className="face stock">{element.name}<br />{Number(level).toFixed(0)}</div>;
  }
  if (look === "branch") return <div className="face branch" />;
  return <div className={"face " + look}>{element.type}<br />{element.name}</div>;
}

export default function Canvas({ running = false }) {
  const { state, dispatch, project, agent } = useStore();
  const [drop, setDrop] = useState(false);
  if (!project) return <div className={"sheet" + (state.ui.panels.grid ? " grid" : "")} />;
  if (!agent) return <div className="sheet" />;
  const runtime = running ? state.runtime : null;
  const problems = state.ui.showErrors ? collectProblems(project, agent, null) : [];
  const bad = new Set(problems.map((item) => item.elementId).filter(Boolean));
  const hot = runtime ? activeStateId(agent, runtime.time) : null;
  const space = agent.space === "GIS" ? " gis" : agent.space === "离散" ? " discrete" : "";

  const onPort = (event, element, dir, index) => {
    event.stopPropagation();
    event.preventDefault();
    if (running) return;
    const frame = event.currentTarget.closest(".frame");
    const point = (ev) => {
      const rect = frame.getBoundingClientRect();
      return { x: ev.clientX - rect.left + frame.scrollLeft, y: ev.clientY - rect.top + frame.scrollTop };
    };
    const start = point(event);
    dispatch({ type: "linkStart", port: { elementId: element.id, port: index, dir }, x: start.x, y: start.y });
    const move = (ev) => {
      const next = point(ev);
      dispatch({ type: "linkMove", x: next.x, y: next.y });
    };
    const up = (ev) => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
      const hit = document.elementFromPoint(ev.clientX, ev.clientY)?.closest?.("[data-port]");
      if (!hit) {
        dispatch({ type: "linkStart", port: null });
        return;
      }
      dispatch({ type: "linkTo", port: { elementId: hit.dataset.element, port: Number(hit.dataset.port), dir: hit.dataset.dir } });
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  };

  const dragElement = (event, element) => {
    if (event.target.closest(".port")) return;
    dispatch({ type: "select", selected: { kind: "element", id: element.id }, keepRun: running });
    if (running || event.target.closest("input, button, textarea, label")) return;
    const origin = { x: element.x, y: element.y, cx: event.clientX, cy: event.clientY };
    let moved = false;
    const move = (ev) => {
      const dx = ev.clientX - origin.cx;
      const dy = ev.clientY - origin.cy;
      if (!moved && Math.hypot(dx, dy) < 3) return;
      if (!moved) dispatch({ type: "checkpoint" });
      moved = true;
      dispatch({ type: "move", id: element.id, x: origin.x + dx, y: origin.y + dy, history: false });
    };
    const up = () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  };

  let draft = null;
  if (!running && state.ui.linking && state.ui.linkCursor) {
    const source = agent.elements.find((item) => item.id === state.ui.linking.elementId);
    if (source) {
      const from = portPoint(source, state.ui.linking.dir, state.ui.linking.port);
      draft = wirePath(from, state.ui.linkCursor, 0);
    }
  }

  const tokens = [];
  if (runtime?.kind === "flow") {
    agent.links.forEach((link, index) => {
      const from = agent.elements.find((item) => item.id === link.from);
      const to = agent.elements.find((item) => item.id === link.to);
      if (!from || !to) return;
      const a = portPoint(from, "out", link.fromPort);
      const b = portPoint(to, "in", link.toPort);
      const p = (runtime.tokenPhase * 0.35 + index * 0.17) % 1;
      tokens.push({ id: link.id, x: a.x + (b.x - a.x) * p, y: a.y + (b.y - a.y) * p });
    });
  }

  return (
    <div className={"sheet" + (state.ui.panels.grid ? " grid" : "")}>
      <div className="ruler">{[0, 100, 200, 300, 400, 500, 600, 700].map((n) => <span key={n}>{n}</span>)}</div>
      <div
        className={"frame" + space + (drop ? " drop" : "")}
        style={{ width: agent.w, height: agent.h }}
        onClick={(event) => { if (event.target.closest(".port, .el")) return; dispatch({ type: "linkStart", port: null }); }}
        onContextMenu={(event) => {
          if (event.target.closest(".el")) return;
          event.preventDefault();
          dispatch({ type: "context", context: { x: event.clientX, y: event.clientY, kind: "canvas", id: agent.id, projectId: project.id } });
        }}
        onDragOver={(event) => { event.preventDefault(); setDrop(true); }}
        onDragLeave={() => setDrop(false)}
        onDrop={(event) => {
          event.preventDefault();
          setDrop(false);
          const type = event.dataTransfer.getData("application/x-sim-type");
          if (!type) return;
          const rect = event.currentTarget.getBoundingClientRect();
          dispatch({ type: "addElement", elementType: type, x: event.clientX - rect.left, y: event.clientY - rect.top });
        }}
      >
        <svg className="links" width={agent.w} height={agent.h}>
          {agent.links.map((link) => {
            const from = agent.elements.find((item) => item.id === link.from);
            const to = agent.elements.find((item) => item.id === link.to);
            if (!from || !to) return null;
            const a = portPoint(from, "out", link.fromPort);
            const b = portPoint(to, "in", link.toPort);
            const siblings = agent.links.filter((item) => item.from === link.from && item.to === link.to);
            const slot = siblings.findIndex((item) => item.id === link.id) - (siblings.length - 1) / 2;
            const d = wirePath(a, b, slot);
            const on = state.ui.selected?.kind === "link" && state.ui.selected.id === link.id;
            return (
              <g key={link.id}>
                <path d={d} className={on ? "on" : ""} />
                <path d={d} className="hit" onClick={(event) => { event.stopPropagation(); dispatch({ type: "select", selected: { kind: "link", id: link.id }, keepRun: running }); }}
                  onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); dispatch({ type: "select", selected: { kind: "link", id: link.id }, keepRun: running }); dispatch({ type: "context", context: { x: event.clientX, y: event.clientY, kind: "link", id: link.id, projectId: project.id } }); }} />
              </g>
            );
          })}
          {draft && <path d={draft} className="draft" />}
        </svg>
        {agent.elements.map((element) => {
          const spec = findType(element.type);
          const outside = element.x < 0 || element.y < 0 || element.x > agent.w || element.y > agent.h;
          const selected = state.ui.selected?.kind === "element" && state.ui.selected.id === element.id;
          return (
            <div key={element.id} className={"el lib-" + element.lib + (selected ? " on" : "") + (outside ? " outside" : "") + (hot === element.id ? " hot" : "")}
              style={{ left: element.x, top: element.y, width: element.w, height: element.h }}
              onPointerDown={(event) => dragElement(event, element)}
              onContextMenu={(event) => { event.preventDefault(); event.stopPropagation(); dispatch({ type: "select", selected: { kind: "element", id: element.id }, keepRun: running }); dispatch({ type: "context", context: { x: event.clientX, y: event.clientY, kind: "element", id: element.id, projectId: project.id } }); }}>
              <Face element={element} runtime={runtime} running={running} dispatch={dispatch} />
              {spec?.ports.inn > 0 && Array.from({ length: spec.ports.inn }, (_, index) => (
                <button type="button" key={"in" + index} className={"port in" + (state.ui.linking?.elementId === element.id && state.ui.linking.dir === "in" && state.ui.linking.port === index ? " arm" : "")}
                  style={{ top: ((index + 1) * element.h) / (spec.ports.inn + 1) - 7 }}
                  data-port={index} data-dir="in" data-element={element.id}
                  aria-label="输入口" onPointerDown={(event) => onPort(event, element, "in", index)} />
              ))}
              {spec?.ports.out > 0 && Array.from({ length: spec.ports.out }, (_, index) => {
                const open = bad.has(element.id);
                return (
                  <button type="button" key={"out" + index} className={"port out" + (open ? " bad" : "") + (state.ui.linking?.elementId === element.id && state.ui.linking.dir === "out" && state.ui.linking.port === index ? " arm" : "")}
                    style={{ top: ((index + 1) * element.h) / (spec.ports.out + 1) - 7 }}
                    data-port={index} data-dir="out" data-element={element.id}
                    aria-label="输出口" onPointerDown={(event) => onPort(event, element, "out", index)} />
                );
              })}
              {runtime?.counts?.[element.id] > 0 && <span className="badge">{Math.round(runtime.counts[element.id])}</span>}
            </div>
          );
        })}
        {tokens.map((token) => <i key={token.id} className="token" style={{ left: token.x, top: token.y }} />)}
      </div>
    </div>
  );
}
