import { describe, expect, test } from 'bun:test';
import { CSASvc, sortChannelsByName, sortMessagesByTime, sortTeamsByName } from '../src/services/csa.js';
import { TeamsToken, TokenBearer, TokenSkype } from '../src/token.js';
import { createFetch, jsonResponse, readFixture } from './helpers.js';

describe('CSA service', () => {
  test('loads conversations', async () => {
    const fixture = await readFixture('resources/chatsvcagg/conversations/conversations-1.json');
    const service = new CSASvc(
      new TeamsToken({ raw: 'csa-token', claims: {} }, TokenBearer),
      new TeamsToken({ raw: 'skype-token', claims: {} }, TokenSkype),
      {
        fetch: createFetch((request) => {
          expect(request.url).toBe('https://teams.microsoft.com/api/csa/api/v1/teams/users/me?isPrefetch=false&enableMembershipSummary=true');
          expect(request.headers.get('Authorization')).toBe('Bearer ' + 'csa-token');
          return jsonResponse(fixture);
        }),
      },
    );

    const conversations = await service.getConversations();
    expect(conversations.teams).toHaveLength(1);
    expect(conversations.teams[0].displayName).toBe('FossTeams');
  });

  test('loads channel messages and pinned channels', async () => {
    const fixture = await readFixture('resources/chatsvcagg/messages/messages-1.json');
    let callCount = 0;
    const service = new CSASvc(
      new TeamsToken({ raw: 'csa-token', claims: {} }, TokenBearer),
      new TeamsToken({ raw: 'skype-token', claims: {} }, TokenSkype),
      {
        fetch: createFetch((request) => {
          callCount += 1;
          if (callCount === 1) {
            expect(request.url).toContain('19%3A10dd444580a348eea3a3a335035aee3d%40thread.tacv2/messages');
            expect(request.headers.get('Authentication')).toBe('skypetoken=skype-token');
            return jsonResponse(fixture);
          }
          expect(request.url).toBe('https://teams.microsoft.com/api/csa/api/v1/teams/users/me/pinnedChannels');
          return jsonResponse({ OrderVersion: 1, PinChannelOrder: ['19:one', '19:two'] });
        }),
      },
    );

    const messages = await service.getMessagesByChannel({ id: '19:10dd444580a348eea3a3a335035aee3d@thread.tacv2' });
    expect(messages).toHaveLength(2);
    expect(messages[0].imdisplayname).toBe('Denys Vitali');

    const pinnedChannels = await service.getPinnedChannels();
    expect(pinnedChannels).toEqual(['19:one', '19:two']);
  });

  test('sort helpers mirror Go behavior', () => {
    expect(sortTeamsByName([{ displayName: 'Zulu' }, { displayName: 'alpha' }])[0].displayName).toBe('alpha');
    expect(sortChannelsByName([{ displayName: 'Zulu' }, { displayName: 'alpha' }])[0].displayName).toBe('alpha');
    expect(sortMessagesByTime([{ composeTime: '2022-01-02T00:00:00Z' }, { composeTime: '2022-01-01T00:00:00Z' }])[0].composeTime).toBe('2022-01-01T00:00:00Z');
  });
});
