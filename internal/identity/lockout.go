package identity

import (
	"context"
	"fmt"

	"koji/internal/caps"
)

func (s *Store) preventDisableLockout(ctx context.Context, id int64) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if err := s.preventFinalSuperAdminDisable(ctx, user); err != nil {
		return err
	}
	return s.preventFinalIdentityManagerDisable(ctx, id)
}

func (s *Store) preventFinalSuperAdminDisable(ctx context.Context, user User) error {
	if !user.IsSuperAdmin {
		return nil
	}
	count, err := s.activeSuperAdminCount(ctx)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrSelfLockout
	}
	return nil
}

func (s *Store) preventFinalIdentityManagerDisable(ctx context.Context, id int64) error {
	hasManage, err := s.userHasCapability(ctx, id, string(caps.IdentityUsersManage))
	if err != nil {
		return err
	}
	if !hasManage {
		return nil
	}
	return s.requireMultipleIdentityManagers(ctx)
}

func (s *Store) preventManageCapabilityLockout(ctx context.Context, id int64) error {
	hasManage, err := s.userHasCapability(ctx, id, string(caps.IdentityUsersManage))
	if err != nil {
		return err
	}
	if !hasManage {
		return nil
	}
	return s.requireMultipleIdentityManagers(ctx)
}

func (s *Store) requireMultipleIdentityManagers(ctx context.Context) error {
	count, err := s.activeIdentityManagerCount(ctx)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrSelfLockout
	}
	return nil
}

func (s *Store) activeSuperAdminCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_super_admin = 1 AND disabled_at IS NULL").Scan(&count); err != nil {
		return 0, fmt.Errorf("count super admins: %w", err)
	}
	return count, nil
}

func (s *Store) activeIdentityManagerCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM users
JOIN user_capabilities ON user_capabilities.user_id = users.id
WHERE users.disabled_at IS NULL
	AND user_capabilities.capability_name = ?`, string(caps.IdentityUsersManage)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count identity managers: %w", err)
	}
	return count, nil
}

func (s *Store) userHasCapability(ctx context.Context, id int64, capability string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_capabilities
WHERE user_id = ? AND capability_name = ?`, id, capability).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("lookup capability: %w", err)
	}
	return count > 0, nil
}
