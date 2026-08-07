package httpapi

import (
	authservice "compliance/api/internal/auth"
	"compliance/api/internal/domain"
	"compliance/api/internal/store"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

type dataStore interface {
	ListFunctions(context.Context) ([]store.Function, error)
	CreateProject(context.Context, string, string) (store.Project, error)
	ListProjects(context.Context) ([]store.Project, error)
	DeleteProject(context.Context, string) error
	GetProject(context.Context, string) (store.Project, error)
	ListProfile(context.Context, string) ([]store.ProfileRow, error)
	UpdateProfile(context.Context, string, string, store.ProfilePatch) (store.ProfileRow, error)
}
type responseStore interface {
	ListResponses(context.Context, string) ([]store.StakeholderResponse, error)
	SaveResponseDraft(context.Context, string, string, string, string) (store.StakeholderResponse, error)
	SubmitResponse(context.Context, string, string, string) (store.StakeholderResponse, error)
	ReviewResponse(context.Context, string, string, string, string, string) (store.StakeholderResponse, error)
}
type documentStore interface {
	CreateResponseDocument(context.Context, string, string, string, string, string, string, int64) (store.ResponseDocument, error)
	GetResponseDocument(context.Context, string, string, string) (store.ResponseDocument, error)
	DeleteResponseDocument(context.Context, string, string, string) (store.ResponseDocument, error)
}
type evidenceStorage interface {
	Save(io.Reader) (string, int64, error)
	Open(string) (io.ReadCloser, error)
	Remove(string) error
}
type evidenceKeyStore interface {
	ListProjectEvidenceKeys(context.Context, string) ([]string, error)
	ListOrganizationEvidenceKeys(context.Context, string) ([]string, error)
}
type Handler struct {
	Store         dataStore
	Responses     responseStore
	Documents     documentStore
	Evidence      evidenceStorage
	Auth          *authservice.Service
	SecureCookies bool
	LoginThrottle *LoginThrottle
	AppOrigin     string
	Invitations   *authservice.InvitationService
}
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func New(s *store.Store, auth *authservice.Service, invitations *authservice.InvitationService, secureCookies bool, appOrigin string) http.Handler {
	return &Handler{Store: s, Responses: s, Documents: s, Evidence: newLocalEvidenceStorage(""), Auth: auth, Invitations: invitations, SecureCookies: secureCookies, LoginThrottle: NewLoginThrottle(), AppOrigin: appOrigin}
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.AppOrigin != "" && r.Header.Get("Origin") == h.AppOrigin {
		w.Header().Set("Access-Control-Allow-Origin", h.AppOrigin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.URL.Path == "/healthz" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if !h.validMutationRequest(r) {
		writeError(w, http.StatusForbidden, "invalid_origin", "Request origin is not allowed")
		return
	}
	if r.URL.Path == "/api/auth/login" && r.Method == http.MethodPost {
		h.login(w, r)
		return
	}
	if r.URL.Path == "/api/auth/logout" && r.Method == http.MethodPost {
		h.logout(w, r)
		return
	}
	if r.URL.Path == "/api/auth/me" && r.Method == http.MethodGet {
		h.me(w, r)
		return
	}
	invitePath := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(invitePath) == 4 && invitePath[0] == "api" && invitePath[1] == "invitations" && invitePath[3] == "accept" && r.Method == http.MethodPost {
		h.acceptInvitation(w, r, invitePath[2])
		return
	}
	var authenticated bool
	r, authenticated = h.authenticate(w, r)
	if !authenticated {
		return
	}
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "organizations" {
		if r.Method == http.MethodGet {
			h.organizations(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.createOrganization(w, r)
			return
		}
	}
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "organizations" {
		id := parts[2]
		if len(parts) == 3 && r.Method == http.MethodDelete {
			h.deleteOrganization(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "projects" && r.Method == http.MethodGet {
			h.organizationProjects(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "projects" && r.Method == http.MethodPost {
			h.createOrganizationProject(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "invitations" && r.Method == http.MethodPost {
			h.inviteStakeholder(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "users" && r.Method == http.MethodGet {
			h.organizationUsers(w, r, id)
			return
		}
		if len(parts) == 5 && parts[3] == "users" && r.Method == http.MethodPatch {
			h.updateOrganizationUser(w, r, id, parts[4])
			return
		}
	}
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "counselor-invitations" && r.Method == http.MethodPost {
		h.inviteCounselor(w, r)
		return
	}
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "counselors" && r.Method == http.MethodGet {
		h.counselors(w, r)
		return
	}
	if len(parts) == 3 && parts[0] == "api" && parts[1] == "counselors" && r.Method == http.MethodPatch {
		h.updateCounselor(w, r, parts[2])
		return
	}
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "functions" && r.Method == http.MethodGet {
		h.functions(w, r)
		return
	}
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "projects" && r.Method == http.MethodGet {
		h.projects(w, r)
		return
	}
	if len(parts) == 2 && parts[0] == "api" && parts[1] == "projects" && r.Method == http.MethodPost {
		if h.Auth != nil && !can(currentUser(r), actionCreateProject) {
			writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
			return
		}
		h.createProject(w, r)
		return
	}
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "projects" {
		id := parts[2]
		if len(parts) == 3 && r.Method == http.MethodDelete {
			action := actionDeleteProject
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.deleteProject(w, r, id)
			return
		}
		if len(parts) == 3 && r.Method == http.MethodGet {
			if !h.authorizeProject(w, r, id, nil) {
				return
			}
			h.project(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "profile" && r.Method == http.MethodGet {
			if !h.authorizeProject(w, r, id, nil) {
				return
			}
			h.profile(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "summary" && r.Method == http.MethodGet {
			if !h.authorizeProject(w, r, id, nil) {
				return
			}
			h.summary(w, r, id)
			return
		}
		if len(parts) == 4 && parts[3] == "responses" && r.Method == http.MethodGet {
			if !h.authorizeProject(w, r, id, nil) {
				return
			}
			h.responses(w, r, id)
			return
		}
		if len(parts) == 5 && parts[3] == "responses" && r.Method == http.MethodPut {
			action := actionSaveResponse
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.saveResponse(w, r, id, parts[4])
			return
		}
		if len(parts) == 6 && parts[3] == "responses" && parts[5] == "submit" && r.Method == http.MethodPost {
			action := actionSubmitResponse
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.submitResponse(w, r, id, parts[4])
			return
		}
		if len(parts) == 6 && parts[3] == "responses" && parts[5] == "review" && r.Method == http.MethodPost {
			action := actionReviewResponse
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.reviewResponse(w, r, id, parts[4])
			return
		}
		if len(parts) == 6 && parts[3] == "responses" && parts[5] == "documents" && r.Method == http.MethodPost {
			action := actionSaveResponse
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.uploadResponseDocument(w, r, id, parts[4])
			return
		}
		if len(parts) == 7 && parts[3] == "responses" && parts[5] == "documents" && r.Method == http.MethodGet {
			if !h.authorizeProject(w, r, id, nil) {
				return
			}
			h.downloadResponseDocument(w, r, id, parts[4], parts[6])
			return
		}
		if len(parts) == 7 && parts[3] == "responses" && parts[5] == "documents" && r.Method == http.MethodDelete {
			action := actionSaveResponse
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.deleteResponseDocument(w, r, id, parts[4], parts[6])
			return
		}
		if len(parts) == 5 && parts[3] == "profile" && r.Method == http.MethodPut {
			action := actionUpdateProfile
			if !h.authorizeProject(w, r, id, &action) {
				return
			}
			h.updateProfile(w, r, id, parts[4])
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (h *Handler) functions(w http.ResponseWriter, r *http.Request) {
	data, err := h.Store.ListFunctions(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", "could not load catalog")
		return
	}
	writeJSON(w, 200, data)
}
func (h *Handler) projects(w http.ResponseWriter, r *http.Request) {
	data, err := h.Store.ListProjects(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", "could not load projects")
		return
	}
	if h.Auth != nil && currentUser(r).UserType == "stakeholder" {
		scoped := data[:0]
		for _, project := range data {
			if canAccessOrganization(currentUser(r), project.OrganizationID) {
				scoped = append(scoped, project)
			}
		}
		data = scoped
	}
	writeJSON(w, 200, data)
}
func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, 404, "not_found", "project not found")
		return
	}
	storageKeys, err := h.projectEvidenceKeys(r.Context(), id)
	if err != nil {
		writeError(w, 500, "internal_error", "could not prepare project deletion")
		return
	}
	err = h.Store.DeleteProject(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 404, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "could not delete project")
		return
	}
	h.removeEvidenceFiles(storageKeys)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) projectEvidenceKeys(ctx context.Context, projectID string) ([]string, error) {
	keyStore, ok := h.Store.(evidenceKeyStore)
	if !ok || h.Evidence == nil {
		return nil, nil
	}
	return keyStore.ListProjectEvidenceKeys(ctx, projectID)
}

func (h *Handler) organizationEvidenceKeys(ctx context.Context, organizationID string) ([]string, error) {
	keyStore, ok := h.Store.(evidenceKeyStore)
	if !ok || h.Evidence == nil {
		return nil, nil
	}
	return keyStore.ListOrganizationEvidenceKeys(ctx, organizationID)
}

func (h *Handler) removeEvidenceFiles(storageKeys []string) {
	for _, storageKey := range storageKeys {
		if err := h.Evidence.Remove(storageKey); err != nil {
			logStorageCleanupFailure(storageKey, err)
		}
	}
}
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name             string `json:"name"`
		OrganizationName string `json:"organizationName"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, 400, "validation_error", "project name is required")
		return
	}
	if strings.TrimSpace(input.OrganizationName) == "" {
		input.OrganizationName = "Unnamed organization"
	}
	p, err := h.Store.CreateProject(r.Context(), input.Name, input.OrganizationName)
	if err != nil {
		writeError(w, 500, "internal_error", "could not create project")
		return
	}
	writeJSON(w, 201, p)
}
func (h *Handler) project(w http.ResponseWriter, r *http.Request, id string) {
	p, err := h.Store.GetProject(r.Context(), id)
	if err != nil {
		writeError(w, 404, "not_found", "project not found")
		return
	}
	writeJSON(w, 200, p)
}
func (h *Handler) profile(w http.ResponseWriter, r *http.Request, id string) {
	p, err := h.Store.ListProfile(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 404, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "could not load profile")
		return
	}
	writeJSON(w, 200, p)
}
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) {
	var patch store.ProfilePatch
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, 400, "invalid_json", err.Error())
		return
	}
	for _, level := range []*string{patch.CurrentCoverageLevel, patch.TargetCoverageLevel} {
		if level != nil {
			if _, err := domain.Score(domain.CoverageLevel(*level)); err != nil {
				writeError(w, 400, "validation_error", err.Error())
				return
			}
		}
	}
	p, err := h.Store.UpdateProfile(r.Context(), projectID, subcategoryID, patch)
	if err != nil {
		log.Printf("profile update failed: %v", err)
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "no fields") {
			writeError(w, 400, "validation_error", err.Error())
		} else {
			writeError(w, 404, "not_found", "profile not found")
		}
		return
	}
	project, _ := h.Store.GetProject(r.Context(), projectID)
	h.writeAudit(currentUser(r), r.Context(), store.AuditEvent{OrganizationID: &project.OrganizationID, ProjectID: &projectID, Action: "profile.updated", EntityType: "profile", EntityID: &subcategoryID})
	writeJSON(w, 200, p)
}

type FunctionSummary struct {
	Code          string  `json:"code"`
	CoveragePct   float64 `json:"coveragePct"`
	IncludedCount int     `json:"includedCount"`
}
type summaryResponse struct {
	domain.Summary
	Functions []FunctionSummary `json:"functions"`
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request, id string) {
	rows, err := h.Store.ListProfile(r.Context(), id)
	if err != nil {
		writeError(w, 500, "internal_error", "could not calculate summary")
		return
	}
	all := make([]domain.ProfileScore, 0, len(rows))
	groups := map[string][]domain.ProfileScore{}
	for _, row := range rows {
		score := domain.ProfileScore{Included: row.Included, Current: domain.CoverageLevel(row.CurrentCoverageLevel), Target: domain.CoverageLevel(row.TargetCoverageLevel)}
		all = append(all, score)
		groups[row.FunctionCode] = append(groups[row.FunctionCode], score)
	}
	out := summaryResponse{Summary: domain.CalculateSummary(all)}
	for code, items := range groups {
		s := domain.CalculateSummary(items)
		out.Functions = append(out.Functions, FunctionSummary{Code: code, CoveragePct: s.CoveragePct, IncludedCount: s.IncludedCount})
	}
	writeJSON(w, 200, out)
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("request must contain one JSON object")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	var body errorBody
	body.Error.Code = code
	body.Error.Message = message
	writeJSON(w, status, body)
}
