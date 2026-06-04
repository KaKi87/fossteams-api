import { GenericHTTPError } from '../errors.js';
import { authString } from '../token.js';
import { parseConversationResponse, parseMessagesResponse } from '../models.js';
import { readResponseText, sortChannelsByName, sortMessagesByTime, sortTeamsByName } from '../utils.js';

export const ChatSvcAgg = 'https://teams.microsoft.com/api/csa/api/';
export const MessagesHost = 'https://emea.ng.msg.teams.microsoft.com/';
export const EndpointChatSvcAgg = 'chatsvcagg';
export const EndpointMessages = 'messages';

export class CSASvc {
  constructor(token, skypeToken, options = {}) {
    if (!token) {
      throw new Error('token cannot be nil');
    }
    if (!skypeToken) {
      throw new Error('skypeToken cannot be nil');
    }
    this.token = token;
    this.skypeToken = skypeToken;
    this.fetch = options.fetch ?? globalThis.fetch;
    this.csaSvcUrl = new URL(options.csaSvcUrl ?? ChatSvcAgg);
    this.msgUrl = new URL(options.msgUrl ?? MessagesHost);
    this._debugSave = false;
    this._debugDisallowUnknownFields = false;
  }

  debugSave(flag) {
    this._debugSave = flag;
  }

  debugDisallowUnknownFields(flag) {
    this._debugDisallowUnknownFields = flag;
  }

  getEndpoint(type, value) {
    const normalized = value.startsWith('/') ? value.slice(1) : value;
    const base = type === EndpointMessages ? this.msgUrl : this.csaSvcUrl;
    return new URL(`v1/${normalized}`, base);
  }

  authenticatedRequest(method, url, body, headers = {}) {
    const finalHeaders = new Headers(headers);
    if (url.startsWith(ChatSvcAgg)) {
      finalHeaders.set('Authorization', authString(this.token));
    } else if (url.startsWith(MessagesHost)) {
      finalHeaders.set('Authentication', authString(this.skypeToken));
    }
    return { method, headers: finalHeaders, body };
  }

  async authenticatedGetRequest(endpointUrl) {
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    if (response.status !== 200) {
      throw new GenericHTTPError(200, response.status, null);
    }
    return readResponseText(response, this._debugSave);
  }

  async getConversations() {
    const endpointUrl = this.getEndpoint(EndpointChatSvcAgg, '/teams/users/me');
    endpointUrl.searchParams.set('isPrefetch', 'false');
    endpointUrl.searchParams.set('enableMembershipSummary', 'true');
    return parseConversationResponse(await this.authenticatedGetRequest(endpointUrl));
  }

  async getMessagesByChannel(channel) {
    const endpointUrl = this.getEndpoint(EndpointMessages, `/users/ME/conversations/${encodeURIComponent(channel.id ?? channel.Id)}/messages`);
    endpointUrl.searchParams.set('view', 'msnp24Equivalent|supportsMessageProperties');
    endpointUrl.searchParams.set('pageSize', '200');
    endpointUrl.searchParams.set('startTime', '1');
    const response = await this.fetch(endpointUrl.toString(), this.authenticatedRequest('GET', endpointUrl.toString(), undefined));
    if (response.status !== 200) {
      throw new GenericHTTPError(200, response.status, null);
    }
    const payload = parseMessagesResponse(await readResponseText(response, this._debugSave));
    return payload.messages;
  }

  async getPinnedChannels() {
    const endpointUrl = this.getEndpoint(EndpointChatSvcAgg, '/teams/users/me/pinnedChannels');
    const payload = JSON.parse(await this.authenticatedGetRequest(endpointUrl));
    return payload.PinChannelOrder ?? payload.pinChannelOrder ?? [];
  }
}

export { sortTeamsByName, sortChannelsByName, sortMessagesByTime };
