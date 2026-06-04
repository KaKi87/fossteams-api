package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseConversationsFixture(t *testing.T) {
	var conversations ConversationResponse
	if err := json.Unmarshal([]byte(loadFixture(t, "resources/chatsvcagg/conversations/conversations-1.json")), &conversations); err != nil {
		t.Fatalf("unable to decode conversations fixture: %v", err)
	}
	if len(conversations.Teams) != 1 {
		t.Fatalf("unexpected team count: %d", len(conversations.Teams))
	}
	team := conversations.Teams[0]
	if team.DisplayName != "FossTeams" {
		t.Fatalf("unexpected team name: %s", team.DisplayName)
	}
	if len(team.Channels) == 0 || team.Channels[0].DisplayName != "General" {
		t.Fatalf("unexpected channels: %#v", team.Channels)
	}
}

func TestRFC3339TimeUnmarshal(t *testing.T) {
	var parsed RFC3339Time
	if err := parsed.UnmarshalJSON([]byte(`"2021-04-11T09:58:29.362Z"`)); err != nil {
		t.Fatalf("unable to unmarshal time: %v", err)
	}
	if time.Time(parsed).UTC().Format(time.RFC3339Nano) != "2021-04-11T09:58:29.362Z" {
		t.Fatalf("unexpected parsed time: %s", time.Time(parsed).UTC().Format(time.RFC3339Nano))
	}
}
