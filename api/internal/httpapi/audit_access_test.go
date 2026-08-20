package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

func TestAuditWriterCapturesRequestAndActorMetadata(t *testing.T) {
	organizationID := "org-1"
	var event store.AuditEvent
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, fakeStore{
		project:      store.Project{ID: "project-1", OrganizationID: organizationID},
		profiles:     []store.ProfileRow{{SubcategoryID: "22222222-2222-2222-2222-222222222222"}},
		updatedPatch: nil,
		auditEvent:   &event,
	})
	request := authenticatedRequest(http.MethodPut, "/api/projects/project-1/profile/22222222-2222-2222-2222-222222222222", `{"included":true}`)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "request-123")
	request.Header.Set("User-Agent", "audit-test/1.0")
	request.RemoteAddr = "192.0.2.10:54321"
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || event.ActorUserID != "counselor-1" || event.ActorRole != "counselor" || event.Result != "success" || event.RequestID != "request-123" || event.IPAddress != "192.0.2.10" || event.UserAgent != "audit-test/1.0" {
		t.Fatalf("unexpected audit context: status=%d event=%#v body=%s", response.Code, event, response.Body.String())
	}
}

func TestAssignedAuditorCanReadProjectAuditLogs(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"},
		auditorProjectAccess: map[string]bool{"project-1": true},
		projectAuditEvents:   []store.AuditTrailEntry{{ID: "event-1", Action: "response.reviewed", ActorRole: "reviewer", Result: "success"}},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-logs", ""))

	var events []store.AuditTrailEntry
	if response.Code != http.StatusOK || json.NewDecoder(response.Body).Decode(&events) != nil || len(events) != 1 || events[0].ActorRole != "reviewer" || events[0].Result != "success" {
		t.Fatalf("unexpected project audit log response: %d events=%#v body=%s", response.Code, events, response.Body.String())
	}
}

func TestUnassignedAuditorCannotReadProjectAuditLogs(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		project:              store.Project{ID: "project-1", OrganizationID: organizationID, Status: "in_review"},
		auditorProjectAccess: map[string]bool{},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/projects/project-1/audit-logs", ""))

	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "event-1") {
		t.Fatalf("expected hidden project audit logs, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAuditorOrganizationAuditLogsAreLimitedToGrantedProjects(t *testing.T) {
	organizationID := "org-1"
	handler := authenticatedHandler(store.User{ID: "auditor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "auditor", Status: "active"}, fakeStore{
		organization:            store.Organization{ID: organizationID},
		organizationAuditEvents: []store.AuditTrailEntry{{ID: "event-1", ProjectID: stringPtr("project-1")}},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, "/api/organizations/org-1/audit-logs", ""))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "event-1") {
		t.Fatalf("expected granted organization audit logs, got %d: %s", response.Code, response.Body.String())
	}
}
