package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestParseMessagesFixture(t *testing.T) {
	var messages MessagesResponse
	if err := json.Unmarshal([]byte(loadFixture(t, "resources/chatsvcagg/messages/messages-1.json")), &messages); err != nil {
		t.Fatalf("unable to decode messages fixture: %v", err)
	}
	if len(messages.Messages) != 2 {
		t.Fatalf("unexpected message count: %d", len(messages.Messages))
	}
	if messages.Messages[0].ImDisplayName != "Denys Vitali" {
		t.Fatalf("unexpected sender: %s", messages.Messages[0].ImDisplayName)
	}
	if messages.Metadata.LastCompleteSegmentEndTime != 1618133509362 {
		t.Fatalf("unexpected metadata: %#v", messages.Metadata)
	}
}

func TestSortMessagesByTime(t *testing.T) {
	messages := []ChatMessage{{Id: "late", ComposeTime: RFC3339Time(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC))}, {Id: "early", ComposeTime: RFC3339Time(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))}}
	sort.Sort(SortMessageByTime(messages))
	if messages[0].Id != "early" {
		t.Fatalf("messages not sorted by time: %#v", messages)
	}
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
