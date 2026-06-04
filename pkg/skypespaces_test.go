package api

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSkypeSpaceGetTenants(t *testing.T) {
	svc := NewSkypeSpaceService(&SkypeToken{Inner: mustParseJWT(t, "skype-token", map[string]any{"email": "user@example.com"}), Type: TokenBearer})
	svc.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != SkypeSpacesEndpoint+"/beta/users/tenants" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer "+"skype-token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"tenantId":"id-1","tenantName":"Tenant","userId":"user-1","isInvitationRedeemed":true,"countryLetterCode":"CH","userType":"member","tenantType":"organization"}]`))}, nil
	})}

	tenants, err := svc.GetTenants()
	if err != nil {
		t.Fatalf("expected tenants to decode: %v", err)
	}
	if len(tenants) != 1 || tenants[0].TenantName != "Tenant" {
		t.Fatalf("unexpected tenants: %#v", tenants)
	}
}

func TestSkypeSpaceGetWithNilClient(t *testing.T) {
	u, err := url.Parse(SkypeSpacesEndpoint)
	if err != nil {
		t.Fatalf("unable to parse endpoint: %v", err)
	}
	svc := SkypeSpaceSvc{token: &SkypeToken{Inner: mustParseJWT(t, "skype-token", map[string]any{"email": "user@example.com"}), Type: TokenBearer}, endpoint: u}
	_, err = svc.get("/beta/users/tenants")
	if err == nil || !strings.Contains(err.Error(), "httpClient is nil") {
		t.Fatalf("expected nil client error, got %v", err)
	}
}
