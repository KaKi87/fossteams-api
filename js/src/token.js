import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

export const TokenSkype = 'skypetoken';
export const TokenBearer = 'Bearer';

function decodeBase64Url(segment) {
  const padded = segment.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(segment.length / 4) * 4, '=');
  return Buffer.from(padded, 'base64').toString('utf8');
}

export function parseJwt(raw = '') {
  const normalized = raw.trim();
  const [, payload = 'e30'] = normalized.split('.');
  let claims = {};
  try {
    claims = JSON.parse(decodeBase64Url(payload));
  } catch {
    claims = {};
  }
  return { raw: normalized, claims };
}

export class TeamsToken {
  constructor(inner, type = TokenBearer) {
    this.inner = typeof inner === 'string' ? parseJwt(inner) : inner;
    this.type = type;
  }
}

export function authString(token) {
  if (!token) {
    return '';
  }
  switch (token.type) {
    case TokenSkype:
      return `skypetoken=${token.inner.raw}`;
    case TokenBearer:
      return 'Bearer ' + token.inner.raw;
    default:
      return '';
  }
}

export async function getToken(tokenType, options = {}) {
  const env = options.env ?? process.env;
  let tokenStr = env[`MS_TEAMS_${tokenType.toUpperCase()}_TOKEN`] ?? '';
  if (!tokenStr) {
    const homeDir = options.homeDir ?? os.homedir();
    const tokenPath = path.join(homeDir, '.config', 'fossteams', `token-${tokenType}.jwt`);
    try {
      tokenStr = await fs.readFile(tokenPath, 'utf8');
    } catch (error) {
      throw new Error(`unable to open ${tokenPath}: ${error.message}`);
    }
  }
  return new TeamsToken(parseJwt(tokenStr), TokenBearer);
}

export async function getRootToken(options = {}) {
  return getToken('skype', options);
}

export async function getSkypeSpacesToken(options = {}) {
  return getToken('skype', options);
}

export async function getTeamsToken(options = {}) {
  return getToken('teams', options);
}

export async function getChatSvcAggToken(options = {}) {
  return getToken('chatsvcagg', options);
}
