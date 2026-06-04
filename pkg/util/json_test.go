package util

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetJSONWithoutDebugSave(t *testing.T) {
	reader, err := GetJSON(&http.Response{Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, false)
	if err != nil {
		t.Fatalf("expected body reader: %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unable to read body: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestGetJSONWithDebugSave(t *testing.T) {
	reader, err := GetJSON(&http.Response{Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, true)
	if err != nil {
		t.Fatalf("expected body reader: %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unable to read debug body: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected debug body: %s", string(body))
	}
}
