package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"compliance/api/internal/domain"
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
	writeJSON(w, http.StatusOK, response)
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
