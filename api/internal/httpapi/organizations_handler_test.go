package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"compliance/api/internal/store"
)

func authenticatedHandler(user store.User, data fakeStore) *Handler {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{user: user, session: store.Session{ExpiresAt: now.Add(time.Hour)}}
	handler := newAuthHandler(repo, now)
	handler.Store = data
	return handler
}

func authenticatedRequest(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	return request
}

func TestCounselorAdminCreatesOrganization(t *testing.T) {
	var createdName string
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor_admin", Status: "active"}, fakeStore{createdOrganizationName: &createdName})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/organizations", `{"name":"Acme"}`))

	if response.Code != http.StatusCreated || createdName != "Acme" {
		t.Fatalf("unexpected response: %d %s name=%s", response.Code, response.Body.String(), createdName)
	}
}

func TestCounselorCannotCreateOrganization(t *testing.T) {
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/organizations", `{"name":"Acme"}`))

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
}

func TestCounselorAdminDeletesOrganization(t *testing.T) {
	var deletedID string
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor_admin", Status: "active"}, fakeStore{deletedID: &deletedID})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodDelete, "/api/organizations/org-1", ""))

	if response.Code != http.StatusNoContent || deletedID != "org-1" {
		t.Fatalf("unexpected response: %d %s id=%s", response.Code, response.Body.String(), deletedID)
	}
}

func TestCounselorCannotDeleteOrganization(t *testing.T) {
	var deletedID string
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{deletedID: &deletedID})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodDelete, "/api/organizations/org-1", ""))

	if response.Code != http.StatusForbidden || deletedID != "" {
		t.Fatalf("unexpected response: %d %s id=%s", response.Code, response.Body.String(), deletedID)
	}
}

func TestStakeholderListsOnlyOwnOrganization(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{organizations: []store.Organization{{ID: "org-1", Name: "Acme"}}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":"org-1"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestCounselorCreatesProjectInsideSelectedOrganization(t *testing.T) {
	var organizationID, projectName string
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{organization: store.Organization{ID: "org-1"}, createdProjectOrganizationID: &organizationID, createdProjectName: &projectName})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/organizations/org-1/projects", `{"name":"Readiness"}`))

	if response.Code != http.StatusCreated || organizationID != "org-1" || projectName != "Readiness" {
		t.Fatalf("unexpected response: %d %s org=%s name=%s", response.Code, response.Body.String(), organizationID, projectName)
	}
}

func TestCounselorCreatesProjectWithMetadata(t *testing.T) {
	var metadata store.ProjectMetadata
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{organization: store.Organization{ID: "org-1"}, createdProjectMetadata: &metadata})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/organizations/org-1/projects", `{"name":"Readiness","objective":"Prepare for registration","assessmentPeriod":"2026","targetCompletionDate":"2026-09-30","scopeBoundary":"Registration systems","complianceDriver":"Regulatory requirement"}`))

	want := store.ProjectMetadata{Objective: "Prepare for registration", AssessmentPeriod: "2026", TargetCompletionDate: "2026-09-30", ScopeBoundary: "Registration systems", ComplianceDriver: "Regulatory requirement"}
	if response.Code != http.StatusCreated || metadata != want {
		t.Fatalf("unexpected response: %d %s metadata=%#v", response.Code, response.Body.String(), metadata)
	}
}

func TestCounselorRejectsInvalidTargetDate(t *testing.T) {
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{organization: store.Organization{ID: "org-1"}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/organizations/org-1/projects", `{"name":"Readiness","targetCompletionDate":"30/09/2026"}`))

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"validation_error"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestCounselorResolvesOrganizationAndProjectSlugs(t *testing.T) {
	organization := store.Organization{ID: "org-1", Name: "Acme Corporation", Slug: "acme-corporation", Type: "client"}
	project := store.Project{ID: "project-1", OrganizationID: "org-1", OrganizationName: organization.Name, Name: "Readiness", Slug: "readiness", Status: "setup"}
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{organizationBySlug: organization, projectBySlug: project})

	organizationResponse := httptest.NewRecorder()
	handler.ServeHTTP(organizationResponse, authenticatedRequest(http.MethodGet, "/api/organizations/by-slug/acme-corporation", ""))
	if organizationResponse.Code != http.StatusOK || !strings.Contains(organizationResponse.Body.String(), `"slug":"acme-corporation"`) {
		t.Fatalf("unexpected organization response: %d %s", organizationResponse.Code, organizationResponse.Body.String())
	}

	projectResponse := httptest.NewRecorder()
	handler.ServeHTTP(projectResponse, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/projects/by-slug/readiness", ""))
	if projectResponse.Code != http.StatusOK || !strings.Contains(projectResponse.Body.String(), `"slug":"readiness"`) {
		t.Fatalf("unexpected project response: %d %s", projectResponse.Code, projectResponse.Body.String())
	}
}

func TestStakeholderCannotResolveForeignOrganizationSlug(t *testing.T) {
	organizationID := "org-2"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{organizationBySlug: store.Organization{ID: "org-1", Name: "Acme", Slug: "acme", Type: "client"}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/by-slug/acme", ""))

	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "Acme") {
		t.Fatalf("expected hidden foreign organization, got %d %s", response.Code, response.Body.String())
	}
}

func TestStateChangingRequestRejectsForeignOrigin(t *testing.T) {
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor_admin", Status: "active"}, fakeStore{})
	handler.AppOrigin = "http://localhost:3000"
	request := authenticatedRequest(http.MethodPost, "/api/organizations", `{"name":"Acme"}`)
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
}
func TestStakeholderCannotResolveSetupProjectSlug(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{
		organization:  store.Organization{ID: organizationID, Name: "Acme"},
		projectBySlug: store.Project{ID: "project-1", OrganizationID: organizationID, Slug: "readiness", Status: "setup"},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/projects/by-slug/readiness", ""))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected setup Project slug to be hidden from Stakeholder: %d %s", response.Code, response.Body.String())
	}
}

func TestCounselorListsSetupProjectsInOrganization(t *testing.T) {
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		organization: store.Organization{ID: "org-1", Name: "Acme"},
		projects: []store.Project{{
			ID: "project-1", OrganizationID: "org-1", Name: "Draft readiness", Status: "setup",
		}},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/projects", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "\"name\":\"Draft readiness\"") {
		t.Fatalf("expected Counselor to see setup project: %d %s", response.Code, response.Body.String())
	}
}

func TestStakeholderOrganizationProjectsHideSetupProjects(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{
		organization: store.Organization{ID: organizationID, Name: "Acme"},
		projects: []store.Project{
			{ID: "project-1", OrganizationID: organizationID, Name: "Draft readiness", Status: "setup"},
			{ID: "project-2", OrganizationID: organizationID, Name: "Published readiness", Status: "in_review"},
		},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/projects", ""))

	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "\"name\":\"Draft readiness\"") || !strings.Contains(response.Body.String(), "\"name\":\"Published readiness\"") {
		t.Fatalf("expected Stakeholder to see only submitted projects: %d %s", response.Code, response.Body.String())
	}
}
