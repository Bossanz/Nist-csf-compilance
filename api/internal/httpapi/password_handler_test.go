package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"compliance/api/internal/notifications"
	"compliance/api/internal/store"
)

type fakePasswordService struct {
	requestUser       store.User
	requestToken      string
	requestFound      bool
	requestErr        error
	confirmErr        error
	changeErr         error
	confirmedToken    string
	confirmedPassword string
	changedUserID     string
	currentPassword   string
	newPassword       string
}

func (f *fakePasswordService) Request(context.Context, string) (store.User, string, bool, error) {
	return f.requestUser, f.requestToken, f.requestFound, f.requestErr
}

func (f *fakePasswordService) Confirm(_ context.Context, token, password string) error {
	f.confirmedToken, f.confirmedPassword = token, password
	return f.confirmErr
}

func (f *fakePasswordService) Change(_ context.Context, userID, currentPassword, newPassword string) error {
	f.changedUserID, f.currentPassword, f.newPassword = userID, currentPassword, newPassword
	return f.changeErr
}

type fakeEmailSender struct {
	messages []notifications.EmailMessage
	err      error
}

func (f *fakeEmailSender) Send(_ context.Context, message notifications.EmailMessage) error {
	f.messages = append(f.messages, message)
	return f.err
}

func TestPasswordResetRequestDoesNotRevealUnknownEmail(t *testing.T) {
	knownService := &fakePasswordService{requestUser: store.User{ID: "user-1", Email: "person@example.com"}, requestToken: "raw-token", requestFound: true}
	unknownService := &fakePasswordService{}
	known := &Handler{Passwords: knownService, EmailSender: &fakeEmailSender{}, AppOrigin: "http://localhost:3000"}
	unknown := &Handler{Passwords: unknownService, EmailSender: &fakeEmailSender{}, AppOrigin: "http://localhost:3000"}

	request := func(handler *Handler) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/auth/password-reset/request", strings.NewReader(`{"email":"person@example.com"}`))
		request.Header.Set("Origin", "http://localhost:3000")
		request.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(response, request)
		return response
	}
	knownResponse, unknownResponse := request(known), request(unknown)
	if knownResponse.Code != http.StatusAccepted || unknownResponse.Code != http.StatusAccepted || knownResponse.Body.String() != unknownResponse.Body.String() {
		t.Fatalf("expected identical generic responses, known=%d %s unknown=%d %s", knownResponse.Code, knownResponse.Body.String(), unknownResponse.Code, unknownResponse.Body.String())
	}
}

func TestPasswordResetConfirmReturnsInvalidToken(t *testing.T) {
	service := &fakePasswordService{confirmErr: store.ErrInvalidPasswordResetToken}
	handler := &Handler{Passwords: service}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/password-reset/confirm", strings.NewReader(`{"token":"bad","password":"NewPassword!2026"}`)))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_reset_token"`) {
		t.Fatalf("expected invalid reset token response, got %d %s", response.Code, response.Body.String())
	}
}

func TestPasswordChangeRequiresAuthentication(t *testing.T) {
	handler := &Handler{Passwords: &fakePasswordService{}}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/api/auth/password", strings.NewReader(`{"currentPassword":"old","newPassword":"new"}`)))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected authentication requirement, got %d %s", response.Code, response.Body.String())
	}
}

func TestPasswordChangeRevokesTheCurrentSession(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{user: store.User{ID: "user-1", UserType: "stakeholder", Role: "assessor", Status: "active"}, session: store.Session{ExpiresAt: now.Add(time.Hour)}}
	service := &fakePasswordService{}
	handler := newAuthHandler(repo, now)
	handler.Passwords = service
	request := httptest.NewRequest(http.MethodPut, "/api/auth/password", strings.NewReader(`{"currentPassword":"old","newPassword":"new"}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "raw-token"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.changedUserID != "user-1" || service.currentPassword != "old" || service.newPassword != "new" {
		t.Fatalf("unexpected password change: status=%d service=%#v body=%s", response.Code, service, response.Body.String())
	}
}

func TestPasswordHandlersMapServiceFailures(t *testing.T) {
	service := &fakePasswordService{confirmErr: errors.New("database unavailable")}
	handler := &Handler{Passwords: service}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/password-reset/confirm", strings.NewReader(`{"token":"raw","password":"NewPassword!2026"}`)))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error, got %d %s", response.Code, response.Body.String())
	}
}
