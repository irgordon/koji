package identity

import (
	"context"
	"fmt"

	"koji/internal/caps"
)

func (s *Store) ListAvailableCapabilities(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT name FROM capabilities ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (s *Store) ListUserCapabilities(ctx context.Context, id int64) ([]string, error) {
	if _, err := s.GetUser(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT capability_name
FROM user_capabilities
WHERE user_id = ?
ORDER BY capability_name`, id)
	if err != nil {
		return nil, fmt.Errorf("list user capabilities: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (s *Store) GrantCapability(ctx context.Context, id int64, capability string) ([]string, error) {
	if err := validateCapability(capability); err != nil {
		return nil, err
	}
	if _, err := s.GetUser(ctx, id); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO user_capabilities (user_id, capability_name)
VALUES (?, ?)`, id, capability); err != nil {
		return nil, fmt.Errorf("grant capability: %w", err)
	}
	return s.ListUserCapabilities(ctx, id)
}

func (s *Store) RevokeCapability(ctx context.Context, id int64, capability string) ([]string, error) {
	if err := validateCapability(capability); err != nil {
		return nil, err
	}
	if err := s.preventIdentityManagerLockout(ctx, id, capability); err != nil {
		return nil, err
	}
	if err := s.deleteUserCapability(ctx, id, capability); err != nil {
		return nil, err
	}
	return s.ListUserCapabilities(ctx, id)
}

func (s *Store) preventIdentityManagerLockout(ctx context.Context, id int64, capability string) error {
	if capability != string(caps.IdentityUsersManage) {
		return nil
	}
	return s.preventManageCapabilityLockout(ctx, id)
}

func (s *Store) deleteUserCapability(ctx context.Context, id int64, capability string) error {
	result, err := s.db.ExecContext(ctx, `
DELETE FROM user_capabilities
WHERE user_id = ? AND capability_name = ?`, id, capability)
	if err != nil {
		return fmt.Errorf("revoke capability: %w", err)
	}
	return requireAffected(result)
}

func validateCapability(capability string) error {
	for _, known := range caps.All() {
		if string(known) == capability {
			return nil
		}
	}
	return ErrInvalidCapability
}
