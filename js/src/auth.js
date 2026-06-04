import { AuthzError, GenericHTTPError } from './errors.js';
import { TeamsToken, TokenBearer, parseJwt } from './token.js';

export const TEAMS_API_ENDPOINT = 'https://teams.microsoft.com/api';
export const Emea = 'emea';
export const Emea1 = 'emea01';
export const AuthzRefresh = 'TokenRefresh';

export class AuthClient {
  constructor(fetchImpl = globalThis.fetch) {
    this.fetch = fetchImpl;
  }

  async authz(token, authzType) {
    const response = await this.fetch(`${TEAMS_API_ENDPOINT}/authsvc/v1.0/authz`, {
      method: 'POST',
      headers: {
        'ms-teams-authz-type': authzType,
        Authorization: 'Bearer ' + token.inner.raw,
      },
    });
    const bodyText = await response.text();
    if (response.status !== 200) {
      try {
        throw new AuthzError(JSON.parse(bodyText));
      } catch (error) {
        if (error instanceof AuthzError) {
          throw error;
        }
        throw new GenericHTTPError(200, response.status, bodyText || null);
      }
    }
    const payload = JSON.parse(bodyText);
    return new TeamsToken(parseJwt(payload.tokens.skypeToken), TokenBearer);
  }
}
