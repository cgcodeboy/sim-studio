import { useEffect, useState } from "react";

let dock = { left: "pages", right: "props", map: true, bottom: true, talkH: 220, leftW: 268, rightW: 320 };
const listeners = new Set();

function publish(next) {
  dock = next;
  listeners.forEach((fn) => fn());
}

export function useDock() {
  const [, bump] = useState(0);
  useEffect(() => {
    const fn = () => bump((n) => n + 1);
    listeners.add(fn);
    return () => listeners.delete(fn);
  }, []);
  return [dock, (patch) => publish({ ...dock, ...patch })];
}

export function rightMode(dock) {
  return dock.right === "runs" ? "runs" : dock.right === "shut" ? "shut" : "props";
}
