package httpapi

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"compliance/api/internal/store"
)

func (h *Handler) uploadRemediationEvidence(w http.ResponseWriter, r *http.Request, projectID, actionID string) {
	if h.RemediationEvidence == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "remediation evidence storage is unavailable")
		return
	}
	if err := r.ParseMultipartForm(maxEvidenceBytes + 64*1024); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "could not read uploaded file")
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "file is required")
		return
	}
	defer file.Close()
	mimeType, err := validateEvidenceFile(file, header)
	if err != nil {
		writeEvidenceError(w, err)
		return
	}
	storagePath, size, err := h.Evidence.Save(file)
	if err != nil {
		writeEvidenceError(w, err)
		return
	}
	evidence, err := h.RemediationEvidence.CreateRemediationEvidence(r.Context(), projectID, actionID, currentUser(r).ID, header.Filename, storagePath, mimeType, size)
	if err != nil {
		if removeErr := h.Evidence.Remove(storagePath); removeErr != nil {
			logStorageCleanupFailure(storagePath, removeErr)
		}
		writeRemediationError(w, err)
		return
	}
	h.writeRemediationEvidenceAudit(r, projectID, actionID, evidence, "remediation.evidence_uploaded")
	writeJSON(w, http.StatusCreated, evidence)
}

func (h *Handler) downloadRemediationEvidence(w http.ResponseWriter, r *http.Request, projectID, actionID, evidenceID string, inline bool) {
	if h.RemediationEvidence == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "remediation evidence storage is unavailable")
		return
	}
	evidence, err := h.RemediationEvidence.GetRemediationEvidence(r.Context(), projectID, actionID, evidenceID)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	file, err := h.Evidence.Open(evidence.StoragePath)
	if errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusNotFound, "not_found", "evidence file not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not open evidence")
		return
	}
	defer file.Close()
	disposition := "attachment"
	if inline {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", evidence.MIMEType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": safeDownloadName(evidence.OriginalName)}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.FormatInt(evidence.SizeBytes, 10))
	readAction := "remediation.evidence_downloaded"
	if inline {
		readAction = "remediation.evidence_viewed"
	}
	h.writeRemediationEvidenceAudit(r, projectID, actionID, evidence, readAction)
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("remediation evidence stream failed for %s: %v", evidence.StoragePath, err)
	}
}

func (h *Handler) deleteRemediationEvidence(w http.ResponseWriter, r *http.Request, projectID, actionID, evidenceID string) {
	if h.RemediationEvidence == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "remediation evidence storage is unavailable")
		return
	}
	evidence, err := h.RemediationEvidence.DeleteRemediationEvidence(r.Context(), projectID, actionID, evidenceID, currentUser(r).ID)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	if err := h.Evidence.Remove(evidence.StoragePath); err != nil {
		logStorageCleanupFailure(evidence.StoragePath, err)
	}
	h.writeRemediationEvidenceAudit(r, projectID, actionID, evidence, "remediation.evidence_deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeRemediationEvidenceAudit(r *http.Request, projectID, actionID string, evidence store.RemediationEvidence, event string) {
	project, err := h.Store.GetProject(r.Context(), projectID)
	if err != nil {
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
		OrganizationID: &project.OrganizationID, ProjectID: &projectID, Action: event,
		EntityType: "remediation_evidence", EntityID: &evidence.ID,
		Metadata: map[string]any{"actionID": actionID, "originalName": evidence.OriginalName, "mimeType": evidence.MIMEType, "sizeBytes": evidence.SizeBytes},
	})
}
