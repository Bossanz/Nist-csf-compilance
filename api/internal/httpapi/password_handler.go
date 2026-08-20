package httpapi

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	authservice "compliance/api/internal/auth"
	"compliance/api/internal/notifications"
	"compliance/api/internal/store"
)

const passwordResetMessage = "If an active account exists, a password reset link will be sent."

func (h *Handler) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if h.Passwords == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Password recovery is unavailable")
		return
	}
	var input struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid password reset request")
		return
	}
	user, rawToken, found, err := h.Passwords.Request(r.Context(), input.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not process password reset request")
		return
	}
	if found && rawToken != "" && h.EmailSender != nil {
		resetURL := strings.TrimRight(h.AppOrigin, "/") + "/reset-password?token=" + url.QueryEscape(rawToken)
		message := notifications.EmailMessage{
			To:      user.Email,
			Subject: "Reset your CSF Compliance password",
			Text:    "Use this link to reset your CSF Compliance password:\n\n" + resetURL + "\n\nThis link expires in 30 minutes and can only be used once.",
		}
		if err := h.EmailSender.Send(r.Context(), message); err != nil {
			log.Printf("password reset email failed to=%s: %v", user.Email, err)
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"message": passwordResetMessage})
}

func (h *Handler) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if h.Passwords == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Password recovery is unavailable")
		return
	}
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid password reset request")
		return
	}
	if strings.TrimSpace(input.Token) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Reset token is required")
		return
	}
	if err := h.Passwords.Confirm(r.Context(), input.Token, input.Password); err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidPasswordResetToken):
			writeError(w, http.StatusBadRequest, "invalid_reset_token", "Reset token is invalid or expired")
		case errors.Is(err, authservice.ErrWeakPassword):
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Could not reset password")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	if h.Passwords == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Password change is unavailable")
		return
	}
	var input struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid password change request")
		return
	}
	if err := h.Passwords.Change(r.Context(), currentUser(r).ID, input.CurrentPassword, input.NewPassword); err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidCurrentPassword):
			writeError(w, http.StatusBadRequest, "invalid_current_password", "Current password is incorrect")
		case errors.Is(err, authservice.ErrWeakPassword):
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Could not change password")
		}
		return
	}
	user := currentUser(r)
	h.writeAudit(user, r.Context(), store.AuditEvent{OrganizationID: user.OrganizationID, Action: "auth.password_changed", EntityType: "user", EntityID: &user.ID})
	w.WriteHeader(http.StatusNoContent)
}
