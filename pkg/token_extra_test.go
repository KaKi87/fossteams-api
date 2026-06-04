package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTokenFromEnvironment(t *testing.T) {
	t.Setenv("MS_TEAMS_SKYPE_TOKEN", mustRawJWT(t, map[string]any{"email": "user@example.com"}))
	token, err := GetSkypeSpacesToken()
	if err != nil {
		t.Fatalf("expected token from env: %v", err)
	}
	if token.Inner == nil || token.Inner.Raw == "" {
		t.Fatal("expected parsed token")
	}
}

func TestGetTokenFromFilesystem(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MS_TEAMS_TEAMS_TOKEN", "")
	tokenPath := filepath.Join(home, ".config", "fossteams")
	if err := os.MkdirAll(tokenPath, 0o755); err != nil {
		t.Fatalf("unable to create token dir: %v", err)
	}
	raw := mustRawJWT(t, map[string]any{"upn": "user@example.com"})
	if err := os.WriteFile(filepath.Join(tokenPath, "token-teams.jwt"), []byte(raw), 0o644); err != nil {
		t.Fatalf("unable to write token file: %v", err)
	}

	token, err := GetTeamsToken()
	if err != nil {
		t.Fatalf("expected token from file: %v", err)
	}
	if token.Inner == nil || token.Inner.Raw != raw {
		t.Fatalf("unexpected token: %#v", token)
	}
}
