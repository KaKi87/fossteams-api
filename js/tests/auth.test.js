import { describe, expect, test } from 'bun:test';
import { promises as fs } from 'node:fs';
import path from 'node:path';
import { AuthClient, AuthzRefresh } from '../src/auth.js';
import { AuthzError, GenericHTTPError } from '../src/errors.js';
import { authString, getTeamsToken, TeamsToken, TokenBearer, TokenSkype } from '../src/token.js';
import { createFetch, jsonResponse, makeJwt } from './helpers.js';

describe('token helpers', () => {
  test('formats auth strings', () => {
    const token = new TeamsToken(makeJwt({ email: 'user@example.com' }), TokenBearer);
    expect(authString(token)).toBe('Bearer ' + token.inner.raw);
    token.type = TokenSkype;
    expect(authString(token)).toBe(`skypetoken=${token.inner.raw}`);
    expect(authString(null)).toBe('');
  });

  test('loads token from filesystem', async () => {
    const homeDir = await fs.mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'teams-js-'));
    const tokenDir = path.join(homeDir, '.config', 'fossteams');
    await fs.mkdir(tokenDir, { recursive: true });
    const raw = makeJwt({ upn: 'user@example.com' });
    await fs.writeFile(path.join(tokenDir, 'token-teams.jwt'), raw, 'utf8');

    const token = await getTeamsToken({ env: {}, homeDir });
    expect(token.inner.raw).toBe(raw);
  });
});

describe('auth client', () => {
  test('refreshes a token', async () => {
    const refreshed = makeJwt({ email: 'user@example.com' });
    const fetch = createFetch((request) => {
      expect(request.method).toBe('POST');
      expect(request.url).toBe('https://teams.microsoft.com/api/authsvc/v1.0/authz');
      expect(request.headers.get('ms-teams-authz-type')).toBe(AuthzRefresh);
      expect(request.headers.get('Authorization')).toBe('Bearer ' + 'root-token');
      return jsonResponse({ tokens: { skypeToken: refreshed, expiresIn: 3600 } });
    });

    const client = new AuthClient(fetch);
    const token = await client.authz(new TeamsToken({ raw: 'root-token', claims: { email: 'user@example.com' } }, TokenBearer), AuthzRefresh);
    expect(token.inner.raw).toBe(refreshed);
    expect(token.type).toBe(TokenBearer);
  });

  test('returns typed authz errors', async () => {
    const client = new AuthClient(createFetch(() => jsonResponse({ errorCode: 'GuestUserNotRedeemed', message: 'select a tenant first' }, 401)));
    await expect(client.authz(new TeamsToken({ raw: 'root-token', claims: {} }, TokenBearer), AuthzRefresh)).rejects.toBeInstanceOf(AuthzError);
  });

  test('falls back to generic http errors', async () => {
    const client = new AuthClient(createFetch(() => new Response('boom', { status: 500 })));
    await expect(client.authz(new TeamsToken({ raw: 'root-token', claims: {} }, TokenBearer), AuthzRefresh)).rejects.toBeInstanceOf(GenericHTTPError);
  });
});
