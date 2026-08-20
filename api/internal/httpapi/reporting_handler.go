package httpapi

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"compliance/api/internal/store"
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
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &report.Project.OrganizationID, ProjectID: &projectID, Action: "report.viewed", EntityType: "final_report", EntityID: &projectID})
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
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &packageData.Project.OrganizationID, ProjectID: &projectID, Action: "audit_package.viewed", EntityType: "audit_package", EntityID: &projectID})
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
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &packageData.Project.OrganizationID, ProjectID: &projectID, Action: "audit_package.downloaded", EntityType: "audit_package", EntityID: &projectID})

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="audit-package.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"record_type", "function", "category", "subcategory", "description", "included", "rationale",
		"assignee", "assignee_email", "response_status", "response_submitted_at", "reviewer",
		"reviewed_at", "review_comment", "evidence_count", "evidence_names", "evidence_types",
		"action_id", "action_title", "action_owner", "action_priority", "action_due_date", "action_status", "action_submitted_at", "action_closed_at", "action_review_comment",
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
			"assessment_outcome", profile.FunctionCode, profile.CategoryCode, profile.SubcategoryCode, profile.Description,
			strconv.FormatBool(profile.Included), profile.Rationale, profile.AssignedUserName,
			profile.AssignedUserEmail, responseStatus, submittedAt, reviewer, reviewedAt, reviewComment,
			strconv.Itoa(len(outcome.Evidence)), strings.Join(names, "; "), strings.Join(types, "; "),
			"", "", "", "", "", "", "", "", "",
		})
	}
	for _, action := range packageData.RemediationActions {
		names := make([]string, 0, len(action.Evidence))
		types := make([]string, 0, len(action.Evidence))
		for _, evidence := range action.Evidence {
			names = append(names, evidence.OriginalName)
			types = append(types, evidence.MIMEType)
		}
		_ = writer.Write([]string{
			"remediation_action", "", "", action.OutcomeCode, action.OutcomeDescription, "", "",
			action.OwnerName, action.OwnerEmail, "", "", "", "", "", strconv.Itoa(len(action.Evidence)), strings.Join(names, "; "), strings.Join(types, "; "),
			action.ID, action.Title, action.OwnerName, action.Priority, action.DueDate.UTC().Format("2006-01-02"), action.Status,
			formatReportTime(action.SubmittedAt), formatReportTime(action.ClosedAt), action.ReviewComment,
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
