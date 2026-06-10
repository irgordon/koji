package observability

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	JobsCreatedTotal                = "jobs_created_total"
	JobsApprovedTotal               = "jobs_approved_total"
	JobsRejectedTotal               = "jobs_rejected_total"
	JobsCompletedTotal              = "jobs_completed_total"
	JobsFailedTotal                 = "jobs_failed_total"
	JobsClaimedTotal                = "jobs_claimed_total"
	WorkerPollsTotal                = "worker_polls_total"
	WorkerErrorsTotal               = "worker_errors_total"
	AgentRPCRequestsTotal           = "agent_rpc_requests_total"
	AgentRPCFailuresTotal           = "agent_rpc_failures_total"
	AuthLoginSuccessTotal           = "auth_login_success_total"
	AuthLoginFailureTotal           = "auth_login_failure_total"
	AuditWritesTotal                = "audit_writes_total"
	AuditWriteFailuresTotal         = "audit_write_failures_total"
	ReadinessChecksTotal            = "readiness_checks_total"
	ReadinessDBFailuresTotal        = "readiness_db_failures_total"
	ReadinessAgentDegradedTotal     = "readiness_agent_degraded_total"
	ReadinessMigrationFailuresTotal = "readiness_migration_failures_total"
)

var defaultRegistry = NewRegistry()

type Registry struct {
	counters sync.Map
}

type Snapshot struct {
	Counters     map[string]uint64 `json:"counters"`
	JobsByStatus map[string]int64  `json:"jobs_by_status"`
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Inc(name string) {
	if r == nil {
		return
	}
	r.counter(name).Add(1)
}

func (r *Registry) Value(name string) uint64 {
	if r == nil {
		return 0
	}
	return r.counter(name).Load()
}

func (r *Registry) Snapshot(ctx context.Context, db *sql.DB) (Snapshot, error) {
	statuses, err := jobStatusCounts(ctx, db)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Counters:     r.counterValues(),
		JobsByStatus: statuses,
	}, nil
}

func (r *Registry) ResetForTest() {
	if r == nil {
		return
	}
	r.counters.Range(func(key any, value any) bool {
		if counter, ok := value.(*atomic.Uint64); ok {
			counter.Store(0)
		}
		return true
	})
}

func (r *Registry) counter(name string) *atomic.Uint64 {
	value, _ := r.counters.LoadOrStore(name, &atomic.Uint64{})
	return value.(*atomic.Uint64)
}

func (r *Registry) counterValues() map[string]uint64 {
	counters := make(map[string]uint64)
	for _, name := range CounterNames() {
		counters[name] = r.Value(name)
	}
	return counters
}

func CounterNames() []string {
	names := []string{
		JobsCreatedTotal,
		JobsApprovedTotal,
		JobsRejectedTotal,
		JobsCompletedTotal,
		JobsFailedTotal,
		JobsClaimedTotal,
		WorkerPollsTotal,
		WorkerErrorsTotal,
		AgentRPCRequestsTotal,
		AgentRPCFailuresTotal,
		AuthLoginSuccessTotal,
		AuthLoginFailureTotal,
		AuditWritesTotal,
		AuditWriteFailuresTotal,
		ReadinessChecksTotal,
		ReadinessDBFailuresTotal,
		ReadinessAgentDegradedTotal,
		ReadinessMigrationFailuresTotal,
	}
	sort.Strings(names)
	return names
}

func jobStatusCounts(ctx context.Context, db *sql.DB) (map[string]int64, error) {
	rows, err := db.QueryContext(ctx, "SELECT status, COUNT(*) FROM jobs GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("query job status metrics: %w", err)
	}
	defer rows.Close()

	return scanJobStatusCounts(rows)
}

func scanJobStatusCounts(rows *sql.Rows) (map[string]int64, error) {
	result := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan job status metrics: %w", err)
		}
		result[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read job status metrics: %w", err)
	}
	return result, nil
}
