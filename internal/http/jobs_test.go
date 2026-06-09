package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
	"koji/internal/jobs"
)

func TestServiceControlCreatesDurableJob(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostServicesControl)

	response := exerciseProtectedServiceControl(fixture, session)
	payload := decodeServiceControlJobResponse(t, response)
	stored := readJob(t, fixture, payload.JobID)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", response.Code)
	}
	if stored.Status != jobs.StatusQueued {
		t.Fatalf("expected queued job, got %q", stored.Status)
	}
	if stored.Action != "service.restart" || stored.Target != "ssh.service" {
		t.Fatalf("unexpected job: %#v", stored)
	}
}

func TestJobsListRequiresCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseJobsList(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestJobDecisionRequiresAuthentication(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()

	response := exerciseAuthGate(fixture.authStore, false, http.MethodPost, "/api/jobs/job-1/approve", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestJobDecisionRequiresCSRF(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	response := exerciseAuthGate(fixture.authStore, false, http.MethodPost, "/api/jobs/job-1/approve", sessionCookieFor(session), "")

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestJobDecisionRequiresCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	job := createQueuedJob(t, fixture, session.UserID)

	response := exerciseJobApprove(fixture, session, job.ID, "reviewed")

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobApproveDeny, audit.OutcomeDenied, "capability_denied") {
		t.Fatal("expected job.approval_denied audit event")
	}
}

func TestJobsListReturnsBoundedJobs(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.JobsRead)
	insertJobs(t, fixture, jobs.DefaultListLimit+5, session.UserID)

	response := exerciseJobsList(fixture, session)
	payload := decodeJobsListResponse(t, response)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if len(payload.Jobs) != jobs.DefaultListLimit {
		t.Fatalf("expected %d jobs, got %d", jobs.DefaultListLimit, len(payload.Jobs))
	}
}

func TestQueuedJobCanBeApproved(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.JobsApprove)
	job := createQueuedJob(t, fixture, session.UserID)

	response := exerciseJobApproveWithMiddleware(fixture, session, job.ID, "approved by operator")
	stored := readJob(t, fixture, job.ID)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if stored.Status != jobs.StatusApproved {
		t.Fatalf("expected approved job, got %#v", stored)
	}
	if stored.ApprovedBy == nil || *stored.ApprovedBy != session.UserID {
		t.Fatalf("expected approving user, got %#v", stored.ApprovedBy)
	}
	if stored.ApprovedAt == "" || stored.DecisionReason != "approved by operator" {
		t.Fatalf("expected durable approval fields, got %#v", stored)
	}
}

func TestQueuedJobCanBeRejected(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.JobsApprove)
	job := createQueuedJob(t, fixture, session.UserID)

	response := exerciseJobReject(fixture, session, job.ID, "unsafe maintenance window")
	stored := readJob(t, fixture, job.ID)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if stored.Status != jobs.StatusRejected {
		t.Fatalf("expected rejected job, got %#v", stored)
	}
	if stored.RejectedBy == nil || *stored.RejectedBy != session.UserID {
		t.Fatalf("expected rejecting user, got %#v", stored.RejectedBy)
	}
	if stored.RejectedAt == "" || stored.DecisionReason != "unsafe maintenance window" {
		t.Fatalf("expected durable rejection fields, got %#v", stored)
	}
}

func TestNonQueuedJobCannotBeDecided(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.JobsApprove)
	job := createQueuedJob(t, fixture, session.UserID)

	firstResponse := exerciseJobApprove(fixture, session, job.ID, "first decision")
	secondResponse := exerciseJobReject(fixture, session, job.ID, "second decision")

	if firstResponse.Code != http.StatusOK {
		t.Fatalf("expected first decision 200, got %d", firstResponse.Code)
	}
	if secondResponse.Code != http.StatusConflict {
		t.Fatalf("expected conflict, got %d", secondResponse.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobApproveDeny, audit.OutcomeDenied, "invalid_status") {
		t.Fatal("expected invalid transition audit event")
	}
}

func TestJobPersistsAcrossStoreRestart(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	store := jobs.NewStore(fixture.database)
	created, err := store.Create(context.Background(), jobs.CreateRequest{
		CreatedBy: &session.UserID,
		Action:    "service.restart",
		Target:    "ssh.service",
		RequestID: "restart-request",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	restartedStore := jobs.NewStore(fixture.database)
	loaded, err := restartedStore.Get(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("load persisted job: %v", err)
	}
	if loaded.ID != created.ID || loaded.Status != jobs.StatusQueued {
		t.Fatalf("unexpected loaded job: %#v", loaded)
	}
}

func TestJobAuditEventsAreWritten(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostServicesControl)
	grantCapability(t, fixture.database, session.UserID, caps.JobsRead)

	createResponse := exerciseProtectedServiceControl(fixture, session)
	payload := decodeServiceControlJobResponse(t, createResponse)
	detailResponse := exerciseJobDetail(fixture, session, payload.JobID)

	if createResponse.Code != http.StatusAccepted {
		t.Fatalf("expected create status 202, got %d", createResponse.Code)
	}
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d", detailResponse.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobCreated, audit.OutcomeAccepted, jobs.StatusQueued) {
		t.Fatal("expected job.created audit event")
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobViewed, audit.OutcomeSuccess, "jobs_read") {
		t.Fatal("expected job.viewed audit event")
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobStatus, audit.OutcomeAccepted, jobs.StatusQueued) {
		t.Fatal("expected job.status_changed audit event")
	}
}

func TestJobDecisionAuditEventsAreWritten(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.JobsApprove)
	job := createQueuedJob(t, fixture, session.UserID)

	response := exerciseJobApprove(fixture, session, job.ID, "approved")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobApproved, audit.OutcomeAccepted, jobs.StatusApproved) {
		t.Fatal("expected job.approved audit event")
	}
	if !auditEventExists(t, fixture.database, audit.ActionJobStatus, audit.OutcomeAccepted, jobs.StatusApproved) {
		t.Fatal("expected approval status audit event")
	}
}

func exerciseJobsList(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodGet, "/api/jobs", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleJobsList(response, request)
	return response
}

func exerciseJobDetail(fixture testFixture, session auth.Session, id string) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodGet, "/api/jobs/"+id, session)
	request.SetPathValue("id", id)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleJobDetail(response, request)
	return response
}

func exerciseJobApprove(fixture testFixture, session auth.Session, id string, reason string) *httptest.ResponseRecorder {
	return exerciseJobDecision(fixture, session, id, reason, protectedHandler(fixture, false).handleJobApprove)
}

func exerciseJobApproveWithMiddleware(fixture testFixture, session auth.Session, id string, reason string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/jobs/{id}/approve", protectedHandler(fixture, false).handleJobApprove)
	handler := applyMiddlewareChain(mux, fixture.authStore, false)
	request := httptest.NewRequest(http.MethodPost, "/api/jobs/"+id+"/approve", strings.NewReader(`{"reason":"`+reason+`"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(auth.CSRFHeaderName, session.CSRFToken)
	request.AddCookie(sessionCookieFor(session))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func exerciseJobReject(fixture testFixture, session auth.Session, id string, reason string) *httptest.ResponseRecorder {
	return exerciseJobDecision(fixture, session, id, reason, protectedHandler(fixture, false).handleJobReject)
}

func exerciseJobDecision(fixture testFixture, session auth.Session, id string, reason string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	request := authenticatedJSONRequest(http.MethodPost, "/api/jobs/"+id, session, `{"reason":"`+reason+`"}`)
	request.SetPathValue("id", id)
	response := httptest.NewRecorder()
	handler(response, request)
	return response
}

func authenticatedJSONRequest(method string, path string, session auth.Session, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	principal := auth.Principal{UserID: session.UserID, Username: session.Username}
	return withPrincipal(request, principal)
}

func createQueuedJob(t *testing.T, fixture testFixture, userID int64) jobs.Job {
	t.Helper()

	job, err := jobs.NewStore(fixture.database).Create(context.Background(), jobs.CreateRequest{
		CreatedBy: &userID,
		Action:    "service.restart",
		Target:    "ssh.service",
		RequestID: "job-decision-test",
	})
	if err != nil {
		t.Fatalf("create queued job: %v", err)
	}
	return job
}

func insertJobs(t *testing.T, fixture testFixture, count int, userID int64) {
	t.Helper()

	store := jobs.NewStore(fixture.database)
	for i := 0; i < count; i++ {
		_, err := store.Create(context.Background(), jobs.CreateRequest{
			CreatedBy: &userID,
			Action:    "service.restart",
			Target:    "ssh.service",
			RequestID: "jobs-list-test",
		})
		if err != nil {
			t.Fatalf("insert job: %v", err)
		}
	}
}

func readJob(t *testing.T, fixture testFixture, id string) jobs.Job {
	t.Helper()

	job, err := jobs.NewStore(fixture.database).Get(context.Background(), id)
	if err != nil {
		t.Fatalf("read job: %v", err)
	}
	return job
}

func decodeServiceControlJobResponse(t *testing.T, response *httptest.ResponseRecorder) serviceControlJobResponse {
	t.Helper()

	var payload serviceControlJobResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode service-control job response: %v", err)
	}
	return payload
}

func decodeJobsListResponse(t *testing.T, response *httptest.ResponseRecorder) jobsListResponse {
	t.Helper()

	var payload jobsListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode jobs list response: %v", err)
	}
	return payload
}
