package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type organizationDataStore interface {
	ListOrganizations(context.Context, *string) ([]store.Organization, error)
	GetOrganization(context.Context, string) (store.Organization, error)
	GetOrganizationBySlug(context.Context, string) (store.Organization, error)
	CreateOrganization(context.Context, string) (store.Organization, error)
	DeleteOrganization(context.Context, string) error
	ListProjectsByOrganization(context.Context, string) ([]store.Project, error)
	GetProjectBySlug(context.Context, string, string) (store.Project, error)
	CreateScopedProject(context.Context, string, string) (store.Project, error)
}

type projectMetadataStore interface {
	CreateScopedProjectWithMetadata(context.Context, string, string, store.ProjectMetadata) (store.Project, error)
}

func (h *Handler) organizationStore() (organizationDataStore, bool) {
	data, ok := h.Store.(organizationDataStore)
	return data, ok
}

func (h *Handler) organizations(w http.ResponseWriter, r *http.Request) {
	data, ok := h.organizationStore()
	if !ok {
		writeError(w, 500, "internal_error", "organization store unavailable")
		return
	}
	var organizationID *string
	user := currentUser(r)
	if h.Auth != nil && user.UserType == "stakeholder" {
		organizationID = user.OrganizationID
	}
	organizations, err := data.ListOrganizations(r.Context(), organizationID)
	if err != nil {
		writeError(w, 500, "internal_error", "could not load organizations")
		return
	}
	writeJSON(w, 200, organizations)
}

func (h *Handler) createOrganization(w http.ResponseWriter, r *http.Request) {
	if h.Auth != nil && !can(currentUser(r), actionCreateOrganization) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid organization request")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		writeError(w, 400, "validation_error", "Organization name is required")
		return
	}
	data, ok := h.organizationStore()
	if !ok {
		writeError(w, 500, "internal_error", "organization store unavailable")
		return
	}
	organization, err := data.CreateOrganization(r.Context(), input.Name)
	if errors.Is(err, store.ErrOrganizationExists) {
		writeError(w, 409, "organization_exists", "Organization already exists")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "Could not create organization")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &organization.ID, Action: "organization.created", EntityType: "organization", EntityID: &organization.ID})
	writeJSON(w, 201, organization)
}

func (h *Handler) authorizeOrganization(w http.ResponseWriter, r *http.Request, id string) (organizationDataStore, bool) {
	data, ok := h.organizationStore()
	if !ok {
		writeError(w, 500, "internal_error", "organization store unavailable")
		return nil, false
	}
	organization, err := data.GetOrganization(r.Context(), id)
	if err != nil || (h.Auth != nil && !canAccessOrganization(currentUser(r), organization.ID)) {
		writeError(w, 404, "not_found", "organization not found")
		return nil, false
	}
	return data, true
}

func (h *Handler) organizationBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	data, ok := h.organizationStore()
	if !ok {
		writeError(w, 500, "internal_error", "organization store unavailable")
		return
	}
	organization, err := data.GetOrganizationBySlug(r.Context(), slug)
	if err != nil || organization.ID == "" || (h.Auth != nil && !canAccessOrganization(currentUser(r), organization.ID)) {
		writeError(w, 404, "not_found", "organization not found")
		return
	}
	writeJSON(w, 200, organization)
}

func (h *Handler) organizationProjects(w http.ResponseWriter, r *http.Request, id string) {
	data, ok := h.authorizeOrganization(w, r, id)
	if !ok {
		return
	}
	projects, err := data.ListProjectsByOrganization(r.Context(), id)
	if err != nil {
		writeError(w, 500, "internal_error", "could not load projects")
		return
	}
	writeJSON(w, 200, projects)
}

func (h *Handler) organizationProjectBySlug(w http.ResponseWriter, r *http.Request, organizationID, slug string) {
	data, ok := h.authorizeOrganization(w, r, organizationID)
	if !ok {
		return
	}
	project, err := data.GetProjectBySlug(r.Context(), organizationID, slug)
	if err != nil || project.ID == "" {
		writeError(w, 404, "not_found", "project not found")
		return
	}
	writeJSON(w, 200, project)
}

func (h *Handler) createOrganizationProject(w http.ResponseWriter, r *http.Request, id string) {
	if h.Auth != nil && !can(currentUser(r), actionCreateProject) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	data, ok := h.authorizeOrganization(w, r, id)
	if !ok {
		return
	}
	var input struct {
		Name                 string `json:"name"`
		Objective            string `json:"objective"`
		AssessmentPeriod     string `json:"assessmentPeriod"`
		TargetCompletionDate string `json:"targetCompletionDate"`
		ScopeBoundary        string `json:"scopeBoundary"`
		ComplianceDriver     string `json:"complianceDriver"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", "Invalid project request")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		writeError(w, 400, "validation_error", "Project name is required")
		return
	}
	input.Objective = strings.TrimSpace(input.Objective)
	input.AssessmentPeriod = strings.TrimSpace(input.AssessmentPeriod)
	input.TargetCompletionDate = strings.TrimSpace(input.TargetCompletionDate)
	input.ScopeBoundary = strings.TrimSpace(input.ScopeBoundary)
	input.ComplianceDriver = strings.TrimSpace(input.ComplianceDriver)
	if input.TargetCompletionDate != "" {
		if _, err := time.Parse("2006-01-02", input.TargetCompletionDate); err != nil {
			writeError(w, 400, "validation_error", "target completion date must use YYYY-MM-DD")
			return
		}
	}
	var project store.Project
	var err error
	if metadataStore, supportsMetadata := data.(projectMetadataStore); supportsMetadata {
		project, err = metadataStore.CreateScopedProjectWithMetadata(r.Context(), id, input.Name, store.ProjectMetadata{
			Objective:            input.Objective,
			AssessmentPeriod:     input.AssessmentPeriod,
			TargetCompletionDate: input.TargetCompletionDate,
			ScopeBoundary:        input.ScopeBoundary,
			ComplianceDriver:     input.ComplianceDriver,
		})
	} else {
		project, err = data.CreateScopedProject(r.Context(), id, input.Name)
	}
	if err != nil {
		writeError(w, 500, "internal_error", "Could not create project")
		return
	}
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &id, ProjectID: &project.ID, Action: "project.created", EntityType: "project", EntityID: &project.ID})
	writeJSON(w, 201, project)
}

func (h *Handler) deleteOrganization(w http.ResponseWriter, r *http.Request, id string) {
	if h.Auth != nil && !can(currentUser(r), actionDeleteOrganization) {
		writeError(w, 403, "forbidden", "Permission denied")
		return
	}
	data, ok := h.authorizeOrganization(w, r, id)
	if !ok {
		return
	}
	storageKeys, err := h.organizationEvidenceKeys(r.Context(), id)
	if err != nil {
		writeError(w, 500, "internal_error", "could not prepare organization deletion")
		return
	}
	if err := data.DeleteOrganization(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 409, "organization_in_use", "Organization still has projects or users")
		return
	} else if err != nil {
		writeError(w, 500, "internal_error", "Could not delete organization")
		return
	}
	h.removeEvidenceFiles(storageKeys)
	w.WriteHeader(http.StatusNoContent)
}
