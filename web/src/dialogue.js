import { findType } from "./catalog.js";
import { addElement, addLink, projectAgent, removeElements, setElementProp } from "./model.js";

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
  if (!project || !task || task.status !== "已写入") return { project, changed: false };
  if (task.intent === "run" || task.intent === "pause" || task.intent === "stop") return { project, changed: false };
  const agent = findAgent(project, task.agent) || projectAgent(project, pageId) || project.agents[0];
  if (!agent) return { project, changed: false, failed: "没有找到智能体" + (task.agent ? " " + task.agent : "") + "。" };
  if (task.intent === "place") return place(project, agent, task);
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
