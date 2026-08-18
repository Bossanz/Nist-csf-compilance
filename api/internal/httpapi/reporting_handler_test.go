package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

func TestFinalReportReturnsReportForAuthorizedReader(t *testing.T) {
	organizationID := "org-1"
	report := store.FinalReport{Project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"}, Outcomes: []store.ReportOutcome{{Profile: store.ProfileRow{SubcategoryCode: "GV.OC-01"}}}}
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: report.Project, finalReport: report})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/final-report", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "GV.OC-01") {
		t.Fatalf("unexpected report response: %d %s", response.Code, response.Body.String())
	}
}

func TestAuditPackageReturnsCSVForAuthorizedReader(t *testing.T) {
	organizationID := "org-1"
	packageData := store.AuditPackage{
		Project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"},
		Outcomes: []store.ReportOutcome{{Profile: store.ProfileRow{FunctionCode: "GV", CategoryCode: "GV.OC", SubcategoryCode: "GV.OC-01"}}},
	}
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: packageData.Project, auditPackage: packageData})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-package.csv", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), "text/csv") || !strings.Contains(response.Body.String(), "GV.OC-01") {
		t.Fatalf("unexpected CSV response: %d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
}

func TestAuditPackageRejectsReaderFromAnotherOrganization(t *testing.T) {
	organizationID := "org-1"
	otherOrganizationID := "org-2"
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &otherOrganizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-package", ""))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", response.Code, response.Body.String())
	}
}
