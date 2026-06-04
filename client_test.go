package teams_api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
	api "github.com/fossteams/teams-api/pkg"
	"github.com/fossteams/teams-api/pkg/csa"
	"github.com/fossteams/teams-api/pkg/mt"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTeamsClientMethods(t *testing.T) {
	defaultClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = defaultClient })
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case csa.ChatSvcAgg + "v1/teams/users/me?enableMembershipSummary=true&isPrefetch=false":
			return response(http.StatusOK, loadFixture(t, "resources/chatsvcagg/conversations/conversations-1.json")), nil
		case csa.ChatSvcAgg + "v1/teams/users/me/pinnedChannels":
			return response(http.StatusOK, `{"OrderVersion":1,"PinChannelOrder":["19:one"]}`), nil
		case csa.MessagesHost + "v1/users/ME/conversations/19:10dd444580a348eea3a3a335035aee3d@thread.tacv2/messages?pageSize=200&startTime=1&view=msnp24Equivalent%7CsupportsMessageProperties":
			return response(http.StatusOK, loadFixture(t, "resources/chatsvcagg/messages/messages-1.json")), nil
		case mt.MiddleTier + "emea/beta/users/user@example.com/?enableGuest=true&includeIBBarredUsers=true&isMailAddress=true&skypeTeamsInfo=true&throwIfNotFound=false":
			return response(http.StatusOK, loadFixture(t, "resources/mt/user/user-1.json")), nil
		case mt.MiddleTier + "emea/beta/users/fetchShortProfile?enableGuest=true&includeIBBarredUsers=false&isMailAddress=false&skypeTeamsInfo=true":
			return response(http.StatusOK, `{"value":[{"displayName":"Denys Vitali","email":"teams-cli@outlook.com","givenName":"Denys","surname":"Vitali","isShortProfile":true,"jobTitle":"","objectId":"fa814989-41d0-4d4b-a365-e5f44e406847","tenantName":"FossTeams","type":"ADUser","userLocation":"Remote","userPrincipalName":"teams-cli@outlook.com"}],"type":"Users"}`), nil
		case mt.MiddleTier + "emea/beta/users/user@example.com/profilepicture?displayname=aaa":
			return response(http.StatusOK, "anBlZy1ieXRlcw=="), nil
		case mt.MiddleTier + "emea/beta/teams/user@example.com/profilepicturev2":
			return response(http.StatusOK, "jpeg-bytes"), nil
		case mt.MiddleTier + "emea/beta/users/tenants":
			return response(http.StatusOK, loadFixture(t, "resources/mt/tenants/tenants-1.json")), nil
		default:
			t.Fatalf("unexpected request: %s", req.URL.String())
			return nil, nil
		}
	})}

	chatSvc, err := csa.NewCSAService(
		&api.TeamsToken{Inner: mustParseJWT(t, "csa-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
		&api.SkypeToken{Inner: mustParseJWT(t, "skype-token", map[string]any{"email": "user@example.com"}), Type: api.TokenSkype},
	)
	if err != nil {
		t.Fatalf("unable to create chat service: %v", err)
	}
	mtSvc, err := mt.NewMiddleTierService(
		api.Emea,
		&api.TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
		&api.TeamsToken{Inner: mustParseJWT(t, "teams-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
	)
	if err != nil {
		t.Fatalf("unable to create MT service: %v", err)
	}

	client := &TeamsClient{chatSvc: chatSvc, mtSvc: mtSvc}
	client.Debug(true)

	conversations, err := client.GetConversations()
	if err != nil || len(conversations.Teams) != 1 {
		t.Fatalf("unexpected conversations result: %#v %v", conversations, err)
	}
	messages, err := client.GetMessages(&conversations.Teams[0].Channels[0])
	if err != nil || len(messages) != 2 {
		t.Fatalf("unexpected messages result: %#v %v", messages, err)
	}
	me, err := client.GetMe()
	if err != nil || me.Email != "teams-cli@outlook.com" {
		t.Fatalf("unexpected me result: %#v %v", me, err)
	}
	shortProfiles, err := client.FetchShortProfile([]string{"8:orgid:fa814989-41d0-4d4b-a365-e5f44e406847"})
	if err != nil || len(shortProfiles) != 1 {
		t.Fatalf("unexpected short profiles: %#v %v", shortProfiles, err)
	}
	profilePicture, err := client.GetProfilePicture("user@example.com")
	if err != nil || string(profilePicture) != "jpeg-bytes" {
		t.Fatalf("unexpected profile picture: %q %v", string(profilePicture), err)
	}
	teamsProfilePicture, err := client.GetTeamsProfilePicture("user@example.com")
	if err != nil || string(teamsProfilePicture) != "jpeg-bytes" {
		t.Fatalf("unexpected teams profile picture: %q %v", string(teamsProfilePicture), err)
	}
	tenants, err := client.GetTenants()
	if err != nil || len(tenants) != 1 {
		t.Fatalf("unexpected tenants result: %#v %v", tenants, err)
	}
	pinnedChannels, err := client.GetPinnedChannels()
	if err != nil || len(pinnedChannels) != 1 || pinnedChannels[0] != "19:one" {
		t.Fatalf("unexpected pinned channels: %#v %v", pinnedChannels, err)
	}
	if client.ChatSvc() == nil {
		t.Fatal("expected chat service")
	}
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func loadFixture(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(filename)))
	body, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatalf("unable to read fixture %s: %v", relativePath, err)
	}
	return string(body)
}

func mustParseJWT(t *testing.T, raw string, claims map[string]any) *jwt.Token {
	t.Helper()
	encodedClaims, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unable to marshal claims: %v", err)
	}
	encoded := fmt.Sprintf("%s.%s.signature", base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)), base64.RawURLEncoding.EncodeToString(encodedClaims))
	if raw == "" {
		raw = encoded
	}
	return &jwt.Token{Raw: raw, Claims: jwt.MapClaims(claims)}
}
