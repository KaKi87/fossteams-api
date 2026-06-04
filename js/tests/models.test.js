import { describe, expect, test } from 'bun:test';
import { parseConversationResponse, parseMessagesResponse } from '../src/models.js';
import { sortMessagesByTime } from '../src/utils.js';
import { readFixture } from './helpers.js';

describe('fixture parsers', () => {
  test('parses conversations and messages fixtures', async () => {
    const conversations = parseConversationResponse(await readFixture('resources/chatsvcagg/conversations/conversations-1.json'));
    const messages = parseMessagesResponse(await readFixture('resources/chatsvcagg/messages/messages-1.json'));

    expect(conversations.teams[0].displayName).toBe('FossTeams');
    expect(messages.messages).toHaveLength(2);
    expect(sortMessagesByTime(messages.messages)[0].id).toBe('1618133498996');
  });
});
