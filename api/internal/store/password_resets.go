package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) FindActiveUserByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(s.DB.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email)=lower($1) AND status='active'`, strings.TrimSpace(email)))
}

func (s *Store) CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id=$1 AND used_at IS NULL`, userID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO password_reset_tokens(user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, userID, tokenHash, expiresAt)
		return err
	})
}

func (s *Store) ConsumePasswordResetToken(ctx context.Context, tokenHash, passwordHash string, now time.Time) (User, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var tokenID, userID string
	var expiresAt time.Time
	var usedAt *time.Time
	if err := tx.QueryRow(ctx, `SELECT id,user_id,expires_at,used_at FROM password_reset_tokens WHERE token_hash=$1 FOR UPDATE`, tokenHash).Scan(&tokenID, &userID, &expiresAt, &usedAt); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrInvalidPasswordResetToken
		}
		return User{}, err
	}
	if usedAt != nil || !expiresAt.After(now) {
		return User{}, ErrInvalidPasswordResetToken
	}

	var user User
	if err := tx.QueryRow(ctx, `UPDATE users SET password_hash=$2,updated_at=now() WHERE id=$1 RETURNING `+userColumns, userID, passwordHash).Scan(
		&user.ID, &user.OrganizationID, &user.Name, &user.Email, &user.UserType, &user.Role, &user.Status, &user.PasswordHash,
	); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrInvalidPasswordResetToken
		}
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at=$2 WHERE id=$1`, tokenID, now); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, userID); err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) ChangePassword(ctx context.Context, userID, passwordHash string) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		command, err := tx.Exec(ctx, `UPDATE users SET password_hash=$2,updated_at=now() WHERE id=$1`, userID, passwordHash)
		if err != nil {
			return err
		}
		if command.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		_, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, userID)
		return err
	})
}
