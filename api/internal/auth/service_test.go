package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	user              store.User
	session           store.Session
	findUserErr       error
	findSessionErr    error
	createdTokenHash  string
	createdExpiry     time.Time
	revokedTokenHash  string
	hasAdmin          bool
	adminCreates      int
	adminEmail        string
	adminPasswordHash string
}

func (f *fakeRepository) FindUserByEmail(context.Context, string) (store.User, error) {
	return f.user, f.findUserErr
}
func (f *fakeRepository) FindUserBySessionHash(context.Context, string) (store.User, store.Session, error) {
	return f.user, f.session, f.findSessionErr
}
func (f *fakeRepository) CreateSession(_ context.Context, _ string, tokenHash string, expires time.Time) error {
	f.createdTokenHash, f.createdExpiry = tokenHash, expires
	return nil
}
func (f *fakeRepository) RevokeSession(_ context.Context, tokenHash string) error {
	f.revokedTokenHash = tokenHash
	return nil
}
func (f *fakeRepository) HasCounselorAdmin(context.Context) (bool, error) { return f.hasAdmin, nil }
func (f *fakeRepository) CreateCounselorAdmin(_ context.Context, email, passwordHash string) (store.User, error) {
	f.adminCreates++
	f.adminEmail, f.adminPasswordHash = email, passwordHash
	return store.User{}, nil
}

func TestLoginReturnsSameErrorForUnknownAndWrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	cases := []fakeRepository{
		{findUserErr: pgx.ErrNoRows},
		{user: store.User{PasswordHash: hash, Status: "active"}},
	}
	for i := range cases {
		_, _, err := NewService(&cases[i], time.Now).Login(context.Background(), "USER@example.com ", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("case %d: expected invalid credentials, got %v", i, err)
		}
	}
}

func TestLoginCreatesHashedSessionForActiveUser(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	hash, _ := HashPassword("correct")
	repo := &fakeRepository{user: store.User{ID: "user-1", PasswordHash: hash, Status: "active"}}
	_, rawToken, err := NewService(repo, func() time.Time { return now }).Login(context.Background(), "USER@example.com ", "correct")
	if err != nil {
		t.Fatal(err)
	}
	if rawToken == "" || repo.createdTokenHash != HashToken(rawToken) {
		t.Fatal("session was not stored as a token hash")
	}
	if !repo.createdExpiry.Equal(now.Add(12 * time.Hour)) {
		t.Fatalf("unexpected expiry: %v", repo.createdExpiry)
	}
}

func TestAuthenticateRejectsExpiredOrDisabledSession(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	cases := []fakeRepository{
		{user: store.User{Status: "active"}, session: store.Session{ExpiresAt: now.Add(-time.Minute)}},
		{user: store.User{Status: "disabled"}, session: store.Session{ExpiresAt: now.Add(time.Hour)}},
	}
	for i := range cases {
		_, err := NewService(&cases[i], func() time.Time { return now }).Authenticate(context.Background(), "raw-token")
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("case %d: expected unauthenticated, got %v", i, err)
		}
	}
}

func TestBootstrapCreatesAdminOnlyWhenNoneExists(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo, time.Now)
	if err := service.Bootstrap(context.Background(), " ADMIN@example.com ", "secret-password"); err != nil {
		t.Fatal(err)
	}
	repo.hasAdmin = true
	if err := service.Bootstrap(context.Background(), "changed@example.com", "changed-password"); err != nil {
		t.Fatal(err)
	}
	if repo.adminCreates != 1 || repo.adminEmail != "admin@example.com" || !CheckPassword(repo.adminPasswordHash, "secret-password") {
		t.Fatalf("unexpected bootstrap state: %#v", repo)
	}
}
