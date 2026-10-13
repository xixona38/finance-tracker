// Local frontend server. API requests stay same-origin and are forwarded to Go.
import http from 'node:http';
import { readFile } from 'node:fs/promises';

const port = Number(process.env.FRONTEND_PORT || 5173);
const target = new URL(process.env.API_TARGET || 'http://127.0.0.1:8080');
if (target.protocol !== 'http:') throw new Error('Local API_TARGET must use http://');
const files = new Map([
  ['/', ['index.html', 'text/html; charset=utf-8']],
  ['/index.html', ['index.html', 'text/html; charset=utf-8']],
  ['/app.js', ['app.js', 'text/javascript; charset=utf-8']],
  ['/lib.js', ['lib.js', 'text/javascript; charset=utf-8']],
  ['/style.css', ['style.css', 'text/css; charset=utf-8']],
  ['/favicon.svg', ['favicon.svg', 'image/svg+xml']],
]);

const server = http.createServer(async (req, res) => {
  const host = req.headers.host || '';
  if (![ `localhost:${port}`, `127.0.0.1:${port}` ].includes(host)) {
    res.writeHead(403).end('Invalid host');
    return;
  }
  const path = new URL(req.url, `http://${host}`).pathname;
  if (path.startsWith('/api/')) {
    // Preserve the browser-facing Host and Origin so Go can enforce its CSRF checks.
    const headers = { ...req.headers, host, connection: 'close' };
    delete headers['forwarded'];
    for (const key of Object.keys(headers)) if (key.startsWith('x-forwarded-')) delete headers[key];
    const upstream = http.request({
      hostname: target.hostname, port: target.port || 80,
      path: req.url, method: req.method, headers, timeout: 30000,
    }, incoming => {
      res.writeHead(incoming.statusCode, incoming.headers);
      incoming.pipe(res);
      incoming.on('error', () => res.destroy());
    });
    upstream.on('timeout', () => upstream.destroy(new Error('API timeout')));
    upstream.on('error', () => {
      if (res.headersSent) return res.destroy();
      res.writeHead(502, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'API unavailable' }));
    });
    res.on('close', () => upstream.destroy());
    req.pipe(upstream);
    return;
  }
  if (!['GET', 'HEAD'].includes(req.method)) return res.writeHead(405).end();
  const entry = files.get(path);
  if (!entry) return res.writeHead(404).end('Not found');
  try {
    const body = await readFile(new URL(entry[0], import.meta.url));
    res.writeHead(200, {
      'Content-Type': entry[1], 'Cache-Control': 'no-store',
      'X-Content-Type-Options': 'nosniff', 'Referrer-Policy': 'same-origin',
      'Content-Security-Policy': "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'",
    });
    res.end(req.method === 'HEAD' ? undefined : body);
  } catch {
    res.writeHead(500).end('Unable to read frontend asset');
  }
});
server.listen(port, '127.0.0.1', () => console.log(`Frontend: http://localhost:${port}\nGo API: ${target.origin}`));
server.on('error', error => { console.error(error.message); process.exitCode = 1; });
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => server.close());
