package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authservice "compliance/api/internal/auth"
	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type fakeAuthRepository struct {
	user             store.User
	session          store.Session
	findUserErr      error
	findSessionErr   error
	createdTokenHash string
	revokedTokenHash string
}

func (f *fakeAuthRepository) FindUserByEmail(context.Context, string) (store.User, error) {
	return f.user, f.findUserErr
}
func (f *fakeAuthRepository) FindUserBySessionHash(context.Context, string) (store.User, store.Session, error) {
	return f.user, f.session, f.findSessionErr
}
func (f *fakeAuthRepository) CreateSession(_ context.Context, _ string, hash string, _ time.Time) error {
	f.createdTokenHash = hash
	return nil
}
func (f *fakeAuthRepository) RevokeSession(_ context.Context, hash string) error {
	f.revokedTokenHash = hash
	return nil
}
func (f *fakeAuthRepository) HasCounselorAdmin(context.Context) (bool, error) { return true, nil }
func (f *fakeAuthRepository) CreateCounselorAdmin(context.Context, string, string) (store.User, error) {
	return store.User{}, nil
}

func newAuthHandler(repo *fakeAuthRepository, now time.Time) *Handler {
	return &Handler{Auth: authservice.NewService(repo, func() time.Time { return now }), SecureCookies: true, LoginThrottle: NewLoginThrottle()}
}

func TestLoginSetsSecureSessionCookie(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	passwordHash, _ := authservice.HashPassword("secret-password")
	repo := &fakeAuthRepository{user: store.User{ID: "user-1", Email: "admin@example.com", Role: "counselor_admin", Status: "active", PasswordHash: passwordHash}}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"secret-password"}`))
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()

	newAuthHandler(repo, now).ServeHTTP(response, request)

	if response.Code != http.StatusOK || repo.createdTokenHash == "" {
		t.Fatalf("unexpected login response: %d %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected session cookie: %#v", cookies)
	}
}

func TestLoginUsesGenericCredentialFailure(t *testing.T) {
	repo := &fakeAuthRepository{findUserErr: pgx.ErrNoRows}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"missing@example.com","password":"wrong"}`))
	response := httptest.NewRecorder()

	newAuthHandler(repo, time.Now()).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) || strings.Contains(response.Body.String(), "missing@example.com") {
		t.Fatalf("unexpected failure response: %d %s", response.Code, response.Body.String())
	}
}

func TestMeRestoresAnActiveSession(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{user: store.User{ID: "user-1", Email: "admin@example.com", Role: "counselor_admin", Status: "active", PasswordHash: "must-not-leak"}, session: store.Session{ExpiresAt: now.Add(time.Hour)}}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	response := httptest.NewRecorder()

	newAuthHandler(repo, now).ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"role":"counselor_admin"`) || strings.Contains(response.Body.String(), "must-not-leak") {
		t.Fatalf("unexpected me response: %d %s", response.Code, response.Body.String())
	}
}

func TestLogoutRevokesSessionAndClearsCookie(t *testing.T) {
	repo := &fakeAuthRepository{}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	response := httptest.NewRecorder()

	newAuthHandler(repo, time.Now()).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || repo.revokedTokenHash != authservice.HashToken("raw-token") {
		t.Fatalf("unexpected logout response: %d hash=%s", response.Code, repo.revokedTokenHash)
	}
	if cookies := response.Result().Cookies(); len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Fatalf("session cookie was not cleared: %#v", cookies)
	}
}

func TestLoginThrottleRejectsSixthFailure(t *testing.T) {
	repo := &fakeAuthRepository{findUserErr: pgx.ErrNoRows}
	handler := newAuthHandler(repo, time.Now())
	for attempt := 1; attempt <= 6; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"missing@example.com","password":"wrong"}`))
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if attempt < 6 && response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", attempt, response.Code)
		}
		if attempt == 6 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("attempt 6: expected 429, got %d", response.Code)
		}
	}
}
