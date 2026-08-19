package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

type fakeRemediationStore struct {
	action        store.RemediationAction
	actions       []store.RemediationAction
	err           error
	createdActor  string
	progressActor string
	submitActor   string
	reviewActor   string
	decision      string
}

func (f *fakeRemediationStore) ListRemediationActions(context.Context, string) ([]store.RemediationAction, error) {
	return f.actions, f.err
}
func (f *fakeRemediationStore) CreateRemediationAction(_ context.Context, _, actor string, _ store.RemediationCreate) (store.RemediationAction, error) {
	f.createdActor = actor
	return f.action, f.err
}
func (f *fakeRemediationStore) UpdateRemediationAction(context.Context, string, string, string, store.RemediationPatch) (store.RemediationAction, error) {
	return f.action, f.err
}
func (f *fakeRemediationStore) UpdateRemediationProgress(_ context.Context, _, _, actor, _ string) (store.RemediationAction, error) {
	f.progressActor = actor
	return f.action, f.err
}
func (f *fakeRemediationStore) SubmitRemediationAction(_ context.Context, _, _, actor string) (store.RemediationAction, error) {
	f.submitActor = actor
	return f.action, f.err
}
func (f *fakeRemediationStore) ReviewRemediationAction(_ context.Context, _, _, actor, decision, _ string) (store.RemediationAction, error) {
	f.reviewActor = actor
	f.decision = decision
	return f.action, f.err
}

func remediationHandler(user store.User, projectStatus string, actions *fakeRemediationStore) *Handler {
	organizationID := "org-1"
	data := fakeStore{project: store.Project{ID: "project-1", OrganizationID: organizationID, Status: projectStatus}}
	handler := authenticatedHandler(user, data)
	handler.Remediations = actions
	return handler
}

func remediationRequest(method, path, body string) *http.Request {
	request := authenticatedRequest(method, path, body)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestCounselorCreatesRemediationAction(t *testing.T) {
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", ProjectID: "project-1", SubcategoryID: "outcome-1", Status: "open"}}
	handler := remediationHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, "in_review", actions)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions", `{"subcategoryID":"outcome-1","title":"Centralize logs","description":"Forward logs","desiredResult":"Searchable events","priority":"high","ownerUserID":"assessor-1","dueDate":"2026-09-30"}`))

	if response.Code != http.StatusCreated || actions.createdActor != "counselor-1" {
		t.Fatalf("unexpected response: %d %s actor=%s", response.Code, response.Body.String(), actions.createdActor)
	}
}

func TestRemediationCreationRemainsAvailableAfterFinalization(t *testing.T) {
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", Status: "open"}}
	handler := remediationHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, "closed", actions)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions", `{"subcategoryID":"outcome-1","title":"Centralize logs","priority":"high","ownerUserID":"assessor-1","dueDate":"2026-09-30"}`))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected remediation after finalization, got %d %s", response.Code, response.Body.String())
	}
}

func TestAssessorCannotCreateRemediationAction(t *testing.T) {
	organizationID := "org-1"
	actions := &fakeRemediationStore{}
	handler := remediationHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, "in_review", actions)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions", `{"subcategoryID":"outcome-1","title":"Not allowed","priority":"high","ownerUserID":"assessor-1","dueDate":"2026-09-30"}`))

	if response.Code != http.StatusForbidden || actions.createdActor != "" {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestAssignedAssessorUpdatesAndSubmitsRemediation(t *testing.T) {
	organizationID := "org-1"
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", Status: "in_progress"}}
	user := store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}
	handler := remediationHandler(user, "closed", actions)
	progressResponse := httptest.NewRecorder()
	handler.ServeHTTP(progressResponse, remediationRequest(http.MethodPatch, "/api/projects/project-1/remediation-actions/action-1/progress", `{"progressNote":"SIEM forwarding enabled."}`))
	if progressResponse.Code != http.StatusOK || actions.progressActor != "assessor-1" {
		t.Fatalf("unexpected progress response: %d %s", progressResponse.Code, progressResponse.Body.String())
	}

	submitResponse := httptest.NewRecorder()
	handler.ServeHTTP(submitResponse, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions/action-1/submit", `{}`))
	if submitResponse.Code != http.StatusOK || actions.submitActor != "assessor-1" {
		t.Fatalf("unexpected submit response: %d %s", submitResponse.Code, submitResponse.Body.String())
	}
}

func TestReviewerCanReadButCannotMutateRemediation(t *testing.T) {
	organizationID := "org-1"
	actions := &fakeRemediationStore{actions: []store.RemediationAction{{ID: "action-1", Status: "open"}}}
	user := store.User{ID: "reviewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "reviewer", Status: "active"}
	handler := remediationHandler(user, "in_review", actions)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, remediationRequest(http.MethodGet, "/api/projects/project-1/remediation-actions", ""))
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "action-1") {
		t.Fatalf("unexpected list response: %d %s", listResponse.Code, listResponse.Body.String())
	}

	progressResponse := httptest.NewRecorder()
	handler.ServeHTTP(progressResponse, remediationRequest(http.MethodPatch, "/api/projects/project-1/remediation-actions/action-1/progress", `{"progressNote":"Not allowed"}`))
	if progressResponse.Code != http.StatusForbidden || actions.progressActor != "" {
		t.Fatalf("unexpected mutation response: %d %s", progressResponse.Code, progressResponse.Body.String())
	}
}

func TestCounselorReviewsRemediationAndWritesAudit(t *testing.T) {
	auditAction := ""
	actions := &fakeRemediationStore{action: store.RemediationAction{ID: "action-1", ProjectID: "project-1", SubcategoryID: "outcome-1", Status: "closed"}}
	data := fakeStore{project: store.Project{ID: "project-1", OrganizationID: "org-1", Status: "in_review"}, auditAction: &auditAction}
	handler := authenticatedHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, data)
	handler.Remediations = actions
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions/action-1/review", `{"decision":"close","comment":"Verified."}`))

	if response.Code != http.StatusOK || actions.reviewActor != "counselor-1" || actions.decision != "close" || auditAction != "remediation.closed" {
		t.Fatalf("unexpected response: %d %s actor=%s decision=%s audit=%s", response.Code, response.Body.String(), actions.reviewActor, actions.decision, auditAction)
	}
}

func TestRemediationStoreErrorsUseStableCodes(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{store.ErrOutcomeNotApproved, "outcome_not_approved"},
		{store.ErrNoCoverageGap, "no_coverage_gap"},
		{store.ErrInvalidRemediationTransition, "invalid_remediation_transition"},
		{store.ErrRemediationClosed, "remediation_closed"},
		{store.ErrInvalidRemediationOwner, "validation_error"},
		{store.ErrInvalidRemediationInput, "validation_error"},
		{store.ErrRemediationForbidden, "forbidden"},
	}
	for _, test := range tests {
		actions := &fakeRemediationStore{err: test.err}
		handler := remediationHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, "in_review", actions)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, remediationRequest(http.MethodPost, "/api/projects/project-1/remediation-actions", `{"subcategoryID":"outcome-1","title":"Centralize logs","priority":"high","ownerUserID":"assessor-1","dueDate":"2026-09-30"}`))
		if !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
			t.Fatalf("error %v: status=%d body=%s", test.err, response.Code, response.Body.String())
		}
	}
}

func TestRemediationUnexpectedStoreFailureIsInternalError(t *testing.T) {
	actions := &fakeRemediationStore{err: errors.New("database unavailable")}
	handler := remediationHandler(store.User{ID: "counselor-1", UserType: "counselor", Role: "counselor", Status: "active"}, "in_review", actions)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, remediationRequest(http.MethodGet, "/api/projects/project-1/remediation-actions", ""))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}
