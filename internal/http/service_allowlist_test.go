package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"koji/internal/auth"
	"koji/internal/caps"
)

func TestAllowlistedServiceStatusAllowed(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostServicesRead)

	request := authenticatedRequest(http.MethodGet, "/api/v1/services", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleServicesList(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected service status list to be allowed, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"services"`) {
		t.Fatalf("expected services response body, got %q", response.Body.String())
	}
}

func TestServiceAllowlistDeniesByDefault(t *testing.T) {
	allowlist := newServiceAllowlist(nil)

	if allowlist.allows("ssh.service") {
		t.Fatal("expected empty service allowlist to deny by default")
	}
}

func TestDevServiceAllowlistDoesNotGrantCapabilities(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	request := authenticatedRequest(http.MethodGet, "/api/v1/services", auth.Session{
		UserID:   session.UserID,
		Username: session.Username,
	})
	response := httptest.NewRecorder()
	protectedHandlers{
		caps:             caps.NewStore(fixture.database),
		audit:            fixture.auditStore,
		devMode:          false,
		serviceAllowlist: newServiceAllowlist([]string{"ssh.service"}),
	}.handleServicesList(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected missing capability to remain denied, got %d", response.Code)
	}
}
