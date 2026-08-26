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
	createdProjectMetadata       *store.ProjectMetadata
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
	bulkScopeFunctionCode        *string
	bulkScopeIncluded            *bool
	bulkScopeRows                []store.ProfileRow
	bulkScopeErr                 error
	submittedScopeProjectID      *string
	submitScopeResult            store.Project
	submitScopeErr               error
	finalizedProject             store.Project
	finalizeApproved             int
	finalizeIncluded             int
	finalizeErr                  error
	finalizedProjectID           *string
	finalizedActorID             *string
	finalReport                  store.FinalReport
	auditPackage                 store.AuditPackage
	reportingErr                 error
	auditorProjectAccess         map[string]bool
	auditorAccessErr             error
	auditEvent                   *store.AuditEvent
	projectAuditEvents           []store.AuditTrailEntry
	organizationAuditEvents      []store.AuditTrailEntry
	invitations                  []store.Invitation
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
func (f fakeStore) CreateScopedProjectWithMetadata(_ context.Context, organizationID, name string, metadata store.ProjectMetadata) (store.Project, error) {
	if f.createdProjectMetadata != nil {
		*f.createdProjectMetadata = metadata
	}
	return f.CreateScopedProject(context.Background(), organizationID, name)
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
func (f fakeStore) UpdateFunctionScope(_ context.Context, _, functionCode string, included bool) ([]store.ProfileRow, error) {
	if f.bulkScopeFunctionCode != nil {
		*f.bulkScopeFunctionCode = functionCode
	}
	if f.bulkScopeIncluded != nil {
		*f.bulkScopeIncluded = included
	}
	return f.bulkScopeRows, f.bulkScopeErr
}
func (f fakeStore) SubmitProjectScope(_ context.Context, id string) (store.Project, error) {
	if f.submittedScopeProjectID != nil {
		*f.submittedScopeProjectID = id
	}
	return f.submitScopeResult, f.submitScopeErr
}
func (f fakeStore) FinalizeProject(_ context.Context, projectID, actorID string) (store.Project, int, int, error) {
	if f.finalizedProjectID != nil {
		*f.finalizedProjectID = projectID
	}
	if f.finalizedActorID != nil {
		*f.finalizedActorID = actorID
	}
	return f.finalizedProject, f.finalizeApproved, f.finalizeIncluded, f.finalizeErr
}
func (f fakeStore) GetFinalReport(context.Context, string) (store.FinalReport, error) {
	return f.finalReport, f.reportingErr
}
func (f fakeStore) GetAuditPackage(context.Context, string) (store.AuditPackage, error) {
	return f.auditPackage, f.reportingErr
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
	if f.auditEvent != nil {
		*f.auditEvent = event
	}
	return nil
}
func (f fakeStore) ListProjectAuditEvents(context.Context, string) ([]store.AuditTrailEntry, error) {
	return f.projectAuditEvents, nil
}
func (f fakeStore) ListOrganizationAuditEvents(context.Context, string, string) ([]store.AuditTrailEntry, error) {
	return f.organizationAuditEvents, nil
}
func (f fakeStore) ListOrganizationInvitations(context.Context, string) ([]store.Invitation, error) {
	return f.invitations, nil
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
func (f fakeStore) HasActiveProjectAuditorAccess(_ context.Context, projectID, _ string) (bool, error) {
	if f.auditorAccessErr != nil {
		return false, f.auditorAccessErr
	}
	return f.auditorProjectAccess[projectID], nil
}

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

func TestCounselorFinalizesProject(t *testing.T) {
	var projectID, actorID string
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		finalizedProject:   store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"},
		finalizeApproved:   1,
		finalizeIncluded:   1,
		finalizedProjectID: &projectID,
		finalizedActorID:   &actorID,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/finalize", ""))

	if response.Code != http.StatusOK || projectID != "project-1" || actorID != "counselor-1" || !strings.Contains(response.Body.String(), `"status":"closed"`) {
		t.Fatalf("unexpected response: %d %s project=%s actor=%s", response.Code, response.Body.String(), projectID, actorID)
	}
}

func TestAssessorCannotFinalizeProject(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/finalize", ""))

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
}

func TestFinalizationReportsProjectNotReady(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		project:          store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"},
		finalizeErr:      store.ErrProjectNotReady,
		finalizeApproved: 0,
		finalizeIncluded: 1,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/finalize", ""))

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"project_not_ready"`) {
		t.Fatalf("expected project_not_ready conflict, got %d: %s", response.Code, response.Body.String())
	}
}

func TestFinalizedProjectRejectsProfileMutation(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", `{"currentPriority":"High"}`))

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"project_finalized"`) {
		t.Fatalf("expected project_finalized conflict, got %d: %s", response.Code, response.Body.String())
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

func TestAssessorCannotUpdateProfileAfterResponseIsSubmitted(t *testing.T) {
	organizationID := "org-1"
	assessorID := "assessor-1"
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: assessorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles:     []store.ProfileRow{{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &assessorID}},
		updatedPatch: &patch,
	})
	handler.Responses = &fakeResponseStore{response: store.StakeholderResponse{ID: "response-1", SubcategoryID: "subcategory-1", Status: "submitted"}}
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", "{\"currentPriority\":\"High\"}")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict || patch.CurrentPriority != nil || !strings.Contains(response.Body.String(), "\"code\":\"invalid_transition\"") {
		t.Fatalf("expected submitted profile lock, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}

func TestAssessorCannotUpdateProfileAfterResponseIsApproved(t *testing.T) {
	organizationID := "org-1"
	assessorID := "assessor-1"
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: assessorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles:     []store.ProfileRow{{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &assessorID}},
		updatedPatch: &patch,
	})
	handler.Responses = &fakeResponseStore{response: store.StakeholderResponse{ID: "response-1", SubcategoryID: "subcategory-1", Status: "reviewed"}}
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", "{\"currentPriority\":\"High\"}")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict || patch.CurrentPriority != nil || !strings.Contains(response.Body.String(), "\"code\":\"invalid_transition\"") {
		t.Fatalf("expected approved profile lock, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}
func TestCounselorCanSubmitProjectScope(t *testing.T) {
	var submittedID string
	project := store.Project{ID: "project-1", OrganizationID: "org-1", Status: "setup"}
	result := project
	result.Status = "in_review"
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		project:                 project,
		submittedScopeProjectID: &submittedID,
		submitScopeResult:       result,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/scope/submit", ""))

	if response.Code != http.StatusOK || submittedID != "project-1" || !strings.Contains(response.Body.String(), "in_review") {
		t.Fatalf("unexpected scope submission: %d %s id=%s", response.Code, response.Body.String(), submittedID)
	}
}

func TestCounselorCanUpdateFunctionScopeInOneRequest(t *testing.T) {
	organizationID := "org-1"
	var functionCode string
	var included bool
	var auditAction string
	handler := authenticatedHandler(store.User{UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		project:               store.Project{ID: "project-1", OrganizationID: organizationID},
		bulkScopeFunctionCode: &functionCode,
		bulkScopeIncluded:     &included,
		bulkScopeRows:         []store.ProfileRow{{SubcategoryID: "subcategory-1", FunctionCode: "GV", Included: true}},
		auditAction:           &auditAction,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/functions/GV/scope", `{"included":true}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || functionCode != "GV" || !included || auditAction != "profile.function_scope_updated" || !strings.Contains(response.Body.String(), "subcategory-1") {
		t.Fatalf("unexpected bulk scope response: %d function=%s included=%v audit=%s body=%s", response.Code, functionCode, included, auditAction, response.Body.String())
	}
}

func TestStakeholderCannotUpdateFunctionScope(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, fakeStore{
		project: store.Project{ID: "project-1", OrganizationID: organizationID},
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/functions/GV/scope", `{"included":true}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
}

func TestStakeholderCannotSubmitProjectScope(t *testing.T) {
	organizationID := "org-1"
	var submittedID string
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{
		project:                 store.Project{ID: "project-1", OrganizationID: organizationID, Status: "setup"},
		submittedScopeProjectID: &submittedID,
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, "/api/projects/project-1/scope/submit", ""))

	if response.Code != http.StatusNotFound || submittedID != "" {
		t.Fatalf("expected stakeholder scope submission to be denied: %d %s id=%s", response.Code, response.Body.String(), submittedID)
	}
}

func TestStakeholderCannotReadSetupProject(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{
		project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "setup"},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1", ""))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected setup Project to be hidden from Stakeholder: %d %s", response.Code, response.Body.String())
	}
}
