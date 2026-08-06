package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

func TestOrgAdminListsOrganizationUsers(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}, fakeStore{organization: store.Organization{ID: organizationID}, users: []store.User{{ID: "user-2", Email: "person@acme.test", Role: "viewer"}}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/users", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"email":"person@acme.test"`) {
		t.Fatalf("unexpected users response: %d %s", response.Code, response.Body.String())
	}
}

func TestOrgAdminChangesStakeholderRole(t *testing.T) {
	organizationID := "org-1"
	var role, status string
	handler := authenticatedHandler(store.User{ID: "admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}, fakeStore{organization: store.Organization{ID: organizationID}, updatedUserRole: &role, updatedUserStatus: &status})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/organizations/org-1/users/user-2", `{"role":"assessor","status":"active"}`))

	if response.Code != http.StatusOK || role != "assessor" || status != "active" {
		t.Fatalf("unexpected role update: %d %s role=%s status=%s", response.Code, response.Body.String(), role, status)
	}
}

func TestOrgAdminCannotGrantCounselorRole(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}, fakeStore{organization: store.Organization{ID: organizationID}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/organizations/org-1/users/user-2", `{"role":"counselor","status":"active"}`))

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
}

func TestCounselorAdminDisablesCounselor(t *testing.T) {
	var role, status string
	handler := authenticatedHandler(store.User{ID: "admin-1", UserType: "counselor", Role: "counselor_admin", Status: "active"}, fakeStore{updatedUserRole: &role, updatedUserStatus: &status})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodPatch, "/api/counselors/user-2", `{"role":"counselor","status":"disabled"}`))

	if response.Code != http.StatusOK || role != "counselor" || status != "disabled" {
		t.Fatalf("unexpected counselor update: %d %s", response.Code, response.Body.String())
	}
}
