package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestParseAuthResponse(t *testing.T) {
	refreshedRaw := mustRawJWT(t, map[string]any{"email": "user@example.com"})
	response := fmt.Sprintf(`{"tokens":{"skypeToken":"%s","expiresIn":86397},"region":"emea","partition":"emea01","regionGtms":{"chatServiceAggregator":"https://chatsvcagg.teams.microsoft.com"},"regionSettings":{"isUnifiedPresenceEnabled":true,"isOutOfOfficeIntegrationEnabled":true,"isContactMigrationEnabled":true,"isAppsDiscoveryEnabled":true,"isFederationEnabled":true},"licenseDetails":{"isFreemium":false,"isBasicLiveEventsEnabled":true,"isTrial":false,"isAdvComms":false}}`, refreshedRaw)

	client := New(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != TEAMS_API_ENDPOINT+"/authsvc/v1.0/authz" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		if got := req.Header.Get("ms-teams-authz-type"); got != AuthzRefresh {
			t.Fatalf("unexpected authz type: %s", got)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer "+"root-token" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(response)),
		}, nil
	})})

	token := &RootSkypeToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: TokenBearer}
	refreshed, err := client.Authz(token, AuthzRefresh)
	if err != nil {
		t.Fatalf("expected authz to succeed: %v", err)
	}
	if refreshed == nil || refreshed.Inner == nil {
		t.Fatal("expected refreshed token")
	}
	if refreshed.Inner.Raw != refreshedRaw {
		t.Fatalf("unexpected token payload: %s", refreshed.Inner.Raw)
	}
	if refreshed.Type != TokenBearer {
		t.Fatalf("unexpected token type: %s", refreshed.Type)
	}
}

func TestAuthzReturnsTypedError(t *testing.T) {
	client := New(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"errorCode":"GuestUserNotRedeemed","message":"select a tenant first"}`)),
		}, nil
	})})

	_, err := client.Authz(&RootSkypeToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: TokenBearer}, AuthzRefresh)
	if err == nil {
		t.Fatal("expected authz to fail")
	}
	authzErr, ok := err.(AuthzError)
	if !ok {
		t.Fatalf("expected AuthzError, got %T", err)
	}
	if authzErr.ErrorCode != GuestUserNotRedeemed {
		t.Fatalf("unexpected error code: %s", authzErr.ErrorCode)
	}
}

func TestAuthString(t *testing.T) {
	root := &TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: TokenBearer}
	if got := AuthString(root); got != "Bearer "+"root-token" {
		t.Fatalf("unexpected bearer auth string: %s", got)
	}
	root.Type = TokenSkype
	if got := AuthString(root); got != "skypetoken=root-token" {
		t.Fatalf("unexpected skype auth string: %s", got)
	}
	if got := AuthString(nil); got != "" {
		t.Fatalf("unexpected empty auth string: %s", got)
	}
}

func mustParseJWT(t *testing.T, raw string, claims map[string]any) *jwt.Token {
	t.Helper()
	if raw == "" {
		raw = mustRawJWT(t, claims)
	}
	return &jwt.Token{Raw: raw, Claims: jwt.MapClaims(claims)}
}

func mustRawJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	encodedClaims, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unable to marshal claims: %v", err)
	}
	return fmt.Sprintf("%s.%s.signature", base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)), base64.RawURLEncoding.EncodeToString(encodedClaims))
}
