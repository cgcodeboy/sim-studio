async function send(path, options) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options
  });
  const text = await response.text();
  const body = text ? JSON.parse(text) : null;
  if (!response.ok) {
    const error = new Error(body?.error || "unavailable");
    error.code = body?.error || "unavailable";
    error.project = body?.project;
    throw error;
  }
  return body;
}

export function listProjects() {
  return send("/api/projects");
}

export function readProject(id) {
  return send("/api/projects/" + encodeURIComponent(id));
}

export function putProject(revision, project) {
  return send("/api/projects/" + encodeURIComponent(project.id), {
    method: "PUT",
    body: JSON.stringify({ revision, project })
  });
}

export function deleteProject(id) {
  return send("/api/projects/" + encodeURIComponent(id), { method: "DELETE" });
}

export function startRecord(body) {
  return send("/api/records", { method: "POST", body: JSON.stringify(body) });
}

export function startLive(body) {
  return send("/api/live", { method: "POST", body: JSON.stringify(body) });
}

export function getLive(id) {
  return send("/api/live/" + encodeURIComponent(id));
}

export function pauseLive(id) {
  return send("/api/live/" + encodeURIComponent(id) + "/pause", { method: "POST", body: "{}" });
}

export function resumeLive(id) {
  return send("/api/live/" + encodeURIComponent(id) + "/resume", { method: "POST", body: "{}" });
}

export function stopLive(id) {
  return send("/api/live/" + encodeURIComponent(id) + "/stop", { method: "POST", body: "{}" });
}

export function speedLive(id, speed) {
  return send("/api/live/" + encodeURIComponent(id) + "/speed", { method: "POST", body: JSON.stringify({ speed }) });
}

export function getRecord(id) {
  return send("/api/records/" + encodeURIComponent(id));
}

export function pauseRecord(id, requestId) {
  return send("/api/records/" + encodeURIComponent(id) + "/pause", { method: "POST", body: JSON.stringify({ requestId }) });
}

export function resumeRecord(id, requestId) {
  return send("/api/records/" + encodeURIComponent(id) + "/resume", { method: "POST", body: JSON.stringify({ requestId }) });
}

export function stopRecord(id, requestId) {
  return send("/api/records/" + encodeURIComponent(id) + "/stop", { method: "POST", body: JSON.stringify({ requestId }) });
}

export function turnDialogue(body) {
  return send("/api/dialogue/turns", { method: "POST", body: JSON.stringify(body) });
}

export function listSessions(projectId) {
  return send("/api/dialogue/sessions?project=" + encodeURIComponent(projectId));
}

export function createSession(projectId) {
  return send("/api/dialogue/sessions", { method: "POST", body: JSON.stringify({ project: projectId }) });
}

export function readSession(id) {
  return send("/api/dialogue/sessions/" + encodeURIComponent(id));
}

export function renameSession(id, title) {
  return send("/api/dialogue/sessions/" + encodeURIComponent(id), { method: "PUT", body: JSON.stringify({ title }) });
}

export function deleteSession(id) {
  return send("/api/dialogue/sessions/" + encodeURIComponent(id), { method: "DELETE" });
}

export function setSpeed(id, requestId, speed) {
  return send("/api/records/" + encodeURIComponent(id) + "/speed", { method: "POST", body: JSON.stringify({ requestId, speed }) });
}
