export class GenericHTTPError extends Error {
  constructor(expectedStatusCode, statusCode, body) {
    const suffix = body == null ? '' : `: ${typeof body === 'string' ? body : Buffer.from(body).toString('utf8')}`;
    super(`invalid status code returned: ${expectedStatusCode} expected but ${statusCode} returned${suffix}`);
    this.name = 'GenericHTTPError';
    this.expectedStatusCode = expectedStatusCode;
    this.statusCode = statusCode;
    this.body = body ?? null;
  }
}

export const GuestUserNotRedeemed = 'GuestUserNotRedeemed';

export class AuthzError extends Error {
  constructor({ errorCode, message }) {
    super(`AuthzError: ${errorCode} - ${message}`);
    this.name = 'AuthzError';
    this.errorCode = errorCode;
    this.message = message;
  }
}
