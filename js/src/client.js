import { AuthClient, AuthzRefresh, Emea } from './auth.js';
import { getChatSvcAggToken, getRootToken, getSkypeSpacesToken, getTeamsToken, TeamsToken, TokenSkype } from './token.js';
import { CSASvc } from './services/csa.js';
import { Service } from './services/mt.js';

export class TeamsClient {
  constructor(chatSvc, mtSvc) {
    this._chatSvc = chatSvc;
    this._mtSvc = mtSvc;
  }

  debug(debugFlag) {
    this._chatSvc.debugSave(debugFlag);
  }

  chatSvc() {
    return this._chatSvc;
  }

  getConversations() {
    return this._chatSvc.getConversations();
  }

  getMessages(channel) {
    return this._chatSvc.getMessagesByChannel(channel);
  }

  getMe() {
    return this._mtSvc.getMe();
  }

  fetchShortProfile(mris) {
    return this._mtSvc.fetchShortProfile(mris);
  }

  getProfilePicture(emailOrId) {
    return this._mtSvc.getProfilePicture(emailOrId);
  }

  getTeamsProfilePicture(emailOrId) {
    return this._mtSvc.getTeamsProfilePicture(emailOrId);
  }

  getTenants() {
    return this._mtSvc.getTenants();
  }

  getPinnedChannels() {
    return this._chatSvc.getPinnedChannels();
  }

  static async create(options = {}) {
    if (options.chatSvc && options.mtSvc) {
      return new TeamsClient(options.chatSvc, options.mtSvc);
    }
    const fetchImpl = options.fetch ?? globalThis.fetch;
    const skypeSpaces = options.skypeSpacesToken ?? await getSkypeSpacesToken(options);
    const chatSvcToken = options.chatSvcToken ?? await getChatSvcAggToken(options);
    const rootToken = options.rootToken ?? await getRootToken(options);
    const teamsToken = options.teamsToken ?? await getTeamsToken(options);
    let skypeToken = options.skypeToken;
    if (!skypeToken) {
      const authClient = new AuthClient(fetchImpl);
      skypeToken = await authClient.authz(rootToken, AuthzRefresh);
      skypeToken.type = TokenSkype;
    }
    const chatSvc = new CSASvc(chatSvcToken, skypeToken, { fetch: fetchImpl });
    const mtSvc = new Service(Emea, skypeSpaces instanceof TeamsToken ? skypeSpaces : new TeamsToken(skypeSpaces.inner ?? skypeSpaces), teamsToken, { fetch: fetchImpl });
    return new TeamsClient(chatSvc, mtSvc);
  }
}

export async function New(options = {}) {
  return TeamsClient.create(options);
}
