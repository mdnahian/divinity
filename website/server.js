import { createServer, request as httpRequest } from 'node:http';
import { readFile } from 'node:fs/promises';
import { resolve, extname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dir = fileURLToPath(new URL('.', import.meta.url));
const sharedDir = resolve(__dir, '..', 'shared');
const PORT = parseInt(process.env.LANDING_PORT || '3001', 10);
const API_TARGET = process.env.API_TARGET || 'http://localhost:8080';

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.css':  'text/css; charset=utf-8',
  '.js':   'application/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png':  'image/png',
  '.svg':  'image/svg+xml',
};

function proxy(req, res) {
  const url = new URL(API_TARGET);
  const opts = {
    hostname: url.hostname,
    port: url.port,
    path: req.url,
    method: req.method,
    headers: { ...req.headers, host: url.host },
  };

  const proxyReq = httpRequest(opts, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers);
    proxyRes.pipe(res);
  });

  proxyReq.on('error', () => {
    res.writeHead(502, { 'Content-Type': 'application/json' });
    res.end('{"error":"backend unavailable"}');
  });

  req.pipe(proxyReq);
}

async function serve(req, res) {
  // Proxy API and health requests to the Go backend
  if (req.url.startsWith('/api/') || req.url.startsWith('/health')) {
    return proxy(req, res);
  }

  let path = req.url.split('?')[0];
  if (path === '/') path = '/index.html';

  const file = path === '/theme.css'
    ? resolve(sharedDir, 'theme.css')
    : resolve(__dir, path.slice(1));

  try {
    const data = await readFile(file);
    const mime = MIME[extname(file)] || 'application/octet-stream';
    res.writeHead(200, { 'Content-Type': mime });
    res.end(data);
  } catch {
    res.writeHead(404, { 'Content-Type': 'text/plain' });
    res.end('Not found');
  }
}

createServer(serve).listen(PORT, () => {
  console.log(`Landing page: http://localhost:${PORT}`);
  console.log(`API proxy:    ${API_TARGET}`);
});
