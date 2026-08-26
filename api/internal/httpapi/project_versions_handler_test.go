package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"compliance/api/internal/store"
)

type versionFakeStore struct {
	fakeStore
	createdSourceID *string
	createdActorID  *string
	createCalls     *int
	createdProject  store.Project
	createErr       error
	versions        []store.Project
	versionsErr     error
}

func authenticatedHandlerWithStore(user store.User, data dataStore) *Handler {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{user: user, session: store.Session{ExpiresAt: now.Add(time.Hour)}}
	handler := newAuthHandler(repo, now)
	handler.Store = data
	return handler
}

func (f versionFakeStore) CreateNextProjectVersion(_ context.Context, sourceProjectID, actorID string) (store.Project, error) {
	if f.createdSourceID != nil {
		*f.createdSourceID = sourceProjectID
	}
	if f.createdActorID != nil {
		*f.createdActorID = actorID
	}
	if f.createCalls != nil {
		*f.createCalls++
	}
	return f.createdProject, f.createErr
}

func (f versionFakeStore) ListProjectVersions(context.Context, string) ([]store.Project, error) {
	return f.versions, f.versionsErr
}

func TestCanCreateProjectVersion(t *testing.T) {
	cases := []struct {
		role    string
		allowed bool
	}{
		{role: "counselor_admin", allowed: true},
		{role: "counselor", allowed: true},
		{role: "org_admin", allowed: false},
		{role: "assessor", allowed: false},
		{role: "reviewer", allowed: false},
		{role: "viewer", allowed: false},
		{role: "auditor", allowed: false},
	}
	for _, test := range cases {
		t.Run(test.role, func(t *testing.T) {
			if got := can(store.User{Role: test.role}, actionCreateProjectVersion); got != test.allowed {
				t.Fatalf("expected can=%v, got %v", test.allowed, got)
			}
		})
	}
}

func TestCounselorCreatesProjectVersionAndWritesAudit(t *testing.T) {
	var sourceID, actorID, auditAction string
	var auditEvent store.AuditEvent
	newProject := store.Project{ID: "project-2", OrganizationID: "org-1", Status: "setup", VersionNumber: 2}
	handler := authenticatedHandlerWithStore(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, versionFakeStore{
		fakeStore:       fakeStore{project: store.Project{ID: "project-1", OrganizationID: "org-1", Status: "closed", VersionNumber: 1}, auditAction: &auditAction, auditEvent: &auditEvent},
		createdSourceID: &sourceID,
		createdActorID:  &actorID,
		createdProject:  newProject,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusCreated || sourceID != "project-1" || actorID != "counselor-1" {
		t.Fatalf("unexpected create response: %d %s source=%s actor=%s", response.Code, response.Body.String(), sourceID, actorID)
	}
	if auditAction != "project.version_created" {
		t.Fatalf("expected version audit event, got %q", auditAction)
	}
	if auditEvent.Metadata["sourceProjectID"] != "project-1" || auditEvent.Metadata["newProjectID"] != "project-2" || auditEvent.Metadata["sourceVersion"] != 1 || auditEvent.Metadata["newVersion"] != 2 {
		t.Fatalf("unexpected version audit metadata: %#v", auditEvent.Metadata)
	}
	if !strings.Contains(response.Body.String(), `"versionNumber":2`) {
		t.Fatalf("response does not contain new version: %s", response.Body.String())
	}
}

func TestCreateProjectVersionMapsUnfinalizedError(t *testing.T) {
	handler := authenticatedHandlerWithStore(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, versionFakeStore{
		fakeStore: fakeStore{project: store.Project{ID: "project-1", OrganizationID: "org-1", Status: "in_review", VersionNumber: 1}},
		createErr: store.ErrProjectVersionNotFinalized,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"project_not_finalized"`) {
		t.Fatalf("expected project_not_finalized conflict, got %d: %s", response.Code, response.Body.String())
	}
}

func TestCreateProjectVersionMapsAllocationConflict(t *testing.T) {
	handler := authenticatedHandlerWithStore(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, versionFakeStore{
		fakeStore: fakeStore{project: store.Project{ID: "project-1", OrganizationID: "org-1", Status: "closed", VersionNumber: 1}},
		createErr: store.ErrProjectVersionConflict,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"version_creation_conflict"`) {
		t.Fatalf("expected version_creation_conflict, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAssessorCannotCreateProjectVersion(t *testing.T) {
	organizationID := "org-1"
	calls := 0
	handler := authenticatedHandlerWithStore(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, versionFakeStore{
		fakeStore:   fakeStore{project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed", VersionNumber: 1}},
		createCalls: &calls,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusForbidden || calls != 0 {
		t.Fatalf("expected assessor denial before store call, got %d calls=%d: %s", response.Code, calls, response.Body.String())
	}
}

func TestListProjectVersionsReturnsVersions(t *testing.T) {
	versions := []store.Project{
		{ID: "project-2", OrganizationID: "org-1", Status: "setup", VersionNumber: 2, IsLatest: true},
		{ID: "project-1", OrganizationID: "org-1", Status: "closed", VersionNumber: 1, IsLatest: false},
	}
	handler := authenticatedHandlerWithStore(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, versionFakeStore{
		fakeStore: fakeStore{project: versions[1]},
		versions:  versions,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var got []store.Project
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].VersionNumber != 2 || got[1].VersionNumber != 1 {
		t.Fatalf("unexpected versions: %#v", got)
	}
}

func TestAuditorOnlySeesProjectVersionsWithAccess(t *testing.T) {
	organizationID := "org-1"
	versions := []store.Project{
		{ID: "project-2", OrganizationID: organizationID, Status: "setup", VersionNumber: 2, IsLatest: true},
		{ID: "project-1", OrganizationID: organizationID, Status: "closed", VersionNumber: 1, IsLatest: false},
	}
	handler := authenticatedHandlerWithStore(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, versionFakeStore{
		fakeStore: fakeStore{
			project:              versions[1],
			auditorProjectAccess: map[string]bool{"project-1": true, "project-2": false},
		},
		versions: versions,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/versions", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var got []store.Project
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "project-1" {
		t.Fatalf("auditor saw versions without access: %#v", got)
	}
}
