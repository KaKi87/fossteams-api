import { describe, expect, test } from 'bun:test';
import { New } from '../src/client.js';
import { createFetch, jsonResponse, makeJwt, readFixture } from './helpers.js';

describe('TeamsClient', () => {
  test('bootstraps from tokens and exposes wrapper methods', async () => {
    const conversationsFixture = await readFixture('resources/chatsvcagg/conversations/conversations-1.json');
    const messagesFixture = await readFixture('resources/chatsvcagg/messages/messages-1.json');
    const userFixture = await readFixture('resources/mt/user/user-1.json');
    const tenantsFixture = await readFixture('resources/mt/tenants/tenants-1.json');
    const refreshed = makeJwt({ email: 'user@example.com' });

    const client = await New({
      env: {
        MS_TEAMS_SKYPE_TOKEN: makeJwt({ email: 'user@example.com' }),
        MS_TEAMS_CHATSVCAGG_TOKEN: makeJwt({ email: 'user@example.com' }),
        MS_TEAMS_TEAMS_TOKEN: makeJwt({ email: 'user@example.com' }),
      },
      fetch: createFetch((request) => {
        if (request.url.endsWith('/authsvc/v1.0/authz')) {
          return jsonResponse({ tokens: { skypeToken: refreshed, expiresIn: 3600 } });
        }
        if (request.url === 'https://teams.microsoft.com/api/csa/api/v1/teams/users/me?isPrefetch=false&enableMembershipSummary=true') {
          return jsonResponse(conversationsFixture);
        }
        if (request.url.includes('/messages?') && request.url.includes('view=msnp24Equivalent%7CsupportsMessageProperties')) {
          return jsonResponse(messagesFixture);
        }
        if (request.url.endsWith('/users/user%40example.com/?throwIfNotFound=false&isMailAddress=true&enableGuest=true&includeIBBarredUsers=true&skypeTeamsInfo=true')) {
          return jsonResponse(userFixture);
        }
        if (request.url.endsWith('/users/tenants')) {
          return jsonResponse(tenantsFixture);
        }
        if (request.url.includes('/fetchShortProfile?')) {
          return jsonResponse({ value: [{ displayName: 'One', email: 'one@example.com', givenName: 'One', surname: 'User', isShortProfile: true, jobTitle: '', objectId: '1', tenantName: 'Tenant', type: 'ADUser', userLocation: 'Remote', userPrincipalName: 'one@example.com' }], type: 'Users' });
        }
        if (request.url.includes('/profilepicture?displayname=aaa')) {
          return new Response(Buffer.from('jpeg-bytes').toString('base64'));
        }
        if (request.url.includes('/profilepicturev2')) {
          return new Response('jpeg-bytes');
        }
        if (request.url.endsWith('/pinnedChannels')) {
          return jsonResponse({ OrderVersion: 1, PinChannelOrder: ['19:one'] });
        }
        throw new Error(`unexpected request: ${request.url}`);
      }),
    });

    const conversations = await client.getConversations();
    expect(conversations.teams).toHaveLength(1);
    const messages = await client.getMessages(conversations.teams[0].channels[0]);
    expect(messages).toHaveLength(2);
    expect((await client.getMe()).email).toBe('teams-cli@outlook.com');
    expect((await client.fetchShortProfile(['8:orgid:user-1']))).toHaveLength(1);
    expect(Buffer.from(await client.getProfilePicture('user@example.com')).toString()).toBe('jpeg-bytes');
    expect(Buffer.from(await client.getTeamsProfilePicture('user@example.com')).toString()).toBe('jpeg-bytes');
    expect((await client.getTenants())).toHaveLength(1);
    expect(await client.getPinnedChannels()).toEqual(['19:one']);
    expect(client.chatSvc()).toBeTruthy();
  });
});
