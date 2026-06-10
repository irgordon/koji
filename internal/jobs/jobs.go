package jobs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"koji/internal/observability"
)

const (
	StatusQueued         = "queued"
	StatusApproved       = "approved"
	StatusRejected       = "rejected"
	StatusRunning        = "running"
	StatusCompleted      = "completed"
	StatusFailed         = "failed"
	StatusNotImplemented = "not_implemented"

	DefaultListLimit = 50
	MaxListLimit     = 100
)

var (
	ErrJobNotFound          = errors.New("job not found")
	ErrInvalidJobTransition = errors.New("invalid job transition")
	ErrNoApprovedJobs       = errors.New("no approved jobs")
)

type CreateRequest struct {
	CreatedBy *int64
	Action    string
	Target    string
	RequestID string
}

type DecisionRequest struct {
	JobID     string
	DecidedBy int64
	Reason    string
}

type Job struct {
	ID             string `json:"id"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      *int64 `json:"created_by,omitempty"`
	Action         string `json:"action"`
	Target         string `json:"target"`
	Status         string `json:"status"`
	StatusReason   string `json:"status_reason"`
	RequestID      string `json:"request_id"`
	ApprovedBy     *int64 `json:"approved_by,omitempty"`
	ApprovedAt     string `json:"approved_at,omitempty"`
	RejectedBy     *int64 `json:"rejected_by,omitempty"`
	RejectedAt     string `json:"rejected_at,omitempty"`
	DecisionReason string `json:"decision_reason"`
	StartedAt      string `json:"started_at,omitempty"`
}

type Store struct {
	db      *sql.DB
	metrics *observability.Registry
}

func NewStore(db *sql.DB) *Store {
	return NewStoreWithMetrics(db, observability.DefaultRegistry())
}

func NewStoreWithMetrics(db *sql.DB, metrics *observability.Registry) *Store {
	return &Store{db: db, metrics: metrics}
}

func (s *Store) Create(ctx context.Context, request CreateRequest) (Job, error) {
	job := newQueuedJob(request)
	if err := s.insert(ctx, job); err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsCreatedTotal)
	return job, nil
}

func (s *Store) ListRecent(ctx context.Context, limit int) ([]Job, error) {
	rows, err := s.queryRecent(ctx, boundedLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (s *Store) Get(ctx context.Context, id string) (Job, error) {
	job, err := s.queryOne(ctx, id)
	if err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Store) Approve(ctx context.Context, request DecisionRequest) (Job, error) {
	job, err := s.decide(ctx, request, approvalDecision())
	if err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsApprovedTotal)
	return job, nil
}

func (s *Store) Reject(ctx context.Context, request DecisionRequest) (Job, error) {
	job, err := s.decide(ctx, request, rejectionDecision())
	if err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsRejectedTotal)
	return job, nil
}

func (s *Store) ClaimApproved(ctx context.Context) (Job, error) {
	row := s.db.QueryRowContext(ctx, claimApprovedSQL(), formatTime(time.Now().UTC()), StatusRunning, StatusRunning, StatusApproved)
	job, err := scanClaimedJob(row)
	if err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsClaimedTotal)
	return job, nil
}

func (s *Store) MarkFailed(ctx context.Context, id string, reason string) (Job, error) {
	job, err := s.finishRunning(ctx, id, StatusFailed, reason)
	if err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsFailedTotal)
	return job, nil
}

func (s *Store) MarkCompleted(ctx context.Context, id string) (Job, error) {
	job, err := s.finishRunning(ctx, id, StatusCompleted, StatusCompleted)
	if err != nil {
		return Job{}, err
	}
	s.metrics.Inc(observability.JobsCompletedTotal)
	return job, nil
}

func newQueuedJob(request CreateRequest) Job {
	return Job{
		ID:           newJobID(),
		CreatedAt:    formatTime(time.Now().UTC()),
		CreatedBy:    request.CreatedBy,
		Action:       request.Action,
		Target:       request.Target,
		Status:       StatusQueued,
		StatusReason: "pending_approval",
		RequestID:    request.RequestID,
	}
}

func (s *Store) insert(ctx context.Context, job Job) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO jobs (
	id,
	created_at,
	created_by,
	action,
	target,
	status,
	status_reason,
	request_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID,
		job.CreatedAt,
		job.CreatedBy,
		job.Action,
		job.Target,
		job.Status,
		job.StatusReason,
		job.RequestID,
	)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (s *Store) queryRecent(ctx context.Context, limit int) (*sql.Rows, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, created_at, created_by, action, target, status, status_reason, request_id,
	approved_by, approved_at, rejected_by, rejected_at, decision_reason, started_at
FROM jobs
ORDER BY created_at DESC, id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return rows, nil
}

func (s *Store) queryOne(ctx context.Context, id string) (Job, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, created_at, created_by, action, target, status, status_reason, request_id,
	approved_by, approved_at, rejected_by, rejected_at, decision_reason, started_at
FROM jobs
WHERE id = ?`, id)
	return scanJob(row)
}

func (s *Store) finishRunning(ctx context.Context, id string, status string, reason string) (Job, error) {
	if err := s.updateRunningJob(ctx, id, status, reason); err != nil {
		return Job{}, err
	}
	return s.Get(ctx, id)
}

func (s *Store) updateRunningJob(ctx context.Context, id string, status string, reason string) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE jobs
SET status = ?,
	status_reason = ?
WHERE id = ? AND status = ?`, status, reason, id, StatusRunning)
	if err != nil {
		return fmt.Errorf("finish job: %w", err)
	}
	return translateDecisionResult(ctx, s, id, result)
}

func (s *Store) decide(ctx context.Context, request DecisionRequest, decision jobDecision) (Job, error) {
	if err := s.updateQueuedJob(ctx, request, decision); err != nil {
		return Job{}, err
	}
	return s.Get(ctx, request.JobID)
}

func (s *Store) updateQueuedJob(ctx context.Context, request DecisionRequest, decision jobDecision) error {
	result, err := s.db.ExecContext(ctx, decision.sql, decision.args(request)...)
	if err != nil {
		return fmt.Errorf("decide job: %w", err)
	}
	return translateDecisionResult(ctx, s, request.JobID, result)
}

func translateDecisionResult(ctx context.Context, store *Store, id string, result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read job decision result: %w", err)
	}
	if affected == 1 {
		return nil
	}
	return store.missingOrInvalidTransition(ctx, id)
}

func (s *Store) missingOrInvalidTransition(ctx context.Context, id string) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return ErrInvalidJobTransition
}

func scanJobs(rows *sql.Rows) ([]Job, error) {
	result := make([]Job, 0, DefaultListLimit)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan jobs: %w", err)
	}
	return result, nil
}

type jobScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner jobScanner) (Job, error) {
	var job Job
	var approvedAt sql.NullString
	var rejectedAt sql.NullString
	var startedAt sql.NullString
	err := scanner.Scan(
		&job.ID,
		&job.CreatedAt,
		&job.CreatedBy,
		&job.Action,
		&job.Target,
		&job.Status,
		&job.StatusReason,
		&job.RequestID,
		&job.ApprovedBy,
		&approvedAt,
		&job.RejectedBy,
		&rejectedAt,
		&job.DecisionReason,
		&startedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, ErrJobNotFound
	}
	if err != nil {
		return Job{}, fmt.Errorf("scan job: %w", err)
	}
	job.ApprovedAt = nullableString(approvedAt)
	job.RejectedAt = nullableString(rejectedAt)
	job.StartedAt = nullableString(startedAt)
	return job, nil
}

func scanClaimedJob(scanner jobScanner) (Job, error) {
	job, err := scanJob(scanner)
	if errors.Is(err, ErrJobNotFound) {
		return Job{}, ErrNoApprovedJobs
	}
	return job, err
}

func claimApprovedSQL() string {
	return `
UPDATE jobs
SET started_at = ?,
	status = ?,
	status_reason = ?
WHERE id = (
	SELECT id
	FROM jobs
	WHERE status = ?
	ORDER BY created_at ASC, id ASC
	LIMIT 1
)
RETURNING id, created_at, created_by, action, target, status, status_reason, request_id,
	approved_by, approved_at, rejected_by, rejected_at, decision_reason, started_at`
}

type jobDecision struct {
	sql  string
	args func(DecisionRequest) []any
}

func approvalDecision() jobDecision {
	return jobDecision{
		sql: `
UPDATE jobs
SET status = ?,
	status_reason = ?,
	approved_by = ?,
	approved_at = ?,
	decision_reason = ?
WHERE id = ? AND status = ?`,
		args: func(request DecisionRequest) []any {
			return []any{
				StatusApproved,
				StatusApproved,
				request.DecidedBy,
				formatTime(time.Now().UTC()),
				request.Reason,
				request.JobID,
				StatusQueued,
			}
		},
	}
}

func rejectionDecision() jobDecision {
	return jobDecision{
		sql: `
UPDATE jobs
SET status = ?,
	status_reason = ?,
	rejected_by = ?,
	rejected_at = ?,
	decision_reason = ?
WHERE id = ? AND status = ?`,
		args: func(request DecisionRequest) []any {
			return []any{
				StatusRejected,
				StatusRejected,
				request.DecidedBy,
				formatTime(time.Now().UTC()),
				request.Reason,
				request.JobID,
				StatusQueued,
			}
		},
	}
}

func boundedLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}

func nullableString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func newJobID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("job-%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
