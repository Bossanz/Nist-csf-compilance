package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

type fakeRemediationEvidenceStore struct {
	evidence       store.RemediationEvidence
	err            error
	createdActor   string
	createdStorage string
	deletedID      string
}

func (f *fakeRemediationEvidenceStore) CreateRemediationEvidence(_ context.Context, _, _, actor, _, storagePath, _ string, _ int64) (store.RemediationEvidence, error) {
	f.createdActor = actor
	f.createdStorage = storagePath
	return f.evidence, f.err
}
func (f *fakeRemediationEvidenceStore) GetRemediationEvidence(context.Context, string, string, string) (store.RemediationEvidence, error) {
	return f.evidence, f.err
}
func (f *fakeRemediationEvidenceStore) DeleteRemediationEvidence(_ context.Context, _, _, evidenceID, _ string) (store.RemediationEvidence, error) {
	f.deletedID = evidenceID
	return f.evidence, f.err
}

func remediationEvidenceHandler(user store.User, documents *fakeRemediationEvidenceStore, files *fakeEvidenceStorage) *Handler {
	handler := remediationHandler(user, "closed", &fakeRemediationStore{})
	handler.RemediationEvidence = documents
	handler.Evidence = files
	return handler
}

func TestAssignedAssessorUploadsRemediationEvidence(t *testing.T) {
	organizationID := "org-1"
	documents := &fakeRemediationEvidenceStore{evidence: store.RemediationEvidence{ID: "evidence-1", ActionID: "action-1", OriginalName: "deployment.pdf"}}
	files := &fakeEvidenceStorage{storageKey: "opaque-key", size: 12}
	handler := remediationEvidenceHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, files)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/remediation-actions/action-1/evidence", "deployment.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusCreated || documents.createdActor != "assessor-1" || documents.createdStorage != "opaque-key" {
		t.Fatalf("unexpected response: %d %s actor=%s storage=%s", response.Code, response.Body.String(), documents.createdActor, documents.createdStorage)
	}
}

func TestRemediationEvidenceUploadCleansStoredFileOnGuardFailure(t *testing.T) {
	organizationID := "org-1"
	documents := &fakeRemediationEvidenceStore{err: store.ErrInvalidRemediationTransition}
	files := &fakeEvidenceStorage{storageKey: "opaque-key", size: 12}
	handler := remediationEvidenceHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, files)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/remediation-actions/action-1/evidence", "deployment.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusConflict || files.removedKey != "opaque-key" {
		t.Fatalf("unexpected response: %d %s removed=%s", response.Code, response.Body.String(), files.removedKey)
	}
}

func TestReviewerPreviewsRemediationEvidenceInline(t *testing.T) {
	organizationID := "org-1"
	documents := &fakeRemediationEvidenceStore{evidence: store.RemediationEvidence{ID: "evidence-1", ActionID: "action-1", StoragePath: "opaque-key", OriginalName: "deployment.pdf", MIMEType: "application/pdf", SizeBytes: 8}}
	files := &fakeEvidenceStorage{open: io.NopCloser(strings.NewReader("%PDF-1.7"))}
	handler := remediationEvidenceHandler(store.User{ID: "reviewer-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "reviewer", Status: "active"}, documents, files)
	response := httptest.NewRecorder()
	request := authenticatedRequest(http.MethodGet, "/api/projects/project-1/remediation-actions/action-1/evidence/evidence-1/preview", "")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "%PDF-1.7" || !strings.HasPrefix(response.Header().Get("Content-Disposition"), "inline") {
		t.Fatalf("unexpected response: %d %s headers=%v", response.Code, response.Body.String(), response.Header())
	}
}

func TestAssignedAssessorDeletesRemediationEvidence(t *testing.T) {
	organizationID := "org-1"
	documents := &fakeRemediationEvidenceStore{evidence: store.RemediationEvidence{ID: "evidence-1", ActionID: "action-1", StoragePath: "opaque-key"}}
	files := &fakeEvidenceStorage{}
	handler := remediationEvidenceHandler(store.User{ID: "assessor-1", OrganizationID: &organizationID, UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, files)
	response := httptest.NewRecorder()
	request := authenticatedRequest(http.MethodDelete, "/api/projects/project-1/remediation-actions/action-1/evidence/evidence-1", "")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || documents.deletedID != "evidence-1" || files.removedKey != "opaque-key" {
		t.Fatalf("unexpected response: %d %s deleted=%s removed=%s", response.Code, response.Body.String(), documents.deletedID, files.removedKey)
	}
}
