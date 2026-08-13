package store

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const userColumns = `id,organization_id,name,email,user_type,role,status,password_hash`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var user User
	var passwordHash sql.NullString
	err := row.Scan(&user.ID, &user.OrganizationID, &user.Name, &user.Email, &user.UserType, &user.Role, &user.Status, &passwordHash)
	user.PasswordHash = passwordHash.String
	return user, err
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(s.DB.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email)=$1`, strings.ToLower(strings.TrimSpace(email))))
}

func (s *Store) FindUserBySessionHash(ctx context.Context, tokenHash string) (User, Session, error) {
	var user User
	var session Session
	var passwordHash sql.NullString
	err := s.DB.QueryRow(ctx, `SELECT u.id,u.organization_id,u.name,u.email,u.user_type,u.role,u.status,u.password_hash,
		s.id,s.user_id,s.token_hash,s.expires_at,s.last_seen_at,s.created_at
		FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1`, tokenHash).Scan(
		&user.ID, &user.OrganizationID, &user.Name, &user.Email, &user.UserType, &user.Role, &user.Status, &passwordHash,
		&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt, &session.LastSeenAt, &session.CreatedAt,
	)
	user.PasswordHash = passwordHash.String
	if err == nil {
		_, _ = s.DB.Exec(ctx, `UPDATE sessions SET last_seen_at=now() WHERE id=$1`, session.ID)
	}
	return user, session, err
}

func (s *Store) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO sessions(user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, userID, tokenHash, expiresAt)
	return err
}

func (s *Store) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, tokenHash)
	return err
}

func (s *Store) HasCounselorAdmin(ctx context.Context) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE user_type='counselor' AND role='counselor_admin')`).Scan(&exists)
	return exists, err
}

func (s *Store) CreateCounselorAdmin(ctx context.Context, email, passwordHash string) (User, error) {
	name := strings.Split(email, "@")[0]
	return scanUser(s.DB.QueryRow(ctx, `INSERT INTO users(name,email,user_type,role,status,password_hash)
		VALUES ($1,$2,'counselor','counselor_admin','active',$3) RETURNING `+userColumns, name, email, passwordHash))
}
