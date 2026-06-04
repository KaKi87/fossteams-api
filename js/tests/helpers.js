import { promises as fs } from 'node:fs';
import path from 'node:path';

export async function readFixture(relativePath) {
  const fixturePath = path.join(import.meta.dir, '..', '..', relativePath);
  return fs.readFile(fixturePath, 'utf8');
}

export function makeJwt(claims = {}) {
  const header = Buffer.from(JSON.stringify({ alg: 'none' })).toString('base64url');
  const payload = Buffer.from(JSON.stringify(claims)).toString('base64url');
  return `${header}.${payload}.signature`;
}

export function createFetch(handler) {
  return async (url, init = {}) => {
    const request = {
      url: typeof url === 'string' ? url : url.toString(),
      method: init.method ?? 'GET',
      headers: new Headers(init.headers ?? {}),
      body: init.body,
    };
    return handler(request);
  };
}

export function jsonResponse(value, status = 200) {
  return new Response(typeof value === 'string' ? value : JSON.stringify(value), {
    status,
    headers: { 'content-type': 'application/json' },
  });
}
