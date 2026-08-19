package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) listRemediationActions(w http.ResponseWriter, r *http.Request, projectID string) {
	if h.Remediations == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "remediation store unavailable")
		return
	}
	actions, err := h.Remediations.ListRemediationActions(r.Context(), projectID)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, actions)
}

func (h *Handler) createRemediationAction(w http.ResponseWriter, r *http.Request, projectID string) {
	if h.Remediations == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "remediation store unavailable")
		return
	}
	var input struct {
		SubcategoryID string `json:"subcategoryID"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		DesiredResult string `json:"desiredResult"`
		Priority      string `json:"priority"`
		OwnerUserID   string `json:"ownerUserID"`
		DueDate       string `json:"dueDate"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "Due date must use YYYY-MM-DD")
		return
	}
	action, err := h.Remediations.CreateRemediationAction(r.Context(), projectID, currentUser(r).ID, store.RemediationCreate{
		SubcategoryID: input.SubcategoryID,
		Title:         input.Title, Description: input.Description, DesiredResult: input.DesiredResult,
		Priority: input.Priority, OwnerUserID: input.OwnerUserID, DueDate: dueDate,
	})
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	h.writeRemediationAudit(r, action, "remediation.created", map[string]any{"status": action.Status, "ownerUserID": action.OwnerUserID})
	writeJSON(w, http.StatusCreated, action)
}

func (h *Handler) updateRemediationAction(w http.ResponseWriter, r *http.Request, projectID, actionID string) {
	var input struct {
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		DesiredResult *string `json:"desiredResult"`
		Priority      *string `json:"priority"`
		OwnerUserID   *string `json:"ownerUserID"`
		DueDate       *string `json:"dueDate"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	patch := store.RemediationPatch{Title: input.Title, Description: input.Description, DesiredResult: input.DesiredResult, Priority: input.Priority, OwnerUserID: input.OwnerUserID}
	if input.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *input.DueDate)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "Due date must use YYYY-MM-DD")
			return
		}
		patch.DueDate = &parsed
	}
	action, err := h.Remediations.UpdateRemediationAction(r.Context(), projectID, actionID, currentUser(r).ID, patch)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	h.writeRemediationAudit(r, action, "remediation.updated", map[string]any{"actionID": action.ID})
	writeJSON(w, http.StatusOK, action)
}

func (h *Handler) updateRemediationProgress(w http.ResponseWriter, r *http.Request, projectID, actionID string) {
	var input struct {
		ProgressNote string `json:"progressNote"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	action, err := h.Remediations.UpdateRemediationProgress(r.Context(), projectID, actionID, currentUser(r).ID, input.ProgressNote)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	h.writeRemediationAudit(r, action, "remediation.progress_updated", map[string]any{"status": action.Status})
	writeJSON(w, http.StatusOK, action)
}

func (h *Handler) submitRemediationAction(w http.ResponseWriter, r *http.Request, projectID, actionID string) {
	action, err := h.Remediations.SubmitRemediationAction(r.Context(), projectID, actionID, currentUser(r).ID)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	h.writeRemediationAudit(r, action, "remediation.submitted", map[string]any{"status": action.Status})
	writeJSON(w, http.StatusOK, action)
}

func (h *Handler) reviewRemediationAction(w http.ResponseWriter, r *http.Request, projectID, actionID string) {
	var input struct {
		Decision string `json:"decision"`
		Comment  string `json:"comment"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	action, err := h.Remediations.ReviewRemediationAction(r.Context(), projectID, actionID, currentUser(r).ID, input.Decision, input.Comment)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	auditAction := "remediation.returned"
	if input.Decision == "close" {
		auditAction = "remediation.closed"
	}
	h.writeRemediationAudit(r, action, auditAction, map[string]any{"decision": input.Decision, "comment": strings.TrimSpace(input.Comment)})
	writeJSON(w, http.StatusOK, action)
}

func (h *Handler) writeRemediationAudit(r *http.Request, action store.RemediationAction, event string, metadata map[string]any) {
	project, err := h.Store.GetProject(r.Context(), action.ProjectID)
	if err != nil {
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
		OrganizationID: &project.OrganizationID, ProjectID: &action.ProjectID,
		Action: event, EntityType: "remediation_action", EntityID: &action.ID, Metadata: metadata,
	})
}

func writeRemediationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrOutcomeNotApproved):
		writeError(w, http.StatusConflict, "outcome_not_approved", "Outcome must be approved before creating an action")
	case errors.Is(err, store.ErrNoCoverageGap):
		writeError(w, http.StatusConflict, "no_coverage_gap", "Current coverage must be below target coverage")
	case errors.Is(err, store.ErrInvalidRemediationTransition):
		writeError(w, http.StatusConflict, "invalid_remediation_transition", "Action cannot change state from its current status")
	case errors.Is(err, store.ErrRemediationClosed):
		writeError(w, http.StatusConflict, "remediation_closed", "Closed actions are read-only")
	case errors.Is(err, store.ErrInvalidRemediationOwner):
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "Owner must be an active organization Admin or Assessor")
	case errors.Is(err, store.ErrInvalidRemediationInput):
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "Remediation action data is invalid")
	case errors.Is(err, store.ErrRemediationForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "Action is not assigned to this user")
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "not_found", "Remediation action not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not process remediation action")
	}
}
