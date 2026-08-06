package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"compliance/api/internal/store"
)

type fakeInvitationRepository struct {
	duplicate         bool
	created           store.Invitation
	acceptErr         error
	acceptedHash      string
	acceptedTokenHash string
}

func (f *fakeInvitationRepository) HasActiveOrPendingEmail(context.Context, string) (bool, error) {
	return f.duplicate, nil
}
func (f *fakeInvitationRepository) CreateInvitation(_ context.Context, invitation store.Invitation) (store.Invitation, error) {
	f.created = invitation
	return invitation, nil
}
func (f *fakeInvitationRepository) AcceptInvitation(_ context.Context, tokenHash, name, passwordHash string, _ time.Time) (store.User, error) {
	f.acceptedTokenHash = tokenHash
	f.acceptedHash = passwordHash
	return store.User{Name: name, Status: "active"}, f.acceptErr
}

func TestOrgAdminCannotInviteCounselorRole(t *testing.T) {
	organizationID := "org-1"
	service := NewInvitationService(&fakeInvitationRepository{}, time.Now)
	_, _, err := service.Invite(context.Background(), store.User{ID: "admin-1", OrganizationID: &organizationID, Role: "org_admin"}, nil, "person@example.com", "counselor")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestCounselorAdminCreatesCounselorInvitationWithoutOrganization(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeInvitationRepository{}
	service := NewInvitationService(repo, func() time.Time { return now })
	invitation, raw, err := service.Invite(context.Background(), store.User{ID: "admin-1", UserType: "counselor", Role: "counselor_admin"}, nil, " PERSON@example.com ", "counselor")
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" || invitation.OrganizationID != nil || invitation.UserType != "counselor" || invitation.Email != "person@example.com" || invitation.TokenHash != HashToken(raw) || !invitation.ExpiresAt.Equal(now.Add(72*time.Hour)) {
		t.Fatalf("unexpected invitation: %#v", invitation)
	}
}

func TestInviteRejectsDuplicateActiveOrPendingEmail(t *testing.T) {
	service := NewInvitationService(&fakeInvitationRepository{duplicate: true}, time.Now)
	_, _, err := service.Invite(context.Background(), store.User{ID: "admin-1", Role: "counselor_admin"}, nil, "person@example.com", "counselor")
	if !errors.Is(err, ErrDuplicateInvitation) {
		t.Fatalf("expected duplicate, got %v", err)
	}
}

func TestAcceptInvitationHashesPasswordAndUsesTokenHash(t *testing.T) {
	repo := &fakeInvitationRepository{}
	service := NewInvitationService(repo, time.Now)
	user, err := service.Accept(context.Background(), "raw-token", "Pat", "secret-password")
	if err != nil || user.Status != "active" || repo.acceptedTokenHash != HashToken("raw-token") || !CheckPassword(repo.acceptedHash, "secret-password") {
		t.Fatalf("unexpected acceptance: user=%#v err=%v", user, err)
	}
}

func TestAcceptInvitationUsesGenericInvalidError(t *testing.T) {
	service := NewInvitationService(&fakeInvitationRepository{acceptErr: store.ErrInvalidInvitation}, time.Now)
	_, err := service.Accept(context.Background(), "expired-token", "Pat", "secret-password")
	if !errors.Is(err, ErrInvalidInvitation) {
		t.Fatalf("expected invalid invitation, got %v", err)
	}
}
