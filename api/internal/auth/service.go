package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"compliance/api/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

type Repository interface {
	FindUserByEmail(context.Context, string) (store.User, error)
	FindUserBySessionHash(context.Context, string) (store.User, store.Session, error)
	CreateSession(context.Context, string, string, time.Time) error
	RevokeSession(context.Context, string) error
	HasCounselorAdmin(context.Context) (bool, error)
	CreateCounselorAdmin(context.Context, string, string) (store.User, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository, now func() time.Time) *Service {
	return &Service{repository: repository, now: now}
}

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (s *Service) Login(ctx context.Context, email, password string) (store.User, string, error) {
	user, err := s.repository.FindUserByEmail(ctx, NormalizeEmail(email))
	if err != nil || user.Status != "active" || !CheckPassword(user.PasswordHash, password) {
		return store.User{}, "", ErrInvalidCredentials
	}
	rawToken, tokenHash, err := NewToken()
	if err != nil {
		return store.User{}, "", err
	}
	if err := s.repository.CreateSession(ctx, user.ID, tokenHash, s.now().Add(12*time.Hour)); err != nil {
		return store.User{}, "", err
	}
	return user, rawToken, nil
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (store.User, error) {
	if rawToken == "" {
		return store.User{}, ErrUnauthenticated
	}
	user, session, err := s.repository.FindUserBySessionHash(ctx, HashToken(rawToken))
	if err != nil || user.Status != "active" || !session.ExpiresAt.After(s.now()) {
		return store.User{}, ErrUnauthenticated
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.repository.RevokeSession(ctx, HashToken(rawToken))
}

func (s *Service) Bootstrap(ctx context.Context, email, password string) error {
	hasAdmin, err := s.repository.HasCounselorAdmin(ctx)
	if err != nil || hasAdmin {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.repository.CreateCounselorAdmin(ctx, NormalizeEmail(email), hash)
	return err
}
