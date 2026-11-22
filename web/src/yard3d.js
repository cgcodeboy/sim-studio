import { axle, box, cyl, glass, lengthOf, paint, rig, sph, steel, tor } from "./meshkit.js";

const RUBBER = () => paint(0x1a1e24, 0.9);

export function car() {
  const group = rig();
  const body = paint(0xd64545, 0.42);
  const glassPane = paint(0x9ec9e0, 0.15);
  box(group, [4.15, 0.48, 1.72], [0.1, 0.52, 0], body);
  box(group, [2.05, 0.48, 1.55], [-0.15, 0.95, 0], body);
  box(group, [1.7, 0.36, 1.42], [-0.15, 1.02, 0], glassPane);
  box(group, [0.7, 0.12, 1.6], [1.55, 0.7, 0], body);
  box(group, [0.08, 0.18, 1.68], [2.05, 0.42, 0], paint(0xf4f1e4, 0.5));
  wheel(group, 1.28, 0.96);
  wheel(group, 1.28, -0.96);
  wheel(group, -1.25, 0.96);
  wheel(group, -1.25, -0.96);
  return group;
}

function wheel(group, x, z) {
  axle(group, [x, 0.3, z], [0.2, 0.3, 3, 0], RUBBER());
  axle(group, [x, 0.3, z], [0.22, 0.16, 0, 0], steel(0xc5ccd4, 0.25, 0.85));
}

export function forklift() {
  const group = rig();
  const body = paint(0xf08a14, 0.42);
  box(group, [1.35, 0.55, 1.05], [-0.15, 0.55, 0], body);
  box(group, [0.55, 0.4, 0.7], [-0.35, 0.95, 0], paint(0x24292f, 0.4));
  box(group, [0.08, 1.7, 0.08], [0.62, 1.05, 0.28], steel(0x333941, 0.35, 0.7));
  box(group, [0.08, 1.7, 0.08], [0.62, 1.05, -0.28], steel(0x333941, 0.35, 0.7));
  box(group, [0.7, 0.04, 0.1], [0.95, 0.22, 0.18], steel(0xb7c0c8, 0.3, 0.8));
  box(group, [0.7, 0.04, 0.1], [0.95, 0.22, -0.18], steel(0xb7c0c8, 0.3, 0.8));
  wheel(group, -0.55, 0.48);
  wheel(group, -0.55, -0.48);
  axle(group, [0.45, 0.22, 0.42], [0.14, 0.2, 3, 0], RUBBER());
  axle(group, [0.45, 0.22, -0.42], [0.14, 0.2, 3, 0], RUBBER());
  return group;
}

export function person() {
  const group = rig();
  const cloth = paint(0x2f5fd0, 0.65);
  const skin = paint(0xe0b090, 0.6);
  box(group, [0.12, 0.72, 0.14], [-0.09, 0.42, 0], cloth);
  box(group, [0.12, 0.72, 0.14], [0.09, 0.42, 0], cloth);
  box(group, [0.14, 0.08, 0.26], [-0.09, 0.04, 0.04], paint(0x24282e, 0.6));
  box(group, [0.14, 0.08, 0.26], [0.09, 0.04, 0.04], paint(0x24282e, 0.6));
  box(group, [0.38, 0.5, 0.2], [0, 1.08, 0], cloth);
  box(group, [0.1, 0.48, 0.1], [-0.26, 1.02, 0], cloth);
  box(group, [0.1, 0.48, 0.1], [0.26, 1.02, 0], cloth);
  sph(group, 0.13, [0, 1.52, 0], skin);
  return group;
}

export function crew() {
  const group = rig();
  box(group, [1.4, 0.06, 0.45], [0, 0.78, 0], paint(0xc5ccd4, 0.5));
  box(group, [0.06, 0.78, 0.06], [-0.62, 0.39, 0.16], steel(0x333941, 0.4, 0.6));
  box(group, [0.06, 0.78, 0.06], [0.62, 0.39, 0.16], steel(0x333941, 0.4, 0.6));
  const left = person();
  left.position.set(-0.4, 0, 0.7);
  const right = person();
  right.position.set(0.4, 0, 0.7);
  group.add(left, right);
  return group;
}

export function train() {
  const group = rig();
  const body = paint(0x2a6f4e, 0.45);
  box(group, [3.6, 0.2, 1.35], [0, 0.55, 0], steel(0x2a3038, 0.45, 0.6));
  box(group, [2.3, 0.85, 1.4], [0.35, 1.05, 0], body);
  box(group, [1.05, 0.7, 1.25], [-1.15, 1.15, 0], paint(0xdfe7ee, 0.4));
  box(group, [0.7, 0.4, 1.05], [-1.15, 1.2, 0], paint(0x9ec9e0, 0.15));
  axle(group, [1.1, 0.42, 0.58], [0.14, 0.24, 2, 0], RUBBER());
  axle(group, [1.1, 0.42, -0.58], [0.14, 0.24, 2, 0], RUBBER());
  axle(group, [-0.9, 0.42, 0.58], [0.14, 0.24, 2, 0], RUBBER());
  axle(group, [-0.9, 0.42, -0.58], [0.14, 0.24, 2, 0], RUBBER());
  const rails = track({ w: 420, h: 28 });
  [...rails.children].forEach((child) => group.add(child));
  return group;
}

export function track(element) {
  const group = rig();
  const length = lengthOf(element, 3);
  const rail = steel(0x8d949c, 0.3, 0.8);
  box(group, [length, 0.08, 0.9], [0, 0.04, 0], paint(0x6a6258, 0.9));
  const ties = Math.max(3, Math.floor(length / 0.55));
  for (let i = 0; i < ties; i++) {
    const x = -length / 2 + (i + 0.5) * (length / ties);
    box(group, [0.18, 0.08, 1.05], [x, 0.1, 0], paint(0x6b4a2e, 0.8));
  }
  box(group, [length, 0.08, 0.06], [0, 0.18, 0.42], rail);
  box(group, [length, 0.08, 0.06], [0, 0.18, -0.42], rail);
  return group;
}

export function road(element) {
  const group = rig();
  const lanes = Math.max(1, Number(element.props?.lanes) || 2);
  const width = 3.5 * lanes;
  const length = lengthOf(element, 6);
  box(group, [length, 0.08, width], [0, 0.04, 0], paint(0x3c424a, 0.92));
  box(group, [length, 0.12, 0.14], [0, 0.08, width / 2], paint(0xc5ccd4, 0.7));
  box(group, [length, 0.12, 0.14], [0, 0.08, -width / 2], paint(0xc5ccd4, 0.7));
  const dashes = Math.max(2, Math.floor(length / 1.8));
  for (let i = 0; i < dashes; i++) {
    const x = -length / 2 + (i + 0.5) * (length / dashes);
    for (let lane = 1; lane < lanes; lane++) box(group, [0.8, 0.02, 0.12], [x, 0.09, -width / 2 + lane * 3.5], paint(0xf4f1e4, 0.5));
  }
  return group;
}

export function crossing() {
  const group = rig();
  box(group, [8, 0.06, 8], [0, 0.03, 0], paint(0x3c424a, 0.92));
  for (let i = 0; i < 6; i++) {
    box(group, [0.45, 0.02, 2.6], [-1.6 + i * 0.7, 0.07, 2.4], paint(0xf4f1e4, 0.5));
    box(group, [0.45, 0.02, 2.6], [-1.6 + i * 0.7, 0.07, -2.4], paint(0xf4f1e4, 0.5));
  }
  return group;
}

export function stopSign() {
  const group = rig();
  cyl(group, [0.06, 0.06, 2.3, 10], [0, 1.15, 0], steel(0x333941, 0.4, 0.65));
  cyl(group, [0.38, 0.38, 0.04, 8], [0, 2.15, 0, Math.PI / 2, 0, 0], paint(0xd23b3b, 0.4));
  cyl(group, [0.28, 0.28, 0.05, 8], [0, 2.15, 0.02, Math.PI / 2, 0, 0], paint(0xf4f1e4, 0.4));
  box(group, [1.6, 0.02, 0.12], [0.9, 0.04, 0], paint(0xf4f1e4, 0.5));
  return group;
}

export function park() {
  const group = rig();
  box(group, [7.2, 0.06, 5], [0, 0.03, 0], paint(0x4a515a, 0.9));
  for (let i = 0; i < 4; i++) box(group, [0.08, 0.02, 4.4], [-2.7 + i * 1.8, 0.07, 0], paint(0xf4f1e4, 0.5));
  return group;
}

export function pipe(element) {
  const group = rig();
  const length = lengthOf(element, 2);
  const metal = steel(0xc5ced6, 0.22, 0.85);
  axle(group, [0, 1.25, 0], [length, 0.09, 0, 1], metal);
  cyl(group, [0.14, 0.14, 0.05, 16], [-length / 2, 1.25, 0, 0, 0, Math.PI / 2], metal);
  cyl(group, [0.14, 0.14, 0.05, 16], [length / 2, 1.25, 0, 0, 0, Math.PI / 2], metal);
  const stands = Math.max(2, Math.round(length / 1.6));
  for (let i = 0; i < stands; i++) {
    const x = -length / 2 + length * (i + 0.5) / stands;
    box(group, [0.06, 1.2, 0.06], [x, 0.6, 0], steel(0x3e4650, 0.45, 0.65));
  }
  return group;
}

export function tank() {
  const group = rig();
  const shell = steel(0xd5dde4, 0.28, 0.72);
  for (const x of [-0.55, 0.55]) {
    for (const z of [-0.55, 0.55]) box(group, [0.08, 0.7, 0.08], [x, 0.35, z], steel(0x3e4650, 0.45, 0.65));
  }
  cyl(group, [0.78, 0.78, 0.1, 28], [0, 0.8, 0], shell);
  cyl(group, [0.76, 0.76, 1.6, 28], [0, 1.65, 0], glass(0x9fd4ea));
  cyl(group, [0.68, 0.68, 0.95, 24], [0, 1.35, 0], paint(0x2f86c4, 0.25));
  cyl(group, [0.8, 0.7, 0.28, 28], [0, 2.55, 0], shell);
  box(group, [0.04, 1.5, 0.08], [0.84, 1.55, 0], steel(0x333941, 0.4, 0.6));
  cyl(group, [0.06, 0.06, 0.35, 10], [0.95, 0.85, 0, 0, 0, Math.PI / 2], shell);
  return group;
}

export function valve(element) {
  const group = rig();
  const metal = steel(0xb7c0c8, 0.25, 0.82);
  const length = Math.min(1.4, lengthOf(element, 0.9));
  axle(group, [0, 0.7, 0], [length, 0.07, 0, 1], metal);
  sph(group, 0.16, [0, 0.7, 0], paint(0xd23b3b, 0.4));
  cyl(group, [0.03, 0.03, 0.22, 8], [0, 0.95, 0], metal);
  tor(group, [0.14, 0.02], [0, 1.08, 0, Math.PI / 2, 0, 0], steel(0x3e4650, 0.35, 0.7));
  return group;
}

export function pump() {
  const group = rig();
  box(group, [1.3, 0.08, 0.7], [0, 0.08, 0], steel(0x333941, 0.5, 0.55));
  cyl(group, [0.22, 0.22, 0.55, 18], [-0.25, 0.42, 0, 0, 0, Math.PI / 2], paint(0x3b6fd8, 0.4));
  cyl(group, [0.18, 0.18, 0.4, 16], [0.35, 0.4, 0], steel(0xc5ced6, 0.25, 0.8));
  axle(group, [0.7, 0.55, 0], [0.45, 0.05, 0, 1], steel(0xb7c0c8, 0.25, 0.8));
  return group;
}

export function sump() {
  const group = rig();
  box(group, [1.3, 0.35, 0.9], [0, 0.05, 0], paint(0x2a3038, 0.7));
  for (let i = 0; i < 5; i++) box(group, [1.15, 0.02, 0.06], [0, 0.24, -0.32 + i * 0.16], steel(0x8d949c, 0.35, 0.7));
  return group;
}

export function wall(element) {
  const group = rig();
  const length = lengthOf(element, 2.2);
  box(group, [length, 2.5, 0.18], [0, 1.25, 0], paint(0xe7e2d6, 0.88));
  box(group, [length, 0.1, 0.24], [0, 2.55, 0], paint(0x8d939c, 0.55));
  return group;
}

export function walk(element) {
  const group = rig();
  const length = lengthOf(element, 2);
  const rail = steel(0x8d949c, 0.32, 0.72);
  box(group, [length, 0.06, 1.25], [0, 0.14, 0], paint(0xd5dde4, 0.6));
  box(group, [length, 0.04, 0.04], [0, 0.95, 0.58], rail);
  box(group, [length, 0.04, 0.04], [0, 0.95, -0.58], rail);
  const posts = Math.max(2, Math.round(length / 1.4));
  for (let i = 0; i < posts; i++) {
    const x = -length / 2 + (i + 0.5) * (length / posts);
    box(group, [0.04, 0.85, 0.04], [x, 0.52, 0.58], rail);
    box(group, [0.04, 0.85, 0.04], [x, 0.52, -0.58], rail);
  }
  return group;
}

export function zone() {
  const group = rig();
  box(group, [2.4, 0.02, 1.8], [0, 0.02, 0], paint(0xf2c14e, 0.7));
  for (const x of [-1.1, 1.1]) {
    for (const z of [-0.8, 0.8]) cyl(group, [0.06, 0.06, 0.9, 10], [x, 0.45, z], paint(0xf08a14, 0.45));
  }
  return group;
}

export function bollard() {
  const group = rig();
  cyl(group, [0.1, 0.12, 0.95, 14], [0, 0.48, 0], steel(0xc5ced6, 0.3, 0.75));
  cyl(group, [0.12, 0.12, 0.08, 14], [0, 0.98, 0], paint(0xf2c14e, 0.4));
  return group;
}

export function bench() {
  const group = rig();
  const wood = paint(0x8a5a32, 0.75);
  box(group, [1.35, 0.06, 0.42], [0, 0.46, 0], wood);
  box(group, [1.35, 0.42, 0.05], [0, 0.72, -0.18], wood);
  box(group, [0.06, 0.46, 0.4], [-0.58, 0.23, 0], steel(0x333941, 0.4, 0.65));
  box(group, [0.06, 0.46, 0.4], [0.58, 0.23, 0], steel(0x333941, 0.4, 0.65));
  const who = person();
  who.position.z = 0.7;
  group.add(who);
  return group;
}

export function barrier(element) {
  const group = rig();
  const length = lengthOf(element, 1.4);
  const post = steel(0x333941, 0.4, 0.65);
  box(group, [0.06, 0.95, 0.06], [-length / 2, 0.48, 0], post);
  box(group, [0.06, 0.95, 0.06], [length / 2, 0.48, 0], post);
  box(group, [length, 0.08, 0.04], [0, 0.9, 0], paint(0xd23b3b, 0.45));
  return group;
}

export function tray(element) {
  const group = rig();
  const length = lengthOf(element, 1.2);
  box(group, [length, 0.06, 0.22], [0, 0.06, 0], steel(0x8d949c, 0.35, 0.7));
  box(group, [length, 0.03, 0.03], [0, 0.1, 0.09], paint(0xf2c14e, 0.45));
  return group;
}

export function vessel() {
  const group = tank();
  group.scale.setScalar(0.58);
  return group;
}
