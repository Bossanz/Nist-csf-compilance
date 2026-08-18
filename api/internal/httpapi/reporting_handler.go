package httpapi

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) finalReport(w http.ResponseWriter, r *http.Request, projectID string) {
	data, ok := h.Store.(reportingStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "reporting store unavailable")
		return
	}
	report, err := data.GetFinalReport(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load final report")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) auditPackage(w http.ResponseWriter, r *http.Request, projectID string) {
	data, ok := h.Store.(reportingStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "reporting store unavailable")
		return
	}
	packageData, err := data.GetAuditPackage(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load audit package")
		return
	}
	writeJSON(w, http.StatusOK, packageData)
}

func (h *Handler) auditPackageCSV(w http.ResponseWriter, r *http.Request, projectID string) {
	data, ok := h.Store.(reportingStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "reporting store unavailable")
		return
	}
	packageData, err := data.GetAuditPackage(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load audit package")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="audit-package.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"function", "category", "subcategory", "description", "included", "rationale",
		"assignee", "assignee_email", "response_status", "response_submitted_at", "reviewer",
		"reviewed_at", "review_comment", "evidence_count", "evidence_names", "evidence_types",
	})
	for _, outcome := range packageData.Outcomes {
		profile := outcome.Profile
		responseStatus, submittedAt, reviewer, reviewedAt, reviewComment := "", "", "", "", ""
		if outcome.Response != nil {
			responseStatus = outcome.Response.Status
			submittedAt = formatReportTime(outcome.Response.SubmittedAt)
			reviewer = stringValue(outcome.Response.ReviewedBy)
			reviewedAt = formatReportTime(outcome.Response.ReviewedAt)
			reviewComment = outcome.Response.ReviewComment
		}
		names := make([]string, 0, len(outcome.Evidence))
		types := make([]string, 0, len(outcome.Evidence))
		for _, evidence := range outcome.Evidence {
			names = append(names, evidence.OriginalName)
			types = append(types, evidence.MIMEType)
		}
		_ = writer.Write([]string{
			profile.FunctionCode, profile.CategoryCode, profile.SubcategoryCode, profile.Description,
			strconv.FormatBool(profile.Included), profile.Rationale, profile.AssignedUserName,
			profile.AssignedUserEmail, responseStatus, submittedAt, reviewer, reviewedAt, reviewComment,
			strconv.Itoa(len(outcome.Evidence)), strings.Join(names, "; "), strings.Join(types, "; "),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return
	}
}

func formatReportTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
