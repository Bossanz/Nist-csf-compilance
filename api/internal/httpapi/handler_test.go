package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type fakeStore struct {
	projects                     []store.Project
	project                      store.Project
	organizations                []store.Organization
	organization                 store.Organization
	createdOrganizationName      *string
	createdProjectOrganizationID *string
	createdProjectName           *string
	auditAction                  *string
	users                        []store.User
	updatedUserRole              *string
	updatedUserStatus            *string
	listErr                      error
	deleteErr                    error
	deletedID                    *string
	profileErr                   error
}

func (f fakeStore) ListProjects(context.Context) ([]store.Project, error) {
	return f.projects, f.listErr
}
func (f fakeStore) ListFunctions(context.Context) ([]store.Function, error) { return nil, nil }
func (f fakeStore) CreateProject(context.Context, string, string) (store.Project, error) {
	return store.Project{}, nil
}
func (f fakeStore) CreateScopedProject(_ context.Context, organizationID, name string) (store.Project, error) {
	if f.createdProjectOrganizationID != nil {
		*f.createdProjectOrganizationID = organizationID
	}
	if f.createdProjectName != nil {
		*f.createdProjectName = name
	}
	return store.Project{}, nil
}
func (f fakeStore) GetProject(context.Context, string) (store.Project, error) {
	return f.project, nil
}
func (f fakeStore) ListProfile(context.Context, string) ([]store.ProfileRow, error) {
	return nil, f.profileErr
}
func (f fakeStore) UpdateProfile(context.Context, string, string, store.ProfilePatch) (store.ProfileRow, error) {
	return store.ProfileRow{}, nil
}
func (f fakeStore) DeleteProject(_ context.Context, id string) error {
	if f.deletedID != nil {
		*f.deletedID = id
	}
	return f.deleteErr
}
func (f fakeStore) ListOrganizations(context.Context, *string) ([]store.Organization, error) {
	return f.organizations, nil
}
func (f fakeStore) GetOrganization(context.Context, string) (store.Organization, error) {
	return f.organization, nil
}
func (f fakeStore) CreateOrganization(_ context.Context, name string) (store.Organization, error) {
	if f.createdOrganizationName != nil {
		*f.createdOrganizationName = name
	}
	return store.Organization{Name: name}, nil
}
func (f fakeStore) DeleteOrganization(context.Context, string) error { return nil }
func (f fakeStore) ListProjectsByOrganization(context.Context, string) ([]store.Project, error) {
	return f.projects, nil
}
func (f fakeStore) WriteAudit(_ context.Context, event store.AuditEvent) error {
	if f.auditAction != nil {
		*f.auditAction = event.Action
	}
	return nil
}
func (f fakeStore) ListOrganizationUsers(context.Context, string) ([]store.User, error) {
	return f.users, nil
}
func (f fakeStore) ListCounselors(context.Context) ([]store.User, error) { return f.users, nil }
func (f fakeStore) UpdateOrganizationUser(_ context.Context, _, _, role, status string) (store.User, error) {
	if f.updatedUserRole != nil {
		*f.updatedUserRole = role
	}
	if f.updatedUserStatus != nil {
		*f.updatedUserStatus = status
	}
	return store.User{Role: role, Status: status}, nil
}
func (f fakeStore) UpdateCounselor(_ context.Context, _, role, status string) (store.User, error) {
	if f.updatedUserRole != nil {
		*f.updatedUserRole = role
	}
	if f.updatedUserStatus != nil {
		*f.updatedUserStatus = status
	}
	return store.User{Role: role, Status: status}, nil
}
func (f fakeStore) RevokeUserSessions(context.Context, string) error { return nil }

func TestHealthz(t *testing.T) {
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestListProjects(t *testing.T) {
	expected := []store.Project{{ID: "project-1", Name: "Readiness", OrganizationName: "Acme", Status: "setup"}}
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{projects: expected}}).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got []store.Project
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, got %#v", expected, got)
	}
}

func TestListProjectsReturnsEmptyArray(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{projects: []store.Project{}}}).ServeHTTP(w, r)

	if w.Code != http.StatusOK || w.Body.String() != "[]\n" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestListProjectsHandlesStoreFailure(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{listErr: errors.New("database unavailable")}}).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProject(t *testing.T) {
	var deletedID string
	projectID := "11111111-1111-1111-1111-111111111111"
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deletedID: &deletedID}}).ServeHTTP(w, r)

	if w.Code != http.StatusNoContent || w.Body.Len() != 0 || deletedID != projectID {
		t.Fatalf("unexpected response: %d %s id=%s", w.Code, w.Body.String(), deletedID)
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deleteErr: pgx.ErrNoRows}}).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProjectHandlesStoreFailure(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deleteErr: errors.New("database unavailable")}}).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProjectRejectsMalformedID(t *testing.T) {
	var deletedID string
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/not-a-uuid", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deletedID: &deletedID}}).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound || deletedID != "" {
		t.Fatalf("unexpected response: %d %s id=%s", w.Code, w.Body.String(), deletedID)
	}
}

func TestDeletedProjectProfileIsNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/projects/11111111-1111-1111-1111-111111111111/profile", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{profileErr: pgx.ErrNoRows}}).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}
