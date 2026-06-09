package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeneratedRequestIDExists(t *testing.T) {
	response, contextID := exerciseRequestIDMiddleware(t, "")

	headerID := response.Header().Get(requestIDHeader)
	if headerID == "" {
		t.Fatal("expected generated request id header")
	}
	if !validRequestID(headerID) {
		t.Fatalf("expected valid generated request id, got %q", headerID)
	}
	if contextID != headerID {
		t.Fatalf("expected context request id %q, got %q", headerID, contextID)
	}
}

func TestValidInboundRequestIDIsPreserved(t *testing.T) {
	inboundID := "phase15-request-0001"

	response, contextID := exerciseRequestIDMiddleware(t, inboundID)

	if response.Header().Get(requestIDHeader) != inboundID {
		t.Fatalf("expected inbound request id to be preserved")
	}
	if contextID != inboundID {
		t.Fatalf("expected context request id %q, got %q", inboundID, contextID)
	}
}

func TestInvalidInboundRequestIDIsReplaced(t *testing.T) {
	inboundID := "bad request id"

	response, contextID := exerciseRequestIDMiddleware(t, inboundID)

	headerID := response.Header().Get(requestIDHeader)
	if headerID == inboundID {
		t.Fatal("expected invalid inbound request id to be replaced")
	}
	if !validRequestID(headerID) {
		t.Fatalf("expected valid replacement request id, got %q", headerID)
	}
	if contextID != headerID {
		t.Fatalf("expected context request id %q, got %q", headerID, contextID)
	}
}

func TestAuditEventIncludesRequestID(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	inboundID := "audit-request-0001"

	handler := RequestIDMiddleware(handleBootstrap(fixture.authStore, fixture.auditStore, true))
	request := httptest.NewRequest(http.MethodPost, "/api/bootstrap", credentialBody("admin", "secret-password"))
	request.Header.Set(requestIDHeader, inboundID)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap status 200, got %d", response.Code)
	}
	if response.Header().Get(requestIDHeader) != inboundID {
		t.Fatalf("expected response request id %q, got %q", inboundID, response.Header().Get(requestIDHeader))
	}
	if auditRequestID(t, fixture, "auth.bootstrap") != inboundID {
		t.Fatal("expected audit event to include response request id")
	}
}

func exerciseRequestIDMiddleware(t *testing.T, inboundID string) (*httptest.ResponseRecorder, string) {
	t.Helper()

	var contextID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = requestID(r)
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequestIDMiddleware(next)
	request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	if inboundID != "" {
		request.Header.Set(requestIDHeader, inboundID)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response, contextID
}

func auditRequestID(t *testing.T, fixture testFixture, action string) string {
	t.Helper()

	var id string
	err := fixture.database.QueryRowContext(
		context.Background(),
		"SELECT request_id FROM audit_events WHERE action = ? ORDER BY id DESC LIMIT 1",
		action,
	).Scan(&id)
	if err != nil {
		t.Fatalf("query audit request id: %v", err)
	}
	return id
}
