import { useEffect, useRef } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/addons/controls/OrbitControls.js";
import { useStore } from "./store.jsx";

const SCALE = 0.012;

function heightOf(type) {
  if (type === "Queue" || type === "Hold") return 0.32;
  if (type === "Source" || type === "Sink") return 0.55;
  if (type === "Service" || type === "Seize" || type === "Delay") return 0.95;
  return 0.68;
}

function colorOf(type) {
  if (type === "Source") return 0x6d5efc;
  if (type === "Sink") return 0x1fa85a;
  if (type === "Queue" || type === "Hold") return 0x3b6fd8;
  if (type === "Delay" || type === "Service") return 0xf5a524;
  return 0x7c6bf2;
}

function groundPoint(x, y) {
  return [(x || 0) * SCALE, (y || 0) * SCALE];
}

function namePlate(text) {
  const canvas = document.createElement("canvas");
  canvas.width = 256;
  canvas.height = 64;
  const pen = canvas.getContext("2d");
  pen.fillStyle = "rgba(255,255,255,0.94)";
  pen.beginPath();
  pen.roundRect(6, 10, 244, 44, 12);
  pen.fill();
  pen.fillStyle = "#1c2130";
  pen.font = "600 28px sans-serif";
  pen.textAlign = "center";
  pen.textBaseline = "middle";
  pen.fillText(String(text || ""), 128, 32);
  const map = new THREE.CanvasTexture(canvas);
  const sprite = new THREE.Sprite(new THREE.SpriteMaterial({ map, transparent: true, depthTest: false }));
  sprite.scale.set(1.5, 0.38, 1);
  return sprite;
}

function stationMesh(element) {
  const w = Math.max(0.45, (element.w || 140) * SCALE);
  const d = Math.max(0.4, (element.h || 72) * SCALE);
  const h = heightOf(element.type);
  const mesh = new THREE.Mesh(
    new THREE.BoxGeometry(w, h, d),
    new THREE.MeshStandardMaterial({ color: colorOf(element.type), roughness: 0.48, metalness: 0.04 })
  );
  const [x, z] = groundPoint((element.x || 0) + (element.w || 140) / 2, (element.y || 0) + (element.h || 72) / 2);
  mesh.position.set(x, h / 2, z);
  mesh.castShadow = true;
  mesh.receiveShadow = true;
  const plate = namePlate(element.name);
  plate.position.set(0, h / 2 + 0.32, 0);
  mesh.add(plate);
  return mesh;
}

function tokenMesh() {
  const mesh = new THREE.Mesh(
    new THREE.BoxGeometry(0.24, 0.24, 0.24),
    new THREE.MeshStandardMaterial({ color: 0x1c2130, roughness: 0.35 })
  );
  mesh.castShadow = true;
  return mesh;
}

function release(object) {
  object.traverse((node) => {
    if (node.geometry) node.geometry.dispose();
    const materials = node.material ? (Array.isArray(node.material) ? node.material : [node.material]) : [];
    materials.forEach((material) => {
      Object.values(material).forEach((value) => { if (value?.isTexture) value.dispose(); });
      material.dispose();
    });
    node.dispose?.();
  });
}

function frameCamera(camera, controls, elements) {
  if (!elements.length) {
    controls.target.set(0, 0.4, 0);
    camera.position.set(6, 5.2, 8);
    return;
  }
  let x = 0;
  let z = 0;
  elements.forEach((element) => {
    const point = groundPoint((element.x || 0) + (element.w || 140) / 2, (element.y || 0) + (element.h || 72) / 2);
    x += point[0];
    z += point[1];
  });
  x /= elements.length;
  z /= elements.length;
  controls.target.set(x, 0.45, z);
  camera.position.set(x + 6.5, 5.4, z + 7.2);
}

export default function View3D() {
  const { state, agent } = useStore();
  const host = useRef(null);
  const live = useRef({ elements: [], movers: [] });
  live.current = { elements: agent?.elements || [], movers: state.runtime?.movers || [] };

  useEffect(() => {
    const el = host.current;
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.shadowMap.enabled = true;
    el.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xd7e0ec);
    scene.fog = new THREE.Fog(0xd7e0ec, 18, 42);
    const camera = new THREE.PerspectiveCamera(42, 1, 0.1, 80);
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.maxPolarAngle = Math.PI / 2.05;
    controls.minDistance = 2;
    controls.maxDistance = 28;

    const sun = new THREE.DirectionalLight(0xffffff, 2.4);
    sun.position.set(8, 12, 6);
    sun.castShadow = true;
    sun.shadow.mapSize.set(1024, 1024);
    sun.shadow.camera.left = -12;
    sun.shadow.camera.right = 12;
    sun.shadow.camera.top = 12;
    sun.shadow.camera.bottom = -12;
    scene.add(sun, new THREE.AmbientLight(0xffffff, 0.55));
    const floor = new THREE.Mesh(
      new THREE.PlaneGeometry(40, 40),
      new THREE.MeshStandardMaterial({ color: 0xe7eef6, roughness: 0.95 })
    );
    floor.rotation.x = -Math.PI / 2;
    floor.receiveShadow = true;
    scene.add(floor, new THREE.GridHelper(24, 24, 0xc5d0de, 0xd5deea));

    const stations = new Map();
    const tokens = new Map();
    let framed = false;
    const timer = new THREE.Timer();
    timer.connect(document);

    const fit = () => {
      const width = el.clientWidth || 1;
      const height = el.clientHeight || 1;
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.setSize(width, height);
    };
    fit();
    const watch = new ResizeObserver(fit);
    watch.observe(el);

    const sync = () => {
      const { elements, movers } = live.current;
      const seen = new Set(elements.map((item) => item.id));
      stations.forEach((mesh, id) => {
        if (seen.has(id)) return;
        scene.remove(mesh);
        release(mesh);
        stations.delete(id);
      });
      elements.forEach((element) => {
        if (stations.has(element.id)) return;
        const mesh = stationMesh(element);
        stations.set(element.id, mesh);
        scene.add(mesh);
      });
      if (!framed && elements.length) {
        frameCamera(camera, controls, elements);
        framed = true;
      }
      const moving = new Set(movers.map((item) => item.id));
      tokens.forEach((mesh, id) => {
        if (moving.has(id)) return;
        scene.remove(mesh);
        release(mesh);
        tokens.delete(id);
      });
      movers.forEach((mover) => {
        let mesh = tokens.get(mover.id);
        if (!mesh) {
          mesh = tokenMesh();
          tokens.set(mover.id, mesh);
          scene.add(mesh);
        }
        const [x, z] = groundPoint(mover.x, mover.y);
        mesh.position.set(x, 0.16 + (mover.z || 0) * 0.006, z);
      });
    };

    renderer.setAnimationLoop(() => {
      timer.update();
      sync();
      controls.update(timer.getDelta());
      renderer.render(scene, camera);
    });

    return () => {
      renderer.setAnimationLoop(null);
      watch.disconnect();
      controls.dispose();
      timer.dispose();
      release(scene);
      renderer.dispose();
      renderer.forceContextLoss();
      renderer.domElement.remove();
    };
  }, []);

  const empty = (agent?.elements || []).length === 0;
  return <div className="view3d" ref={host}>{empty && <p className="empty-note">这一页还没有模块，三维里没有东西可看。</p>}</div>;
}
