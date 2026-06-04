package mt

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
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetUserAndGetMe(t *testing.T) {
	svc := mustService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("Authorization"); got != "Bearer "+"root-token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		expected := MiddleTier + "emea/beta/users/user@example.com/?enableGuest=true&includeIBBarredUsers=true&isMailAddress=true&skypeTeamsInfo=true&throwIfNotFound=false"
		if req.URL.String() != expected {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		return response(http.StatusOK, loadFixture(t, "resources/mt/user/user-1.json")), nil
	}))

	user, err := svc.GetUser("user@example.com")
	if err != nil {
		t.Fatalf("expected user lookup to succeed: %v", err)
	}
	if user.DisplayName != "Denys Vitali" {
		t.Fatalf("unexpected display name: %s", user.DisplayName)
	}

	me, err := svc.GetMe()
	if err != nil {
		t.Fatalf("expected get me to succeed: %v", err)
	}
	if me.Email != user.Email {
		t.Fatalf("expected get me to use token email, got %s", me.Email)
	}
}

func TestFetchShortProfile(t *testing.T) {
	svc := mustService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if ct := req.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("unexpected content-type: %s", ct)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("unable to read body: %v", err)
		}
		if string(body) != `["8:orgid:user-1","8:orgid:user-2"]` {
			t.Fatalf("unexpected body: %s", string(body))
		}
		return response(http.StatusOK, `{"value":[{"displayName":"One","email":"one@example.com","givenName":"One","surname":"User","isShortProfile":true,"jobTitle":"","objectId":"1","tenantName":"Tenant","type":"ADUser","userLocation":"Remote","userPrincipalName":"one@example.com"},{"displayName":"Two","email":"two@example.com","givenName":"Two","surname":"User","isShortProfile":true,"jobTitle":"","objectId":"2","tenantName":"Tenant","type":"ADUser","userLocation":"Remote","userPrincipalName":"two@example.com"}],"type":"Users"}`), nil
	}))

	users, err := svc.FetchShortProfile("8:orgid:user-1", "8:orgid:user-2")
	if err != nil {
		t.Fatalf("expected short profile lookup to succeed: %v", err)
	}
	if len(users) != 2 || users[1].Email != "two@example.com" {
		t.Fatalf("unexpected users: %#v", users)
	}
}

func TestGetProfilePictureAndTeamsProfilePicture(t *testing.T) {
	jpg := []byte("jpeg-bytes")
	base64Payload := base64.StdEncoding.EncodeToString(jpg)
	callCount := 0
	svc := mustService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		switch callCount {
		case 1:
			if req.Header.Get("Authorization") == "" {
				t.Fatal("expected authorization header")
			}
			return response(http.StatusOK, base64Payload), nil
		case 2:
			if cookie := req.Header.Get("Cookie"); cookie != "TSAUTHCOOKIE=teams-token" {
				t.Fatalf("unexpected cookie header: %s", cookie)
			}
			return response(http.StatusOK, string(jpg)), nil
		default:
			t.Fatalf("unexpected extra request: %s", req.URL.String())
			return nil, nil
		}
	}))

	profilePicture, err := svc.GetProfilePicture("user@example.com")
	if err != nil {
		t.Fatalf("expected profile picture to decode: %v", err)
	}
	if string(profilePicture) != string(jpg) {
		t.Fatalf("unexpected picture bytes: %q", string(profilePicture))
	}

	teamsPicture, err := svc.GetTeamsProfilePicture("user@example.com")
	if err != nil {
		t.Fatalf("expected teams profile picture to decode: %v", err)
	}
	if string(teamsPicture) != string(jpg) {
		t.Fatalf("unexpected teams picture bytes: %q", string(teamsPicture))
	}
}

func TestGetTokenEmail(t *testing.T) {
	email, err := GetTokenEmail(&api.TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer})
	if err != nil || email != "user@example.com" {
		t.Fatalf("unexpected email claim result: %s %v", email, err)
	}
	upn, err := GetTokenEmail(&api.TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"upn": "user@example.com"}), Type: api.TokenBearer})
	if err != nil || upn != "user@example.com" {
		t.Fatalf("unexpected upn claim result: %s %v", upn, err)
	}
	_, err = GetTokenEmail(&api.TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"name": "user"}), Type: api.TokenBearer})
	if err == nil || !strings.Contains(err.Error(), "email nor upn") {
		t.Fatalf("expected missing claim error, got %v", err)
	}
}

func mustService(t *testing.T, transport http.RoundTripper) *Service {
	t.Helper()
	svc, err := NewMiddleTierService(
		api.Emea,
		&api.TeamsToken{Inner: mustParseJWT(t, "root-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
		&api.TeamsToken{Inner: mustParseJWT(t, "teams-token", map[string]any{"email": "user@example.com"}), Type: api.TokenBearer},
	)
	if err != nil {
		t.Fatalf("unable to create service: %v", err)
	}
	svc.client = &http.Client{Transport: transport}
	return svc
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func loadFixture(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
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
