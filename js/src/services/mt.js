import { GenericHTTPError } from '../errors.js';
import { authString } from '../token.js';
import { parseTenantsResponse, parseUserResponse } from '../models.js';
import { readResponseText, toByteArray } from '../utils.js';

export const MiddleTier = 'https://teams.microsoft.com/api/mt/';

export function getTokenEmail(token) {
  if (!token) {
    throw new Error('invalid token provided (nil)');
  }
  if (token.inner?.claims?.email) {
    return token.inner.claims.email;
  }
  if (token.inner?.claims?.upn) {
    return token.inner.claims.upn;
  }
  throw new Error("JWT doesn't contain email nor upn");
}

export class Service {
  constructor(region, token, teamsToken, options = {}) {
    this.region = region;
    this.token = token;
    this.teamsToken = teamsToken;
    this.fetch = options.fetch ?? globalThis.fetch;
    this.middleTierUrl = new URL(options.middleTierUrl ?? MiddleTier);
    this._debugSave = false;
    this._debugDisallowUnknownFields = false;
  }

  debugSave(flag) {
    this._debugSave = flag;
  }

  debugDisallowUnknownFields(flag) {
    this._debugDisallowUnknownFields = flag;
  }

  authenticatedRequest(method, url, body, headers = {}) {
    const finalHeaders = new Headers(headers);
    finalHeaders.set('Authorization', authString(this.token));
    return { method, headers: finalHeaders, body };
  }

  cookieRequest(method, url, body, headers = {}) {
    const finalHeaders = new Headers(headers);
    finalHeaders.set('Cookie', `TSAUTHCOOKIE=${this.teamsToken.inner.raw}`);
    return { method, headers: finalHeaders, body };
  }

  getEndpoint(value) {
    const normalized = value.startsWith('/') ? value.slice(1) : value;
    return new URL(`${this.region}/beta/${normalized}`, this.middleTierUrl);
  }

  async getTenants() {
    const endpointUrl = this.getEndpoint('/users/tenants');
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    const text = await readResponseText(response, this._debugSave);
    if (response.status !== 200) {
      throw new Error(`invalid status code ${response.status}: resp = ${text}`);
    }
    return parseTenantsResponse(text);
  }

  async getUser(email) {
    const endpointUrl = this.getEndpoint(`/users/${encodeURIComponent(email)}/`);
    endpointUrl.searchParams.set('throwIfNotFound', 'false');
    endpointUrl.searchParams.set('isMailAddress', 'true');
    endpointUrl.searchParams.set('enableGuest', 'true');
    endpointUrl.searchParams.set('includeIBBarredUsers', 'true');
    endpointUrl.searchParams.set('skypeTeamsInfo', 'true');
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    if (response.status !== 200) {
      throw new GenericHTTPError(200, response.status, null);
    }
    return parseUserResponse(await readResponseText(response, this._debugSave)).value;
  }

  async getMe() {
    return this.getUser(getTokenEmail(this.token));
  }

  async fetchShortProfile(...mris) {
    const values = Array.isArray(mris[0]) ? mris[0] : mris;
    const endpointUrl = this.getEndpoint('/users/fetchShortProfile');
    endpointUrl.searchParams.set('isMailAddress', 'false');
    endpointUrl.searchParams.set('enableGuest', 'true');
    endpointUrl.searchParams.set('includeIBBarredUsers', 'false');
    endpointUrl.searchParams.set('skypeTeamsInfo', 'true');
    const response = await this.fetch(
      endpointUrl.toString(),
      this.authenticatedRequest('POST', endpointUrl.toString(), JSON.stringify(values), { 'Content-Type': 'application/json' }),
    );
    if (response.status !== 200) {
      throw new GenericHTTPError(200, response.status, null);
    }
    return parseUserResponse(await readResponseText(response, this._debugSave)).value;
  }

  async getProfilePicture(emailOrId) {
    const endpointUrl = this.getEndpoint(`/users/${encodeURIComponent(emailOrId)}/profilepicture?displayname=aaa`);
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    const text = await readResponseText(response, this._debugSave, 'txt');
    if (response.status !== 200) {
      throw new Error(`invalid status code ${response.status}: resp = ${text}`);
    }
    return toByteArray(Buffer.from(text, 'base64'));
  }

  async getTeamsProfilePicture(emailOrId) {
    const endpointUrl = this.getEndpoint(`/teams/${encodeURIComponent(emailOrId)}/profilepicturev2`);
    const response = await this.fetch(endpointUrl.toString(), this.cookieRequest('GET', endpointUrl.toString(), undefined));
    const bytes = Buffer.from(await response.arrayBuffer());
    if (response.status !== 200) {
      throw new Error(`invalid status code ${response.status}: resp = ${bytes.toString('utf8')}`);
    }
    return toByteArray(bytes);
  }

  async getVerifiedDomains() {
    const endpointUrl = this.getEndpoint('/tenant/verifiedDomains');
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    const text = await readResponseText(response, this._debugSave);
    if (response.status !== 200) {
      throw new Error(`invalid status code ${response.status}: resp = ${text}`);
    }
    return JSON.parse(text);
  }
}
