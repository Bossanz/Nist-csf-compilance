package httpapi

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	authservice "compliance/api/internal/auth"
	"compliance/api/internal/store"
)

const sessionCookieName = "compliance_session"

type throttleEntry struct {
	failures int
	first    time.Time
}

type LoginThrottle struct {
	mu      sync.Mutex
	entries map[string]throttleEntry
}

func NewLoginThrottle() *LoginThrottle { return &LoginThrottle{entries: map[string]throttleEntry{}} }

func (t *LoginThrottle) Allow(key string, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry := t.entries[key]
	if entry.first.IsZero() || now.Sub(entry.first) >= 5*time.Minute {
		delete(t.entries, key)
		return true
	}
	return entry.failures < 5
}

func (t *LoginThrottle) Fail(key string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry := t.entries[key]
	if entry.first.IsZero() || now.Sub(entry.first) >= 5*time.Minute {
		entry = throttleEntry{first: now}
	}
	entry.failures++
	t.entries[key] = entry
}

func (t *LoginThrottle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, key)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid login request")
		return
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	key := authservice.NormalizeEmail(input.Email) + "|" + host
	now := time.Now()
	if h.LoginThrottle != nil && !h.LoginThrottle.Allow(key, now) {
		h.writeAudit(store.User{}, r.Context(), store.AuditEvent{Action: "auth.login_throttled", EntityType: "session", Result: "failure", Metadata: map[string]any{"email": authservice.NormalizeEmail(input.Email)}})
		writeError(w, http.StatusTooManyRequests, "too_many_attempts", "Too many login attempts")
		return
	}
	user, rawToken, err := h.Auth.Login(r.Context(), input.Email, input.Password)
	if errors.Is(err, authservice.ErrInvalidCredentials) {
		if h.LoginThrottle != nil {
			h.LoginThrottle.Fail(key, now)
		}
		h.writeAudit(store.User{}, r.Context(), store.AuditEvent{Action: "auth.login_failed", EntityType: "session", Result: "failure", Metadata: map[string]any{"email": authservice.NormalizeEmail(input.Email)}})
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}
	if err != nil {
		h.writeAudit(store.User{}, r.Context(), store.AuditEvent{Action: "auth.login_failed", EntityType: "session", Result: "failure"})
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not log in")
		return
	}
	if h.LoginThrottle != nil {
		h.LoginThrottle.Reset(key)
	}
	h.setSessionCookie(w, rawToken, 12*time.Hour)
	h.writeAudit(user, r.Context(), store.AuditEvent{OrganizationID: user.OrganizationID, Action: "auth.login_succeeded", EntityType: "session", Result: "success", Metadata: map[string]any{"email": user.Email}})
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required")
		return
	}
	user, err := h.Auth.Authenticate(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var user store.User
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		user, _ = h.Auth.Authenticate(r.Context(), cookie.Value)
		_ = h.Auth.Logout(r.Context(), cookie.Value)
	}
	if user.ID != "" {
		h.writeAudit(user, r.Context(), store.AuditEvent{OrganizationID: user.OrganizationID, Action: "auth.logout", EntityType: "session", Result: "success"})
	}
	h.setSessionCookie(w, "", -time.Hour)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, value string, lifetime time.Duration) {
	maxAge := int(lifetime.Seconds())
	if lifetime < 0 {
		maxAge = -1
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: value, Path: "/", HttpOnly: true, Secure: h.SecureCookies,
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	})
}

func isAuthPath(path string) bool { return strings.HasPrefix(path, "/api/auth/") }
