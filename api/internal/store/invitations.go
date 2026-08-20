package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidInvitation = errors.New("invalid invitation")

// InvitationStatus derives the lifecycle state from persisted timestamps.
// Persisted status is intentionally not trusted so a stale response cannot make
// a cancelled, superseded, or expired invitation usable again.
func InvitationStatus(inv Invitation, now time.Time) string {
	switch {
	case inv.AcceptedAt != nil:
		return "accepted"
	case inv.CancelledAt != nil:
		return "cancelled"
	case inv.SupersededAt != nil:
		return "superseded"
	case !inv.ExpiresAt.After(now):
		return "expired"
	default:
		return "pending"
	}
}

func (s *Store) HasActiveOrPendingEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM users WHERE lower(email)=lower($1)
		UNION ALL SELECT 1 FROM invitations
		WHERE lower(email)=lower($1)
		  AND accepted_at IS NULL
		  AND cancelled_at IS NULL
		  AND superseded_at IS NULL
		  AND expires_at>now()
	)`, strings.TrimSpace(email)).Scan(&exists)
	return exists, err
}

func (s *Store) CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	projectIDs := uniqueStrings(invitation.ProjectIDs)
	if len(projectIDs) > 0 {
		if invitation.OrganizationID == nil {
			return Invitation{}, ErrInvalidProjectAccess
		}
		var projectCount int
		if err := s.DB.QueryRow(ctx, `SELECT count(*) FROM projects WHERE organization_id=$1 AND id=ANY($2::uuid[])`, *invitation.OrganizationID, projectIDs).Scan(&projectCount); err != nil {
			return Invitation{}, err
		}
		if projectCount != len(projectIDs) {
			return Invitation{}, ErrInvalidProjectAccess
		}
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Invitation{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at`, invitation.OrganizationID, invitation.Email, invitation.UserType, invitation.Role, invitation.TokenHash, invitation.InvitedBy, invitation.ExpiresAt).Scan(&invitation.ID, &invitation.CreatedAt)
	if err != nil {
		return Invitation{}, err
	}
	for _, projectID := range projectIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO invitation_project_access(invitation_id,project_id) VALUES($1,$2)`, invitation.ID, projectID); err != nil {
			return Invitation{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}
	invitation.ProjectIDs = projectIDs
	invitation.Status = InvitationStatus(invitation, time.Now())
	return invitation, nil
}

func (s *Store) AcceptInvitation(ctx context.Context, tokenHash, name, passwordHash string, now time.Time) (User, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var invitation Invitation
	err = tx.QueryRow(ctx, `SELECT id,organization_id,email,user_type,role,token_hash,invited_by,expires_at,
		accepted_at,cancelled_at,cancelled_by,superseded_at,superseded_by,created_at
		FROM invitations WHERE token_hash=$1 FOR UPDATE`, tokenHash).Scan(
		&invitation.ID, &invitation.OrganizationID, &invitation.Email, &invitation.UserType, &invitation.Role,
		&invitation.TokenHash, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.AcceptedAt,
		&invitation.CancelledAt, &invitation.CancelledBy, &invitation.SupersededAt, &invitation.SupersededBy,
		&invitation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidInvitation
	}
	if err != nil {
		return User{}, err
	}
	if InvitationStatus(invitation, now) != "pending" {
		return User{}, ErrInvalidInvitation
	}
	var user User
	err = tx.QueryRow(ctx, `INSERT INTO users(organization_id,name,email,user_type,role,status,password_hash)
		VALUES($1,$2,$3,$4,$5,'active',$6) RETURNING `+userColumns, invitation.OrganizationID, name, invitation.Email, invitation.UserType, invitation.Role, passwordHash).Scan(&user.ID, &user.OrganizationID, &user.Name, &user.Email, &user.UserType, &user.Role, &user.Status, &user.PasswordHash)
	if err != nil {
		return User{}, err
	}
	if invitation.Role == "auditor" {
		if invitation.OrganizationID == nil {
			return User{}, ErrInvalidProjectAccess
		}
		var accessCount int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM invitation_project_access WHERE invitation_id=$1`, invitation.ID).Scan(&accessCount); err != nil {
			return User{}, err
		}
		if accessCount == 0 {
			return User{}, ErrInvalidProjectAccess
		}
		if _, err := tx.Exec(ctx, `INSERT INTO project_auditor_access(project_id,user_id,granted_by)
			SELECT project_id,$2,$3 FROM invitation_project_access WHERE invitation_id=$1
			ON CONFLICT(project_id,user_id) DO UPDATE SET revoked_at=NULL, granted_by=EXCLUDED.granted_by`, invitation.ID, user.ID, invitation.InvitedBy); err != nil {
			return User{}, err
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE invitations SET accepted_at=$2 WHERE id=$1`, invitation.ID, now); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) ResendInvitation(ctx context.Context, organizationID, invitationID, tokenHash, invitedBy string, expiresAt, now time.Time) (Invitation, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Invitation{}, err
	}
	defer tx.Rollback(ctx)

	var old Invitation
	err = tx.QueryRow(ctx, `SELECT id,organization_id,email,user_type,role,token_hash,invited_by,expires_at,
		accepted_at,cancelled_at,cancelled_by,superseded_at,superseded_by,created_at
		FROM invitations WHERE id=$1 AND organization_id=$2 FOR UPDATE`, invitationID, organizationID).Scan(
		&old.ID, &old.OrganizationID, &old.Email, &old.UserType, &old.Role, &old.TokenHash, &old.InvitedBy,
		&old.ExpiresAt, &old.AcceptedAt, &old.CancelledAt, &old.CancelledBy, &old.SupersededAt,
		&old.SupersededBy, &old.CreatedAt)
	if err != nil {
		return Invitation{}, err
	}
	if InvitationStatus(old, now) != "pending" {
		return Invitation{}, ErrInvitationNotPending
	}
	projectIDs, err := invitationProjectIDs(ctx, tx, old.ID)
	if err != nil {
		return Invitation{}, err
	}

	var replacement Invitation
	err = tx.QueryRow(ctx, `INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at`, old.OrganizationID, old.Email, old.UserType, old.Role, tokenHash, invitedBy, expiresAt).Scan(&replacement.ID, &replacement.CreatedAt)
	if err != nil {
		return Invitation{}, err
	}
	for _, projectID := range projectIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO invitation_project_access(invitation_id,project_id) VALUES($1,$2)`, replacement.ID, projectID); err != nil {
			return Invitation{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE invitations SET superseded_at=$2,superseded_by=$3 WHERE id=$1`, old.ID, now, replacement.ID); err != nil {
		return Invitation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}
	replacement.OrganizationID = old.OrganizationID
	replacement.Email = old.Email
	replacement.UserType = old.UserType
	replacement.Role = old.Role
	replacement.TokenHash = tokenHash
	replacement.InvitedBy = invitedBy
	replacement.ExpiresAt = expiresAt
	replacement.ProjectIDs = projectIDs
	replacement.Status = InvitationStatus(replacement, now)
	return replacement, nil
}

func (s *Store) CancelInvitation(ctx context.Context, organizationID, invitationID, cancelledBy string, now time.Time) (Invitation, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Invitation{}, err
	}
	defer tx.Rollback(ctx)

	var invitation Invitation
	err = tx.QueryRow(ctx, `SELECT id,organization_id,email,user_type,role,token_hash,invited_by,expires_at,
		accepted_at,cancelled_at,cancelled_by,superseded_at,superseded_by,created_at
		FROM invitations WHERE id=$1 AND organization_id=$2 FOR UPDATE`, invitationID, organizationID).Scan(
		&invitation.ID, &invitation.OrganizationID, &invitation.Email, &invitation.UserType, &invitation.Role,
		&invitation.TokenHash, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.AcceptedAt,
		&invitation.CancelledAt, &invitation.CancelledBy, &invitation.SupersededAt, &invitation.SupersededBy,
		&invitation.CreatedAt)
	if err != nil {
		return Invitation{}, err
	}
	if InvitationStatus(invitation, now) != "pending" {
		return Invitation{}, ErrInvitationNotPending
	}
	if _, err := tx.Exec(ctx, `UPDATE invitations SET cancelled_at=$2,cancelled_by=$3 WHERE id=$1`, invitation.ID, now, cancelledBy); err != nil {
		return Invitation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}
	invitation.CancelledAt = &now
	invitation.CancelledBy = &cancelledBy
	invitation.Status = "cancelled"
	invitation.ProjectIDs, err = invitationProjectIDs(ctx, s.DB, invitation.ID)
	if err != nil {
		return Invitation{}, err
	}
	return invitation, nil
}

type invitationProjectQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func invitationProjectIDs(ctx context.Context, querier invitationProjectQuerier, invitationID string) ([]string, error) {
	rows, err := querier.Query(ctx, `SELECT project_id FROM invitation_project_access WHERE invitation_id=$1 ORDER BY project_id`, invitationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
