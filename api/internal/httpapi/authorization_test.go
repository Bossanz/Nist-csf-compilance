package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

func TestAuthorizationMatrix(t *testing.T) {
	cases := []struct {
		role    string
		action  action
		allowed bool
	}{
		{"counselor_admin", actionCreateOrganization, true},
		{"counselor", actionCreateOrganization, false},
		{"counselor", actionCreateProject, true},
		{"org_admin", actionCreateProject, false},
		{"org_admin", actionUpdateProfile, false},
		{"assessor", actionUpdateProfile, false},
		{"assessor", actionSaveResponse, true},
		{"assessor", actionSubmitResponse, true},
		{"reviewer", actionReviewResponse, true},
		{"reviewer", actionUpdateProfile, false},
		{"viewer", actionUpdateProfile, false},
	}
	for _, test := range cases {
		if got := can(store.User{Role: test.role}, test.action); got != test.allowed {
			t.Errorf("role=%s action=%s: expected %v, got %v", test.role, test.action, test.allowed, got)
		}
	}
}

func TestProtectedRouteRequiresAuthentication(t *testing.T) {
	repo := &fakeAuthRepository{findSessionErr: pgx.ErrNoRows}
	handler := newAuthHandler(repo, time.Now())
	handler.Store = fakeStore{projects: []store.Project{}}
	request := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", response.Code, response.Body.String())
	}
}

func TestStakeholderCannotAccessAnotherOrganizationProject(t *testing.T) {
	organizationID := "org-1"
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{
		user:    store.User{ID: "user-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"},
		session: store.Session{ExpiresAt: now.Add(time.Hour)},
	}
	handler := newAuthHandler(repo, now)
	handler.Store = fakeStore{project: store.Project{ID: "project-2", OrganizationID: "org-2"}}
	request := httptest.NewRequest(http.MethodGet, "/api/projects/project-2", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", response.Code, response.Body.String())
	}
}

func TestProfileUpdateWritesAuditEvent(t *testing.T) {
	organizationID := "org-1"
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{user: store.User{ID: "user-1", UserType: "counselor", Role: "counselor", Status: "active"}, session: store.Session{ExpiresAt: now.Add(time.Hour)}}
	var action string
	handler := newAuthHandler(repo, now)
	handler.Store = fakeStore{project: store.Project{ID: "11111111-1111-1111-1111-111111111111", OrganizationID: organizationID}, auditAction: &action}
	request := httptest.NewRequest(http.MethodPut, "/api/projects/11111111-1111-1111-1111-111111111111/profile/22222222-2222-2222-2222-222222222222", strings.NewReader(`{"included":true}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || action != "profile.updated" {
		t.Fatalf("expected audited update, got %d action=%s body=%s", response.Code, action, response.Body.String())
	}
}
