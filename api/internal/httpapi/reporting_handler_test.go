package httpapi

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
		Project:  store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"},
		Outcomes: []store.ReportOutcome{{Profile: store.ProfileRow{FunctionCode: "GV", CategoryCode: "GV.OC", SubcategoryCode: "GV.OC-01"}}},
	}
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: packageData.Project, auditPackage: packageData})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-package.csv", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), "text/csv") || !strings.Contains(response.Body.String(), "GV.OC-01") {
		t.Fatalf("unexpected CSV response: %d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
}

func TestAuditPackageCSVIncludesRemediationRegister(t *testing.T) {
	organizationID := "org-1"
	packageData := store.AuditPackage{
		Project:            store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed"},
		RemediationActions: []store.RemediationAction{{ID: "action-1", OutcomeCode: "GV.OC-01", Title: "Centralize logs", OwnerName: "Assessor", Priority: "high", Status: "in_progress", DueDate: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)}},
	}
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: packageData.Project, auditPackage: packageData})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-package.csv", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "record_type") || !strings.Contains(response.Body.String(), "remediation_action") || !strings.Contains(response.Body.String(), "Centralize logs") {
		t.Fatalf("unexpected remediation CSV: %d %s", response.Code, response.Body.String())
	}
	records, err := csv.NewReader(strings.NewReader(response.Body.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for index, record := range records {
		if len(record) != len(records[0]) {
			t.Fatalf("CSV row %d has %d columns; expected %d", index, len(record), len(records[0]))
		}
	}
}

func TestAuditPackageCSVIncludesAssessmentVersion(t *testing.T) {
	organizationID := "org-1"
	packageData := store.AuditPackage{
		Project:  store.Project{ID: "project-1", OrganizationID: organizationID, Status: "closed", VersionNumber: 2},
		Outcomes: []store.ReportOutcome{{Profile: store.ProfileRow{FunctionCode: "GV", CategoryCode: "GV.OC", SubcategoryCode: "GV.OC-01"}}},
	}
	handler := authenticatedHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, fakeStore{project: packageData.Project, auditPackage: packageData})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-package.csv", ""))

	records, err := csv.NewReader(strings.NewReader(response.Body.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	versionColumn := -1
	for index, column := range records[0] {
		if column == "assessment_version" {
			versionColumn = index
			break
		}
	}
	if versionColumn < 0 || len(records) < 2 || records[1][versionColumn] != "2" {
		t.Fatalf("expected assessment version 2 in CSV, got %v", records)
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
