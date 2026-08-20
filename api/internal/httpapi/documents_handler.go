package httpapi

import (
	"compliance/api/internal/store"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) uploadResponseDocument(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) {
	if h.Documents == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "document storage is unavailable")
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
	storageKey, size, err := h.Evidence.Save(file)
	if err != nil {
		writeEvidenceError(w, err)
		return
	}
	document, err := h.Documents.CreateResponseDocument(r.Context(), projectID, subcategoryID, currentUser(r).ID, header.Filename, storageKey, mimeType, size)
	if err != nil {
		if removeErr := h.Evidence.Remove(storageKey); removeErr != nil {
			logStorageCleanupFailure(storageKey, removeErr)
		}
		writeDocumentError(w, err, "could not save document")
		return
	}
	if project, projectErr := h.Store.GetProject(r.Context(), projectID); projectErr == nil {
		h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
			OrganizationID: &project.OrganizationID,
			ProjectID:      &projectID,
			Action:         "evidence.uploaded",
			EntityType:     "evidence",
			EntityID:       &document.ID,
			Metadata:       map[string]any{"subcategoryID": subcategoryID, "originalName": header.Filename, "mimeType": mimeType, "sizeBytes": size},
		})
	}
	writeJSON(w, http.StatusCreated, document)
}

func (h *Handler) downloadResponseDocument(w http.ResponseWriter, r *http.Request, projectID, subcategoryID, documentID string) {
	if h.Documents == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "document storage is unavailable")
		return
	}
	document, err := h.Documents.GetResponseDocument(r.Context(), projectID, subcategoryID, documentID)
	if err != nil {
		writeDocumentError(w, err, "could not load document")
		return
	}
	file, err := h.Evidence.Open(document.StorageKey)
	if errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusNotFound, "not_found", "document file not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not open document")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", document.MIMEType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": safeDownloadName(document.OriginalName)}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if document.SizeBytes >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(document.SizeBytes, 10))
	}
	if project, projectErr := h.Store.GetProject(r.Context(), projectID); projectErr == nil {
		h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
			OrganizationID: &project.OrganizationID,
			ProjectID:      &projectID,
			Action:         "evidence.downloaded",
			EntityType:     "evidence",
			EntityID:       &document.ID,
			Metadata:       map[string]any{"subcategoryID": subcategoryID, "originalName": document.OriginalName, "mimeType": document.MIMEType, "sizeBytes": document.SizeBytes},
		})
	}
	if _, err := io.Copy(w, file); err != nil {
		logStorageCleanupFailure(document.StorageKey, err)
	}
}

func (h *Handler) deleteResponseDocument(w http.ResponseWriter, r *http.Request, projectID, subcategoryID, documentID string) {
	if h.Documents == nil || h.Evidence == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "document storage is unavailable")
		return
	}
	document, err := h.Documents.DeleteResponseDocument(r.Context(), projectID, subcategoryID, documentID)
	if err != nil {
		writeDocumentError(w, err, "could not delete document")
		return
	}
	if err := h.Evidence.Remove(document.StorageKey); err != nil {
		logStorageCleanupFailure(document.StorageKey, err)
	}
	if project, projectErr := h.Store.GetProject(r.Context(), projectID); projectErr == nil {
		h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
			OrganizationID: &project.OrganizationID,
			ProjectID:      &projectID,
			Action:         "evidence.deleted",
			EntityType:     "evidence",
			EntityID:       &document.ID,
			Metadata:       map[string]any{"subcategoryID": subcategoryID, "originalName": document.OriginalName},
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateEvidenceFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > maxEvidenceBytes {
		return "", ErrEvidenceTooLarge
	}
	extension := strings.ToLower(filepath.Ext(header.Filename))
	canonical, ok := map[string]string{
		".pdf":  "application/pdf",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
	}[extension]
	if !ok {
		return "", ErrUnsupportedEvidence
	}
	declared := strings.ToLower(strings.TrimSpace(strings.Split(header.Header.Get("Content-Type"), ";")[0]))
	if declared != "" && declared != "application/octet-stream" && declared != "application/zip" && declared != canonical {
		return "", ErrUnsupportedEvidence
	}

	buffer := make([]byte, 512)
	read, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	detected := http.DetectContentType(buffer[:read])
	switch extension {
	case ".pdf":
		if detected != "application/pdf" {
			return "", ErrUnsupportedEvidence
		}
	case ".png":
		if detected != "image/png" {
			return "", ErrUnsupportedEvidence
		}
	case ".jpg", ".jpeg":
		if detected != "image/jpeg" {
			return "", ErrUnsupportedEvidence
		}
	case ".docx", ".xlsx":
		if read < 2 || buffer[0] != 'P' || buffer[1] != 'K' {
			return "", ErrUnsupportedEvidence
		}
	}
	return canonical, nil
}

func writeEvidenceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEvidenceTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "evidence file must be 20 MB or smaller")
	case errors.Is(err, ErrUnsupportedEvidence):
		writeError(w, http.StatusBadRequest, "unsupported_file", "only PDF, DOCX, XLSX, PNG, and JPEG files are allowed")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "could not store document")
	}
}

func writeDocumentError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "document not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", message)
}

func safeDownloadName(name string) string {
	name = filepath.Base(name)
	name = strings.NewReplacer("\r", "", "\n", "", `"`, "'").Replace(name)
	if name == "" || name == "." {
		return "evidence"
	}
	return name
}

func logStorageCleanupFailure(storageKey string, err error) {
	// Keep the API response successful after database deletion; cleanup can be retried safely.
	log.Printf("evidence cleanup failed for %s: %v", storageKey, err)
}
