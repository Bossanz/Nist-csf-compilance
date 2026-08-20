package httpapi

import (
	"context"
	"errors"
	"net/http"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type auditLogStore interface {
	ListProjectAuditEvents(context.Context, string) ([]store.AuditTrailEntry, error)
	ListOrganizationAuditEvents(context.Context, string, string) ([]store.AuditTrailEntry, error)
}

func (h *Handler) projectAuditLogs(w http.ResponseWriter, r *http.Request, projectID string) {
	data, ok := h.Store.(auditLogStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "audit log store unavailable")
		return
	}
	events, err := data.ListProjectAuditEvents(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load audit logs")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{ProjectID: &projectID, Action: "audit_logs.viewed", EntityType: "audit_log"})
	writeJSON(w, http.StatusOK, events)
}

func (h *Handler) organizationAuditLogs(w http.ResponseWriter, r *http.Request, organizationID string) {
	data, ok := h.Store.(auditLogStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "audit log store unavailable")
		return
	}
	if _, ok := h.authorizeOrganization(w, r, organizationID); !ok {
		return
	}
	auditorUserID := ""
	if currentUser(r).Role == "auditor" {
		auditorUserID = currentUser(r).ID
	}
	events, err := data.ListOrganizationAuditEvents(r.Context(), organizationID, auditorUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load audit logs")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &organizationID, Action: "audit_logs.viewed", EntityType: "audit_log"})
	writeJSON(w, http.StatusOK, events)
}
