package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authservice "compliance/api/internal/auth"
	"compliance/api/internal/store"
)

type fakeHTTPInvitationRepository struct {
	created           store.Invitation
	acceptedTokenHash string
	acceptErr         error
}

func (f *fakeHTTPInvitationRepository) HasActiveOrPendingEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (f *fakeHTTPInvitationRepository) CreateInvitation(_ context.Context, invitation store.Invitation) (store.Invitation, error) {
	f.created = invitation
	invitation.ID = "invite-1"
	return invitation, nil
}
func (f *fakeHTTPInvitationRepository) AcceptInvitation(_ context.Context, tokenHash, name, passwordHash string, _ time.Time) (store.User, error) {
	f.acceptedTokenHash = tokenHash
	return store.User{ID: "user-2", Name: name, Status: "active"}, f.acceptErr
}

func TestCounselorCreatesOrganizationAdminInvitation(t *testing.T) {
	invitationRepo := &fakeHTTPInvitationRepository{}
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{organization: store.Organization{ID: "org-1"}})
	handler.Invitations = authservice.NewInvitationService(invitationRepo, time.Now)
	handler.AppOrigin = "http://localhost:3000"
	request := authenticatedRequest(http.MethodPost, "/api/organizations/org-1/invitations", `{"email":"admin@acme.test","role":"org_admin"}`)
	request.Header.Set("Origin", handler.AppOrigin)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || invitationRepo.created.Role != "org_admin" || !strings.Contains(response.Body.String(), `"invitationURL":"http://localhost:3000/invite/`) {
		t.Fatalf("unexpected invitation: %d %s %#v", response.Code, response.Body.String(), invitationRepo.created)
	}
}

func TestInvitationAcceptanceDoesNotRequireExistingSession(t *testing.T) {
	invitationRepo := &fakeHTTPInvitationRepository{}
	handler := &Handler{Invitations: authservice.NewInvitationService(invitationRepo, time.Now), AppOrigin: "http://localhost:3000"}
	request := httptest.NewRequest(http.MethodPost, "/api/invitations/raw-token/accept", strings.NewReader(`{"name":"Pat","password":"secret-password"}`))
	request.Header.Set("Origin", handler.AppOrigin)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || invitationRepo.acceptedTokenHash != authservice.HashToken("raw-token") {
		t.Fatalf("unexpected acceptance: %d %s hash=%s", response.Code, response.Body.String(), invitationRepo.acceptedTokenHash)
	}
}

func TestInvalidInvitationUsesGenericResponse(t *testing.T) {
	invitationRepo := &fakeHTTPInvitationRepository{acceptErr: store.ErrInvalidInvitation}
	handler := &Handler{Invitations: authservice.NewInvitationService(invitationRepo, time.Now), AppOrigin: "http://localhost:3000"}
	request := httptest.NewRequest(http.MethodPost, "/api/invitations/expired/accept", strings.NewReader(`{"name":"Pat","password":"secret-password"}`))
	request.Header.Set("Origin", handler.AppOrigin)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_invitation"`) {
		t.Fatalf("unexpected invalid response: %d %s", response.Code, response.Body.String())
	}
}
