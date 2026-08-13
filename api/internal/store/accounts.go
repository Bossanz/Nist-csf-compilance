package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func scanUsers(rows pgx.Rows) ([]User, error) {
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) ListOrganizationUsers(ctx context.Context, organizationID string) ([]User, error) {
	rows, err := s.DB.Query(ctx, `SELECT `+userColumns+` FROM users WHERE organization_id=$1 ORDER BY lower(name),lower(email)`, organizationID)
	if err != nil {
		return nil, err
	}
	return scanUsers(rows)
}
func (s *Store) ListCounselors(ctx context.Context) ([]User, error) {
	rows, err := s.DB.Query(ctx, `SELECT `+userColumns+` FROM users WHERE user_type='counselor' ORDER BY lower(name),lower(email)`)
	if err != nil {
		return nil, err
	}
	return scanUsers(rows)
}

func (s *Store) UpdateOrganizationUser(ctx context.Context, organizationID, userID, role, status string) (User, error) {
	return scanUser(s.DB.QueryRow(ctx, `UPDATE users SET role=$3,status=$4,updated_at=now() WHERE id=$1 AND organization_id=$2 AND user_type='stakeholder' RETURNING `+userColumns, userID, organizationID, role, status))
}
func (s *Store) UpdateCounselor(ctx context.Context, userID, role, status string) (User, error) {
	return scanUser(s.DB.QueryRow(ctx, `UPDATE users SET role=$2,status=$3,updated_at=now() WHERE id=$1 AND user_type='counselor' RETURNING `+userColumns, userID, role, status))
}
func (s *Store) RevokeUserSessions(ctx context.Context, userID string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, userID)
	return err
}
