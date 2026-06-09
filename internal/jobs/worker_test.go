package jobs

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"koji/internal/agent"
	"koji/internal/audit"
	"koji/internal/db"
)

type fakeServiceController struct {
	err error
}

func (f fakeServiceController) ControlService(ctx context.Context, request agent.ServiceControlRequest) error {
	return f.err
}

func TestOnlyApprovedJobsCanBeClaimed(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	createQueuedJob(t, store, userID)

	_, err := store.ClaimApproved(context.Background())
	if !errors.Is(err, ErrNoApprovedJobs) {
		t.Fatalf("expected no approved jobs, got %v", err)
	}

	approved := createApprovedJob(t, store, userID)
	claimed, err := store.ClaimApproved(context.Background())
	if err != nil {
		t.Fatalf("claim approved job: %v", err)
	}
	if claimed.ID != approved.ID || claimed.Status != StatusRunning || claimed.StartedAt == "" {
		t.Fatalf("unexpected claimed job: %#v", claimed)
	}
}

func TestDoubleClaimCannotHappen(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	createApprovedJob(t, store, userID)

	if _, err := store.ClaimApproved(context.Background()); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	_, err := store.ClaimApproved(context.Background())
	if !errors.Is(err, ErrNoApprovedJobs) {
		t.Fatalf("expected no second claim, got %v", err)
	}
}

func TestWorkerMarksNotImplementedAgentResponse(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	job := createApprovedJob(t, store, userID)
	worker := NewWorker(store, fakeServiceController{err: agent.ErrNotImplemented}, audit.NewStore(database), time.Millisecond)

	if err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatalf("process job: %v", err)
	}

	stored := readJobsTestJob(t, store, job.ID)
	if stored.Status != StatusNotImplemented || stored.StatusReason != StatusNotImplemented {
		t.Fatalf("expected not implemented job, got %#v", stored)
	}
}

func TestWorkerMarksAgentUnavailableFailureSafely(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	job := createApprovedJob(t, store, userID)
	worker := NewWorker(store, fakeServiceController{err: agent.ErrAgentUnavailable}, audit.NewStore(database), time.Millisecond)

	if err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatalf("process job: %v", err)
	}

	stored := readJobsTestJob(t, store, job.ID)
	if stored.Status != StatusFailed || stored.StatusReason != ReasonAgentUnavailable {
		t.Fatalf("expected safe failure job, got %#v", stored)
	}
}

func TestWorkerPreservesAgentCommandFailureReason(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	job := createApprovedJob(t, store, userID)
	worker := NewWorker(store, fakeServiceController{err: agent.ErrCommandFailed}, audit.NewStore(database), time.Millisecond)

	if err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatalf("process job: %v", err)
	}

	stored := readJobsTestJob(t, store, job.ID)
	if stored.Status != StatusFailed || stored.StatusReason != ReasonCommandFailed {
		t.Fatalf("expected command failure job, got %#v", stored)
	}
}

func TestWorkerMapsMutationDisabledSafely(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	job := createApprovedJob(t, store, userID)
	worker := NewWorker(store, fakeServiceController{err: agent.ErrMutationDisabled}, audit.NewStore(database), time.Millisecond)

	if err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatalf("process job: %v", err)
	}

	stored := readJobsTestJob(t, store, job.ID)
	if stored.Status != StatusNotImplemented || stored.StatusReason != ReasonMutationDisabled {
		t.Fatalf("expected mutation disabled job, got %#v", stored)
	}
}

func TestWorkerStopsOnContextCancellation(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	worker := NewWorker(NewStore(database), fakeServiceController{}, audit.NewStore(database), time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- worker.Start(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean worker stop, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestWorkerAuditEventsAreWritten(t *testing.T) {
	database := openJobsTestDB(t)
	defer database.Close()
	store := NewStore(database)
	userID := insertJobsTestUser(t, database)
	createApprovedJob(t, store, userID)
	worker := NewWorker(store, fakeServiceController{err: agent.ErrNotImplemented}, audit.NewStore(database), time.Millisecond)

	if err := worker.ProcessOne(context.Background()); err != nil {
		t.Fatalf("process job: %v", err)
	}
	assertJobsAuditEvent(t, database, audit.ActionJobStarted, audit.OutcomeAccepted)
	assertJobsAuditEvent(t, database, audit.ActionJobNotImpl, audit.OutcomeRejected)
	assertJobsAuditEvent(t, database, audit.ActionJobStatus, audit.OutcomeAccepted)
}

func openJobsTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "koji.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return database
}

func insertJobsTestUser(t *testing.T, database *sql.DB) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
INSERT INTO users (username, password_hash)
VALUES (?, ?)`, "operator", "hash")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read inserted user id: %v", err)
	}
	return id
}

func createQueuedJob(t *testing.T, store *Store, userID int64) Job {
	t.Helper()

	job, err := store.Create(context.Background(), CreateRequest{
		CreatedBy: &userID,
		Action:    "service.restart",
		Target:    "ssh.service",
		RequestID: "worker-test",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	return job
}

func createApprovedJob(t *testing.T, store *Store, userID int64) Job {
	t.Helper()

	job := createQueuedJob(t, store, userID)
	approved, err := store.Approve(context.Background(), DecisionRequest{
		JobID:     job.ID,
		DecidedBy: userID,
		Reason:    "approved",
	})
	if err != nil {
		t.Fatalf("approve job: %v", err)
	}
	return approved
}

func readJobsTestJob(t *testing.T, store *Store, id string) Job {
	t.Helper()

	job, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("read job: %v", err)
	}
	return job
}

func assertJobsAuditEvent(t *testing.T, database *sql.DB, action string, outcome string) {
	t.Helper()

	var count int
	err := database.QueryRowContext(context.Background(), `
SELECT COUNT(*)
FROM audit_events
WHERE action = ? AND outcome = ?`, action, outcome).Scan(&count)
	if err != nil {
		t.Fatalf("query audit events: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected audit event %s %s", action, outcome)
	}
}
