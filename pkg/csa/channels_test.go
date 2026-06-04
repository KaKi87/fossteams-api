package csa

import (
	"net/http"
	"testing"

	"github.com/fossteams/teams-api/pkg/models"
)

func TestGetMessagesByChannelUsesSkypeToken(t *testing.T) {
	svc := mustTestService(t, fixtureTransport(t, func(req *http.Request) (*http.Response, error) {
		expected := MessagesHost + "v1/users/ME/conversations/19:10dd444580a348eea3a3a335035aee3d@thread.tacv2/messages?pageSize=200&startTime=1&view=msnp24Equivalent%7CsupportsMessageProperties"
		if req.URL.String() != expected {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		if got := req.Header.Get("Authentication"); got != "skypetoken=skype-token" {
			t.Fatalf("unexpected skype auth header: %s", got)
		}
		return jsonResponse(t, http.StatusOK, loadFixture(t, "resources/chatsvcagg/messages/messages-1.json")), nil
	}))

	messages, err := svc.GetMessagesByChannel(&models.Channel{Id: "19:10dd444580a348eea3a3a335035aee3d@thread.tacv2"})
	if err != nil {
		t.Fatalf("expected messages to decode: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("unexpected message count: %d", len(messages))
	}
	if messages[0].ImDisplayName != "Denys Vitali" {
		t.Fatalf("unexpected sender: %s", messages[0].ImDisplayName)
	}
}

func TestGetPinnedChannels(t *testing.T) {
	svc := mustTestService(t, fixtureTransport(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != ChatSvcAgg+"v1/teams/users/me/pinnedChannels" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		return jsonResponse(t, http.StatusOK, `{"OrderVersion":1,"PinChannelOrder":["19:one","19:two"]}`), nil
	}))

	channels, err := svc.GetPinnedChannels()
	if err != nil {
		t.Fatalf("expected pinned channels to decode: %v", err)
	}
	if len(channels) != 2 || channels[0] != "19:one" || channels[1] != "19:two" {
		t.Fatalf("unexpected pinned channels: %#v", channels)
	}
}
