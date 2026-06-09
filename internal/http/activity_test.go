package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
)

func TestActivityRequiresAuthentication(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()

	response := exerciseAuthGate(fixture.authStore, false, http.MethodGet, "/api/activity", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestActivityRequiresCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseActivityList(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestActivityWithCapabilityReturnsBoundedEvents(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.AuditEventsRead)
	insertAuditEvents(t, fixture, audit.DefaultRecentLimit+5)

	response := exerciseActivityList(fixture, session)
	payload := decodeActivityResponse(t, response)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if len(payload.Events) != audit.DefaultRecentLimit {
		t.Fatalf("expected %d events, got %d", audit.DefaultRecentLimit, len(payload.Events))
	}
	if payload.Events[0].Action != "test.action.54" {
		t.Fatalf("expected newest event first, got %q", payload.Events[0].Action)
	}
}

func TestActivityResponseExcludesSensitiveDetails(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.AuditEventsRead)
	insertSensitiveAuditEvent(t, fixture)

	response := exerciseActivityList(fixture, session)
	body := response.Body.String()

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	for _, token := range []string{"actor", "user_id", "remote_addr", "dev_bypass", "secret-token"} {
		if strings.Contains(body, token) {
			t.Fatalf("activity response exposed sensitive token %q in %q", token, body)
		}
	}
}

func exerciseActivityList(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodGet, "/api/activity", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleActivityList(response, request)
	return response
}

func insertAuditEvents(t *testing.T, fixture testFixture, count int) {
	t.Helper()

	for i := 0; i < count; i++ {
		_, err := fixture.database.ExecContext(context.Background(), `
INSERT INTO audit_events (
	actor,
	action,
	target,
	status,
	message,
	outcome,
	reason_code,
	request_id,
	remote_addr
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"user:1",
			fmt.Sprintf("test.action.%02d", i),
			"target",
			audit.OutcomeSuccess,
			"message",
			audit.OutcomeSuccess,
			"reason",
			fmt.Sprintf("request-%02d", i),
			"192.0.2.1",
		)
		if err != nil {
			t.Fatalf("insert audit event: %v", err)
		}
	}
}

func insertSensitiveAuditEvent(t *testing.T, fixture testFixture) {
	t.Helper()

	_, err := fixture.database.ExecContext(context.Background(), `
INSERT INTO audit_events (
	actor,
	action,
	target,
	status,
	message,
	outcome,
	reason_code,
	request_id,
	remote_addr,
	dev_bypass
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"user:99",
		"secret.action",
		"target",
		audit.OutcomeFailure,
		"secret-token",
		audit.OutcomeFailure,
		"safe_reason",
		"request-safe",
		"198.51.100.10",
		1,
	)
	if err != nil {
		t.Fatalf("insert sensitive audit event: %v", err)
	}
}

func decodeActivityResponse(t *testing.T, response *httptest.ResponseRecorder) activityResponse {
	t.Helper()

	var payload activityResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode activity response: %v", err)
	}
	return payload
}
