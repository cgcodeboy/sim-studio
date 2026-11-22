import * as THREE from "three";
import { RoundedBoxGeometry } from "three/addons/geometries/RoundedBoxGeometry.js";

export const METER = 0.012;

export function steel(color, rough = 0.4, metal = 0.62) {
  return new THREE.MeshStandardMaterial({ color, roughness: rough, metalness: metal });
}

export function paint(color, rough = 0.55) {
  return steel(color, rough, 0.06);
}

export function glow(color) {
  return new THREE.MeshStandardMaterial({ color, emissive: color, emissiveIntensity: 0.75, roughness: 0.4, metalness: 0.04 });
}

export function glass(color) {
  return new THREE.MeshPhysicalMaterial({
    color, roughness: 0.08, metalness: 0, transmission: 0.45, thickness: 0.15, transparent: true, opacity: 0.72
  });
}

export function rig() {
  const group = new THREE.Group();
  group.userData.roll = [];
  return group;
}

function pose(mesh, at) {
  mesh.position.set(at[0] || 0, at[1] || 0, at[2] || 0);
  mesh.rotation.set(at[3] || 0, at[4] || 0, at[5] || 0);
  mesh.castShadow = true;
  mesh.receiveShadow = true;
  return mesh;
}

export function box(group, size, at, mat) {
  const span = Math.min(size[0], size[1], size[2]);
  const radius = Math.min(0.01, span * 0.28);
  const geometry = radius < 0.0015
    ? new THREE.BoxGeometry(size[0], size[1], size[2])
    : new RoundedBoxGeometry(size[0], size[1], size[2], 2, radius);
  const mesh = pose(new THREE.Mesh(geometry, mat), at);
  group.add(mesh);
  return mesh;
}

export function cyl(group, size, at, mat) {
  const mesh = pose(new THREE.Mesh(new THREE.CylinderGeometry(size[0], size[1], size[2], size[3] || 16), mat), at);
  group.add(mesh);
  return mesh;
}

export function sph(group, radius, at, mat) {
  const mesh = pose(new THREE.Mesh(new THREE.SphereGeometry(radius, 18, 12), mat), at);
  group.add(mesh);
  return mesh;
}

export function tor(group, size, at, mat) {
  const mesh = pose(new THREE.Mesh(new THREE.TorusGeometry(size[0], size[1], 8, 18), mat), at);
  group.add(mesh);
  return mesh;
}

export function axle(group, at, size, mat) {
  const holder = new THREE.Group();
  holder.position.set(at[0], at[1], at[2]);
  if (size[3] === 1) holder.rotation.z = Math.PI / 2;
  else holder.rotation.x = Math.PI / 2;
  const mesh = new THREE.Mesh(new THREE.CylinderGeometry(size[1], size[1], size[0], 16), mat);
  mesh.castShadow = true;
  mesh.receiveShadow = true;
  holder.add(mesh);
  group.add(holder);
  if (size[2]) group.userData.roll.push({ mesh, speed: size[2], axis: "y" });
  return mesh;
}

export function spin(group, mesh, speed) {
  group.userData.roll.push({ mesh, speed, axis: "y" });
}

export function beltTexture(length) {
  const canvas = document.createElement("canvas");
  canvas.width = 64;
  canvas.height = 32;
  const pen = canvas.getContext("2d");
  pen.fillStyle = "#c5cbcf";
  pen.fillRect(0, 0, 64, 32);
  pen.strokeStyle = "#6d747b";
  pen.lineWidth = 4;
  pen.beginPath();
  pen.moveTo(4, 26);
  pen.lineTo(20, 6);
  pen.lineTo(36, 26);
  pen.stroke();
  const map = new THREE.CanvasTexture(canvas);
  map.wrapS = THREE.RepeatWrapping;
  map.colorSpace = THREE.SRGBColorSpace;
  map.repeat.set(Math.max(2, length / 0.4), 1);
  return map;
}

export function lengthOf(element, min) {
  return Math.max(min, Math.max(element.w || 0, element.h || 0) * METER);
}

export function spot(element) {
  const pose = element?.pose;
  if (pose && Number.isFinite(pose.x) && Number.isFinite(pose.z)) {
    return { x: pose.x, z: pose.z, yaw: Number.isFinite(pose.yaw) ? pose.yaw : 0 };
  }
  return { x: 0, z: 0, yaw: 0 };
}

export function place(element, group) {
  const at = spot(element);
  group.position.set(at.x, 0, at.z);
  group.rotation.y = at.yaw;
  group.userData.elementId = element.id;
  return group;
}

export function tickModels(root, time) {
  root.traverse((node) => {
    const jobs = node.userData?.roll;
    if (!jobs?.length) return;
    for (const job of jobs) job.mesh.rotation[job.axis] = time * job.speed;
    if (node.userData.belt) node.userData.belt.offset.x = time * 0.45;
  });
}

export function release(object) {
  const materials = new Set();
  const geometries = new Set();
  object.traverse((node) => {
    if (node.geometry) geometries.add(node.geometry);
    const list = !node.material ? [] : Array.isArray(node.material) ? node.material : [node.material];
    list.forEach((material) => materials.add(material));
  });
  geometries.forEach((geometry) => geometry.dispose());
  materials.forEach((material) => {
    Object.values(material).forEach((value) => { if (value?.isTexture) value.dispose(); });
    material.dispose();
  });
}

export function carton() {
  const group = rig();
  const board = paint(0xd7b07a, 0.72);
  const edge = paint(0x8d6238, 0.6);
  box(group, [0.36, 0.22, 0.28], [0, 0.13, 0], board);
  box(group, [0.37, 0.012, 0.29], [0, 0.246, 0], edge);
  box(group, [0.012, 0.22, 0.29], [0.18, 0.13, 0], edge);
  box(group, [0.012, 0.22, 0.29], [-0.18, 0.13, 0], edge);
  box(group, [0.37, 0.04, 0.012], [0, 0.25, 0.08, 0.7, 0, 0], board);
  box(group, [0.37, 0.04, 0.012], [0, 0.25, -0.08, -0.7, 0, 0], board);
  return group;
}
