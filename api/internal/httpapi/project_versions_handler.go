package httpapi

import (
	"errors"
	"net/http"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) createProjectVersion(w http.ResponseWriter, r *http.Request, sourceProjectID string) {
	data, ok := h.Store.(projectVersionStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "project version store unavailable")
		return
	}
	sourceProject, err := h.Store.GetProject(r.Context(), sourceProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load source project")
		return
	}
	created, err := data.CreateNextProjectVersion(r.Context(), sourceProjectID, currentUser(r).ID)
	switch {
	case errors.Is(err, store.ErrProjectVersionNotFinalized):
		writeError(w, http.StatusConflict, "project_not_finalized", "Only a finalized project can start a new assessment version")
		return
	case errors.Is(err, store.ErrProjectVersionNotLatest):
		writeError(w, http.StatusConflict, "project_version_not_latest", "Start the next version from the latest project version")
		return
	case errors.Is(err, store.ErrProjectVersionConflict):
		writeError(w, http.StatusConflict, "version_creation_conflict", "The next project version could not be created")
		return
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create project version")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{
		OrganizationID: &created.OrganizationID,
		ProjectID:      &created.ID,
		Action:         "project.version_created",
		EntityType:     "project",
		EntityID:       &created.ID,
		Metadata: map[string]any{
			"sourceProjectID": sourceProject.ID,
			"newProjectID":    created.ID,
			"sourceVersion":   sourceProject.VersionNumber,
			"newVersion":      created.VersionNumber,
		},
	})
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listProjectVersions(w http.ResponseWriter, r *http.Request, projectID string) {
	data, ok := h.Store.(projectVersionStore)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "project version store unavailable")
		return
	}
	versions, err := data.ListProjectVersions(r.Context(), projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load project versions")
		return
	}
	if h.Auth != nil && currentUser(r).UserType == "stakeholder" {
		visible := make([]store.Project, 0, len(versions))
		for _, version := range versions {
			if currentUser(r).Role == "auditor" {
				allowed, accessErr := h.hasActiveProjectAuditorAccess(r.Context(), version.ID, currentUser(r).ID)
				if accessErr != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", "could not check Auditor project access")
					return
				}
				if allowed {
					visible = append(visible, version)
				}
				continue
			}
			if version.Status != "setup" {
				visible = append(visible, version)
			}
		}
		versions = visible
	}
	writeJSON(w, http.StatusOK, versions)
}
