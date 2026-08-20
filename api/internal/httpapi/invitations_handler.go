package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authservice "compliance/api/internal/auth"
	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type invitationResponse struct {
	store.Invitation
	InvitationURL string `json:"invitationURL"`
}

type invitationListStore interface {
	ListOrganizationInvitations(context.Context, string) ([]store.Invitation, error)
}

func (h *Handler) inviteStakeholder(w http.ResponseWriter, r *http.Request, organizationID string) {
	if h.Invitations == nil {
		writeError(w, 500, "internal_error", "invitation service unavailable")
		return
	}
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	if !can(currentUser(r), actionInviteStakeholder) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	var input struct {
		Email      string   `json:"email"`
		Role       string   `json:"role"`
		ProjectIDs []string `json:"projectIDs"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid invitation request")
		return
	}
	invitation, raw, err := h.Invitations.Invite(r.Context(), currentUser(r), &organizationID, input.Email, input.Role, input.ProjectIDs...)
	h.writeInvitationResult(w, r, invitation, raw, err)
}

func (h *Handler) listInvitations(w http.ResponseWriter, r *http.Request, organizationID string) {
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	if !can(currentUser(r), actionInviteStakeholder) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	data, ok := h.Store.(invitationListStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "invitation store unavailable")
		return
	}
	invitations, err := data.ListOrganizationInvitations(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not load invitations")
		return
	}
	writeJSON(w, http.StatusOK, invitations)
}

func (h *Handler) resendInvitation(w http.ResponseWriter, r *http.Request, organizationID, invitationID string) {
	if h.Invitations == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "invitation service unavailable")
		return
	}
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	if !can(currentUser(r), actionInviteStakeholder) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	invitation, raw, err := h.Invitations.Resend(r.Context(), currentUser(r), organizationID, invitationID)
	if errors.Is(err, authservice.ErrInvitationNotPending) {
		writeError(w, http.StatusConflict, "invitation_not_pending", "Invitation is no longer pending")
		return
	}
	if errors.Is(err, authservice.ErrForbidden) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "Invitation not found")
		return
	}
	h.writeInvitationResultWithStatus(w, r, invitation, raw, err, http.StatusOK, "user.invitation_resent")
}

func (h *Handler) cancelInvitation(w http.ResponseWriter, r *http.Request, organizationID, invitationID string) {
	if h.Invitations == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "invitation service unavailable")
		return
	}
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	if !can(currentUser(r), actionInviteStakeholder) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	invitation, err := h.Invitations.Cancel(r.Context(), currentUser(r), organizationID, invitationID)
	if errors.Is(err, authservice.ErrInvitationNotPending) {
		writeError(w, http.StatusConflict, "invitation_not_pending", "Invitation is no longer pending")
		return
	}
	if errors.Is(err, authservice.ErrForbidden) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "Invitation not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not cancel invitation")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: invitation.OrganizationID, Action: "user.invitation_cancelled", EntityType: "invitation", EntityID: &invitation.ID})
	writeJSON(w, http.StatusOK, invitation)
}

func (h *Handler) inviteCounselor(w http.ResponseWriter, r *http.Request) {
	if h.Invitations == nil {
		writeError(w, 500, "internal_error", "invitation service unavailable")
		return
	}
	if !can(currentUser(r), actionManageCounselor) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	var input struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid invitation request")
		return
	}
	invitation, raw, err := h.Invitations.Invite(r.Context(), currentUser(r), nil, input.Email, input.Role)
	h.writeInvitationResult(w, r, invitation, raw, err)
}

func (h *Handler) writeInvitationResult(w http.ResponseWriter, r *http.Request, invitation store.Invitation, raw string, err error) {
	h.writeInvitationResultWithStatus(w, r, invitation, raw, err, http.StatusCreated, "user.invited")
}

func (h *Handler) writeInvitationResultWithStatus(w http.ResponseWriter, r *http.Request, invitation store.Invitation, raw string, err error, status int, action string) {
	if errors.Is(err, authservice.ErrForbidden) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	if errors.Is(err, authservice.ErrDuplicateInvitation) {
		writeError(w, 409, "duplicate_invitation", "An active user or invitation already exists")
		return
	}
	if errors.Is(err, authservice.ErrInvalidProjectAccess) {
		writeError(w, http.StatusBadRequest, "invalid_project_access", "Auditor invitation must include valid Project access")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "Could not create invitation")
		return
	}
	base := strings.TrimRight(h.AppOrigin, "/")
	response := invitationResponse{Invitation: invitation, InvitationURL: base + "/invite/" + raw}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: invitation.OrganizationID, Action: action, EntityType: "invitation", EntityID: &invitation.ID, Metadata: map[string]any{"email": invitation.Email, "role": invitation.Role}})
	h.notifyInvitation(r.Context(), invitation, raw)
	writeJSON(w, status, response)
}

func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request, rawToken string) {
	if h.Invitations == nil {
		writeError(w, 500, "internal_error", "invitation service unavailable")
		return
	}
	var input struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid invitation request")
		return
	}
	user, err := h.Invitations.Accept(r.Context(), rawToken, input.Name, input.Password)
	if errors.Is(err, authservice.ErrInvalidInvitation) {
		writeError(w, 400, "invalid_invitation", "Invitation is invalid or expired")
		return
	}
	if errors.Is(err, authservice.ErrWeakPassword) {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "Could not accept invitation")
		return
	}
	writeJSON(w, 200, user)
}
