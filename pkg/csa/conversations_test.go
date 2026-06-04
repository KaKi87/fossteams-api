package csa

import (
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"

	api "github.com/fossteams/teams-api/pkg"
)

func fixtureTransport(t *testing.T, fn func(*http.Request) (*http.Response, error)) http.RoundTripper {
	t.Helper()
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return fn(req)
	})
}

func TestGetConversationsUsesChatSvcAggToken(t *testing.T) {
	svc := mustTestService(t, fixtureTransport(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != ChatSvcAgg+"v1/teams/users/me?enableMembershipSummary=true&isPrefetch=false" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer "+"csa-token" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		return jsonResponse(t, http.StatusOK, loadFixture(t, "resources/chatsvcagg/conversations/conversations-1.json")), nil
	}))

	conversations, err := svc.GetConversations()
	if err != nil {
		t.Fatalf("expected conversations to decode: %v", err)
	}
	if len(conversations.Teams) == 0 || len(conversations.Teams[0].Channels) == 0 {
		t.Fatal("expected teams and channels")
	}

	sort.Sort(TeamsByName(conversations.Teams))
	if conversations.Teams[0].DisplayName != "FossTeams" {
		t.Fatalf("unexpected first team: %s", conversations.Teams[0].DisplayName)
	}
}

func TestGetConversationsReturnsHTTPError(t *testing.T) {
	svc := mustTestService(t, fixtureTransport(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(t, http.StatusUnauthorized, `{"error":"denied"}`), nil
	}))

	_, err := svc.GetConversations()
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
}

func mustTestService(t *testing.T, transport http.RoundTripper) *CSASvc {
	t.Helper()
	svc, err := NewCSAService(
		&api.TeamsToken{Inner: mustParseJWT(t, "csa-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
		&api.SkypeToken{Inner: mustParseJWT(t, "skype-token", map[string]any{"email": "user@example.com"}), Type: api.TokenSkype},
	)
	if err != nil {
		t.Fatalf("unable to create CSA service: %v", err)
	}
	svc.client = &http.Client{Transport: transport}
	return svc
}

func jsonResponse(t *testing.T, status int, body string) *http.Response {
	t.Helper()
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
