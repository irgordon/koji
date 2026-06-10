package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
	"koji/internal/jobs"
	"koji/internal/observability"
)

func TestObservabilityMetricsRequiresAuthentication(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()

	response := exerciseObservabilityMetrics(fixture, auth.Session{})

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestObservabilityMetricsRequiresCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseObservabilityMetrics(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestObservabilityMetricsReturnsCountersAndJobStatus(t *testing.T) {
	observability.DefaultRegistry().ResetForTest()
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.ObservabilityRead)
	createObservedJob(t, fixture.database, session.UserID)

	response := exerciseObservabilityMetrics(fixture, session)
	payload := decodeObservabilitySnapshot(t, response)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if payload.Counters[observability.JobsCreatedTotal] != 1 {
		t.Fatalf("expected jobs_created_total 1, got %d", payload.Counters[observability.JobsCreatedTotal])
	}
	if payload.JobsByStatus[jobs.StatusQueued] != 1 {
		t.Fatalf("expected queued job count 1, got %d", payload.JobsByStatus[jobs.StatusQueued])
	}
}

func TestAuditWriteMetricsIncrement(t *testing.T) {
	observability.DefaultRegistry().ResetForTest()
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.ObservabilityRead)

	if err := fixture.auditStore.Record(context.Background(), audit.Event{
		UserID:     &session.UserID,
		Action:     audit.ActionJobViewed,
		Target:     "jobs",
		Outcome:    audit.OutcomeSuccess,
		ReasonCode: "test",
	}); err != nil {
		t.Fatalf("record audit event: %v", err)
	}

	response := exerciseObservabilityMetrics(fixture, session)
	payload := decodeObservabilitySnapshot(t, response)

	if payload.Counters[observability.AuditWritesTotal] != 1 {
		t.Fatalf("expected audit_writes_total 1, got %d", payload.Counters[observability.AuditWritesTotal])
	}
}

func exerciseObservabilityMetrics(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/api/observability/metrics", nil)
	if session.ID != "" {
		request = authenticatedRequest(http.MethodGet, "/api/observability/metrics", session)
	}

	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleObservabilityMetrics(response, request)
	return response
}

func createObservedJob(t *testing.T, database *sql.DB, userID int64) {
	t.Helper()

	createdBy := userID
	_, err := jobs.NewStore(database).Create(context.Background(), jobs.CreateRequest{
		CreatedBy: &createdBy,
		Action:    "restart",
		Target:    "ssh.service",
		RequestID: "test-request-id",
	})
	if err != nil {
		t.Fatalf("create observed job: %v", err)
	}
}

func decodeObservabilitySnapshot(t *testing.T, response *httptest.ResponseRecorder) observability.Snapshot {
	t.Helper()

	var payload observability.Snapshot
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode observability response: %v", err)
	}
	return payload
}
