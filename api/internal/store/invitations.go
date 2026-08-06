package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidInvitation = errors.New("invalid invitation")

func (s *Store) HasActiveOrPendingEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM users WHERE lower(email)=lower($1)
		UNION ALL SELECT 1 FROM invitations WHERE lower(email)=lower($1) AND accepted_at IS NULL AND expires_at>now()
	)`, strings.TrimSpace(email)).Scan(&exists)
	return exists, err
}

func (s *Store) CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	err := s.DB.QueryRow(ctx, `INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at`, invitation.OrganizationID, invitation.Email, invitation.UserType, invitation.Role, invitation.TokenHash, invitation.InvitedBy, invitation.ExpiresAt).Scan(&invitation.ID, &invitation.CreatedAt)
	return invitation, err
}

func (s *Store) AcceptInvitation(ctx context.Context, tokenHash, name, passwordHash string, now time.Time) (User, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)
	var invitation Invitation
	err = tx.QueryRow(ctx, `SELECT id,organization_id,email,user_type,role,token_hash,invited_by,expires_at,accepted_at,created_at
		FROM invitations WHERE token_hash=$1 FOR UPDATE`, tokenHash).Scan(&invitation.ID, &invitation.OrganizationID, &invitation.Email, &invitation.UserType, &invitation.Role, &invitation.TokenHash, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.AcceptedAt, &invitation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && (invitation.AcceptedAt != nil || !invitation.ExpiresAt.After(now)) {
		return User{}, ErrInvalidInvitation
	}
	if err != nil {
		return User{}, err
	}
	var user User
	err = tx.QueryRow(ctx, `INSERT INTO users(organization_id,name,email,user_type,role,status,password_hash)
		VALUES($1,$2,$3,$4,$5,'active',$6) RETURNING `+userColumns, invitation.OrganizationID, name, invitation.Email, invitation.UserType, invitation.Role, passwordHash).Scan(&user.ID, &user.OrganizationID, &user.Name, &user.Email, &user.UserType, &user.Role, &user.Status, &user.PasswordHash)
	if err != nil {
		return User{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE invitations SET accepted_at=$2 WHERE id=$1`, invitation.ID, now); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}
