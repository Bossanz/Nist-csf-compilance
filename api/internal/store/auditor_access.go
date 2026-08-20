package store

import "context"

func (s *Store) HasActiveProjectAuditorAccess(ctx context.Context, projectID, userID string) (bool, error) {
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1
		FROM project_auditor_access access
		JOIN users u ON u.id=access.user_id
		WHERE access.project_id=$1
		  AND access.user_id=$2
		  AND access.revoked_at IS NULL
		  AND u.user_type='stakeholder'
		  AND u.role='auditor'
		  AND u.status='active'
	)`, projectID, userID).Scan(&allowed)
	return allowed, err
}
