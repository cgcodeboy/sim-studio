import * as THREE from "three";
import { axle, beltTexture, box, cyl, glass, glow, lengthOf, paint, rig, sph, spin, steel, tor } from "./meshkit.js";

const BED = 0.78;

export function belt(element, bare) {
  const group = rig();
  const length = lengthOf(element, bare ? 1.15 : 2.2);
  rails(group, length);
  rollerBed(group, length);
  band(group, length);
  if (!bare) {
    stands(group, length);
    drive(group, length);
  }
  return group;
}

function rails(group, length) {
  const shell = paint(0xd7dbd4, 0.42);
  const lip = paint(0xe6b23c, 0.46);
  for (const z of [0.32, -0.32]) {
    box(group, [length, 0.02, 0.04], [0, BED - 0.06, z], shell);
    box(group, [length, 0.022, 0.028], [0, BED + 0.052, z], shell);
    box(group, [length, 0.008, 0.012], [0, BED + 0.066, z], lip);
  }
  const posts = Math.max(4, Math.round(length / 0.36));
  for (let i = 0; i < posts; i++) {
    const x = -length / 2 + (i + 0.5) * (length / posts);
    box(group, [0.014, 0.09, 0.014], [x, BED, 0.32], shell);
    box(group, [0.014, 0.09, 0.014], [x, BED, -0.32], shell);
  }
}

function rollerBed(group, length) {
  const roller = steel(0xe7ebef, 0.22, 0.72);
  const count = Math.max(8, Math.floor((length - 0.2) / 0.11));
  for (let i = 0; i < count; i++) {
    const x = -length / 2 + 0.1 + i * ((length - 0.2) / Math.max(1, count - 1));
    axle(group, [x, BED - 0.002, 0], [0.58, 0.026, 4.5, 0], roller);
  }
  const drum = steel(0xc5c9ce, 0.3, 0.6);
  axle(group, [-length / 2 + 0.06, BED, 0], [0.6, 0.04, 2.6, 0], drum);
  axle(group, [length / 2 - 0.06, BED, 0], [0.6, 0.04, 2.6, 0], drum);
}

function band(group, length) {
  const map = beltTexture(length);
  box(group, [length - 0.14, 0.006, 0.4], [0, BED + 0.03, 0], new THREE.MeshStandardMaterial({ map, color: 0xffffff, roughness: 0.9, metalness: 0 }));
  group.userData.belt = map;
}

function stands(group, length) {
  const tube = paint(0xc9cec6, 0.5);
  const steps = Math.max(2, Math.round(length / 1.1));
  for (let i = 0; i < steps; i++) stand(group, -length / 2 + length * (i + 0.5) / steps, tube);
}

function stand(group, x, mat) {
  box(group, [0.04, 0.72, 0.04], [x, 0.36, 0.32], mat);
  box(group, [0.04, 0.72, 0.04], [x, 0.36, -0.32], mat);
  box(group, [0.028, 0.028, 0.64], [x, 0.22, 0], mat);
  box(group, [0.016, 0.46, 0.016], [x, 0.46, 0.1, 0, 0, 0.7], mat);
  box(group, [0.016, 0.46, 0.016], [x, 0.46, -0.1, 0, 0, -0.7], mat);
  box(group, [0.14, 0.012, 0.1], [x, 0.01, 0.32], mat);
  box(group, [0.14, 0.012, 0.1], [x, 0.01, -0.32], mat);
}

function drive(group, length) {
  const x = length / 2 - 0.02;
  const iron = paint(0x3e444c, 0.42);
  box(group, [0.16, 0.14, 0.14], [x, 0.34, 0.4], iron);
  cyl(group, [0.05, 0.05, 0.16, 18], [x + 0.15, 0.34, 0.4, 0, 0, Math.PI / 2], iron);
  cyl(group, [0.058, 0.058, 0.018, 18], [x + 0.24, 0.34, 0.4, 0, 0, Math.PI / 2], paint(0x2c3138, 0.5));
  box(group, [0.03, 0.07, 0.03], [0, BED + 0.1, 0.33], iron);
  cyl(group, [0.01, 0.01, 0.05, 10], [0, BED + 0.1, 0.29, Math.PI / 2, 0, 0], paint(0x22262c, 0.4));
}

export function buffer(element) {
  const group = belt(element);
  const x = lengthOf(element, 1.8) / 2 - 0.05;
  const stop = paint(0xe07a22, 0.42);
  box(group, [0.016, 0.14, 0.03], [x, BED + 0.1, 0.22], stop);
  box(group, [0.016, 0.14, 0.03], [x, BED + 0.1, -0.22], stop);
  box(group, [0.016, 0.1, 0.44], [x, BED + 0.12, 0], stop);
  return group;
}

export function hopper() {
  const group = rig();
  const frame = paint(0xd7dbd4, 0.46);
  const body = paint(0xf3f1ea, 0.4);
  const rib = paint(0xe07a22, 0.42);
  for (const x of [-0.36, 0.36]) {
    for (const z of [-0.36, 0.36]) box(group, [0.04, 0.86, 0.04], [x, 0.43, z], frame);
  }
  for (const x of [-0.36, 0.36]) box(group, [0.04, 0.035, 0.76], [x, 0.86, 0], frame);
  for (const z of [-0.36, 0.36]) box(group, [0.76, 0.035, 0.04], [0, 0.86, z], frame);
  cyl(group, [0.38, 0.11, 0.58, 4], [0, 1.2, 0, 0, Math.PI / 4, 0], body);
  cyl(group, [0.44, 0.44, 0.035, 4], [0, 1.5, 0, 0, Math.PI / 4, 0], frame);
  box(group, [0.02, 0.46, 0.06], [0.26, 1.2, 0], rib);
  box(group, [0.06, 0.46, 0.02], [0, 1.2, 0.26], rib);
  cyl(group, [0.09, 0.07, 0.14, 16], [0, 0.84, 0], steel(0xd5d8dc, 0.3, 0.45));
  cyl(group, [0.12, 0.12, 0.018, 16], [0, 0.76, 0], frame);
  const pan = rig();
  box(pan, [0.62, 0.01, 0.4], [0, BED, 0], frame);
  box(pan, [0.62, 0.05, 0.012], [0, BED + 0.025, 0.19], frame);
  box(pan, [0.62, 0.05, 0.012], [0, BED + 0.025, -0.19], frame);
  pan.position.x = 0.78;
  group.add(pan);
  return group;
}

export function bin() {
  const group = rig();
  const wall = paint(0xe7eee6, 0.48);
  const rim = paint(0x2f6f45, 0.42);
  box(group, [1.12, 0.035, 0.84], [0, 0.14, 0], paint(0xc9cec6, 0.55));
  box(group, [1.12, 0.58, 0.018], [0, 0.45, 0.41], wall);
  box(group, [1.12, 0.58, 0.018], [0, 0.45, -0.41], wall);
  box(group, [0.018, 0.58, 0.84], [0.55, 0.45, 0], wall);
  box(group, [0.018, 0.38, 0.8], [-0.55, 0.35, 0], wall);
  box(group, [1.16, 0.028, 0.04], [0, 0.75, 0.41], rim);
  box(group, [1.16, 0.028, 0.04], [0, 0.75, -0.41], rim);
  box(group, [0.04, 0.028, 0.84], [0.55, 0.75, 0], rim);
  box(group, [0.28, 0.07, 0.07], [-0.1, 0.07, 0.16], paint(0x3a3d42, 0.5));
  box(group, [0.28, 0.07, 0.07], [-0.1, 0.07, -0.16], paint(0x3a3d42, 0.5));
  chute(group);
  return group;
}

function chute(group) {
  const pan = new THREE.Group();
  const metal = steel(0xd5d8dc, 0.32, 0.4);
  box(pan, [0.66, 0.01, 0.38], [0, 0, 0], metal);
  box(pan, [0.66, 0.055, 0.012], [0, 0.028, 0.18], metal);
  box(pan, [0.66, 0.055, 0.012], [0, 0.028, -0.18], metal);
  pan.position.set(-0.5, 1.02, 0);
  pan.rotation.z = 0.62;
  group.add(pan);
}

export function machine() {
  const group = rig();
  const body = paint(0xf4f4f1, 0.38);
  box(group, [1.36, 0.07, 0.88], [0, 0.035, 0], paint(0xc9cec6, 0.48));
  box(group, [1.16, 1.16, 0.02], [0, 0.68, -0.35], body);
  box(group, [0.02, 1.16, 0.7], [-0.57, 0.68, 0], body);
  box(group, [0.02, 1.16, 0.7], [0.57, 0.68, 0], body);
  box(group, [1.2, 0.03, 0.74], [0, 1.27, 0], body);
  door(group);
  louvers(group);
  cyl(group, [0.035, 0.045, 0.06, 12], [0.36, 1.38, 0], paint(0xc9cec6, 0.4));
  cyl(group, [0.055, 0.06, 0.07, 16], [0.36, 1.45, 0], glow(0x3ecf6e));
  const feed = belt({ w: 80, h: 36 }, true);
  feed.position.x = -1.02;
  group.add(feed);
  return group;
}

function door(group) {
  const face = 0.37;
  const trim = paint(0xd0d4cc, 0.4);
  const ink = paint(0x3a3d42, 0.38);
  box(group, [1.08, 0.02, 0.016], [0, 1.2, face], trim);
  box(group, [1.08, 0.02, 0.016], [0, 0.16, face], trim);
  box(group, [0.02, 1.06, 0.016], [-0.53, 0.68, face], trim);
  box(group, [0.02, 1.06, 0.016], [0.53, 0.68, face], trim);
  box(group, [1.0, 1.0, 0.014], [0, 0.68, face - 0.01], paint(0xf7f7f4, 0.32));
  box(group, [0.02, 0.22, 0.02], [0.42, 0.68, face + 0.012], ink);
  cyl(group, [0.012, 0.012, 0.08, 8], [-0.5, 0.98, face], ink);
  cyl(group, [0.012, 0.012, 0.08, 8], [-0.5, 0.38, face], ink);
  box(group, [0.36, 0.22, 0.008], [0, 0.9, face + 0.01], glass(0xd7e7ec));
  box(group, [0.4, 0.014, 0.012], [0, 1.02, face + 0.008], ink);
  box(group, [0.4, 0.014, 0.012], [0, 0.78, face + 0.008], ink);
  box(group, [0.014, 0.26, 0.012], [-0.19, 0.9, face + 0.008], ink);
  box(group, [0.014, 0.26, 0.012], [0.19, 0.9, face + 0.008], ink);
}

function louvers(group) {
  const slat = paint(0xb7bbb4, 0.45);
  for (let i = 0; i < 5; i++) box(group, [0.02, 0.012, 0.22], [0.59, 0.46 + i * 0.04, 0, 0, 0, 0.55], slat);
}

export function divert(element) {
  const group = belt(element);
  box(group, [0.62, 0.035, 0.1], [0.05, BED + 0.12, 0.02, 0, 0.6, 0], paint(0xf2c14e, 0.4));
  const spur = belt({ w: 120, h: 36 }, true);
  spur.position.set(0.35, 0, 0.72);
  spur.rotation.y = -0.7;
  group.add(spur);
  return group;
}

export function merge() {
  const group = rig();
  const main = belt({ w: 180, h: 40 });
  main.position.x = 0.55;
  group.add(main);
  const left = belt({ w: 130, h: 36 }, true);
  left.position.set(-0.85, 0, 0.62);
  left.rotation.y = 0.55;
  group.add(left);
  const right = belt({ w: 130, h: 36 }, true);
  right.position.set(-0.85, 0, -0.62);
  right.rotation.y = -0.55;
  group.add(right);
  return group;
}

export function gate() {
  const group = rig();
  const post = steel(0x343b44, 0.4, 0.65);
  box(group, [0.08, 1.55, 0.08], [-0.58, 0.78, 0], post);
  box(group, [0.08, 1.55, 0.08], [0.58, 0.78, 0], post);
  box(group, [1.28, 0.08, 0.08], [0, 1.52, 0], post);
  box(group, [1.05, 0.035, 0.06], [0.08, 0.98, 0, 0, 0, 0.18], paint(0xd23b3b, 0.42));
  box(group, [0.9, 0.015, 0.55], [0, 0.01, 0], paint(0xf2c14e, 0.6));
  return group;
}

export function robot() {
  const group = rig();
  const arm = paint(0xf4f4f1, 0.36);
  const joint = paint(0x3a3d42, 0.34);
  const accent = paint(0xe07a22, 0.4);
  cyl(group, [0.26, 0.3, 0.06, 24], [0, 0.03, 0], joint);
  cyl(group, [0.2, 0.22, 0.08, 24], [0, 0.1, 0], arm);
  for (const turn of [0, 1.57, 3.14, 4.71]) {
    cyl(group, [0.014, 0.014, 0.016, 8], [Math.cos(turn) * 0.16, 0.07, Math.sin(turn) * 0.16], joint);
  }
  cyl(group, [0.07, 0.08, 0.36, 18], [0, 0.32, 0], accent);
  const shoulder = new THREE.Group();
  shoulder.position.set(0, 0.54, 0);
  shoulder.rotation.z = -0.7;
  group.add(shoulder);
  sph(shoulder, 0.08, [0, 0, 0], joint);
  cyl(shoulder, [0.048, 0.048, 0.4, 16], [0.22, 0, 0, 0, 0, Math.PI / 2], arm);
  const elbow = new THREE.Group();
  elbow.position.set(0.42, 0, 0);
  elbow.rotation.z = 1.15;
  shoulder.add(elbow);
  sph(elbow, 0.062, [0, 0, 0], joint);
  cyl(elbow, [0.036, 0.036, 0.32, 16], [0.18, 0, 0, 0, 0, Math.PI / 2], arm);
  box(elbow, [0.05, 0.09, 0.01], [0.36, 0, 0.028], joint);
  box(elbow, [0.05, 0.09, 0.01], [0.36, 0, -0.028], joint);
  return group;
}

export function crane() {
  const group = rig();
  const mast = paint(0xf0c31f, 0.42);
  box(group, [0.9, 0.08, 0.9], [0, 0.04, 0], paint(0x3a3d42, 0.45));
  box(group, [0.08, 3.05, 0.08], [-0.08, 1.56, 0.08], mast);
  box(group, [0.08, 3.05, 0.08], [0.08, 1.56, -0.08], mast);
  for (let i = 0; i < 6; i++) box(group, [0.02, 0.42, 0.02], [0, 0.4 + i * 0.46, 0, 0, 0, i % 2 ? 0.5 : -0.5], mast);
  jib(group, mast);
  box(group, [0.14, 0.08, 0.16], [1.55, 2.92, 0], paint(0x3a3d42, 0.4));
  box(group, [0.012, 1.2, 0.012], [1.55, 2.28, 0], paint(0x2c3138, 0.4));
  tor(group, [0.07, 0.012], [1.55, 1.64, 0, Math.PI / 2, 0, 0], paint(0x2c3138, 0.35));
  return group;
}

function jib(group, mat) {
  box(group, [2.3, 0.035, 0.035], [1.15, 3.12, 0.05], mat);
  box(group, [2.3, 0.035, 0.035], [1.15, 2.9, -0.05], mat);
  for (let i = 0; i < 5; i++) box(group, [0.4, 0.016, 0.016], [0.28 + i * 0.4, 3.01, 0, 0, 0, i % 2 ? 0.5 : -0.5], mat);
}

export function rack() {
  const group = rig();
  const upright = paint(0x3b6fd8, 0.45);
  const beam = paint(0xf08a14, 0.45);
  for (const x of [-0.7, 0.7]) {
    for (const z of [-0.32, 0.32]) box(group, [0.06, 2.5, 0.06], [x, 1.25, z], upright);
  }
  for (const y of [0.7, 1.4, 2.1]) {
    box(group, [1.46, 0.06, 0.05], [0, y, 0.32], beam);
    box(group, [1.46, 0.06, 0.05], [0, y, -0.32], beam);
    if (y < 2) palletDeck(group, y + 0.08);
  }
  return group;
}

function palletDeck(group, y) {
  const wood = paint(0xb08968, 0.75);
  box(group, [0.7, 0.08, 0.5], [-0.32, y, 0], wood);
  box(group, [0.7, 0.08, 0.5], [0.32, y, 0], wood);
  box(group, [0.42, 0.28, 0.32], [-0.32, y + 0.2, 0], paint(0xc4a574, 0.7));
  box(group, [0.42, 0.22, 0.32], [0.32, y + 0.16, 0], paint(0xd7c4a0, 0.7));
}

export function turntable() {
  const group = rig();
  cyl(group, [0.98, 0.98, 0.08, 32], [0, 0.06, 0], steel(0x2c333b, 0.5, 0.55));
  const disc = cyl(group, [0.84, 0.84, 0.1, 32], [0, 0.15, 0], steel(0x8b949e, 0.32, 0.75));
  spin(group, disc, 0.5);
  box(group, [1.45, 0.02, 0.18], [0, 0.22, 0], paint(0xf2c14e, 0.45));
  cyl(group, [0.12, 0.12, 0.16, 14], [0, 0.22, 0], steel(0x3a424c, 0.35, 0.7));
  return group;
}

export function cart() {
  const group = rig();
  const rubber = paint(0x1c2128, 0.85);
  box(group, [0.95, 0.08, 0.58], [0, 0.38, 0], paint(0x3b6fd8, 0.45));
  box(group, [0.7, 0.16, 0.42], [0, 0.48, 0], paint(0xc4a574, 0.72));
  box(group, [0.04, 0.72, 0.04], [-0.42, 0.74, 0], steel(0x333941, 0.4, 0.65));
  axle(group, [0.32, 0.16, 0.3], [0.08, 0.16, 4, 0], rubber);
  axle(group, [0.32, 0.16, -0.3], [0.08, 0.16, 4, 0], rubber);
  axle(group, [-0.32, 0.16, 0.3], [0.08, 0.16, 4, 0], rubber);
  axle(group, [-0.32, 0.16, -0.3], [0.08, 0.16, 4, 0], rubber);
  return group;
}
