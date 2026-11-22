import * as THREE from "three";
import { box, cyl, glow, paint, rig, steel } from "./meshkit.js";

export function screen() {
  const group = rig();
  const pole = steel(0x2c333b, 0.4, 0.65);
  cyl(group, [0.08, 0.1, 1.05, 12], [0, 0.52, 0], pole);
  box(group, [0.45, 0.05, 0.32], [0, 0.03, 0], pole);
  box(group, [1.2, 0.72, 0.06], [0, 1.28, 0], paint(0x1c2128, 0.4));
  box(group, [1.05, 0.56, 0.02], [0, 1.28, 0.04], paint(0x10161d, 0.35));
  const colors = [0x6d5efc, 0x3b6fd8, 0xf5a524, 0x1fa85a];
  for (let i = 0; i < colors.length; i++) {
    const h = 0.14 + i * 0.08;
    box(group, [0.14, h, 0.03], [-0.36 + i * 0.24, 1.08 + h / 2, 0.06], paint(colors[i], 0.4));
  }
  return group;
}

export function pie() {
  const group = screen();
  const colors = [0x6d5efc, 0xf5a524, 0x1fa85a];
  colors.forEach((color, index) => {
    const mesh = new THREE.Mesh(
      new THREE.CylinderGeometry(0.22, 0.22, 0.06, 18, 1, false, index * Math.PI * 2 / 3, Math.PI * 2 / 3 - 0.04),
      paint(color, 0.4)
    );
    mesh.position.set(0.28, 1.32, 0.08);
    mesh.castShadow = true;
    group.add(mesh);
  });
  return group;
}

export function panel() {
  const group = rig();
  const desk = new THREE.Group();
  desk.position.y = 0.95;
  desk.rotation.x = -0.5;
  group.add(desk);
  box(desk, [0.78, 0.08, 0.46], [0, 0, 0], paint(0xdfe6ee, 0.48));
  cyl(desk, [0.05, 0.05, 0.03, 14], [-0.2, 0.055, 0.08], glow(0x21c45a));
  cyl(desk, [0.05, 0.05, 0.03, 14], [0, 0.055, 0.08], glow(0xd23b3b));
  cyl(desk, [0.05, 0.05, 0.03, 14], [0.2, 0.055, 0.08], glow(0xf2c14e));
  box(desk, [0.36, 0.025, 0.05], [0, 0.05, -0.1], paint(0x1c2128, 0.45));
  box(desk, [0.08, 0.05, 0.07], [0.06, 0.08, -0.1], paint(0xf2c14e, 0.4));
  cyl(group, [0.04, 0.05, 0.9, 10], [-0.26, 0.45, 0], steel(0x333941, 0.4, 0.65));
  cyl(group, [0.04, 0.05, 0.9, 10], [0.26, 0.45, 0], steel(0x333941, 0.4, 0.65));
  return group;
}

export function beacon() {
  const group = rig();
  cyl(group, [0.07, 0.09, 1.15, 12], [0, 0.58, 0], steel(0x333941, 0.4, 0.65));
  cyl(group, [0.14, 0.14, 0.16, 16], [0, 1.22, 0], glow(0xd23b3b));
  cyl(group, [0.14, 0.14, 0.16, 16], [0, 1.4, 0], glow(0xf2c14e));
  cyl(group, [0.14, 0.14, 0.16, 16], [0, 1.58, 0], glow(0x21c45a));
  cyl(group, [0.16, 0.16, 0.05, 16], [0, 1.7, 0], steel(0x24282e, 0.4, 0.6));
  return group;
}

export function crate() {
  const group = rig();
  const wood = paint(0xa87443, 0.78);
  box(group, [0.72, 0.06, 0.52], [0, 0.03, 0], wood);
  box(group, [0.72, 0.4, 0.04], [0, 0.26, 0.24], wood);
  box(group, [0.72, 0.4, 0.04], [0, 0.26, -0.24], wood);
  box(group, [0.04, 0.4, 0.52], [0.34, 0.26, 0], wood);
  box(group, [0.04, 0.4, 0.52], [-0.34, 0.26, 0], wood);
  box(group, [0.72, 0.04, 0.52], [0, 0.48, 0], wood);
  return group;
}

export function pallet() {
  const group = rig();
  const wood = paint(0xb08968, 0.78);
  box(group, [1.2, 0.1, 0.8], [0, 0.12, 0], wood);
  box(group, [1.05, 0.06, 0.1], [0, 0.04, 0.28], wood);
  box(group, [1.05, 0.06, 0.1], [0, 0.04, -0.28], wood);
  const load = crate();
  load.position.y = 0.16;
  group.add(load);
  return group;
}

export function cameraRig() {
  const group = rig();
  const metal = steel(0x2a3038, 0.32, 0.72);
  cyl(group, [0.05, 0.07, 1.2, 12], [0, 0.6, 0], metal);
  box(group, [0.28, 0.16, 0.16], [0, 1.28, 0], metal);
  cyl(group, [0.07, 0.08, 0.18, 14], [0.2, 1.28, 0, 0, 0, Math.PI / 2], paint(0x111418, 0.3));
  box(group, [0.1, 0.08, 0.08], [0, 1.4, -0.02], metal);
  return group;
}

export function lamp() {
  const group = rig();
  const metal = steel(0x333941, 0.35, 0.7);
  cyl(group, [0.05, 0.08, 1.7, 12], [0, 0.85, 0], metal);
  box(group, [0.28, 0.16, 0.2], [0.1, 1.75, 0, 0, 0, -0.4], metal);
  box(group, [0.22, 0.04, 0.16], [0.16, 1.66, 0.02, 0, 0, -0.4], glow(0xfff1c9));
  return group;
}

export function mapTable() {
  const group = rig();
  const top = paint(0xdfe6ee, 0.5);
  box(group, [2.4, 0.08, 1.5], [0, 0.78, 0], top);
  for (const x of [-1.05, 1.05]) {
    for (const z of [-0.6, 0.6]) box(group, [0.06, 0.78, 0.06], [x, 0.39, z], steel(0x333941, 0.4, 0.6));
  }
  box(group, [2.1, 0.04, 1.2], [0, 0.84, 0], paint(0x8fbf7a, 0.8));
  box(group, [1.5, 0.02, 0.22], [0, 0.88, 0.15], paint(0x4a515a, 0.7));
  box(group, [0.35, 0.28, 0.28], [-0.55, 1.02, -0.25], paint(0xe7edf3, 0.5));
  box(group, [0.28, 0.45, 0.28], [0.45, 1.1, -0.2], paint(0xc5ccd4, 0.5));
  return group;
}

export function easel() {
  const group = rig();
  const wood = paint(0x6b4a2e, 0.75);
  box(group, [0.06, 1.5, 0.06], [-0.28, 0.75, 0.1, 0.15, 0, 0], wood);
  box(group, [0.06, 1.5, 0.06], [0.28, 0.75, 0.1, -0.15, 0, 0], wood);
  box(group, [0.06, 1.35, 0.06], [0, 0.7, -0.2, 0.35, 0, 0], wood);
  box(group, [0.7, 0.5, 0.04], [0, 1.05, 0], paint(0xf4f1e4, 0.6));
  box(group, [0.62, 0.36, 0.02], [0, 1.05, 0.03], paint(0x3b6fd8, 0.45));
  return group;
}

export function table() {
  const group = rig();
  cyl(group, [0.08, 0.1, 0.72, 12], [0, 0.36, 0], steel(0x333941, 0.4, 0.65));
  cyl(group, [0.62, 0.62, 0.06, 28], [0, 0.75, 0], paint(0xdfe6ee, 0.45));
  cyl(group, [0.28, 0.28, 0.04, 20], [0, 0.04, 0], steel(0x333941, 0.45, 0.6));
  return group;
}

export function dock() {
  const group = rig();
  box(group, [2.4, 0.9, 1.5], [0, 0.45, 0], paint(0xc5ccd4, 0.6));
  box(group, [2.4, 0.08, 0.08], [0, 0.92, 0.74], paint(0xf2c14e, 0.45));
  box(group, [0.7, 0.08, 0.28], [1.35, 0.28, 0.9], steel(0x333941, 0.45, 0.6));
  box(group, [0.7, 0.08, 0.28], [1.35, 0.48, 0.7], steel(0x333941, 0.45, 0.6));
  box(group, [0.7, 0.08, 0.28], [1.35, 0.68, 0.5], steel(0x333941, 0.45, 0.6));
  return group;
}
