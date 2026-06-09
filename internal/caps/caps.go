package caps

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Capability string

const (
	HostMetricsRead     Capability = "host.metrics.read"
	HostDiskRead        Capability = "host.disk.read"
	HostServicesRead    Capability = "host.services.read"
	HostProcessesRead   Capability = "host.processes.read"
	HostServicesControl Capability = "host.services.control"
	AuditEventsRead     Capability = "audit.events.read"
	JobsRead            Capability = "jobs.read"
	JobsApprove         Capability = "jobs.approve"
)

var ErrCapabilityDenied = errors.New("capability denied")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Require(ctx context.Context, userID int64, capability Capability) error {
	allowed, err := s.UserHasCapability(ctx, userID, capability)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrCapabilityDenied
	}
	return nil
}

func (s *Store) UserHasCapability(ctx context.Context, userID int64, capability Capability) (bool, error) {
	if userID <= 0 {
		return false, nil
	}

	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_capabilities
WHERE user_id = ? AND capability_name = ?`, userID, string(capability)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("lookup user capability: %w", err)
	}
	return count > 0, nil
}
