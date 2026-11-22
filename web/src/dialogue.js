import { findType } from "./catalog.js";
import { addElement, addLink, clone, moveElement, projectAgent, removeElements, setElementProp } from "./model.js";

export function outline(project, pageId, text, sessionId) {
  const page = projectAgent(project, pageId);
  return {
    text,
    page: page?.name || "Main",
    project: project?.id || "",
    session: sessionId || "",
    agents: (project?.agents || []).map((agent) => ({
      name: agent.name,
      elements: agent.elements.map((item) => ({ name: item.name, type: item.type }))
    }))
  };
}

export function applyTask(project, task, pageId) {
  let applied = applyOne(project, task, pageId);
  for (const extra of task?.more || []) {
    if (applied.failed) return applied;
    const base = applied.project || project;
    const next = applyOne(base, { ...extra, status: extra.status || "已写入" }, applied.pageId || pageId);
    if (next.failed) return { ...applied, failed: next.failed };
    applied = { ...next, changed: applied.changed || next.changed };
  }
  return applied;
}

function applyOne(project, task, pageId) {
  if (!project || !task || task.status !== "已写入") return { project, changed: false };
  if (task.intent === "run" || task.intent === "pause" || task.intent === "stop") return { project, changed: false };
  const agent = findAgent(project, task.agent) || projectAgent(project, pageId) || project.agents[0];
  if (!agent) return { project, changed: false, failed: "没有找到智能体" + (task.agent ? " " + task.agent : "") + "。" };
  if (task.intent === "place") return place(project, agent, task);
  if (task.intent === "layout") return arrange(project, agent, task);
  if (task.intent === "tie") return tieLoose(project, agent);
  if (task.intent === "set") return setProp(project, agent, task);
  if (task.intent === "link") return link(project, agent, task);
  if (task.intent === "remove") return remove(project, agent, task);
  return { project, changed: false, failed: "没有这种操作。" };
}

function place(project, agent, task) {
  let next = project;
  const ids = [];
  for (const type of task.types || []) {
    const added = addElement(next, agent.id, type);
    if (!added.id) return { project, changed: false, failed: "画布上没有 " + type + "。" };
    next = applyProps(added.project, agent.id, added.id, task.props);
    ids.push(added.id);
  }
  if (task.chain && ids.length > 1) next = chain(next, agent.id, ids);
  return { project: next, changed: true, pageId: agent.id, selected: { kind: "element", id: ids[ids.length - 1] } };
}

function arrange(project, agent, task) {
  if (!agent.elements.length) return { project, changed: false, failed: "页面上还没有模块。" };
  const spots = placeSlots(agent, gapOf(task));
  let next = project;
  let right = 0;
  let bottom = 0;
  for (const spot of spots) {
    next = moveElement(next, agent.id, spot.id, spot.x, spot.y);
    right = Math.max(right, spot.x + spot.w);
    bottom = Math.max(bottom, spot.y + spot.h);
  }
  next = fitFrame(next, agent.id, right + 48, bottom + 48);
  return { project: next, changed: true, pageId: agent.id, selected: { kind: "agent", id: agent.id } };
}

function gapOf(task) {
  const gap = Number(task && task.gap);
  if (gap >= 4 && gap <= 40) return gap;
  return 8;
}

function placeSlots(agent, gap) {
  const links = linkEnds(agent);
  const spine = spineOf(agent.elements, links);
  const onSpine = new Set(spine);
  const box = modulePitch(agent, gap);
  const extra = sideGroups(agent, onSpine, links, spine[0]);
  const stride = extra.size ? box.y * 2 : box.y;
  const pos = new Map();
  spine.forEach((id, index) => pos.set(id, spinePoint(index, box, stride)));
  extra.forEach((list, host) => tuckBeside(list, pos.get(host), box, pos));
  shiftInto(pos);
  return agent.elements.map((element) => ({
    id: element.id,
    x: pos.get(element.id).x,
    y: pos.get(element.id).y,
    w: element.w || 104,
    h: element.h || 52
  }));
}

function sideGroups(agent, onSpine, links, fallback) {
  const extra = new Map();
  agent.elements.forEach((element) => {
    if (onSpine.has(element.id)) return;
    const host = hostOnSpine(element.id, onSpine, links, fallback);
    if (!extra.has(host)) extra.set(host, []);
    extra.get(host).push(element);
  });
  return extra;
}

function modulePitch(agent, gap) {
  const width = Math.max(...agent.elements.map((element) => element.w || 104));
  const height = Math.max(...agent.elements.map((element) => element.h || 52));
  return { x: width + gap, y: height + gap };
}

function spinePoint(index, box, stride) {
  const perRow = 8;
  const band = Math.floor(index / perRow);
  let xCol = index % perRow;
  if (band % 2 === 1) xCol = perRow - 1 - xCol;
  return { x: 36 + xCol * box.x, y: 48 + band * stride };
}

function tuckBeside(list, host, box, pos) {
  let above = 1;
  let below = 1;
  list.forEach((element) => {
    const waste = /sink|dispose|exit/i.test(element.type || "");
    const step = waste ? below : above;
    if (waste) below += 1;
    else above += 1;
    pos.set(element.id, { x: host.x, y: host.y + (waste ? step : -step) * box.y });
  });
}

function shiftInto(pos) {
  let minX = Infinity;
  let minY = Infinity;
  pos.forEach((point) => {
    minX = Math.min(minX, point.x);
    minY = Math.min(minY, point.y);
  });
  pos.forEach((point) => {
    point.x += 36 - minX;
    point.y += 48 - minY;
  });
}

function linkEnds(agent) {
  const out = new Map(agent.elements.map((element) => [element.id, []]));
  const inn = new Map(agent.elements.map((element) => [element.id, []]));
  for (const link of agent.links || []) {
    if (!out.has(link.from) || !inn.has(link.to)) continue;
    out.get(link.from).push(link.to);
    inn.get(link.to).push(link.from);
  }
  return { out, inn };
}

function spineOf(elements, links) {
  const col = columnOf(elements, links);
  let best = elements[0].id;
  elements.forEach((element) => {
    if (col.get(element.id) > col.get(best)) best = element.id;
  });
  const path = [];
  const seen = new Set();
  let cur = best;
  while (cur && !seen.has(cur)) {
    seen.add(cur);
    path.push(cur);
    const preds = (links.inn.get(cur) || []).slice().sort((a, b) => col.get(b) - col.get(a));
    cur = preds[0];
  }
  path.reverse();
  return path;
}

function hostOnSpine(id, onSpine, links, fallback) {
  const seen = new Set();
  let cur = id;
  while (cur && !onSpine.has(cur) && !seen.has(cur)) {
    seen.add(cur);
    const preds = links.inn.get(cur) || [];
    if (!preds.length) break;
    cur = preds[0];
  }
  return onSpine.has(cur) ? cur : fallback;
}

function columnOf(elements, links) {
  const col = new Map();
  const visiting = new Set();
  const depth = (id) => {
    if (col.has(id)) return col.get(id);
    if (visiting.has(id)) return 0;
    visiting.add(id);
    const preds = links.inn.get(id) || [];
    const value = preds.length ? Math.max(...preds.map(depth)) + 1 : 0;
    visiting.delete(id);
    col.set(id, value);
    return value;
  };
  elements.forEach((element) => depth(element.id));
  return col;
}

function fitFrame(project, agentId, w, h) {
  const next = clone(project);
  const page = projectAgent(next, agentId);
  if (!page) return project;
  page.w = Math.max(800, Math.ceil(w / 10) * 10);
  page.h = Math.max(600, Math.ceil(h / 10) * 10);
  return next;
}

function chain(project, agentId, ids) {
  let next = project;
  for (let i = 0; i < ids.length - 1; i += 1) {
    const linked = addLink(next, agentId, ids[i], 0, ids[i + 1], 0);
    if (linked.id) next = linked.project;
  }
  return tieExtraOuts(next, agentId, ids);
}

function tieLoose(project, agent) {
  const before = projectAgent(project, agent.id).links.length;
  const next = tieExtraOuts(project, agent.id, null);
  if (projectAgent(next, agent.id).links.length === before) return { project, changed: false, failed: "输出口都已经接上。" };
  return { project: next, changed: true, pageId: agent.id, selected: { kind: "agent", id: agent.id } };
}

function tieExtraOuts(project, agentId, ids) {
  let next = project;
  const wanted = ids ? new Set(ids) : null;
  const elements = projectAgent(project, agentId).elements.filter((element) => !wanted || wanted.has(element.id));
  for (const element of elements) {
    const spec = findType(element.type);
    if (!spec || spec.kind !== "flow" || spec.ports.out <= 1) continue;
    for (let port = 1; port < spec.ports.out; port += 1) {
      const agent = projectAgent(next, agentId);
      if (agent.links.some((link) => link.from === element.id && link.fromPort === port)) continue;
      const sink = addElement(next, agentId, "Sink", element.x + 150, element.y + 80);
      if (!sink.id) continue;
      const linked = addLink(sink.project, agentId, element.id, port, sink.id, 0);
      next = linked.id ? linked.project : sink.project;
    }
  }
  return next;
}

function setProp(project, agent, task) {
  const element = findElement(agent, task.element);
  if (!element) return { project, changed: false, failed: "没有找到 " + task.element + "。" };
  const value = task.key === "capacity" ? Number(task.value) : task.value;
  return { project: setElementProp(project, agent.id, element.id, task.key, value), changed: true, pageId: agent.id, selected: { kind: "element", id: element.id } };
}

function link(project, agent, task) {
  const from = findElement(agent, task.from);
  const to = findElement(agent, task.to);
  if (!from || !to) return { project, changed: false, failed: "没有找到要连的模块。" };
  const linked = addLink(project, agent.id, from.id, 0, to.id, 0);
  if (!linked.id) return { project, changed: false, failed: "这根线没有连上。" };
  return { project: linked.project, changed: true, pageId: agent.id, selected: { kind: "link", id: linked.id } };
}

function remove(project, agent, task) {
  const element = findElement(agent, task.element);
  if (!element) return { project, changed: false, failed: "没有找到 " + task.element + "。" };
  return { project: removeElements(project, agent.id, [element.id]), changed: true, pageId: agent.id, selected: { kind: "agent", id: agent.id } };
}

function applyProps(project, agentId, id, props) {
  let next = project;
  Object.entries(props || {}).forEach(([key, value]) => {
    next = setElementProp(next, agentId, id, key, key === "capacity" ? Number(value) : value);
  });
  return next;
}

function findAgent(project, name) {
  const needle = String(name || "").trim().toLowerCase();
  return project.agents.find((agent) => agent.name.toLowerCase() === needle || agent.id.toLowerCase() === needle) || null;
}

function findElement(agent, name) {
  const needle = String(name || "").trim().toLowerCase();
  const flat = needle.replace(/\s+/g, "");
  return agent.elements.find((item) => item.name.toLowerCase() === needle)
    || agent.elements.find((item) => item.type.toLowerCase() === needle)
    || agent.elements.find((item) => item.type.toLowerCase().replace(/\s+/g, "") === flat)
    || null;
}
