const PATHS = {
  spark: "M12 3l1.6 5.2L19 10l-5.4 1.8L12 17l-1.6-5.2L5 10l5.4-1.8L12 3z",
  folder: "M3 7.5h6l2 2H21v9.5H3z",
  box: "M4 8l8-4 8 4-8 4-8-4zm0 0v8l8 4 8-4V8",
  play: "M8 6l10 6-10 6z",
  sliders: "M4 8h10M18 8h2M4 16h2M10 16h10M14 6v4M8 14v4",
  message: "M5 6h14v9H8l-3 3z",
  inbox: "M4 6h16v12H4zM4 14h4l2 2h4l2-2h4",
  flag: "M6 4v16M6 5h10l-2 3 2 3H6",
  layers: "M12 4l8 4-8 4-8-4 8-4zm-8 8l8 4 8-4M4 16l8 4 8-4",
  split: "M6 6h4v4H6zM14 14h4v4h-4zM8 10v2a4 4 0 0 0 4 4h2",
  pause: "M9 7v10M15 7v10",
  chevronLeft: "M14 6l-6 6 6 6",
  chevronRight: "M10 6l6 6-6 6",
  chevronUp: "M6 14l6-6 6 6",
  chevronDown: "M6 10l6 6 6-6"
};

function svg(size, children) {
  return (
    <svg className="ico" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {children}
    </svg>
  );
}

export function Icon({ name, size = 16 }) {
  if (name === "layout") return svg(size, <><rect x="4" y="4" width="16" height="16" rx="2" /><path d="M4 10h16M10 10v10" /></>);
  if (name === "clock") return svg(size, <><circle cx="12" cy="12" r="8" /><path d="M12 8v5l3 2" /></>);
  if (name === "users") return svg(size, <><circle cx="9" cy="9" r="3" /><circle cx="16" cy="10" r="2" /><path d="M4 18c1-3 3-4 5-4s4 1 5 4M14 14c2 0 3 1 4 3" /></>);
  return svg(size, <path d={PATHS[name] || PATHS.box} />);
}

const FLOW = {
  Source: "inbox",
  Sink: "flag",
  Delay: "clock",
  Queue: "layers",
  Hold: "pause",
  Service: "users",
  "Select Output": "split",
  "Select Output 5": "split"
};

export function moduleIcon(type, look) {
  if (FLOW[type]) return FLOW[type];
  if (look === "resource") return "users";
  if (look === "flow") return "box";
  return "layout";
}
