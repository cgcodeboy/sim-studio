import { createServer } from "node:http";
import { request as proxy } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";

const root = join(import.meta.dirname, "dist");
const origin = new URL(process.env.API_ORIGIN || "http://api:8080");
const types = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".json": "application/json" };

createServer((req, res) => {
  if (req.url.startsWith("/api")) {
    const out = proxy({ hostname: origin.hostname, port: origin.port, path: req.url, method: req.method, headers: req.headers }, (upstream) => {
      res.writeHead(upstream.statusCode, upstream.headers);
      upstream.pipe(res);
    });
    out.on("error", () => {
      res.writeHead(502);
      res.end("连不上运行服务");
    });
    req.pipe(out);
    return;
  }
  const path = normalize(decodeURIComponent(req.url.split("?")[0]));
  const file = join(root, path === "/" ? "index.html" : path);
  if (!file.startsWith(root)) {
    res.writeHead(404);
    res.end();
    return;
  }
  readFile(file).then(
    (body) => {
      res.writeHead(200, { "Content-Type": types[extname(file)] || "application/octet-stream" });
      res.end(body);
    },
    () => readFile(join(root, "index.html")).then((body) => {
      res.writeHead(200, { "Content-Type": "text/html" });
      res.end(body);
    })
  );
}).listen(8080);
