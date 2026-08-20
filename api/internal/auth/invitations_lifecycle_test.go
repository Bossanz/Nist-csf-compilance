package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"compliance/api/internal/store"
)

type lifecycleInvitationRepository struct {
	created            store.Invitation
	projectIDs         []string
	resendOrganization string
	resendInvitation   string
	cancelOrganization string
	cancelInvitation   string
	acceptErr          error
}

func (f *lifecycleInvitationRepository) HasActiveOrPendingEmail(context.Context, string) (bool, error) {
	return false, nil
}

func (f *lifecycleInvitationRepository) CreateInvitation(_ context.Context, invitation store.Invitation) (store.Invitation, error) {
	if invitation.ID == "" {
		invitation.ID = "invite-1"
	}
	f.created = invitation
	return invitation, nil
}

func (f *lifecycleInvitationRepository) AcceptInvitation(_ context.Context, _, name, _ string, _ time.Time) (store.User, error) {
	return store.User{Name: name, Status: "active"}, f.acceptErr
}

func (f *lifecycleInvitationRepository) CreateInvitationProjectAccess(_ context.Context, _ string, projectIDs []string) error {
	f.projectIDs = append([]string(nil), projectIDs...)
	return nil
}

func (f *lifecycleInvitationRepository) ResendInvitation(_ context.Context, organizationID, invitationID, _, _ string, _, _ time.Time) (store.Invitation, error) {
	f.resendOrganization = organizationID
	f.resendInvitation = invitationID
	return store.Invitation{ID: "invite-2"}, nil
}

func (f *lifecycleInvitationRepository) CancelInvitation(_ context.Context, organizationID, invitationID, _ string, _ time.Time) (store.Invitation, error) {
	f.cancelOrganization = organizationID
	f.cancelInvitation = invitationID
	return store.Invitation{ID: invitationID}, nil
}

func (f *lifecycleInvitationRepository) ListInvitations(context.Context, string) ([]store.Invitation, error) {
	return nil, nil
}

func (f *lifecycleInvitationRepository) GetInvitation(context.Context, string, string) (store.Invitation, error) {
	return store.Invitation{}, nil
}

func TestInviteAuditorRequiresOrganizationAdminAndProjectIDs(t *testing.T) {
	organizationID := "org-1"
	repo := &lifecycleInvitationRepository{}
	service := NewInvitationService(repo, time.Now)

	invitation, raw, err := service.Invite(context.Background(), store.User{ID: "admin-1", UserType: "stakeholder", OrganizationID: &organizationID, Role: "org_admin"}, &organizationID, "auditor@example.com", "auditor", "project-1")
	if err != nil || raw == "" || invitation.Role != "auditor" || len(invitation.ProjectIDs) != 1 || invitation.ProjectIDs[0] != "project-1" {
		t.Fatalf("unexpected Auditor invitation: invitation=%#v raw=%q projects=%v err=%v", invitation, raw, invitation.ProjectIDs, err)
	}
}

func TestInviteAuditorRequiresProjectIDs(t *testing.T) {
	organizationID := "org-1"
	service := NewInvitationService(&lifecycleInvitationRepository{}, time.Now)
	_, _, err := service.Invite(context.Background(), store.User{ID: "admin-1", UserType: "stakeholder", OrganizationID: &organizationID, Role: "org_admin"}, &organizationID, "auditor@example.com", "auditor")
	if !errors.Is(err, ErrInvalidProjectAccess) {
		t.Fatalf("expected invalid project access, got %v", err)
	}
}

func TestResendSupersedesPendingInvitationAndReturnsNewToken(t *testing.T) {
	organizationID := "org-1"
	repo := &lifecycleInvitationRepository{}
	service := NewInvitationService(repo, time.Now)
	invitation, raw, err := service.Resend(context.Background(), store.User{ID: "counselor-admin-1", UserType: "counselor", Role: "counselor_admin"}, organizationID, "invite-1")
	if err != nil || raw == "" || invitation.ID != "invite-2" || repo.resendOrganization != organizationID || repo.resendInvitation != "invite-1" {
		t.Fatalf("unexpected resend: invitation=%#v raw=%q repo=%#v err=%v", invitation, raw, repo, err)
	}
}

func TestCancelRejectsAcceptanceOfInvitation(t *testing.T) {
	organizationID := "org-1"
	repo := &lifecycleInvitationRepository{}
	service := NewInvitationService(repo, time.Now)
	invitation, err := service.Cancel(context.Background(), store.User{ID: "counselor-admin-1", UserType: "counselor", Role: "counselor_admin"}, organizationID, "invite-1")
	if err != nil || invitation.ID != "invite-1" || repo.cancelOrganization != organizationID || repo.cancelInvitation != "invite-1" {
		t.Fatalf("unexpected cancel: invitation=%#v repo=%#v err=%v", invitation, repo, err)
	}
}

func TestInvitationLifecyclePermissions(t *testing.T) {
	organizationID := "org-1"
	otherOrganizationID := "org-2"
	orgAdmin := store.User{ID: "org-admin-1", UserType: "stakeholder", OrganizationID: &organizationID, Role: "org_admin"}
	tests := []struct {
		name    string
		actor   store.User
		resend  bool
		wantErr bool
	}{
		{name: "Counselor Admin can resend", actor: store.User{ID: "counselor-admin-1", UserType: "counselor", Role: "counselor_admin"}, resend: true},
		{name: "Counselor Admin can cancel", actor: store.User{ID: "counselor-admin-1", UserType: "counselor", Role: "counselor_admin"}},
		{name: "Counselor cannot resend", actor: store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor"}, resend: true, wantErr: true},
		{name: "Counselor cannot cancel", actor: store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor"}, wantErr: true},
		{name: "Org Admin can resend", actor: orgAdmin, resend: true},
		{name: "Org Admin can cancel", actor: orgAdmin},
		{name: "Org Admin from another organization cannot resend", actor: store.User{ID: "other-org-admin-1", UserType: "stakeholder", OrganizationID: &otherOrganizationID, Role: "org_admin"}, resend: true, wantErr: true},
		{name: "Org Admin from another organization cannot cancel", actor: store.User{ID: "other-org-admin-1", UserType: "stakeholder", OrganizationID: &otherOrganizationID, Role: "org_admin"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInvitationService(&lifecycleInvitationRepository{}, time.Now)
			var err error
			if tt.resend {
				_, _, err = service.Resend(context.Background(), tt.actor, organizationID, "invite-1")
			} else {
				_, err = service.Cancel(context.Background(), tt.actor, organizationID, "invite-1")
			}
			if tt.wantErr && !errors.Is(err, ErrForbidden) {
				t.Fatalf("expected forbidden, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
