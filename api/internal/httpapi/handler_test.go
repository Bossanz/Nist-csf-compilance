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
	organizationBySlug           store.Organization
	projectBySlug                store.Project
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
	profiles                     []store.ProfileRow
	updatedPatch                 *store.ProfilePatch
	profileUpdateResult          store.ProfileRow
	profileUpdateErr             error
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
	return f.profiles, f.profileErr
}
func (f fakeStore) UpdateProfile(_ context.Context, _, _ string, patch store.ProfilePatch) (store.ProfileRow, error) {
	if f.updatedPatch != nil {
		*f.updatedPatch = patch
	}
	return f.profileUpdateResult, f.profileUpdateErr
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
func (f fakeStore) GetOrganizationBySlug(context.Context, string) (store.Organization, error) {
	return f.organizationBySlug, nil
}
func (f fakeStore) CreateOrganization(_ context.Context, name string) (store.Organization, error) {
	if f.createdOrganizationName != nil {
		*f.createdOrganizationName = name
	}
	return store.Organization{Name: name}, nil
}
func (f fakeStore) DeleteOrganization(_ context.Context, id string) error {
	if f.deletedID != nil {
		*f.deletedID = id
	}
	return f.deleteErr
}
func (f fakeStore) ListProjectsByOrganization(context.Context, string) ([]store.Project, error) {
	return f.projects, nil
}
func (f fakeStore) GetProjectBySlug(context.Context, string, string) (store.Project, error) {
	return f.projectBySlug, nil
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

func TestStakeholderProfileOnlyReturnsIncludedOutcomes(t *testing.T) {
	organizationID := "org-1"
	assessorID := "assessor-1"
	otherAssessorID := "assessor-2"
	rows := []store.ProfileRow{
		{SubcategoryCode: "GV.OC-01", Included: true, AssignedUserID: &assessorID},
		{SubcategoryCode: "GV.OC-02", Included: true, AssignedUserID: &otherAssessorID},
	}
	handler := authenticatedHandler(store.User{ID: assessorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project:  store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles: rows,
	})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/profile", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var got []store.ProfileRow
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].SubcategoryCode != "GV.OC-01" {
		t.Fatalf("expected only assigned included outcome, got %#v", got)
	}
}

func TestCounselorCannotUpdateProfileFields(t *testing.T) {
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: "org-1"},
		updatedPatch: &patch,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", `{"included":true,"rationale":"Relevant","assignedUserID":"assessor-1","currentPriority":"High"}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || patch.CurrentPriority != nil {
		t.Fatalf("expected scope-only authorization, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}

func TestAssignedAssessorCanUpdateProfileOnly(t *testing.T) {
	organizationID := "org-1"
	assessorID := "assessor-1"
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: assessorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles:     []store.ProfileRow{{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &assessorID}},
		updatedPatch: &patch,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", `{"currentPriority":"High","targetCoverageLevel":"full"}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || patch.CurrentPriority == nil || *patch.CurrentPriority != "High" || patch.Included != nil || patch.AssignedUserID != nil {
		t.Fatalf("expected assigned profile update, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}

func TestUnassignedAssessorCannotUpdateProfile(t *testing.T) {
	organizationID := "org-1"
	assessorID := "assessor-1"
	otherAssessorID := "assessor-2"
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: assessorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles:     []store.ProfileRow{{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &otherAssessorID}},
		updatedPatch: &patch,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", `{"currentPriority":"High"}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "assigned") || patch.CurrentPriority != nil {
		t.Fatalf("expected unassigned profile denial, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}
