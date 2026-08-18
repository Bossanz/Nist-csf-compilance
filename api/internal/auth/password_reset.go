package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidCurrentPassword = errors.New("invalid current password")

type PasswordRepository interface {
	FindActiveUserByEmail(context.Context, string) (store.User, error)
	FindActiveUserByID(context.Context, string) (store.User, error)
	CreatePasswordResetToken(context.Context, string, string, time.Time) error
	ConsumePasswordResetToken(context.Context, string, string, time.Time) (store.User, error)
	ChangePassword(context.Context, string, string) error
}

type PasswordService struct {
	repository PasswordRepository
	now        func() time.Time
}

func NewPasswordService(repository PasswordRepository, now func() time.Time) *PasswordService {
	return &PasswordService{repository: repository, now: now}
}

func (s *PasswordService) Request(ctx context.Context, email string) (store.User, string, bool, error) {
	user, err := s.repository.FindActiveUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, pgx.ErrNoRows) {
		return store.User{}, "", false, nil
	}
	if err != nil || user.Status != "active" {
		return store.User{}, "", false, err
	}
	rawToken, tokenHash, err := NewToken()
	if err != nil {
		return store.User{}, "", false, err
	}
	if err := s.repository.CreatePasswordResetToken(ctx, user.ID, tokenHash, s.now().Add(30*time.Minute)); err != nil {
		return store.User{}, "", false, err
	}
	return user, rawToken, true, nil
}

func (s *PasswordService) Confirm(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = s.repository.ConsumePasswordResetToken(ctx, HashToken(rawToken), passwordHash, s.now())
	return err
}

func (s *PasswordService) Change(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}
	user, err := s.repository.FindActiveUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidCurrentPassword
	}
	if err != nil {
		return err
	}
	if !CheckPassword(user.PasswordHash, currentPassword) {
		return ErrInvalidCurrentPassword
	}
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repository.ChangePassword(ctx, userID, passwordHash)
}
