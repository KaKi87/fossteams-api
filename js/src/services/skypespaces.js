import { GenericHTTPError } from '../errors.js';

export const SkypeSpacesEndpoint = 'https://api.teams.skype.com';

export class SkypeSpaceSvc {
  constructor(token, options = {}) {
    this.token = token;
    this.endpoint = new URL(options.endpoint ?? SkypeSpacesEndpoint);
    this.fetch = options.fetch ?? globalThis.fetch;
  }

  authenticateRequest(headers = new Headers()) {
    headers.set('Authorization', 'Bearer ' + this.token.inner.raw);
    return headers;
  }

  async get(pathname) {
    const response = await this.fetch(new URL(pathname, this.endpoint).toString(), {
      method: 'GET',
      headers: this.authenticateRequest(),
    });
    return response;
  }

  async getTenants() {
    const response = await this.get('/beta/users/tenants');
    const text = await response.text();
    if (response.status !== 200) {
      throw new GenericHTTPError(200, response.status, text || null);
    }
    return JSON.parse(text);
  }
}

export function NewSkypeSpaceService(token, options = {}) {
  return new SkypeSpaceSvc(token, options);
}
