package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type fakePasswordRepository struct {
	user             store.User
	findErr          error
	createdUserID    string
	createdTokenHash string
	createdExpiry    time.Time
	consumeTokenHash string
	consumePassword  string
	consumeErr       error
	changedUserID    string
	changedPassword  string
}

func (f *fakePasswordRepository) FindActiveUserByEmail(context.Context, string) (store.User, error) {
	return f.user, f.findErr
}

func (f *fakePasswordRepository) FindActiveUserByID(context.Context, string) (store.User, error) {
	return f.user, f.findErr
}

func (f *fakePasswordRepository) CreatePasswordResetToken(_ context.Context, userID, tokenHash string, expiresAt time.Time) error {
	f.createdUserID, f.createdTokenHash, f.createdExpiry = userID, tokenHash, expiresAt
	return nil
}

func (f *fakePasswordRepository) ConsumePasswordResetToken(_ context.Context, tokenHash, passwordHash string, _ time.Time) (store.User, error) {
	f.consumeTokenHash, f.consumePassword = tokenHash, passwordHash
	return f.user, f.consumeErr
}

func (f *fakePasswordRepository) ChangePassword(_ context.Context, userID, passwordHash string) error {
	f.changedUserID, f.changedPassword = userID, passwordHash
	return nil
}

func TestRequestPasswordResetReturnsNoUserForUnknownEmail(t *testing.T) {
	repo := &fakePasswordRepository{findErr: pgx.ErrNoRows}
	service := NewPasswordService(repo, func() time.Time { return time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC) })

	user, rawToken, found, err := service.Request(context.Background(), "unknown@example.com")
	if err != nil || found || rawToken != "" || user.ID != "" {
		t.Fatalf("expected non-enumerating empty result, user=%#v token=%q found=%v err=%v", user, rawToken, found, err)
	}
}

func TestRequestPasswordResetCreatesHashedTokenForActiveUser(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	repo := &fakePasswordRepository{user: store.User{ID: "user-1", Email: "person@example.com", Status: "active"}}
	service := NewPasswordService(repo, func() time.Time { return now })

	user, rawToken, found, err := service.Request(context.Background(), " PERSON@example.com ")
	if err != nil || !found || user.ID != "user-1" || rawToken == "" {
		t.Fatalf("unexpected request result: user=%#v token=%q found=%v err=%v", user, rawToken, found, err)
	}
	if repo.createdUserID != "user-1" || repo.createdTokenHash != HashToken(rawToken) || !repo.createdExpiry.Equal(now.Add(30*time.Minute)) {
		t.Fatalf("unexpected token persistence: %#v", repo)
	}
}

func TestConfirmPasswordResetRejectsWeakPassword(t *testing.T) {
	repo := &fakePasswordRepository{}
	service := NewPasswordService(repo, time.Now)
	if err := service.Confirm(context.Background(), "raw-token", "short"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected weak password error, got %v", err)
	}
	if repo.consumeTokenHash != "" {
		t.Fatal("weak password should not consume a reset token")
	}
}

func TestConfirmPasswordResetPassesOnlyTokenHashToRepository(t *testing.T) {
	repo := &fakePasswordRepository{user: store.User{ID: "user-1"}}
	service := NewPasswordService(repo, time.Now)
	if err := service.Confirm(context.Background(), "raw-token", "NewPassword!2026"); err != nil {
		t.Fatal(err)
	}
	if repo.consumeTokenHash != HashToken("raw-token") || repo.consumeTokenHash == "raw-token" || !CheckPassword(repo.consumePassword, "NewPassword!2026") {
		t.Fatal("reset service did not hash token/password before persistence")
	}
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	oldHash, _ := HashPassword("OldPassword!2026")
	repo := &fakePasswordRepository{user: store.User{ID: "user-1", PasswordHash: oldHash}}
	service := NewPasswordService(repo, time.Now)
	if err := service.Change(context.Background(), "user-1", "wrong", "NewPassword!2026"); !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("expected invalid current password, got %v", err)
	}
	if repo.changedUserID != "" {
		t.Fatal("wrong current password should not update the repository")
	}
}

func TestChangePasswordUpdatesHashForCorrectCurrentPassword(t *testing.T) {
	oldHash, _ := HashPassword("OldPassword!2026")
	repo := &fakePasswordRepository{user: store.User{ID: "user-1", PasswordHash: oldHash}}
	service := NewPasswordService(repo, time.Now)
	if err := service.Change(context.Background(), "user-1", "OldPassword!2026", "NewPassword!2026"); err != nil {
		t.Fatal(err)
	}
	if repo.changedUserID != "user-1" || !CheckPassword(repo.changedPassword, "NewPassword!2026") {
		t.Fatal("expected a new password hash for the authenticated user")
	}
}
