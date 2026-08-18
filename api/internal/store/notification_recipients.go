package store

import (
	"context"
)

func (s *Store) ListProjectReviewerEmails(ctx context.Context, projectID string) ([]string, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT DISTINCT lower(u.email)
		FROM projects p
		JOIN users u ON u.organization_id=p.organization_id
		WHERE p.id=$1 AND u.user_type='stakeholder' AND u.role='reviewer' AND u.status='active'
		ORDER BY lower(u.email)`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (s *Store) GetAssignedAssessorEmail(ctx context.Context, projectID, subcategoryID string) (string, error) {
	var email string
	err := s.DB.QueryRow(ctx, `
		SELECT lower(u.email)
		FROM project_subcategory_profiles profile
		JOIN projects p ON p.id=profile.project_id
		JOIN users u ON u.id=profile.assigned_user_id
		WHERE profile.project_id=$1 AND profile.subcategory_id=$2
		  AND u.organization_id=p.organization_id
		  AND u.user_type='stakeholder' AND u.role IN ('org_admin','assessor') AND u.status='active'`, projectID, subcategoryID).Scan(&email)
	return email, err
}

func (s *Store) ListOrganizationEmailsByRoles(ctx context.Context, organizationID string, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}
	rows, err := s.DB.Query(ctx, `
		SELECT DISTINCT lower(email)
		FROM users
		WHERE organization_id=$1 AND user_type='stakeholder' AND status='active' AND role=ANY($2::text[])
		ORDER BY lower(email)`, organizationID, roles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}
