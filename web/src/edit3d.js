import * as THREE from "three";

const ground = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
const size = new THREE.Vector3();
const center = new THREE.Vector3();

function aim(dom, event, pointer) {
  const rect = dom.getBoundingClientRect();
  const width = rect.width || 1;
  const height = rect.height || 1;
  pointer.set(((event.clientX - rect.left) / width) * 2 - 1, -((event.clientY - rect.top) / height) * 2 + 1);
}

function picked(ctx, event) {
  aim(ctx.dom, event, ctx.pointer);
  ctx.raycaster.setFromCamera(ctx.pointer, ctx.camera);
  const hit = ctx.raycaster.intersectObjects([...ctx.bag.stations.values()], true)[0];
  let node = hit?.object || null;
  while (node && !node.userData.elementId) node = node.parent;
  return node;
}

export function attachPick(group) {
  group.updateWorldMatrix(true, true);
  const box = new THREE.Box3().setFromObject(group);
  if (box.isEmpty()) return;
  box.getSize(size);
  group.worldToLocal(box.getCenter(center));
  const mesh = new THREE.Mesh(
    new THREE.BoxGeometry(Math.max(size.x, 0.6), Math.max(size.y, 0.4), Math.max(size.z, 0.6)),
    new THREE.MeshBasicMaterial({ transparent: true, opacity: 0, depthWrite: false })
  );
  mesh.position.copy(center);
  mesh.userData.pick = true;
  group.add(mesh);
}

function placeHold(ctx, pose) {
  const mesh = ctx.bag.stations.get(ctx.drag.id);
  ctx.bag.hold = { id: ctx.drag.id, ...pose };
  if (!mesh) return;
  mesh.position.set(pose.x, 0, pose.z);
  mesh.rotation.y = pose.yaw;
}

function onDown(ctx, event) {
  if (!ctx.live.current.editing || event.button !== 0) return;
  const node = picked(ctx, event);
  if (!node) return;
  event.preventDefault();
  event.stopPropagation();
  ctx.bag.controls.enabled = false;
  ctx.dom.setPointerCapture?.(event.pointerId);
  const id = node.userData.elementId;
  ctx.drag = {
    id, shift: event.shiftKey, pointer: event.pointerId,
    x: node.position.x, z: node.position.z, yaw: node.rotation.y,
    cx: event.clientX, cy: event.clientY, moved: false, grab: null
  };
  aim(ctx.dom, event, ctx.pointer);
  ctx.raycaster.setFromCamera(ctx.pointer, ctx.camera);
  if (ctx.raycaster.ray.intersectPlane(ground, ctx.hit)) ctx.drag.grab = { x: ctx.hit.x - node.position.x, z: ctx.hit.z - node.position.z };
  ctx.dom.style.cursor = "grabbing";
  ctx.live.current.dispatch({ type: "select", selected: { kind: "element", id } });
}

function onMove(ctx, event) {
  const drag = ctx.drag;
  if (!drag || (event.pointerId && drag.pointer !== event.pointerId)) return;
  const dx = event.clientX - drag.cx;
  const dy = event.clientY - drag.cy;
  if (!drag.moved && Math.hypot(dx, dy) < 2) return;
  if (!drag.moved) ctx.live.current.dispatch({ type: "checkpoint" });
  drag.moved = true;
  const pose = drag.shift
    ? { x: drag.x, z: drag.z, yaw: drag.yaw + dx * 0.01 }
    : groundPose(ctx, event, drag);
  if (!pose) return;
  placeHold(ctx, pose);
  ctx.live.current.dispatch({ type: "pose", id: drag.id, pose, history: false });
}

function groundPose(ctx, event, drag) {
  aim(ctx.dom, event, ctx.pointer);
  ctx.raycaster.setFromCamera(ctx.pointer, ctx.camera);
  if (!drag.grab || ctx.raycaster.ray.intersectPlane(ground, ctx.hit) === null) return null;
  return { x: ctx.hit.x - drag.grab.x, z: ctx.hit.z - drag.grab.z, yaw: drag.yaw };
}

function onUp(ctx, event) {
  if (ctx.drag?.pointer != null && event.pointerId && ctx.drag.pointer !== event.pointerId) return;
  const drag = ctx.drag;
  const hold = ctx.bag.hold;
  ctx.drag = null;
  ctx.bag.controls.enabled = true;
  ctx.dom.style.cursor = "";
  if (!drag?.moved) return;
  if (hold && hold.id === drag.id) {
    ctx.live.current.dispatch({ type: "pose", id: drag.id, pose: { x: hold.x, z: hold.z, yaw: hold.yaw }, history: false });
  }
  requestAnimationFrame(() => ctx.live.current.dispatch({ type: "save" }));
}

export function bindEdit(dom, camera, bag, live) {
  const ctx = {
    dom, camera, bag, live, drag: null,
    raycaster: new THREE.Raycaster(),
    pointer: new THREE.Vector2(),
    hit: new THREE.Vector3()
  };
  const down = (event) => onDown(ctx, event);
  const move = (event) => onMove(ctx, event);
  const up = (event) => onUp(ctx, event);
  dom.addEventListener("pointerdown", down, true);
  window.addEventListener("pointermove", move);
  window.addEventListener("pointerup", up);
  return () => {
    dom.removeEventListener("pointerdown", down, true);
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", up);
  };
}
