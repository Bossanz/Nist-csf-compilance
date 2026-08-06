package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type fakeStore struct {
	projects []store.Project
	listErr  error
	deleteErr error
	deletedID *string
}

func (f fakeStore) ListProjects(context.Context) ([]store.Project, error) {
	return f.projects, f.listErr
}
func (f fakeStore) ListFunctions(context.Context) ([]store.Function, error) { return nil, nil }
func (f fakeStore) CreateProject(context.Context, string, string) (store.Project, error) {
	return store.Project{}, nil
}
func (f fakeStore) GetProject(context.Context, string) (store.Project, error) {
	return store.Project{}, nil
}
func (f fakeStore) ListProfile(context.Context, string) ([]store.ProfileRow, error) { return nil, nil }
func (f fakeStore) UpdateProfile(context.Context, string, string, store.ProfilePatch) (store.ProfileRow, error) {
	return store.ProfileRow{}, nil
}
func (f fakeStore) DeleteProject(_ context.Context, id string) error { if f.deletedID != nil { *f.deletedID = id }; return f.deleteErr }

func TestHealthz(t *testing.T) {
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestListProjects(t *testing.T) {
	expected := []store.Project{{ID: "project-1", Name: "Readiness", OrganizationName: "Acme", Status: "setup"}}
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{projects: expected}}).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got []store.Project
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, got %#v", expected, got)
	}
}

func TestListProjectsReturnsEmptyArray(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{projects: []store.Project{}}}).ServeHTTP(w, r)

	if w.Code != http.StatusOK || w.Body.String() != "[]\n" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestListProjectsHandlesStoreFailure(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{listErr: errors.New("database unavailable")}}).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProject(t *testing.T) {
	var deletedID string
	projectID := "11111111-1111-1111-1111-111111111111"
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deletedID: &deletedID}}).ServeHTTP(w, r)

	if w.Code != http.StatusNoContent || w.Body.Len() != 0 || deletedID != projectID {
		t.Fatalf("unexpected response: %d %s id=%s", w.Code, w.Body.String(), deletedID)
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deleteErr: pgx.ErrNoRows}}).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProjectHandlesStoreFailure(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deleteErr: errors.New("database unavailable")}}).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestDeleteProjectRejectsMalformedID(t *testing.T) {
	var deletedID string
	r := httptest.NewRequest(http.MethodDelete, "/api/projects/not-a-uuid", nil)
	w := httptest.NewRecorder()

	(&Handler{Store: fakeStore{deletedID: &deletedID}}).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound || deletedID != "" {
		t.Fatalf("unexpected response: %d %s id=%s", w.Code, w.Body.String(), deletedID)
	}
}
