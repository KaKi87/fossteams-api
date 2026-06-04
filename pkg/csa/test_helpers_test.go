package csa

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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
	if claims == nil {
		claims = map[string]any{}
	}
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
