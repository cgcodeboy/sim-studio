import { createContext, useCallback, useContext, useEffect, useMemo, useReducer, useRef } from "react";
import { deleteProject, deleteSession, getLive, getRecord, listProjects, listSessions, createSession, pauseLive, pauseRecord, putProject, readProject, readSession, renameSession, resumeLive, resumeRecord, setSpeed, speedLive, startLive, startRecord, stopLive, stopRecord, turnDialogue } from "./api.js";
import { applyTask, outline } from "./dialogue.js";
import { advanceExperiment, advanceFlow, flowMovers, initialRuntime } from "./sim.js";
import { cardTalk, talkFromSession } from "./talks.js";
import { blankProject, clone, collectProblems, pasteElements, projectAgent, projectExperiment, experimentsOn, readSaved, SAVE_KEY, uid, addAgent, addElement, addExperiment, addLink, addPin, freezePoses, moveElement, removeAgent, removeElements, removeExperiment, removeLink, renameElement, setElementProp, setPose } from "./model.js";

const Store = createContext(null);
export function useStore() { return useContext(Store); }

function load() {
  return {
    session: {},
    saved: {},
    talks: {},
    openId: null,
    past: [],
    future: [],
    runtime: null,
    ui: {
      menu: null,
      dialog: null,
      context: null,
      palette: "pml",
      panels: { tree: true, props: true, problems: true, grid: true, talk: true },
      tabs: [],
      pageId: null,
      selected: null,
      linking: null,
      linkCursor: null,
      mode: "edit",
      play: false,
      speed: 1,
      view: "2d",
      showErrors: false,
      notice: "",
      clipboard: null,
      help: false
    }
  };
}

function current(state) {
  return state.openId ? state.session[state.openId] || null : null;
}

function snapshot(state) {
  const project = current(state);
  if (!project) return null;
  return { project: clone(project), tabs: state.ui.tabs.slice(), pageId: state.ui.pageId, selected: state.ui.selected };
}

function restore(state, shot) {
  return {
    ...state,
    openId: shot.project.id,
    session: { ...state.session, [shot.project.id]: shot.project },
    ui: { ...state.ui, tabs: shot.tabs, pageId: shot.pageId, selected: shot.selected, mode: "edit", play: false }
  };
}

function commit(state, project, uiPatch = {}) {
  const shot = snapshot(state);
  return {
    ...state,
    openId: project.id,
    session: { ...state.session, [project.id]: project },
    past: shot ? [...state.past, shot].slice(-50) : state.past,
    future: [],
    ui: { ...state.ui, ...uiPatch, notice: "" }
  };
}

function pageOf(project, pageId) {
  return projectAgent(project, pageId) || projectExperiment(project, pageId);
}

function talkTitle(items) {
  const used = new Set(items.map((item) => item.title));
  let n = items.length + 1;
  while (used.has("会话 " + n)) n += 1;
  return "会话 " + n;
}

function blankTalk(items = []) {
  return { id: uid(), title: talkTitle(items), named: false, messages: [] };
}

function withTalk(state, projectId, bag) {
  return { ...state, talks: { ...state.talks, [projectId]: bag } };
}

function reducer(state, action) {
  switch (action.type) {
    case "menu":
      return { ...state, ui: { ...state.ui, menu: state.ui.menu === action.id ? null : action.id, context: null } };
    case "closePop":
      return { ...state, ui: { ...state.ui, menu: null, context: null } };
    case "dialog":
      return { ...state, ui: { ...state.ui, dialog: action.dialog, menu: null, context: null } };
    case "context":
      return { ...state, ui: { ...state.ui, context: action.context, menu: null } };
    case "help":
      return { ...state, ui: { ...state.ui, help: action.value, menu: null } };
    case "notice":
      return { ...state, ui: { ...state.ui, notice: action.text } };
    case "palette":
      return { ...state, ui: { ...state.ui, palette: action.id } };
    case "panel":
      return { ...state, ui: { ...state.ui, panels: { ...state.ui.panels, [action.key]: !state.ui.panels[action.key] } } };
    case "speed":
      return { ...state, ui: { ...state.ui, speed: action.value } };
    case "view":
      return { ...state, ui: { ...state.ui, view: action.value } };
    case "talksLoaded": {
      if (action.projectId !== state.openId) return state;
      const opened = talkFromSession(action.opened);
      const items = (action.cards || []).map((card) => card.id === opened.id ? opened : cardTalk(card));
      if (!items.some((item) => item.id === opened.id)) items.unshift(opened);
      return withTalk(state, action.projectId, { activeId: opened.id, items });
    }
    case "talkOpened": {
      if (action.projectId !== state.openId) return state;
      const bag = state.talks[action.projectId] || { activeId: null, items: [] };
      const talk = talkFromSession(action.session);
      const items = [talk, ...bag.items.filter((item) => item.id !== talk.id)];
      return withTalk(state, action.projectId, { activeId: talk.id, items });
    }
    case "talkFilled": {
      if (action.projectId !== state.openId) return state;
      const bag = state.talks[action.projectId];
      if (!bag) return state;
      const filled = talkFromSession(action.session);
      const items = bag.items.map((item) => item.id === filled.id && !item.pending ? filled : item);
      return withTalk(state, action.projectId, { ...bag, items });
    }
    case "newTalk": {
      if (!state.openId) return state;
      const bag = state.talks[state.openId] || { activeId: null, items: [] };
      const talk = blankTalk(bag.items);
      return withTalk(state, state.openId, { activeId: talk.id, items: [...bag.items, talk] });
    }
    case "selectTalk": {
      const bag = state.talks[state.openId];
      if (!bag || !bag.items.some((item) => item.id === action.id)) return state;
      return withTalk(state, state.openId, { ...bag, activeId: action.id });
    }
    case "closeTalk": {
      const bag = state.talks[state.openId];
      if (!bag) return state;
      const items = bag.items.filter((item) => item.id !== action.id);
      const activeId = bag.activeId === action.id ? (items.at(-1)?.id || null) : bag.activeId;
      return withTalk(state, state.openId, { activeId, items });
    }
    case "talkUser": {
      const bag = state.talks[state.openId];
      const text = (action.text || "").trim();
      if (!bag || !text) return state;
      const items = bag.items.map((item) => {
        if (item.id !== action.talkId) return item;
        const title = !item.named && item.messages.length === 0 ? text.slice(0, 18) : item.title;
        return {
          ...item,
          title,
          pending: true,
          messages: [...item.messages, { id: uid(), role: "user", text }, { id: "pending", role: "assistant", text: "正在识别。", status: "进行中" }]
        };
      });
      return withTalk(state, state.openId, { ...bag, items });
    }
    case "talkResult": {
      const bag = state.talks[state.openId];
      if (!bag) return state;
      const task = action.task || { reply: "没有结果。", status: "未写入" };
      const applied = action.applied || { project: current(state), changed: false };
      const reply = applied.failed || task.reply;
      const status = applied.failed ? "未写入" : task.status;
      const items = bag.items.map((item) => {
        if (item.id !== action.talkId) return item;
        const messages = item.messages.map((message) => message.id === "pending" ? { id: uid(), role: "assistant", text: reply, status } : message);
        return { ...item, pending: false, messages };
      });
      const next = withTalk(state, state.openId, { ...bag, items });
      if (!applied.changed) return next;
      const tabs = next.ui.tabs.includes(applied.pageId) ? next.ui.tabs : [...next.ui.tabs, applied.pageId];
      return commit(next, applied.project, { pageId: applied.pageId, tabs, selected: applied.selected, mode: "edit", play: false });
    }
    case "renameTalk": {
      const bag = state.talks[state.openId];
      const name = (action.name || "").trim();
      if (!bag || !name) return { ...state, ui: { ...state.ui, dialog: null } };
      const items = bag.items.map((item) => item.id === action.id ? { ...item, title: name, named: true } : item);
      return { ...withTalk(state, state.openId, { ...bag, items }), ui: { ...state.ui, dialog: null, context: null } };
    }
    case "focusProject": {
      const source = state.session[action.id] || state.saved[action.id];
      if (!source) return state;
      const session = state.session[action.id] ? state.session : { ...state.session, [action.id]: clone(source) };
      const same = state.openId === action.id;
      const pageId = action.pageId || (same ? state.ui.pageId : "main") || "main";
      const selected = action.selected || (same ? state.ui.selected : { kind: "agent", id: "main" });
      const tabs = same ? (state.ui.tabs.includes(pageId) ? state.ui.tabs : [...state.ui.tabs, pageId]) : [pageId];
      return {
        ...state,
        openId: action.id,
        session,
        past: same ? state.past : [],
        future: same ? state.future : [],
        runtime: null,
        ui: { ...state.ui, pageId, tabs, selected, linking: null, linkCursor: null, menu: null, context: null, mode: "edit", play: false }
      };
    }
    case "createProject": {
      const project = blankProject(action.name.trim(), action.unit || "分钟");
      return {
        ...state,
        openId: project.id,
        session: { ...state.session, [project.id]: project },
        past: [],
        future: [],
        runtime: null,
        ui: { ...state.ui, dialog: null, tabs: ["main"], pageId: "main", selected: { kind: "agent", id: "main" }, mode: "edit", play: false, showErrors: false, notice: "" }
      };
    }
    case "openProject": {
      const project = state.session[action.id] || state.saved[action.id];
      if (!project) return state;
      const copy = clone(project);
      return {
        ...state,
        openId: copy.id,
        session: { ...state.session, [copy.id]: copy },
        past: [],
        future: [],
        runtime: null,
        ui: { ...state.ui, dialog: null, tabs: ["main"], pageId: "main", selected: { kind: "agent", id: "main" }, mode: "edit", play: false, showErrors: false }
      };
    }
    case "closeProject": {
      const id = action.id || state.openId;
      const project = id && (state.session[id] || state.saved[id]);
      if (!project) return state;
      if (!state.saved[id]) return { ...state, ui: { ...state.ui, context: null, notice: "「" + project.name + "」还没保存，仍留在列表里。" } };
      const session = { ...state.session };
      delete session[id];
      const closing = state.openId === id;
      const nextId = closing ? (Object.keys(session)[0] || null) : state.openId;
      return {
        ...state,
        session,
        openId: nextId,
        runtime: closing ? null : state.runtime,
        past: closing ? [] : state.past,
        future: closing ? [] : state.future,
        ui: {
          ...state.ui,
          context: null,
          dialog: null,
          mode: closing ? "edit" : state.ui.mode,
          play: closing ? false : state.ui.play,
          tabs: closing ? (nextId ? ["main"] : []) : state.ui.tabs,
          pageId: closing ? (nextId ? "main" : null) : state.ui.pageId,
          selected: closing ? (nextId ? { kind: "agent", id: "main" } : null) : state.ui.selected
        }
      };
    }
    case "save":
    case "saveAs":
    case "deleteProject":
      return state;
    case "stored": {
      const project = action.project;
      return {
        ...state,
        openId: action.openId || state.openId,
        session: { ...state.session, [project.id]: project },
        saved: { ...state.saved, [project.id]: project },
        ui: { ...state.ui, notice: action.notice, dialog: null, context: null }
      };
    }
    case "conflict": {
      const project = action.project;
      return {
        ...state,
        session: { ...state.session, [project.id]: project },
        saved: { ...state.saved, [project.id]: project },
        ui: { ...state.ui, notice: "这份工程在数据库里已经更新，已改回数据库中的内容。" }
      };
    }
    case "hydrate": {
      const saved = {};
      for (const project of action.projects) saved[project.id] = project;
      return { ...state, saved };
    }
    case "forget": {
      const session = { ...state.session };
      const saved = { ...state.saved };
      delete session[action.id];
      delete saved[action.id];
      const closing = state.openId === action.id;
      const nextId = closing ? (Object.keys(session)[0] || null) : state.openId;
      return {
        ...state,
        session,
        saved,
        openId: nextId,
        runtime: closing ? null : state.runtime,
        ui: {
          ...state.ui,
          context: null,
          dialog: null,
          mode: closing ? "edit" : state.ui.mode,
          play: closing ? false : state.ui.play,
          tabs: closing ? (nextId ? ["main"] : []) : state.ui.tabs,
          pageId: closing ? (nextId ? "main" : null) : state.ui.pageId,
          selected: closing ? (nextId ? { kind: "agent", id: "main" } : null) : state.ui.selected
        }
      };
    }
    case "select":
      return { ...state, ui: { ...state.ui, selected: action.selected, pageId: action.pageId || state.ui.pageId, tabs: action.tab && !state.ui.tabs.includes(action.tab) ? [...state.ui.tabs, action.tab] : state.ui.tabs, linking: null, linkCursor: null, mode: action.keepRun ? state.ui.mode : "edit", play: action.keepRun ? state.ui.play : false } };
    case "tab":
      return { ...state, ui: { ...state.ui, pageId: action.id, selected: { kind: pageKind(current(state), action.id), id: action.id }, mode: "edit", play: false, linking: null, linkCursor: null } };
    case "closeTab": {
      const tabs = state.ui.tabs.filter((id) => id !== action.id);
      const pageId = state.ui.pageId === action.id ? (tabs[tabs.length - 1] || "main") : state.ui.pageId;
      const nextTabs = tabs.length ? tabs : ["main"];
      return { ...state, ui: { ...state.ui, tabs: nextTabs, pageId, selected: { kind: pageKind(current(state), pageId), id: pageId }, mode: "edit", play: false } };
    }
    case "createAgent": {
      const project = current(state);
      if (!project) return state;
      const { project: next, id } = addAgent(project, action.name);
      return commit(state, next, { dialog: null, tabs: [...state.ui.tabs, id], pageId: id, selected: { kind: "agent", id }, mode: "edit", play: false });
    }
    case "createExperiment": {
      const project = current(state);
      if (!project) return state;
      const page = projectAgent(project, state.ui.pageId)?.id || "main";
      const { project: next, id } = addExperiment(project, action.name, action.expType, page);
      return commit(state, next, { dialog: null, selected: { kind: "experiment", id } });
    }
    case "rename": {
      const project = current(state);
      if (!project) return state;
      const next = clone(project);
      const name = action.name.trim();
      if (!name) return state;
      if (action.target.kind === "project") next.name = name;
      if (action.target.kind === "agent") {
        const agent = projectAgent(next, action.target.id);
        if (agent) agent.name = name;
      }
      if (action.target.kind === "experiment") {
        const exp = projectExperiment(next, action.target.id);
        if (exp) exp.name = name;
      }
      if (action.target.kind === "element") {
        const renamed = renameElement(next, state.ui.pageId, action.target.id, name);
        if (action.history === false) return { ...state, session: { ...state.session, [renamed.id]: renamed } };
        return commit(state, renamed, { dialog: null });
      }
      if (action.history === false) return { ...state, session: { ...state.session, [next.id]: next } };
      return commit(state, next, { dialog: null });
    }
    case "patchProject": {
      const project = current(state);
      if (!project) return state;
      const next = clone(project);
      next[action.key] = action.value;
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "patchAgent": {
      const project = current(state);
      const next = clone(project);
      const agent = projectAgent(next, action.agentId || state.ui.pageId);
      if (!agent) return state;
      agent[action.key] = action.value;
      if (action.key === "space" && action.value === "GIS") {
        return commit(state, next, { palette: "space" });
      }
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "patchExp": {
      const project = current(state);
      const next = clone(project);
      const exp = projectExperiment(next, action.id || state.ui.pageId);
      if (!exp) return state;
      exp.props = { ...exp.props, [action.key]: action.value };
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "checkpoint": {
      const shot = snapshot(state);
      if (!shot) return state;
      return { ...state, past: [...state.past, shot].slice(-50), future: [] };
    }
    case "addElement": {
      const project = current(state);
      const agentId = action.agentId && projectAgent(project, action.agentId) ? action.agentId : state.ui.pageId;
      const agent = projectAgent(project, agentId);
      if (!agent) return { ...state, ui: { ...state.ui, notice: project ? "先打开一个页面。" : "先新建工程。", dialog: null } };
      const { project: next, id } = addElement(project, agent.id, action.elementType, action.x, action.y);
      const tabs = state.ui.tabs.includes(agent.id) ? state.ui.tabs : [...state.ui.tabs, agent.id];
      return commit(state, next, { selected: { kind: "element", id }, pageId: agent.id, tabs, linking: null, linkCursor: null, dialog: null, context: null });
    }
    case "move": {
      const project = current(state);
      const next = moveElement(project, state.ui.pageId, action.id, action.x, action.y);
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "pose": {
      const project = current(state);
      const next = setPose(project, state.ui.pageId, action.id, action.pose);
      if (next === project) return state;
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "freezePoses": {
      const project = current(state);
      if (!project) return state;
      const next = freezePoses(project);
      if (next === project) return state;
      return { ...state, session: { ...state.session, [next.id]: next } };
    }
    case "setProp": {
      const project = current(state);
      const next = setElementProp(project, state.ui.pageId, action.id, action.key, action.value);
      return action.history === false ? { ...state, session: { ...state.session, [next.id]: next } } : commit(state, next);
    }
    case "addPin": {
      const project = current(state);
      return commit(state, addPin(project, state.ui.pageId, action.id, action.label));
    }
    case "linkStart":
      return { ...state, ui: { ...state.ui, linking: action.port, linkCursor: action.port ? { x: action.x, y: action.y } : null } };
    case "linkMove":
      if (!state.ui.linking) return state;
      return { ...state, ui: { ...state.ui, linkCursor: { x: action.x, y: action.y } } };
    case "linkTo": {
      const project = current(state);
      const a = state.ui.linking;
      const b = action.port;
      const clear = { linking: null, linkCursor: null };
      if (!project || !a || !b || a.dir === b.dir || a.elementId === b.elementId) return { ...state, ui: { ...state.ui, ...clear } };
      const from = a.dir === "out" ? a : b;
      const to = a.dir === "in" ? a : b;
      const { project: next, id } = addLink(project, state.ui.pageId, from.elementId, from.port, to.elementId, to.port);
      if (!id) return { ...state, ui: { ...state.ui, ...clear } };
      return commit(state, next, { ...clear, selected: { kind: "link", id } });
    }
    case "removeLink": {
      const project = current(state);
      return commit(state, removeLink(project, state.ui.pageId, action.id), { selected: null });
    }
    case "remove": {
      const project = current(state);
      if (!project || !state.ui.selected) return state;
      const sel = state.ui.selected;
      if (sel.kind === "link") return commit(state, removeLink(project, state.ui.pageId, sel.id), { selected: null, context: null });
      if (sel.kind === "element") return commit(state, removeElements(project, state.ui.pageId, [sel.id]), { selected: { kind: "agent", id: state.ui.pageId }, context: null });
      if (sel.kind === "agent") {
        if (project.agents.length <= 1) return { ...state, ui: { ...state.ui, notice: "工程里要留一个页面。", context: null } };
        const next = removeAgent(project, sel.id);
        if (next === project) return { ...state, ui: { ...state.ui, notice: "这一页上有工程里仅剩的仿真，先在别的页面建一个。", context: null } };
        const pageId = state.ui.pageId === sel.id ? next.agents[0].id : state.ui.pageId;
        return commit(state, next, { pageId, tabs: state.ui.tabs.filter((id) => id !== sel.id), selected: { kind: "agent", id: pageId }, context: null });
      }
      if (sel.kind === "experiment") {
        if (removeExperiment(project, sel.id) === project) return { ...state, ui: { ...state.ui, notice: "工程里要留一个仿真。", context: null } };
        const next = removeExperiment(project, sel.id);
        const pageId = state.ui.pageId === sel.id ? "main" : state.ui.pageId;
        return commit(state, next, { pageId, tabs: state.ui.tabs.filter((id) => id !== sel.id), selected: { kind: pageKind(next, pageId), id: pageId }, context: null });
      }
      return state;
    }
    case "copy": {
      const project = current(state);
      const agent = projectAgent(project, state.ui.pageId);
      const sel = state.ui.selected;
      if (!agent || sel?.kind !== "element") return state;
      const element = agent.elements.find((item) => item.id === sel.id);
      const links = agent.links.filter((link) => link.from === sel.id || link.to === sel.id);
      return { ...state, ui: { ...state.ui, clipboard: { elements: [clone(element)], links: clone(links) }, notice: "已复制 " + element.name, context: null } };
    }
    case "paste": {
      const project = current(state);
      if (!project || !state.ui.clipboard || !projectAgent(project, state.ui.pageId)) return state;
      const { project: next, ids } = pasteElements(project, state.ui.pageId, state.ui.clipboard);
      return commit(state, next, { selected: { kind: "element", id: ids[0] }, context: null });
    }
    case "cut": {
      const copied = reducer(state, { type: "copy" });
      return reducer(copied, { type: "remove" });
    }
    case "undo": {
      const shot = state.past[state.past.length - 1];
      if (!shot) return state;
      const cur = snapshot(state);
      return restore({ ...state, past: state.past.slice(0, -1), future: cur ? [...state.future, cur] : state.future, runtime: null }, shot);
    }
    case "redo": {
      const shot = state.future[state.future.length - 1];
      if (!shot) return state;
      const cur = snapshot(state);
      return restore({ ...state, future: state.future.slice(0, -1), past: cur ? [...state.past, cur] : state.past, runtime: null }, shot);
    }
    case "run":
      return state;
    case "started":
      return {
        ...state,
        runtime: { ...action.result.sample, recordId: action.result.recordId, status: action.result.status, live: action.result.live, stamp: Date.now() },
        ui: { ...state.ui, mode: "run", play: true, showErrors: false, notice: "", context: null, dialog: null }
      };
    case "sample": {
      const status = action.result.status;
      const finished = status === "finished" || status === "interrupted";
      const runtime = { ...action.result.sample, recordId: action.result.recordId, status, live: state.runtime?.live, stamp: Date.now() };
      return {
        ...state,
        runtime,
        ui: {
          ...state.ui,
          play: status === "running",
          mode: status === "paused" ? "pause" : finished ? "edit" : "run"
        }
      };
    }
    case "pause":
    case "resume":
    case "stop":
      return state;
    case "problems":
      return { ...state, ui: { ...state.ui, showErrors: true, panels: { ...state.ui.panels, problems: true }, mode: "edit", play: false, notice: "有没接上的口，运行没有开始。" } };
    case "localHold":
      return { ...state, ui: { ...state.ui, play: action.play, mode: action.mode } };
    case "localStop":
      return { ...state, runtime: null, ui: { ...state.ui, play: false, mode: "edit" } };
    case "clearProblems":
      return { ...state, ui: { ...state.ui, showErrors: false } };
    default:
      return state;
  }
}

function pageKind(project, id) {
  if (!project) return "agent";
  if (projectAgent(project, id)) return "agent";
  if (projectExperiment(project, id)) return "experiment";
  return "project";
}

function speedName(value) {
  if (value === 10) return "x10";
  if (value === 0) return "asap";
  return "x1";
}

function localSample(project, runtime, speed) {
  const experiment = projectExperiment(project, runtime.experimentId);
  const agent = projectAgent(project, experiment?.page || "main");
  const dt = speed === 0 ? 2 : speed === 10 ? 1 : 0.1;
  if (experiment && experiment.type !== "simulation") return { ...advanceExperiment(experiment, runtime, dt), experimentId: runtime.experimentId };
  if (!agent) return runtime;
  const next = advanceFlow(agent, runtime, dt);
  return { ...next, experimentId: runtime.experimentId, movers: flowMovers(agent, next) };
}

function beginLocal(rawDispatch, experiment) {
  const sample = experiment?.type && experiment.type !== "simulation" ? initialRuntime("exp") : initialRuntime("flow");
  rawDispatch({ type: "started", result: { recordId: "local", status: "running", sample: { ...sample, experimentId: experiment.id, movers: [] } } });
}

function explain(error) {
  if (error.code === "rejected") return "先停止运行再保存。";
  if (error.code === "invalid") return "文档无效。";
  if (error.code === "conflict") return "这份工程在数据库里已经更新，已改回数据库中的内容。";
  return "连不上运行服务。";
}

export function StoreProvider({ children }) {
  const [state, rawDispatch] = useReducer(reducer, null, load);
  const stateRef = useRef(state);
  const loadingTalks = useRef(false);
  stateRef.current = state;
  const dispatch = useCallback((action) => {
    if (action.type === "save" || action.type === "saveAs") {
      const snapshot = stateRef.current;
      const project = snapshot.session[snapshot.openId];
      if (!project) {
        rawDispatch({ type: "notice", text: "没有打开的工程。" });
        return;
      }
      const next = action.type === "saveAs" ? { ...clone(project), id: uid("p"), name: action.name.trim(), revision: 0 } : project;
      putProject(next.revision || 0, next).then(
        (stored) => rawDispatch({ type: "stored", project: stored, openId: action.type === "saveAs" ? stored.id : snapshot.openId, notice: action.type === "saveAs" ? "已另存为「" + stored.name + "」。" : "「" + stored.name + "」已保存。" }),
        (error) => rawDispatch(error.project ? { type: "conflict", project: error.project } : { type: "notice", text: explain(error) })
      );
      return;
    }
    if (action.type === "deleteProject") {
      const snapshot = stateRef.current;
      const id = action.id || snapshot.openId;
      const known = snapshot.saved[id];
      const done = () => rawDispatch({ type: "forget", id });
      if (!known) {
        done();
        return;
      }
      deleteProject(id).then(done, (error) => rawDispatch({ type: "notice", text: explain(error) }));
      return;
    }
    if (action.type === "sendTalk") {
      const snapshot = stateRef.current;
      const text = (action.text || "").trim();
      const bag = snapshot.talks[snapshot.openId];
      const project = snapshot.session[snapshot.openId];
      if (!bag || !project || !text || bag.items.some((item) => item.pending)) return;
      const talkId = bag.activeId;
      rawDispatch({ type: "talkUser", text, talkId });
      turnDialogue(outline(project, snapshot.ui.pageId, text, talkId)).then(
        (task) => {
          const live = stateRef.current;
          const applied = applyTask(live.session[live.openId] || project, task, live.ui.pageId);
          rawDispatch({ type: "talkResult", talkId, task, applied });
          if (!applied.failed && task.status === "已写入" && (task.intent === "run" || task.run)) setTimeout(() => dispatch({ type: "run" }), 0);
          if (!applied.failed && task.intent === "pause") setTimeout(() => dispatch({ type: "pause" }), 0);
          if (!applied.failed && task.intent === "stop") setTimeout(() => dispatch({ type: "stop" }), 0);
        },
        () => rawDispatch({ type: "talkResult", talkId, task: { intent: "clarify", reply: "对话服务没连上，工程没有改。", status: "未写入" }, applied: { changed: false } })
      );
      return;
    }
    if (action.type === "run") {
      const snapshot = stateRef.current;
      const project = snapshot.session[snapshot.openId];
      if (!project) {
        rawDispatch({ type: "notice", text: "先新建工程。" });
        return;
      }
      const page = projectAgent(project, snapshot.ui.pageId)?.id || "main";
      const onPage = experimentsOn(project, page);
      const picked = action.experimentId ? projectExperiment(project, action.experimentId) : onPage.find((item) => item.type === "simulation") || onPage[0];
      if (!picked) {
        rawDispatch({ type: "notice", text: "先在这个页面建一个仿真。" });
        return;
      }
      const agent = projectAgent(project, picked.page || page) || projectAgent(project, "main");
      const problems = collectProblems(project, agent, picked?.type === "rl" ? picked : null);
      if (problems.length) {
        rawDispatch({ type: "problems" });
        return;
      }
      const experimentId = picked.id;
      if (picked.type === "simulation") {
        startLive({ project, experimentId, speed: speedName(snapshot.ui.speed) }).then(
          (result) => rawDispatch({ type: "started", result: { ...result, live: true } }),
          () => rawDispatch({ type: "notice", text: "连不上运行服务。" })
        );
        return;
      }
      putProject(project.revision || 0, project).then(
        (stored) => {
          rawDispatch({ type: "stored", project: stored, openId: stored.id, notice: "" });
          return startRecord({ requestId: uid("q"), projectId: stored.id, experimentId, speed: speedName(snapshot.ui.speed) });
        }
      ).then(
        (result) => result && rawDispatch({ type: "started", result }),
        (error) => {
          if (error.project) rawDispatch({ type: "conflict", project: error.project });
          else beginLocal(rawDispatch, picked);
        }
      );
      return;
    }
    if (action.type === "pause" || action.type === "resume" || action.type === "stop" || action.type === "speed") {
      const snapshot = stateRef.current;
      if (snapshot.runtime?.recordId === "local") {
        if (action.type === "speed") rawDispatch(action);
        if (action.type === "pause") rawDispatch({ type: "localHold", play: false, mode: "pause" });
        if (action.type === "resume") rawDispatch({ type: "localHold", play: true, mode: "run" });
        if (action.type === "stop") rawDispatch({ type: "localStop" });
        return;
      }
      if (snapshot.runtime?.live) {
        const liveId = snapshot.runtime.recordId;
        if (action.type === "speed") rawDispatch(action);
        const call = action.type === "pause" ? pauseLive(liveId)
          : action.type === "resume" ? resumeLive(liveId)
          : action.type === "stop" ? stopLive(liveId)
          : speedLive(liveId, speedName(action.value));
        call.then(
          (result) => rawDispatch({ type: "sample", result }),
          () => rawDispatch({ type: "notice", text: "连不上运行服务。" })
        );
        return;
      }
      if (action.type === "speed") rawDispatch(action);
      const recordId = snapshot.runtime?.recordId;
      if (!recordId) return;
      const requestId = uid("q");
      const call = action.type === "pause" ? pauseRecord(recordId, requestId)
        : action.type === "resume" ? resumeRecord(recordId, requestId)
        : action.type === "stop" ? stopRecord(recordId, requestId)
        : setSpeed(recordId, requestId, speedName(action.value));
      call.then(
        (result) => rawDispatch({ type: "sample", result }),
        () => rawDispatch({ type: "notice", text: "连不上运行服务。" })
      );
      return;
    }
    if (action.type === "newTalk" || action.type === "closeTalk" || action.type === "renameTalk" || action.type === "selectTalk") {
      const snapshot = stateRef.current;
      const projectId = snapshot.openId;
      if (!projectId) return;
      if (action.type === "newTalk") {
        if (loadingTalks.current) return;
        createSession(projectId).then(
          (session) => rawDispatch({ type: "talkOpened", projectId, session }),
          () => rawDispatch({ type: "notice", text: "会话没有建起来。" })
        );
        return;
      }
      if (action.type === "renameTalk") {
        renameSession(action.id, action.name).then(
          (session) => rawDispatch({ type: "renameTalk", id: action.id, name: session.title }),
          () => rawDispatch({ type: "notice", text: "会话没有改名。" })
        );
        return;
      }
      if (action.type === "closeTalk") {
        const bag = snapshot.talks[projectId];
        deleteSession(action.id).then(
          () => {
            rawDispatch({ type: "closeTalk", id: action.id });
            const items = (bag?.items || []).filter((item) => item.id !== action.id);
            const activeId = bag?.activeId === action.id ? (items.at(-1)?.id || null) : bag?.activeId;
            const active = items.find((item) => item.id === activeId);
            if (active && !active.loaded) {
              readSession(active.id).then(
                (session) => rawDispatch({ type: "talkFilled", projectId, session }),
                () => rawDispatch({ type: "notice", text: "这条会话没有读回来。" })
              );
            }
          },
          () => rawDispatch({ type: "notice", text: "会话没有删掉。" })
        );
        return;
      }
      const bag = snapshot.talks[projectId];
      const item = bag?.items.find((entry) => entry.id === action.id);
      if (!item) return;
      rawDispatch(action);
      if (item.loaded) return;
      readSession(action.id).then(
        (session) => rawDispatch({ type: "talkFilled", projectId, session }),
        () => rawDispatch({ type: "notice", text: "这条会话没有读回来。" })
      );
      return;
    }
    rawDispatch(action);
  }, []);
  useEffect(() => {
    let cancel = false;
    (async () => {
      try {
        let items = await listProjects();
        if (items.length === 0) {
          const legacy = Object.values(readSaved());
          for (const project of legacy) items.push(await putProject(0, { ...project, revision: 0 }));
          if (legacy.length) localStorage.removeItem(SAVE_KEY);
        }
        const projects = [];
        for (const item of items) projects.push(item.agents ? item : await readProject(item.id));
        if (!cancel) rawDispatch({ type: "hydrate", projects });
      } catch {
        if (!cancel) rawDispatch({ type: "notice", text: "连不上运行服务。刷新后读不到已保存的工程。" });
      }
    })();
    return () => { cancel = true; };
  }, []);
  useEffect(() => {
    const projectId = state.openId;
    if (!projectId) return undefined;
    let cancel = false;
    loadingTalks.current = true;
    (async () => {
      try {
        let cards = await listSessions(projectId);
        if (cancel) return;
        if (!cards || cards.length === 0) {
          const created = await createSession(projectId);
          if (cancel) return;
          rawDispatch({ type: "talksLoaded", projectId, cards: [created], opened: created });
          return;
        }
        const opened = await readSession(cards[0].id);
        if (cancel) return;
        rawDispatch({ type: "talksLoaded", projectId, cards, opened });
      } catch {
        if (!cancel) rawDispatch({ type: "notice", text: "会话没有读回来。" });
      } finally {
        if (!cancel) loadingTalks.current = false;
      }
    })();
    return () => { cancel = true; };
  }, [state.openId]);
  useEffect(() => {
    if (!state.ui.play || !state.runtime?.recordId) return undefined;
    const recordId = state.runtime.recordId;
    if (recordId === "local") {
      const id = setInterval(() => {
        const snap = stateRef.current;
        const project = snap.session[snap.openId];
        if (!snap.ui.play || !project || snap.runtime?.recordId !== "local") return;
        const sample = localSample(project, snap.runtime, snap.ui.speed);
        rawDispatch({ type: "sample", result: { recordId: "local", status: "running", sample } });
      }, 100);
      return () => clearInterval(id);
    }
    const id = setInterval(() => {
      const read = state.runtime?.live ? getLive(recordId) : getRecord(recordId);
      read.then(
        (result) => rawDispatch({ type: "sample", result }),
        () => rawDispatch({ type: "notice", text: "连不上运行服务。" })
      );
    }, 200);
    return () => clearInterval(id);
  }, [state.ui.play, state.runtime?.recordId]);
  useEffect(() => {
    const onKey = (event) => {
      if (document.querySelector(".modal")) return;
      const tag = document.activeElement?.tagName;
      const typing = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
      const meta = event.metaKey || event.ctrlKey;
      if (meta && event.key.toLowerCase() === "s") { event.preventDefault(); dispatch({ type: "save" }); return; }
      if (meta && event.key.toLowerCase() === "z") { event.preventDefault(); dispatch({ type: event.shiftKey ? "redo" : "undo" }); return; }
      if (typing) return;
      if (meta && event.key.toLowerCase() === "c") { dispatch({ type: "copy" }); return; }
      if (meta && event.key.toLowerCase() === "x") { dispatch({ type: "cut" }); return; }
      if (meta && event.key.toLowerCase() === "v") { dispatch({ type: "paste" }); return; }
      if (event.key === "Escape") { dispatch({ type: "closePop" }); dispatch({ type: "dialog", dialog: null }); return; }
      if (event.key === "Delete" || event.key === "Backspace") dispatch({ type: "remove" });
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [dispatch]);
  const api = useMemo(() => {
    const project = current(state);
    const experiment = projectExperiment(project, state.ui.pageId);
    let agent = projectAgent(project, state.ui.pageId);
    if (!agent && project && state.ui.mode !== "edit") agent = projectAgent(project, "main");
    return { state, dispatch, project, agent, experiment };
  }, [state, dispatch]);
  return <Store.Provider value={api}>{children}</Store.Provider>;
}
