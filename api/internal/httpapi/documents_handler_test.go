package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"compliance/api/internal/store"
)

type fakeDocumentStore struct {
	document       store.ResponseDocument
	createErr      error
	getErr         error
	deleteErr      error
	createdActor   string
	createdStorage string
	deletedID      string
}

func (f *fakeDocumentStore) CreateResponseDocument(_ context.Context, _, _, actor, _, storageKey, _ string, _ int64) (store.ResponseDocument, error) {
	f.createdActor = actor
	f.createdStorage = storageKey
	return f.document, f.createErr
}
func (f *fakeDocumentStore) GetResponseDocument(context.Context, string, string, string) (store.ResponseDocument, error) {
	return f.document, f.getErr
}
func (f *fakeDocumentStore) DeleteResponseDocument(_ context.Context, _, _, documentID string) (store.ResponseDocument, error) {
	f.deletedID = documentID
	return f.document, f.deleteErr
}

type fakeEvidenceStorage struct {
	storageKey string
	size       int64
	saveErr    error
	open       io.ReadCloser
	openErr    error
	removeErr  error
	removedKey string
}

func (f *fakeEvidenceStorage) Save(io.Reader) (string, int64, error) {
	return f.storageKey, f.size, f.saveErr
}
func (f *fakeEvidenceStorage) Open(string) (io.ReadCloser, error) {
	return f.open, f.openErr
}
func (f *fakeEvidenceStorage) Remove(storageKey string) error {
	f.removedKey = storageKey
	return f.removeErr
}

func documentHandler(user store.User, documents *fakeDocumentStore, evidence *fakeEvidenceStorage) *Handler {
	row := store.ProfileRow{SubcategoryID: "subcategory-1", Included: true}
	if user.Role == "org_admin" || user.Role == "assessor" {
		assignedUserID := user.ID
		row.AssignedUserID = &assignedUserID
	}
	return documentHandlerWithProfiles(user, documents, evidence, []store.ProfileRow{row})
}

func documentHandlerWithProfiles(user store.User, documents *fakeDocumentStore, evidence *fakeEvidenceStorage, profiles []store.ProfileRow) *Handler {
	handler := responseHandlerWithProfiles(user, &fakeResponseStore{}, profiles)
	handler.Documents = documents
	handler.Evidence = evidence
	return handler
}

func multipartRequest(t *testing.T, path, fileName, contentType string, contents []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-Upload-Content-Type", contentType)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	return request
}

func TestAssessorCanUploadEvidence(t *testing.T) {
	documents := &fakeDocumentStore{document: store.ResponseDocument{ID: "doc-1", OriginalName: "evidence.pdf"}}
	evidence := &fakeEvidenceStorage{storageKey: "opaque-key", size: 12}
	handler := documentHandler(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence)
	response := httptest.NewRecorder()

	request := multipartRequest(t, "/api/projects/project-1/responses/subcategory-1/documents", "evidence.pdf", "application/pdf", []byte("%PDF-1.7"))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || documents.createdActor != "assessor-1" || documents.createdStorage != "opaque-key" {
		t.Fatalf("unexpected response: %d %s actor=%s key=%s", response.Code, response.Body.String(), documents.createdActor, documents.createdStorage)
	}
}

func TestStakeholderCannotUploadEvidenceForExcludedOutcome(t *testing.T) {
	documents := &fakeDocumentStore{}
	evidence := &fakeEvidenceStorage{storageKey: "opaque-key"}
	handler := documentHandlerWithProfiles(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence, []store.ProfileRow{{SubcategoryID: "subcategory-1", Included: false}})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/responses/subcategory-1/documents", "evidence.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusNotFound || documents.createdActor != "" || evidence.removedKey != "" {
		t.Fatalf("unexpected response: %d %s actor=%s removed=%s", response.Code, response.Body.String(), documents.createdActor, evidence.removedKey)
	}
}

func TestUnassignedAssessorCannotUploadEvidence(t *testing.T) {
	documents := &fakeDocumentStore{}
	evidence := &fakeEvidenceStorage{storageKey: "opaque-key"}
	assessorID := "assessor-1"
	otherAssessorID := "assessor-2"
	handler := documentHandlerWithProfiles(store.User{ID: assessorID, OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence, []store.ProfileRow{
		{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &otherAssessorID},
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/responses/subcategory-1/documents", "evidence.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusForbidden || documents.createdActor != "" || evidence.removedKey != "" {
		t.Fatalf("expected unassigned upload denial, got %d actor=%s removed=%s body=%s", response.Code, documents.createdActor, evidence.removedKey, response.Body.String())
	}
}

func TestViewerCannotUploadEvidence(t *testing.T) {
	documents := &fakeDocumentStore{}
	evidence := &fakeEvidenceStorage{storageKey: "opaque-key"}
	handler := documentHandler(store.User{ID: "viewer-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "viewer", Status: "active"}, documents, evidence)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/responses/subcategory-1/documents", "evidence.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusForbidden || documents.createdActor != "" || evidence.storageKey == evidence.removedKey {
		t.Fatalf("unexpected response: %d %s actor=%s", response.Code, response.Body.String(), documents.createdActor)
	}
}

func TestUploadCleansFileWhenMetadataInsertFails(t *testing.T) {
	documents := &fakeDocumentStore{createErr: errors.New("database unavailable")}
	evidence := &fakeEvidenceStorage{storageKey: "opaque-key", size: 12}
	handler := documentHandler(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, multipartRequest(t, "/api/projects/project-1/responses/subcategory-1/documents", "evidence.pdf", "application/pdf", []byte("%PDF-1.7")))

	if response.Code != http.StatusInternalServerError || evidence.removedKey != "opaque-key" {
		t.Fatalf("unexpected response: %d %s removed=%s", response.Code, response.Body.String(), evidence.removedKey)
	}
}

func TestReviewerCanDownloadEvidence(t *testing.T) {
	documents := &fakeDocumentStore{document: store.ResponseDocument{ID: "doc-1", StorageKey: "opaque-key", OriginalName: "evidence.pdf", MIMEType: "application/pdf", SizeBytes: 8}}
	evidence := &fakeEvidenceStorage{open: io.NopCloser(strings.NewReader("%PDF-1.7"))}
	handler := documentHandler(store.User{ID: "reviewer-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "reviewer", Status: "active"}, documents, evidence)
	response := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/responses/subcategory-1/documents/doc-1", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "%PDF-1.7" || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("unexpected response: %d %s headers=%v", response.Code, response.Body.String(), response.Header())
	}
}

func TestAssessorCanDeleteEvidence(t *testing.T) {
	documents := &fakeDocumentStore{document: store.ResponseDocument{ID: "doc-1", StorageKey: "opaque-key", CreatedAt: time.Now()}}
	evidence := &fakeEvidenceStorage{}
	handler := documentHandler(store.User{ID: "assessor-1", OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence)
	response := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodDelete, "/api/projects/project-1/responses/subcategory-1/documents/doc-1", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || documents.deletedID != "doc-1" || evidence.removedKey != "opaque-key" {
		t.Fatalf("unexpected response: %d %s deleted=%s removed=%s", response.Code, response.Body.String(), documents.deletedID, evidence.removedKey)
	}
}

func TestUnassignedAssessorCannotDeleteEvidence(t *testing.T) {
	documents := &fakeDocumentStore{document: store.ResponseDocument{ID: "doc-1", StorageKey: "opaque-key", CreatedAt: time.Now()}}
	evidence := &fakeEvidenceStorage{}
	assessorID := "assessor-1"
	otherAssessorID := "assessor-2"
	handler := documentHandlerWithProfiles(store.User{ID: assessorID, OrganizationID: stringPtr("org-1"), UserType: "stakeholder", Role: "assessor", Status: "active"}, documents, evidence, []store.ProfileRow{
		{SubcategoryID: "subcategory-1", Included: true, AssignedUserID: &otherAssessorID},
	})
	response := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodDelete, "/api/projects/project-1/responses/subcategory-1/documents/doc-1", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || documents.deletedID != "" {
		t.Fatalf("expected unassigned delete denial, got %d deleted=%s body=%s", response.Code, documents.deletedID, response.Body.String())
	}
}
