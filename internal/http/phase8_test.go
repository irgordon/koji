package http

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
	"koji/internal/config"
	"koji/internal/jobs"
	"koji/internal/observability"
)

func TestAuthenticatedUserWithoutCapabilityReceivesForbidden(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseProtectedDiskRead(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestUserWithCapabilityReceivesAllowedResponse(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostDiskRead)

	response := exerciseProtectedDiskRead(fixture, session)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestDeniedServiceControlIsAudited(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseProtectedServiceControl(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionServiceControl, audit.OutcomeDenied, "capability_denied") {
		t.Fatal("expected denied service-control audit event")
	}
}

func TestAcceptedServiceControlIntentIsAudited(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostServicesControl)

	response := exerciseProtectedServiceControl(fixture, session)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobCreated, audit.OutcomeAccepted, jobs.StatusQueued) {
		t.Fatal("expected queued job creation audit event")
	}
}

func TestNonAllowlistedServiceControlIntentIsAudited(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostServicesControl)

	response := exerciseProtectedServiceControlForService(fixture, session, "nginx.service")

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionServiceControl, audit.OutcomeDenied, "service_not_allowlisted") {
		t.Fatal("expected non-allowlisted service-control audit event")
	}
}

func TestAuthEventsAreAudited(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()

	exerciseAuditedBootstrap(fixture, "admin", "secret-password")
	exerciseAuditedLogin(fixture, "admin", "bad-password")
	sessionResponse := exerciseAuditedLogin(fixture, "admin", "secret-password")
	session := auth.Session{ID: sessionCookie(sessionResponse).Value, CSRFToken: csrfTokenFromResponse(t, sessionResponse)}
	exerciseAuditedLogout(fixture, session)

	assertAuditEvent(t, fixture.database, audit.ActionBootstrap, audit.OutcomeSuccess)
	assertAuditEvent(t, fixture.database, audit.ActionLogin, audit.OutcomeFailure)
	assertAuditEvent(t, fixture.database, audit.ActionLogin, audit.OutcomeSuccess)
	assertAuditEvent(t, fixture.database, audit.ActionLogout, audit.OutcomeSuccess)
}

func TestDevBypassAuditMarkerExists(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/disk", nil)
	response := httptest.NewRecorder()
	protectedHandler(fixture, true).handleDiskFetch(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !devBypassAuditExists(t, fixture.database) {
		t.Fatal("expected dev bypass audit marker")
	}
}

func exerciseProtectedDiskRead(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodGet, "/api/v1/disk", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleDiskFetch(response, request)
	return response
}

func exerciseProtectedServiceControl(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	return exerciseProtectedServiceControlForService(fixture, session, "ssh.service")
}

func exerciseProtectedServiceControlForService(fixture testFixture, session auth.Session, service string) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodPost, "/api/services/"+service+"/restart", session)
	request.SetPathValue("name", service)
	request.SetPathValue("action", "restart")

	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleServiceControl(response, request)
	return response
}

func protectedHandler(fixture testFixture, devMode bool) protectedHandlers {
	return protectedHandlers{
		caps:             caps.NewStore(fixture.database),
		audit:            fixture.auditStore,
		devMode:          devMode,
		serviceAllowlist: newServiceAllowlist([]string{"ssh.service"}),
		processPolicy:    processVisibilityPolicy{mode: config.ProcessVisibilitySummary, maxProcesses: config.DefaultMaxProcesses},
		jobs:             jobs.NewStore(fixture.database),
		metrics:          observability.DefaultRegistry(),
		database:         fixture.database,
	}
}

func authenticatedRequest(method string, path string, session auth.Session) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	principal := auth.Principal{UserID: session.UserID, Username: session.Username}
	return withPrincipal(request, principal)
}

func grantCapability(t *testing.T, database *sql.DB, userID int64, capability caps.Capability) {
	t.Helper()

	_, err := database.ExecContext(context.Background(), `
INSERT INTO user_capabilities (user_id, capability_name)
VALUES (?, ?)`, userID, string(capability))
	if err != nil {
		t.Fatalf("grant capability: %v", err)
	}
}

func exerciseAuditedBootstrap(fixture testFixture, username string, password string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/bootstrap", credentialBody(username, password))
	response := httptest.NewRecorder()
	handleBootstrap(fixture.authStore, fixture.auditStore, true).ServeHTTP(response, request)
	return response
}

func exerciseAuditedLogin(fixture testFixture, username string, password string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/login", credentialBody(username, password))
	response := httptest.NewRecorder()
	handleLogin(fixture.authStore, fixture.auditStore, true).ServeHTTP(response, request)
	return response
}

func exerciseAuditedLogout(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	request.AddCookie(sessionCookieFor(session))
	request.Header.Set(auth.CSRFHeaderName, session.CSRFToken)

	response := httptest.NewRecorder()
	handleLogout(fixture.authStore, fixture.auditStore).ServeHTTP(response, request)
	return response
}

func csrfTokenFromResponse(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()

	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == auth.CSRFCookieName {
			return cookie.Value
		}
	}
	t.Fatal("expected csrf cookie")
	return ""
}

func assertAuditEvent(t *testing.T, database *sql.DB, action string, outcome string) {
	t.Helper()

	if !auditEventExists(t, database, action, outcome, "") {
		t.Fatalf("expected audit event %s %s", action, outcome)
	}
}

func auditEventExists(t *testing.T, database *sql.DB, action string, outcome string, reasonCode string) bool {
	t.Helper()

	query := "SELECT COUNT(*) FROM audit_events WHERE action = ? AND outcome = ?"
	args := []any{action, outcome}
	if reasonCode != "" {
		query += " AND reason_code = ?"
		args = append(args, reasonCode)
	}

	var count int
	if err := database.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query audit event: %v", err)
	}
	return count > 0
}

func devBypassAuditExists(t *testing.T, database *sql.DB) bool {
	t.Helper()

	var count int
	err := database.QueryRowContext(context.Background(), `
SELECT COUNT(*)
FROM audit_events
WHERE dev_bypass = 1 AND reason_code = ?`, string(caps.HostDiskRead)).Scan(&count)
	if err != nil {
		t.Fatalf("query dev bypass audit event: %v", err)
	}
	return count > 0
}
