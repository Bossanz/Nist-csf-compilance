package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/domain"
	"compliance/api/internal/store"
)

func stringPtr(value string) *string { return &value }

type fakeResponseStore struct {
	response      store.StakeholderResponse
	listErr       error
	saveErr       error
	submitErr     error
	reviewErr     error
	savedActor    string
	reviewedActor string
	reviewedState string
}

func (f *fakeResponseStore) ListResponses(context.Context, string) ([]store.StakeholderResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []store.StakeholderResponse{f.response}, nil
}
func (f *fakeResponseStore) SaveResponseDraft(_ context.Context, _, _, actor, _ string) (store.StakeholderResponse, error) {
	f.savedActor = actor
	return f.response, f.saveErr
}
func (f *fakeResponseStore) SubmitResponse(context.Context, string, string, string) (store.StakeholderResponse, error) {
	return f.response, f.submitErr
}
func (f *fakeResponseStore) ReviewResponse(_ context.Context, _, _, actor, status, _ string) (store.StakeholderResponse, error) {
	f.reviewedActor = actor
	f.reviewedState = status
	return f.response, f.reviewErr
}

func responseHandler(user store.User, responses *fakeResponseStore) *Handler {
	organizationID := "org-1"
	data := fakeStore{project: store.Project{ID: "project-1", OrganizationID: organizationID}}
	handler := authenticatedHandler(user, data)
	handler.Responses = responses
	return handler
}

func responseRequest(method, path, body string) *http.Request {
	request := authenticatedRequest(method, path, body)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestAssessorCanSaveResponse(t *testing.T) {
	responses := &fakeResponseStore{response: store.StakeholderResponse{ID: "response-1", Status: string(domain.ResponseDraft)}}
	handler := responseHandler(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, responses)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, responseRequest(http.MethodPut, "/api/projects/project-1/responses/subcategory-1", `{"responseText":"Quarterly review"}`))

	if response.Code != http.StatusOK || responses.savedActor != "assessor-1" {
		t.Fatalf("unexpected response: %d %s actor=%s", response.Code, response.Body.String(), responses.savedActor)
	}
}

func TestReviewerCanReviewSubmittedResponse(t *testing.T) {
	responses := &fakeResponseStore{response: store.StakeholderResponse{ID: "response-1", Status: string(domain.ResponseSubmitted)}}
	handler := responseHandler(store.User{ID: "reviewer-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "reviewer", Status: "active"}, responses)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, responseRequest(http.MethodPost, "/api/projects/project-1/responses/subcategory-1/review", `{"status":"reviewed","comment":"Evidence is sufficient"}`))

	if response.Code != http.StatusOK || responses.reviewedActor != "reviewer-1" || responses.reviewedState != "reviewed" {
		t.Fatalf("unexpected response: %d %s actor=%s state=%s", response.Code, response.Body.String(), responses.reviewedActor, responses.reviewedState)
	}
}

func TestViewerCannotMutateResponse(t *testing.T) {
	responses := &fakeResponseStore{}
	handler := responseHandler(store.User{ID: "viewer-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "viewer", Status: "active"}, responses)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, responseRequest(http.MethodPut, "/api/projects/project-1/responses/subcategory-1", `{"responseText":"Not allowed"}`))

	if response.Code != http.StatusForbidden || responses.savedActor != "" {
		t.Fatalf("unexpected response: %d %s actor=%s", response.Code, response.Body.String(), responses.savedActor)
	}
}

func TestDraftCannotBeReviewed(t *testing.T) {
	responses := &fakeResponseStore{reviewErr: domain.ErrInvalidResponseTransition}
	handler := responseHandler(store.User{ID: "reviewer-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "reviewer", Status: "active"}, responses)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, responseRequest(http.MethodPost, "/api/projects/project-1/responses/subcategory-1/review", `{"status":"reviewed"}`))

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"invalid_transition"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestResponseStoreFailureIsInternalError(t *testing.T) {
	responses := &fakeResponseStore{saveErr: errors.New("database unavailable")}
	handler := responseHandler(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, responses)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, responseRequest(http.MethodPut, "/api/projects/project-1/responses/subcategory-1", `{"responseText":"Answer"}`))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}
