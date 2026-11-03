import { defaultProps, experimentSpec, findType, isEntry, isExit } from "./catalog.js";

export function uid(prefix) {
  return prefix + Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
}

export function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

export function blankProject(name, unit = "分钟") {
  return {
    id: uid("p"),
    name,
    unit,
    agents: [{ id: "main", name: "Main", space: "连续", w: 800, h: 600, elements: [], links: [] }],
    experiments: [{ id: "sim", name: "Simulation", type: "simulation", props: defaultProps(experimentSpec("simulation").fields) }]
  };
}

export function projectAgent(project, id) {
  return project?.agents.find((agent) => agent.id === id) || null;
}

export function projectExperiment(project, id) {
  return project?.experiments.find((item) => item.id === id) || null;
}

function uniqueName(owner, base) {
  const used = new Set((owner.elements || owner.agents || owner.experiments || []).map((item) => item.name));
  const root = String(base || "item").replace(/\s+/g, "") || "item";
  if (!used.has(root)) return root;
  let n = 2;
  while (used.has(root + n)) n += 1;
  return root + n;
}

export function addAgent(project, name) {
  const next = clone(project);
  const id = uid("a");
  next.agents.push({ id, name: name.trim(), space: "连续", w: 800, h: 600, elements: [], links: [] });
  return { project: next, id };
}

export function addExperiment(project, name, type) {
  const next = clone(project);
  const spec = experimentSpec(type);
  const id = uid("e");
  next.experiments.push({ id, name: name.trim() || spec.name, type: spec.type, props: defaultProps(spec.fields) });
  return { project: next, id };
}

export function addElement(project, agentId, type, x, y) {
  const spec = findType(type);
  if (!spec) return { project, id: null };
  const next = clone(project);
  const agent = projectAgent(next, agentId);
  const id = uid("n");
  const flows = agent.elements.filter((item) => findType(item.type)?.kind === "flow").length;
  const element = {
    id,
    name: uniqueName(agent, type.toLowerCase().replace(/[^a-z0-9]+/gi, "") || "item"),
    type,
    lib: spec.lib,
    x: x == null ? 36 + (flows % 5) * 128 : x,
    y: y == null ? 64 + Math.floor(flows / 5) * 88 : y,
    w: spec.w,
    h: spec.h,
    props: defaultProps(spec.fields)
  };
  agent.elements.push(element);
  return { project: next, id };
}

export function moveElement(project, agentId, elementId, x, y) {
  const next = clone(project);
  const element = projectAgent(next, agentId)?.elements.find((item) => item.id === elementId);
  if (element) { element.x = Math.round(x / 10) * 10; element.y = Math.round(y / 10) * 10; }
  return next;
}

export function renameElement(project, agentId, elementId, name) {
  const next = clone(project);
  const element = projectAgent(next, agentId)?.elements.find((item) => item.id === elementId);
  if (element && name.trim()) element.name = name.trim();
  return next;
}

export function setElementProp(project, agentId, elementId, key, value) {
  const next = clone(project);
  const element = projectAgent(next, agentId)?.elements.find((item) => item.id === elementId);
  if (element) element.props = { ...element.props, [key]: value };
  return next;
}

export function addPin(project, agentId, elementId, label) {
  const next = clone(project);
  const element = projectAgent(next, agentId)?.elements.find((item) => item.id === elementId);
  if (element && label.trim()) {
    const pins = Array.isArray(element.props.pins) ? element.props.pins.slice() : [];
    pins.push(label.trim());
    element.props = { ...element.props, pins, query: label.trim() };
  }
  return next;
}

export function addLink(project, agentId, from, fromPort, to, toPort) {
  if (!project || from === to) return { project, id: null };
  const next = clone(project);
  const agent = projectAgent(next, agentId);
  if (!agent) return { project, id: null };
  const id = uid("l");
  agent.links.push({ id, from, fromPort, to, toPort });
  return { project: next, id };
}

export function removeLink(project, agentId, linkId) {
  const next = clone(project);
  const agent = projectAgent(next, agentId);
  agent.links = agent.links.filter((link) => link.id !== linkId);
  return next;
}

export function removeElements(project, agentId, ids) {
  const next = clone(project);
  const agent = projectAgent(next, agentId);
  if (!agent) return project;
  const drop = new Set(ids);
  agent.elements = agent.elements.filter((item) => !drop.has(item.id));
  agent.links = agent.links.filter((link) => !drop.has(link.from) && !drop.has(link.to));
  return next;
}

export function removeAgent(project, agentId) {
  if (project.agents.length <= 1) return project;
  const next = clone(project);
  next.agents = next.agents.filter((agent) => agent.id !== agentId);
  return next;
}

export function removeExperiment(project, experimentId) {
  if (project.experiments.length <= 1) return project;
  const next = clone(project);
  next.experiments = next.experiments.filter((item) => item.id !== experimentId);
  return next;
}

export function pasteElements(project, agentId, payload, dx = 24, dy = 24) {
  const next = clone(project);
  const agent = projectAgent(next, agentId);
  const map = {};
  const created = [];
  for (const element of payload.elements || []) {
    const id = uid("n");
    map[element.id] = id;
    const copy = clone(element);
    copy.id = id;
    copy.name = uniqueName(agent, element.name);
    copy.x += dx;
    copy.y += dy;
    agent.elements.push(copy);
    created.push(id);
  }
  for (const link of payload.links || []) {
    if (!map[link.from] || !map[link.to]) continue;
    agent.links.push({ id: uid("l"), from: map[link.from], fromPort: link.fromPort, to: map[link.to], toPort: link.toPort });
  }
  return { project: next, ids: created };
}

export function collectProblems(project, agent, experiment) {
  const list = [];
  if (agent) {
    const flows = agent.elements.filter((item) => findType(item.type)?.kind === "flow");
    if (flows.length) {
      for (const element of flows) {
        const spec = findType(element.type);
        for (let port = 0; port < spec.ports.out; port += 1) {
          if (!agent.links.some((link) => link.from === element.id && link.fromPort === port)) {
            const which = spec.ports.out > 1 ? "输出口 " + (port + 1) : "输出口";
            list.push({ text: element.name + " 的" + which + "没有接上", where: agent.name + " / " + element.name, elementId: element.id });
          }
        }
      }
      if (!flows.some((item) => isEntry(item.type)) || !flows.some((item) => isExit(item.type))) {
        list.push({ text: "流程缺少入口或出口", where: agent.name });
      }
    }
  }
  if (experiment?.type === "rl" && !String(experiment.props.observation || "").trim()) {
    list.push({ text: "Observation 还是空的", where: experiment.name });
  }
  return list;
}

export const SAVE_KEY = "sim-studio-projects-v1";

export function readSaved() {
  try { return JSON.parse(localStorage.getItem(SAVE_KEY) || "{}"); }
  catch { return {}; }
}

export function writeSaved(saved) {
  try { localStorage.setItem(SAVE_KEY, JSON.stringify(saved)); }
  catch { /* 浏览器不允许写入时，这次会话里仍然留着。 */ }
}
