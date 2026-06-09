package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	ActionLogin          = "auth.login"
	ActionLogout         = "auth.logout"
	ActionBootstrap      = "auth.bootstrap"
	ActionCapabilityDeny = "capability.denied"
	ActionCapabilityPass = "capability.bypass"
	ActionServiceControl = "service.control"
	ActionProcessList    = "process.list"
	ActionJobCreated     = "job.created"
	ActionJobViewed      = "job.viewed"
	ActionJobStatus      = "job.status_changed"
	ActionJobApproved    = "job.approved"
	ActionJobRejected    = "job.rejected"
	ActionJobApproveDeny = "job.approval_denied"
	ActionJobStarted     = "job.started"
	ActionJobNotImpl     = "job.not_implemented"
	ActionJobFailed      = "job.failed"
)

const (
	OutcomeSuccess  = "success"
	OutcomeFailure  = "failure"
	OutcomeDenied   = "denied"
	OutcomeAccepted = "accepted"
	OutcomeRejected = "rejected"
)

const (
	DefaultRecentLimit = 50
	MaxRecentLimit     = 100
)

type Event struct {
	UserID     *int64
	Action     string
	Target     string
	Outcome    string
	ReasonCode string
	RequestID  string
	RemoteAddr string
	DevBypass  bool
}

type ActivityEvent struct {
	Timestamp  string `json:"timestamp"`
	Action     string `json:"action"`
	Target     string `json:"target"`
	Outcome    string `json:"outcome"`
	ReasonCode string `json:"reason_code"`
	RequestID  string `json:"request_id"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Record(ctx context.Context, event Event) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO audit_events (
	actor,
	action,
	target,
	status,
	message,
	user_id,
	outcome,
	reason_code,
	request_id,
	remote_addr,
	dev_bypass,
	created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		actorValue(event.UserID),
		event.Action,
		event.Target,
		event.Outcome,
		event.ReasonCode,
		event.UserID,
		event.Outcome,
		event.ReasonCode,
		event.RequestID,
		event.RemoteAddr,
		boolInt(event.DevBypass),
		formatTime(time.Now().UTC()),
	)
	if err != nil {
		return fmt.Errorf("record audit event: %w", err)
	}
	return nil
}

func (s *Store) ListRecent(ctx context.Context, limit int) ([]ActivityEvent, error) {
	eventsLimit := boundedLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
SELECT created_at, action, target, outcome, reason_code, request_id
FROM audit_events
ORDER BY id DESC
LIMIT ?`, eventsLimit)
	if err != nil {
		return nil, fmt.Errorf("list recent audit events: %w", err)
	}
	defer rows.Close()

	return scanActivityEvents(rows)
}

func scanActivityEvents(rows *sql.Rows) ([]ActivityEvent, error) {
	events := make([]ActivityEvent, 0, DefaultRecentLimit)
	for rows.Next() {
		event, err := scanActivityEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan audit events: %w", err)
	}
	return events, nil
}

func scanActivityEvent(rows *sql.Rows) (ActivityEvent, error) {
	var event ActivityEvent
	if err := rows.Scan(
		&event.Timestamp,
		&event.Action,
		&event.Target,
		&event.Outcome,
		&event.ReasonCode,
		&event.RequestID,
	); err != nil {
		return ActivityEvent{}, fmt.Errorf("scan audit event: %w", err)
	}
	return event, nil
}

func boundedLimit(limit int) int {
	if limit <= 0 {
		return DefaultRecentLimit
	}
	if limit > MaxRecentLimit {
		return MaxRecentLimit
	}
	return limit
}

func actorValue(userID *int64) string {
	if userID == nil {
		return "anonymous"
	}
	return fmt.Sprintf("user:%d", *userID)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
