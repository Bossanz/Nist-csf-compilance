package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestPasswordResetTokenIsSingleUseAndRevokesSessions(t *testing.T) {
	data, ctx, userID, _ := passwordResetFixture(t)
	now := time.Now().UTC()
	if err := data.CreatePasswordResetToken(ctx, userID, "reset-hash", now.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := data.CreateSession(ctx, userID, "session-one", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := data.CreateSession(ctx, userID, "session-two", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	if _, err := data.ConsumePasswordResetToken(ctx, "reset-hash", "new-password-hash", now); err != nil {
		t.Fatalf("consume reset token: %v", err)
	}
	if _, err := data.ConsumePasswordResetToken(ctx, "reset-hash", "another-password-hash", now); err != ErrInvalidPasswordResetToken {
		t.Fatalf("expected single-use error, got %v", err)
	}
	var passwordHash string
	if err := data.DB.QueryRow(ctx, "SELECT password_hash FROM users WHERE id=$1", userID).Scan(&passwordHash); err != nil {
		t.Fatal(err)
	}
	if passwordHash != "new-password-hash" {
		t.Fatalf("expected updated password hash, got %q", passwordHash)
	}
	var sessionCount int
	if err := data.DB.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id=$1", userID).Scan(&sessionCount); err != nil {
		t.Fatal(err)
	}
	if sessionCount != 0 {
		t.Fatalf("expected sessions to be revoked, got %d", sessionCount)
	}
}

func TestPasswordResetTokenExpires(t *testing.T) {
	data, ctx, userID, _ := passwordResetFixture(t)
	now := time.Now().UTC()
	if err := data.CreatePasswordResetToken(ctx, userID, "expired-reset-hash", now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := data.ConsumePasswordResetToken(ctx, "expired-reset-hash", "new-password-hash", now); err != ErrInvalidPasswordResetToken {
		t.Fatalf("expected expired token error, got %v", err)
	}
}

func TestChangePasswordRevokesAllSessions(t *testing.T) {
	data, ctx, userID, _ := passwordResetFixture(t)
	now := time.Now().UTC()
	if err := data.CreateSession(ctx, userID, "change-session-one", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := data.CreateSession(ctx, userID, "change-session-two", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := data.ChangePassword(ctx, userID, "changed-password-hash"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	var sessionCount int
	if err := data.DB.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id=$1", userID).Scan(&sessionCount); err != nil {
		t.Fatal(err)
	}
	if sessionCount != 0 {
		t.Fatalf("expected sessions to be revoked, got %d", sessionCount)
	}
}

func TestFindActiveUserByEmailIgnoresDisabledUsers(t *testing.T) {
	data, ctx, userID, email := passwordResetFixture(t)
	if _, err := data.DB.Exec(ctx, "UPDATE users SET status='disabled' WHERE id=$1", userID); err != nil {
		t.Fatal(err)
	}
	if _, err := data.FindActiveUserByEmail(ctx, email); err == nil {
		t.Fatal("expected disabled user to be excluded")
	}
}

func passwordResetFixture(t *testing.T) (*Store, context.Context, string, string) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := "reset-user-" + suffix + "@example.com"
	var organizationID, userID string
	if err := data.DB.QueryRow(ctx, `INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, "Password Reset Test "+suffix).Scan(&organizationID); err != nil {
		data.Close()
		t.Fatal(err)
	}
	if err := data.DB.QueryRow(ctx, `INSERT INTO users(organization_id,name,email,user_type,role,status,password_hash)
		VALUES ($1,'Reset User',$2,'stakeholder','assessor','active','old-password-hash') RETURNING id`, organizationID, email).Scan(&userID); err != nil {
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organizationID)
		data.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM users WHERE id=$1", userID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organizationID)
		data.Close()
	})
	return data, ctx, userID, email
}
