package httpapi

import (
	"context"
	"net/http"

	"compliance/api/internal/store"
)

type accountDataStore interface {
	ListOrganizationUsers(context.Context, string) ([]store.User, error)
	ListCounselors(context.Context) ([]store.User, error)
	UpdateOrganizationUser(context.Context, string, string, string, string) (store.User, error)
	UpdateCounselor(context.Context, string, string, string) (store.User, error)
	RevokeUserSessions(context.Context, string) error
}

func (h *Handler) accountStore() (accountDataStore, bool) {
	data, ok := h.Store.(accountDataStore)
	return data, ok
}

func validStatus(status string) bool { return status == "active" || status == "disabled" }
func isStakeholderRole(role string) bool {
	return role == "org_admin" || role == "assessor" || role == "reviewer" || role == "viewer"
}
func isCounselorRole(role string) bool { return role == "counselor_admin" || role == "counselor" }

func (h *Handler) organizationUsers(w http.ResponseWriter, r *http.Request, organizationID string) {
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	data, ok := h.accountStore()
	if !ok {
		writeError(w, 500, "internal_error", "account store unavailable")
		return
	}
	users, err := data.ListOrganizationUsers(r.Context(), organizationID)
	if err != nil {
		writeError(w, 500, "internal_error", "Could not load users")
		return
	}
	writeJSON(w, 200, users)
}

func (h *Handler) updateOrganizationUser(w http.ResponseWriter, r *http.Request, organizationID, userID string) {
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	if !can(currentUser(r), actionInviteStakeholder) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	var input struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid user update")
		return
	}
	if !isStakeholderRole(input.Role) || !validStatus(input.Status) {
		writeError(w, 403, "forbidden", "Invalid stakeholder role or status")
		return
	}
	if currentUser(r).ID == userID && input.Status == "disabled" {
		writeError(w, 400, "validation_error", "You cannot disable your own account")
		return
	}
	data, ok := h.accountStore()
	if !ok {
		writeError(w, 500, "internal_error", "account store unavailable")
		return
	}
	user, err := data.UpdateOrganizationUser(r.Context(), organizationID, userID, input.Role, input.Status)
	if err != nil {
		writeError(w, 404, "not_found", "user not found")
		return
	}
	if input.Status == "disabled" {
		_ = data.RevokeUserSessions(r.Context(), userID)
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &organizationID, Action: "user.role_changed", EntityType: "user", EntityID: &userID, Metadata: map[string]any{"role": input.Role, "status": input.Status}})
	writeJSON(w, 200, user)
}

func (h *Handler) counselors(w http.ResponseWriter, r *http.Request) {
	if !can(currentUser(r), actionManageCounselor) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	data, ok := h.accountStore()
	if !ok {
		writeError(w, 500, "internal_error", "account store unavailable")
		return
	}
	users, err := data.ListCounselors(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", "Could not load counselors")
		return
	}
	writeJSON(w, 200, users)
}

func (h *Handler) updateCounselor(w http.ResponseWriter, r *http.Request, userID string) {
	if !can(currentUser(r), actionManageCounselor) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	var input struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid counselor update")
		return
	}
	if !isCounselorRole(input.Role) || !validStatus(input.Status) {
		writeError(w, 403, "forbidden", "Invalid counselor role or status")
		return
	}
	if currentUser(r).ID == userID && input.Status == "disabled" {
		writeError(w, 400, "validation_error", "You cannot disable your own account")
		return
	}
	data, ok := h.accountStore()
	if !ok {
		writeError(w, 500, "internal_error", "account store unavailable")
		return
	}
	user, err := data.UpdateCounselor(r.Context(), userID, input.Role, input.Status)
	if err != nil {
		writeError(w, 404, "not_found", "user not found")
		return
	}
	if input.Status == "disabled" {
		_ = data.RevokeUserSessions(r.Context(), userID)
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{Action: "user.role_changed", EntityType: "user", EntityID: &userID, Metadata: map[string]any{"role": input.Role, "status": input.Status}})
	writeJSON(w, 200, user)
}
