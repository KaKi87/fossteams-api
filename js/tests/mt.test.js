import { describe, expect, test } from 'bun:test';
import { Service, getTokenEmail } from '../src/services/mt.js';
import { TeamsToken, TokenBearer } from '../src/token.js';
import { createFetch, jsonResponse, readFixture } from './helpers.js';

describe('MT service', () => {
  test('loads users and current user', async () => {
    const fixture = await readFixture('resources/mt/user/user-1.json');
    const service = new Service(
      'emea',
      new TeamsToken({ raw: 'root-token', claims: { email: 'user@example.com' } }, TokenBearer),
      new TeamsToken({ raw: 'teams-token', claims: {} }, TokenBearer),
      {
        fetch: createFetch((request) => {
          expect(request.url).toBe('https://teams.microsoft.com/api/mt/emea/beta/users/user%40example.com/?throwIfNotFound=false&isMailAddress=true&enableGuest=true&includeIBBarredUsers=true&skypeTeamsInfo=true');
          expect(request.headers.get('Authorization')).toBe('Bearer ' + 'root-token');
          return jsonResponse(fixture);
        }),
      },
    );

    const user = await service.getUser('user@example.com');
    expect(user.displayName).toBe('Denys Vitali');

    const me = await service.getMe();
    expect(me.email).toBe('teams-cli@outlook.com');
  });

  test('loads short profiles, tenants and verified domains', async () => {
    const tenantsFixture = await readFixture('resources/mt/tenants/tenants-1.json');
    let callCount = 0;
    const service = new Service(
      'emea',
      new TeamsToken({ raw: 'root-token', claims: { email: 'user@example.com' } }, TokenBearer),
      new TeamsToken({ raw: 'teams-token', claims: {} }, TokenBearer),
      {
        fetch: createFetch(async (request) => {
          callCount += 1;
          if (callCount === 1) {
            expect(request.method).toBe('POST');
            expect(request.body).toBe('["8:orgid:user-1"]');
            return jsonResponse({ value: [{ displayName: 'One', email: 'one@example.com', givenName: 'One', surname: 'User', isShortProfile: true, jobTitle: '', objectId: '1', tenantName: 'Tenant', type: 'ADUser', userLocation: 'Remote', userPrincipalName: 'one@example.com' }], type: 'Users' });
          }
          if (callCount === 2) {
            return jsonResponse(tenantsFixture);
          }
          return jsonResponse([{ Name: 'example.com' }]);
        }),
      },
    );

    const users = await service.fetchShortProfile('8:orgid:user-1');
    expect(users[0].email).toBe('one@example.com');
    const tenants = await service.getTenants();
    expect(tenants[0].tenantName).toBe('FossTeams');
    const domains = await service.getVerifiedDomains();
    expect(domains[0].Name).toBe('example.com');
  });

  test('loads picture bytes and token emails', async () => {
    let callCount = 0;
    const service = new Service(
      'emea',
      new TeamsToken({ raw: 'root-token', claims: { upn: 'user@example.com' } }, TokenBearer),
      new TeamsToken({ raw: 'teams-token', claims: {} }, TokenBearer),
      {
        fetch: createFetch((request) => {
          callCount += 1;
          if (callCount === 1) {
            expect(request.url).toContain('/users/user%40example.com/profilepicture?displayname=aaa');
            return new Response(Buffer.from('jpeg-bytes').toString('base64'));
          }
          expect(request.headers.get('Cookie')).toBe('TSAUTHCOOKIE=teams-token');
          return new Response('jpeg-bytes');
        }),
      },
    );

    expect(Buffer.from(await service.getProfilePicture('user@example.com')).toString()).toBe('jpeg-bytes');
    expect(Buffer.from(await service.getTeamsProfilePicture('user@example.com')).toString()).toBe('jpeg-bytes');
    expect(getTokenEmail(service.token)).toBe('user@example.com');
  });
});
