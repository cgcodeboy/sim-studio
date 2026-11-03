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

export function setSpeed(id, requestId, speed) {
  return send("/api/records/" + encodeURIComponent(id) + "/speed", { method: "POST", body: JSON.stringify({ requestId, speed }) });
}
