import { useEffect, useLayoutEffect, useRef } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/addons/controls/OrbitControls.js";
import { RoomEnvironment } from "three/addons/environments/RoomEnvironment.js";
import Canvas from "./canvas.jsx";
import { bindEdit, attachPick } from "./edit3d.js";
import { spot, stampOf } from "./meshkit.js";
import { buildModel, carton, release, tickModels } from "./models3d.js";
import { car, forklift, person, train } from "./yard3d.js";
import { useStore } from "./store.jsx";

function frameCamera(camera, controls, elements) {
  if (!elements.length) {
    controls.target.set(0, 0.8, 0);
    camera.position.set(4.2, 3.2, 5.4);
    return;
  }
  let minX = Infinity;
  let maxX = -Infinity;
  let minZ = Infinity;
  let maxZ = -Infinity;
  elements.forEach((element) => {
    const at = spot(element);
    minX = Math.min(minX, at.x);
    maxX = Math.max(maxX, at.x);
    minZ = Math.min(minZ, at.z);
    maxZ = Math.max(maxZ, at.z);
  });
  const dist = Math.max(maxX - minX, maxZ - minZ, 2) * 0.85 + 3.2;
  controls.target.set((minX + maxX) / 2, 0.85, (minZ + maxZ) / 2);
  camera.position.set(controls.target.x + dist * 0.72, dist * 0.48, controls.target.z + dist * 0.78);
}

function makeScene(renderer) {
  const scene = new THREE.Scene();
  scene.background = new THREE.Color(0xf4f6f8);
  scene.fog = new THREE.Fog(0xf4f6f8, 40, 90);
  scene.environmentIntensity = 0.35;
  const pmrem = new THREE.PMREMGenerator(renderer);
  const room = new RoomEnvironment();
  scene.environment = pmrem.fromScene(room, 0.04).texture;
  room.dispose();
  pmrem.dispose();
  const sun = new THREE.DirectionalLight(0xfff8ee, 2.6);
  sun.position.set(9, 14, 7);
  sun.castShadow = true;
  sun.shadow.mapSize.set(1024, 1024);
  sun.shadow.camera.left = -24;
  sun.shadow.camera.right = 24;
  sun.shadow.camera.top = 24;
  sun.shadow.camera.bottom = -24;
  sun.shadow.bias = -0.0004;
  scene.add(sun, new THREE.HemisphereLight(0xffffff, 0xc5ccc4, 0.85));
  const floor = new THREE.Mesh(
    new THREE.PlaneGeometry(80, 80),
    new THREE.MeshStandardMaterial({ color: 0xe4ebf3, roughness: 0.96, metalness: 0 })
  );
  floor.rotation.x = -Math.PI / 2;
  floor.receiveShadow = true;
  const grid = new THREE.GridHelper(40, 40, 0xc5d0de, 0xd5deea);
  grid.position.y = 0.01;
  scene.add(floor, grid);
  return scene;
}

function watchSize(el, camera, renderer) {
  const fit = () => {
    camera.aspect = (el.clientWidth || 1) / (el.clientHeight || 1);
    camera.updateProjectionMatrix();
    renderer.setSize(el.clientWidth || 1, el.clientHeight || 1);
  };
  fit();
  const watch = new ResizeObserver(fit);
  watch.observe(el);
  return watch;
}

function syncWorld(scene, live, bag) {
  const { elements, movers } = live.current;
  const rider = live.current.editing ? new Set() : riding(movers);
  const seen = new Set(elements.map((item) => item.id));
  bag.stations.forEach((mesh, id) => {
    if (seen.has(id)) return;
    scene.remove(mesh);
    release(mesh);
    bag.stations.delete(id);
  });
  elements.forEach((element) => {
    let mesh = bag.stations.get(element.id);
    if (mesh && mesh.userData.stamp !== stampOf(element)) {
      scene.remove(mesh);
      release(mesh);
      bag.stations.delete(element.id);
      mesh = null;
    }
    if (!mesh) {
      mesh = buildModel(element);
      bag.stations.set(element.id, mesh);
      scene.add(mesh);
      attachPick(mesh);
    }
    const held = bag.hold?.id === element.id ? bag.hold : null;
    const at = held || spot(element);
    mesh.position.set(at.x, at.y || 0, at.z);
    mesh.rotation.y = at.yaw;
    mesh.visible = live.current.editing || !rider.has(element.type);
    if (held && Math.hypot(spot(element).x - held.x, spot(element).z - held.z) < 0.02) bag.hold = null;
  });
  const chosen = bag.stations.get(live.current.selectedId);
  bag.ring.visible = !!chosen;
  if (chosen) bag.ring.position.set(chosen.position.x, 0.03, chosen.position.z);
  if (!bag.framed && elements.length) {
    frameCamera(bag.camera, bag.controls, elements);
    bag.framed = true;
  }
  const moving = new Set(movers.map((item) => item.id));
  bag.tokens.forEach((mesh, id) => {
    if (moving.has(id)) return;
    scene.remove(mesh);
    release(mesh);
    bag.tokens.delete(id);
  });
  movers.forEach((mover) => {
    let mesh = bag.tokens.get(mover.id);
    if (!mesh || mesh.userData.kind !== mover.kind) {
      if (mesh) {
        scene.remove(mesh);
        release(mesh);
      }
      mesh = tokenBody(mover.kind);
      mesh.userData.kind = mover.kind;
      bag.tokens.set(mover.id, mesh);
      scene.add(mesh);
    }
    const at = played(mover, live.current.runtime, live.current.unit, live.current.speed);
    mesh.position.set(at.x, at.y, at.z);
    mesh.rotation.y = at.heading;
    if (mover.kind === "person") swingLegs(mesh);
  });
}

function played(mover, runtime, unit, speed) {
  const pace = speed === 0 ? 60 : speed === 10 ? 10 : 1;
  const wall = runtime?.stamp ? Math.max(0, (Date.now() - runtime.stamp) / 1000) : 0;
  const perSecond = unit === "秒" ? 1 : unit === "小时" ? 1 / 3600 : 1 / 60;
  const now = (runtime?.time || 0) + wall * pace * perSecond;
  const span = (mover.t1 || 0) - (mover.t0 || 0);
  let u = span > 1e-6 ? (now - mover.t0) / span : 1;
  if (u < 0) u = 0;
  if (u > 1) u = 1;
  const fromX = Number.isFinite(mover.fromX) ? mover.fromX : mover.x || 0;
  const fromY = Number.isFinite(mover.fromY) ? mover.fromY : mover.y || 0;
  const fromZ = Number.isFinite(mover.fromZ) ? mover.fromZ : mover.z || 0;
  const toX = Number.isFinite(mover.toX) ? mover.toX : fromX;
  const toY = Number.isFinite(mover.toY) ? mover.toY : fromY;
  const toZ = Number.isFinite(mover.toZ) ? mover.toZ : fromZ;
  return {
    x: fromX + (toX - fromX) * u,
    y: fromY + (toY - fromY) * u,
    z: fromZ + (toZ - fromZ) * u,
    heading: Number.isFinite(mover.heading) ? mover.heading : 0
  };
}

function riding(movers) {
  const live = new Set(movers.map((item) => item.kind));
  const types = new Set();
  if (live.has("car")) ["CarSource", "CarEnter", "CarMoveTo", "CarDispose"].forEach((type) => types.add(type));
  if (live.has("train")) ["TrainSource", "TrainEnter", "TrainMoveTo", "TrainDispose", "TrainExit"].forEach((type) => types.add(type));
  if (live.has("person")) ["PedSource", "PedEnter", "PedGoTo", "PedWait", "PedService", "PedSink", "PedExit"].forEach((type) => types.add(type));
  if (live.has("fork")) ["TransporterFleet", "MoveByTransporter", "SeizeTransporter", "ReleaseTransporter"].forEach((type) => types.add(type));
  return types;
}

function tokenBody(kind) {
  if (kind === "person") return person();
  if (kind === "car") return car();
  if (kind === "train") return train();
  if (kind === "fork") return forklift();
  return carton();
}

function swingLegs(mesh) {
  const step = Math.sin(performance.now() / 180) * 0.45;
  if (mesh.children[0]) mesh.children[0].rotation.x = step;
  if (mesh.children[1]) mesh.children[1].rotation.x = -step;
}

function selectionRing() {
  const ring = new THREE.Mesh(
    new THREE.RingGeometry(0.42, 0.5, 40),
    new THREE.MeshBasicMaterial({ color: 0x6d5efc, side: THREE.DoubleSide, depthWrite: false })
  );
  ring.rotation.x = -Math.PI / 2;
  ring.position.y = 0.03;
  ring.visible = false;
  return ring;
}

function mount(el, live) {
  const renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.outputColorSpace = THREE.SRGBColorSpace;
  renderer.toneMapping = THREE.ACESFilmicToneMapping;
  renderer.toneMappingExposure = 1.05;
  renderer.shadowMap.enabled = true;
  renderer.setSize(el.clientWidth || 640, el.clientHeight || 480);
  el.appendChild(renderer.domElement);
  const scene = makeScene(renderer);
  const camera = new THREE.PerspectiveCamera(40, 1, 0.08, 120);
  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.maxPolarAngle = Math.PI / 2.08;
  controls.minDistance = 1.1;
  controls.maxDistance = 48;
  const ring = selectionRing();
  scene.add(ring);
  const bag = { stations: new Map(), tokens: new Map(), camera, controls, framed: false, ring };
  const timer = new THREE.Timer();
  timer.connect(document);
  const watch = watchSize(el, camera, renderer);
  const unbind = bindEdit(renderer.domElement, camera, bag, live);
  renderer.setAnimationLoop(() => {
    timer.update();
    syncWorld(scene, live, bag);
    if (!live.current.editing) tickModels(scene, timer.getElapsed());
    controls.update(timer.getDelta());
    renderer.render(scene, camera);
  });
  return () => {
    renderer.setAnimationLoop(null);
    unbind();
    watch.disconnect();
    controls.dispose();
    timer.dispose();
    scene.environment?.dispose();
    release(scene);
    renderer.dispose();
    renderer.forceContextLoss();
    renderer.domElement.remove();
  };
}

export default function View3D() {
  const { state, agent, project, dispatch } = useStore();
  const host = useRef(null);
  const live = useRef({ elements: [], movers: [], editing: false, selectedId: null, dispatch });
  const editing = state.ui.mode === "edit";
  const selectedId = state.ui.selected?.kind === "element" ? state.ui.selected.id : null;
  live.current = {
    elements: agent?.elements || [],
    movers: editing ? [] : (state.runtime?.movers || []),
    runtime: state.runtime,
    unit: project?.unit,
    speed: state.ui.speed,
    editing,
    selectedId,
    dispatch
  };
  useEffect(() => mount(host.current, live), []);
  const empty = (agent?.elements || []).length === 0;
  return <div className="view3d" ref={host}>{empty && <p className="empty-note">这一页还没有模块，三维里没有东西可看。</p>}</div>;
}

export function MapStage() {
  const { state, dispatch, agent } = useStore();
  const three = state.ui.view === "3d";
  const loose = (agent?.elements || []).some((item) => !item.pose || !Number.isFinite(item.pose.x));
  useLayoutEffect(() => {
    if (loose) dispatch({ type: "freezePoses" });
  }, [loose, dispatch]);
  return (
    <div className="mapstage">
      <div className="mapbar">
        <button type="button" aria-pressed={!three} onClick={() => dispatch({ type: "view", value: "2d" })}>2D</button>
        <button type="button" aria-pressed={three} onClick={() => dispatch({ type: "view", value: "3d" })}>3D</button>
        {three && <span>拖模型布置位置，按住 Shift 再拖是旋转</span>}
      </div>
      {three ? <View3D /> : <Canvas />}
    </div>
  );
}
