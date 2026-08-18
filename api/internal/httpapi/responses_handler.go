package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"compliance/api/internal/domain"
	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) responses(w http.ResponseWriter, r *http.Request, projectID string) {
	if h.Responses == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "response store is unavailable")
		return
	}
	data, err := h.Responses.ListResponses(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load responses")
		return
	}
	if h.Auth != nil && currentUser(r).UserType == "stakeholder" {
		profiles, err := h.Store.ListProfile(r.Context(), projectID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not check outcome access")
			return
		}
		visible := make(map[string]struct{}, len(profiles))
		for _, row := range profiles {
			if stakeholderCanReadProfile(currentUser(r), row) {
				visible[row.SubcategoryID] = struct{}{}
			}
		}
		scoped := make([]store.StakeholderResponse, 0, len(data))
		for _, response := range data {
			if _, ok := visible[response.SubcategoryID]; ok {
				scoped = append(scoped, response)
			}
		}
		data = scoped
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) saveResponse(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) {
	var input struct {
		ResponseText string `json:"responseText"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if h.Responses == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "response store is unavailable")
		return
	}
	response, err := h.Responses.SaveResponseDraft(r.Context(), projectID, subcategoryID, currentUser(r).ID, input.ResponseText)
	if err != nil {
		writeResponseError(w, err, "could not save response")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) submitResponse(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) {
	if h.Responses == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "response store is unavailable")
		return
	}
	response, err := h.Responses.SubmitResponse(r.Context(), projectID, subcategoryID, currentUser(r).ID)
	if err != nil {
		writeResponseError(w, err, "could not submit response")
		return
	}
	h.writeResponseAudit(r, projectID, subcategoryID, response.ID, "response.submitted")
	h.notifyResponseSubmitted(r.Context(), projectID, subcategoryID)
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) reviewResponse(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) {
	var input struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(input.Status) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "review status is required")
		return
	}
	if h.Responses == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "response store is unavailable")
		return
	}
	response, err := h.Responses.ReviewResponse(r.Context(), projectID, subcategoryID, currentUser(r).ID, input.Status, input.Comment)
	if err != nil {
		writeResponseError(w, err, "could not review response")
		return
	}
	action := "response.needs_more_info"
	if input.Status == string(domain.ResponseReviewed) {
		action = "response.reviewed"
	}
	h.writeResponseAudit(r, projectID, subcategoryID, response.ID, action)
	h.notifyResponseReviewed(r.Context(), projectID, subcategoryID, input.Status, input.Comment)
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) writeResponseAudit(r *http.Request, projectID, subcategoryID, responseID, action string) {
	project, err := h.Store.GetProject(r.Context(), projectID)
	if err != nil {
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
		OrganizationID: &project.OrganizationID,
		ProjectID:      &projectID,
		Action:         action,
		EntityType:     "response",
		EntityID:       &responseID,
		Metadata:       map[string]any{"subcategoryID": subcategoryID},
	})
}

func writeResponseError(w http.ResponseWriter, err error, message string) {
	switch {
	case errors.Is(err, domain.ErrInvalidResponseTransition):
		writeError(w, http.StatusConflict, "invalid_transition", "response cannot move to the requested status")
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "not_found", "response not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", message)
	}
}
