package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

func TestAuditorCanReadAnAssignedSetupProject(t *testing.T) {
	organizationID := "org-1"
	auditorID := "auditor-1"
	handler := authenticatedHandler(store.User{ID: auditorID, OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-1", OrganizationID: organizationID, Status: "setup"},
		auditorProjectAccess: map[string]bool{"project-1": true},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("expected assigned Auditor to read setup Project, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAuditorCannotReadAnUnassignedProject(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-2", OrganizationID: organizationID, Status: "in_review"},
		auditorProjectAccess: map[string]bool{},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-2", ""))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected unassigned Project to be hidden, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAuditorCannotMutateAssignedProject(t *testing.T) {
	organizationID := "org-1"
	var patch store.ProfilePatch
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"},
		auditorProjectAccess: map[string]bool{"project-1": true},
		updatedPatch:         &patch,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/subcategory-1", `{"currentPriority":"High"}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || patch.CurrentPriority != nil {
		t.Fatalf("expected Auditor mutation to be denied, got %d patch=%#v body=%s", response.Code, patch, response.Body.String())
	}
}

func TestAuditorReadsIncludedProfileRowsInAssignedProject(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"},
		auditorProjectAccess: map[string]bool{"project-1": true},
		profiles: []store.ProfileRow{
			{SubcategoryID: "subcategory-1", SubcategoryCode: "GV.OC-01", Included: true},
			{SubcategoryID: "subcategory-2", SubcategoryCode: "GV.OC-02", Included: false},
		},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/profile", ""))

	var rows []store.ProfileRow
	if response.Code != http.StatusOK || json.NewDecoder(response.Body).Decode(&rows) != nil || len(rows) != 1 || rows[0].SubcategoryCode != "GV.OC-01" {
		t.Fatalf("expected included profile rows for Auditor, got %d rows=%#v body=%s", response.Code, rows, response.Body.String())
	}
}

func TestAuditorProjectListOnlyReturnsGrantedProjects(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		organization: store.Organization{ID: organizationID, Name: "Acme"},
		projects: []store.Project{
			{ID: "project-1", OrganizationID: organizationID, Name: "Granted", Status: "setup"},
			{ID: "project-2", OrganizationID: organizationID, Name: "Hidden", Status: "in_review"},
		},
		auditorProjectAccess: map[string]bool{"project-1": true},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/projects", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"name":"Granted"`) || strings.Contains(response.Body.String(), `"name":"Hidden"`) {
		t.Fatalf("expected only granted Projects, got %d: %s", response.Code, response.Body.String())
	}
}
