import { findType, isEntry, isExit } from "./catalog.js";

function minutesOf(text, fallback) {
  const match = String(text ?? "").match(/[\d.]+/);
  const n = match ? Number(match[0]) : NaN;
  if (!Number.isFinite(n)) return fallback;
  const raw = String(text);
  if (raw.includes("秒")) return n / 60;
  if (raw.includes("小时")) return n * 60;
  return n;
}

function chainFrom(agent, startId) {
  const chain = [];
  const seen = new Set();
  let current = startId;
  while (current && !seen.has(current)) {
    seen.add(current);
    const element = agent.elements.find((item) => item.id === current);
    if (!element) break;
    chain.push(element);
    const link = agent.links.find((item) => item.from === current && item.fromPort === 0);
    current = link?.to;
  }
  return chain;
}

export function mainChain(agent) {
  const entry = agent.elements.find((item) => findType(item.type)?.kind === "flow" && isEntry(item.type));
  if (!entry) return [];
  return chainFrom(agent, entry.id);
}

export function arrivalRate(agent) {
  const slider = agent.elements.find((item) => item.type === "Slider");
  if (slider) return Math.max(0.1, Number(slider.props.value) || 1);
  const source = agent.elements.find((item) => isEntry(item.type));
  return Math.max(0.1, minutesOf(source?.props.rate, 1) || Number(source?.props.rate) || 1);
}

export function flowSnapshot(agent, time) {
  const chain = mainChain(agent);
  const counts = {};
  const empty = { arrived: 0, exited: 0, queued: 0, overflow: 0, inProcess: 0, counts, delay: 0 };
  if (!chain.length) return empty;
  const rate = arrivalRate(agent);
  let delay = 0;
  let queueCap = Infinity;
  let processCap = Infinity;
  let queueId = null;
  for (const element of chain) {
    if (element.type === "Delay" || element.type === "Service" || element.type === "PedWait" || element.type === "PedService" || element.type === "ProcessByRobot") {
      delay += minutesOf(element.props.delay, 1);
      const cap = Number(element.props.capacity);
      if (Number.isFinite(cap)) processCap = Math.min(processCap, cap);
    }
    if (element.type === "Queue") {
      const cap = Number(element.props.capacity);
      if (Number.isFinite(cap)) { queueCap = cap; queueId = element.id; }
    }
  }
  const throughput = delay > 0 && Number.isFinite(processCap) ? processCap / delay : rate;
  const outRate = Math.min(rate, throughput);
  const arrived = rate * time;
  const exited = Math.max(0, outRate * Math.max(0, time - delay));
  const busy = Math.max(0, arrived - exited);
  const inProcess = Math.min(busy, Number.isFinite(processCap) ? processCap : busy);
  let queued = Math.max(0, busy - inProcess);
  let overflow = 0;
  if (Number.isFinite(queueCap) && queued > queueCap) {
    overflow = queued - queueCap;
    queued = queueCap;
  }
  for (const element of chain) {
    if (isExit(element.type)) counts[element.id] = Math.floor(exited);
    else if (element.id === queueId) counts[element.id] = Math.ceil(queued);
    else if (element.type === "Delay" || element.type === "Service" || element.type === "PedService" || element.type === "ProcessByRobot") counts[element.id] = Math.ceil(inProcess);
    else if (isEntry(element.type)) counts[element.id] = Math.floor(arrived);
    else counts[element.id] = Math.ceil(busy / chain.length);
  }
  return { arrived, exited, queued, overflow, inProcess, counts, delay, rate };
}

export function advanceFlow(agent, runtime, dt) {
  const time = runtime.time + dt;
  const snap = flowSnapshot(agent, time);
  const series = runtime.series.concat([{ t: time, out: snap.exited, q: snap.queued }]).slice(-60);
  return { ...runtime, time, ...snap, series, tokenPhase: time };
}

export function advanceExperiment(experiment, runtime, dt) {
  const time = runtime.time + dt;
  const progress = Math.min(100, runtime.progress + dt * 8);
  const next = { ...runtime, time, progress };
  if (experiment.type === "variation") {
    const min = Number(experiment.props.min) || 1;
    const max = Number(experiment.props.max) || 10;
    const step = Math.abs(Number(experiment.props.step) || 1) || 1;
    const values = [];
    for (let v = min, guard = 0; v <= max + 1e-6 && guard < 40; v += step, guard += 1) values.push(Number(v.toFixed(2)));
    const count = Math.min(values.length, Math.floor(progress / 8) + (progress > 0 ? 1 : 0));
    next.rows = values.slice(0, count).map((value, index) => ({ value, output: Math.round(20 + value * 7 + (index % 3)) }));
  } else if (experiment.type === "optimization") {
    const iterations = Number(experiment.props.iterations) || 40;
    const done = Math.round((progress / 100) * iterations);
    next.progress = Math.min(100, (done / iterations) * 100);
    next.best = Number((30 + done * 1.4).toFixed(1));
    next.rows = [{ iteration: done, best: next.best, parameter: experiment.props.parameter }];
  } else if (experiment.type === "compare") {
    next.left = Math.round(time * 4.2);
    next.right = Math.round(time * 6.5);
  } else if (experiment.type === "montecarlo") {
    const reps = Number(experiment.props.replications) || 30;
    const done = Math.min(reps, Math.floor(progress / 3));
    const bins = Array(8).fill(0);
    for (let i = 0; i < done; i += 1) bins[i % 8] += 1 + (i % 3 === 0 ? 1 : 0);
    next.bins = bins;
    next.rows = [{ done, total: reps }];
  } else if (experiment.type === "sensitivity") {
    const min = Number(experiment.props.min) || 1;
    const max = Number(experiment.props.max) || 12;
    next.rows = [
      { name: experiment.props.parameter || "到达", low: Math.round(20 + min * 3), high: Math.round(20 + max * 6) },
      { name: "容量", low: 28, high: 46 },
      { name: "延迟", low: 18, high: 33 }
    ];
  } else if (experiment.type === "calibration") {
    next.error = Number((20 * Math.exp(-progress / 28)).toFixed(2));
    next.best = Number((Number(experiment.props.target) || 80) * (1 - next.error / 40)).toFixed(1);
  } else if (experiment.type === "rl") {
    const episodes = Number(experiment.props.episodes) || 40;
    const done = Math.round((progress / 100) * episodes);
    next.reward = Array.from({ length: done }, (_, i) => Number((10 + i * 1.6 + (i % 4) - 2).toFixed(1)));
    next.rows = [{ episode: done, reward: next.reward[next.reward.length - 1] || 0 }];
  } else if (experiment.type === "custom") {
    const lines = runtime.log.slice();
    if (Math.floor(time) !== Math.floor(runtime.time)) lines.push("t=" + time.toFixed(1) + "  " + (experiment.props.note || "自定义步骤"));
    next.log = lines.slice(-12);
  }
  return next;
}

export function initialRuntime(kind) {
  if (kind === "exp") {
    return { kind, time: 0, progress: 0, rows: [], bins: Array(8).fill(0), best: null, error: 20, reward: [], log: [], left: 0, right: 0, counts: {}, series: [] };
  }
  return { kind: "flow", time: 0, arrived: 0, exited: 0, queued: 0, overflow: 0, inProcess: 0, counts: {}, series: [], tokenPhase: 0, rate: 1, delay: 0 };
}

export function activeStateId(agent, time) {
  const states = agent.elements.filter((item) => item.type === "State" || item.type === "Final State");
  if (!states.length) return null;
  return states[Math.floor(time) % states.length].id;
}

export function stockLevel(agent, time) {
  const stock = agent.elements.find((item) => item.type === "Stock");
  if (!stock) return null;
  const flow = agent.elements.find((item) => item.type === "Flow");
  const rate = Number(flow?.props.rate) || 10;
  const initial = Number(stock.props.initial) || 100;
  return Math.max(0, initial + Math.sin(time / 3) * rate);
}
