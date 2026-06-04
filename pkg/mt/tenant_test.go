package mt

import (
	"net/http"
	"testing"

	"github.com/fossteams/teams-api/pkg/models"
)

func TestGetTenants(t *testing.T) {
	svc := mustService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return response(http.StatusOK, loadFixture(t, "resources/mt/tenants/tenants-1.json")), nil
	}))

	tenants, err := svc.GetTenants()
	if err != nil {
		t.Fatalf("expected tenants to decode: %v", err)
	}
	if len(tenants) != 1 {
		t.Fatalf("unexpected tenant count: %d", len(tenants))
	}
	if tenants[0].TenantName != "FossTeams" || tenants[0].TenantType != models.Organization {
		t.Fatalf("unexpected tenant: %#v", tenants[0])
	}
}

func TestGetVerifiedDomains(t *testing.T) {
	svc := mustService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return response(http.StatusOK, `[{"Name":"example.com"},{"Name":"example.org"}]`), nil
	}))

	verifiedDomains, err := svc.GetVerifiedDomains()
	if err != nil {
		t.Fatalf("expected domains to decode: %v", err)
	}
	if len(*verifiedDomains) != 2 || (*verifiedDomains)[0].Name != "example.com" {
		t.Fatalf("unexpected domains: %#v", verifiedDomains)
	}
}
