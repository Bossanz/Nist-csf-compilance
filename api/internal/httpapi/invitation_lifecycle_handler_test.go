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

type invitationLifecycleHTTPRepository struct {
	created   store.Invitation
	resendID  string
	cancelID  string
	resendOrg string
	cancelOrg string
}

func (f *invitationLifecycleHTTPRepository) HasActiveOrPendingEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (f *invitationLifecycleHTTPRepository) CreateInvitation(_ context.Context, invitation store.Invitation) (store.Invitation, error) {
	invitation.ID = "invite-1"
	f.created = invitation
	return invitation, nil
}
func (f *invitationLifecycleHTTPRepository) AcceptInvitation(context.Context, string, string, string, time.Time) (store.User, error) {
	return store.User{}, nil
}
func (f *invitationLifecycleHTTPRepository) ResendInvitation(_ context.Context, organizationID, invitationID, _, _ string, _, _ time.Time) (store.Invitation, error) {
	f.resendOrg = organizationID
	f.resendID = invitationID
	return store.Invitation{ID: "invite-2", OrganizationID: &organizationID, Email: "auditor@example.com", Role: "auditor", Status: "pending", ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (f *invitationLifecycleHTTPRepository) CancelInvitation(_ context.Context, organizationID, invitationID, _ string, _ time.Time) (store.Invitation, error) {
	f.cancelOrg = organizationID
	f.cancelID = invitationID
	return store.Invitation{ID: invitationID, OrganizationID: &organizationID, Status: "cancelled"}, nil
}

func TestOrgAdminCreatesAuditorInvitationWithProjectScope(t *testing.T) {
	organizationID := "org-1"
	repository := &invitationLifecycleHTTPRepository{}
	handler := authenticatedHandler(store.User{ID: "admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}, fakeStore{organization: store.Organization{ID: organizationID}})
	handler.Invitations = authservice.NewInvitationService(repository, time.Now)
	handler.AppOrigin = "http://localhost:3000"
	request := authenticatedRequest(http.MethodPost, "/api/organizations/org-1/invitations", `{"email":"auditor@example.com","role":"auditor","projectIDs":["project-1","project-2"]}`)
	request.Header.Set("Origin", handler.AppOrigin)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || repository.created.Role != "auditor" || len(repository.created.ProjectIDs) != 2 || !strings.Contains(response.Body.String(), "/invite/") {
		t.Fatalf("unexpected Auditor invitation: %d %#v %s", response.Code, repository.created, response.Body.String())
	}
}

func TestOrganizationInvitationLifecycleEndpoints(t *testing.T) {
	organizationID := "org-1"
	repository := &invitationLifecycleHTTPRepository{}
	handler := authenticatedHandler(store.User{ID: "admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}, fakeStore{
		organization: store.Organization{ID: organizationID},
		invitations:  []store.Invitation{{ID: "invite-1", OrganizationID: &organizationID, Email: "auditor@example.com", Role: "auditor", Status: "pending", ExpiresAt: time.Now().Add(time.Hour)}},
	})
	handler.Invitations = authservice.NewInvitationService(repository, time.Now)
	handler.AppOrigin = "http://localhost:3000"

	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/invitations", ""))
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "auditor@example.com") {
		t.Fatalf("unexpected invitation list: %d %s", listResponse.Code, listResponse.Body.String())
	}

	resendRequest := authenticatedRequest(http.MethodPost, "/api/organizations/org-1/invitations/invite-1/resend", "")
	resendRequest.Header.Set("Origin", handler.AppOrigin)
	resendResponse := httptest.NewRecorder()
	handler.ServeHTTP(resendResponse, resendRequest)
	if resendResponse.Code != http.StatusOK || repository.resendOrg != organizationID || repository.resendID != "invite-1" || !strings.Contains(resendResponse.Body.String(), "/invite/") {
		t.Fatalf("unexpected resend: %d %s repo=%#v", resendResponse.Code, resendResponse.Body.String(), repository)
	}

	cancelRequest := authenticatedRequest(http.MethodPost, "/api/organizations/org-1/invitations/invite-1/cancel", "")
	cancelRequest.Header.Set("Origin", handler.AppOrigin)
	cancelResponse := httptest.NewRecorder()
	handler.ServeHTTP(cancelResponse, cancelRequest)
	if cancelResponse.Code != http.StatusOK || repository.cancelOrg != organizationID || repository.cancelID != "invite-1" || !strings.Contains(cancelResponse.Body.String(), `"status":"cancelled"`) {
		t.Fatalf("unexpected cancel: %d %s repo=%#v", cancelResponse.Code, cancelResponse.Body.String(), repository)
	}
}
