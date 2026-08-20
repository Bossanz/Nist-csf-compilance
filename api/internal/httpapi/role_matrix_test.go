package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"compliance/api/internal/store"
)

func TestViewerCannotMutateRemediation(t *testing.T) {
	organizationID := "org-1"
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", Status: "open"}}
	handler := remediationHandler(store.User{ID: "viewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "viewer", Status: "active"}, "in_review", actions)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPatch, "/api/projects/project-1/remediation-actions/action-1/progress", `{"progressNote":"Not allowed"}`))

	if response.Code != http.StatusForbidden || actions.progressActor != "" {
		t.Fatalf("viewer mutation was not denied: %d %s", response.Code, response.Body.String())
	}
}

func TestOrgAdminCanProgressAssignedRemediationButCannotReviewIt(t *testing.T) {
	organizationID := "org-1"
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", Status: "in_progress"}}
	user := store.User{ID: "org-admin-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "org_admin", Status: "active"}
	handler := remediationHandler(user, "closed", actions)

	progressResponse := httptest.NewRecorder()
	handler.ServeHTTP(progressResponse, remediationRequest(http.MethodPatch, "/api/projects/project-1/remediation-actions/action-1/progress", `{"progressNote":"Progress recorded"}`))
	if progressResponse.Code != http.StatusOK || actions.progressActor != user.ID {
		t.Fatalf("org admin progress was not allowed: %d %s", progressResponse.Code, progressResponse.Body.String())
	}

	reviewResponse := httptest.NewRecorder()
	handler.ServeHTTP(reviewResponse, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions/action-1/review", `{"decision":"close","comment":"Not allowed"}`))
	if reviewResponse.Code != http.StatusForbidden || actions.reviewActor != "" {
		t.Fatalf("org admin review was not denied: %d %s", reviewResponse.Code, reviewResponse.Body.String())
	}
}

func TestCounselorAdminCanManageRemediationAfterFinalization(t *testing.T) {
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", Status: "awaiting_review"}}
	handler := remediationHandler(store.User{ID: "admin-1", UserType: "counselor", Role: "counselor_admin", Status: "active"}, "closed", actions)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions/action-1/review", `{"decision":"close","comment":"Verified"}`))

	if response.Code != http.StatusOK || actions.reviewActor != "admin-1" {
		t.Fatalf("counselor admin could not manage finalized remediation: %d %s", response.Code, response.Body.String())
	}
}
