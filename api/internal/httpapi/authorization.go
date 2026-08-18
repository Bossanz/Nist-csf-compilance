package httpapi

import (
	"context"
	"net/http"
	"strings"

	"compliance/api/internal/domain"
	"compliance/api/internal/store"
)

func (h *Handler) validMutationRequest(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || h.AppOrigin == "" {
		return true
	}
	if r.Header.Get("Origin") != h.AppOrigin {
		return false
	}
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	return r.ContentLength == 0 || strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "multipart/form-data")
}

type action string

const (
	actionCreateOrganization action = "create_organization"
	actionDeleteOrganization action = "delete_organization"
	actionCreateProject      action = "create_project"
	actionDeleteProject      action = "delete_project"
	actionSubmitScope        action = "submit_scope"
	actionFinalizeProject    action = "finalize_project"
	actionUpdateProfile      action = "update_profile"
	actionSaveResponse       action = "save_response"
	actionSubmitResponse     action = "submit_response"
	actionReviewResponse     action = "review_response"
	actionInviteStakeholder  action = "invite_stakeholder"
	actionManageCounselor    action = "manage_counselor"
)

func can(user store.User, requested action) bool {
	switch requested {
	case actionCreateOrganization, actionDeleteOrganization, actionManageCounselor:
		return user.Role == "counselor_admin"
	case actionCreateProject, actionDeleteProject, actionSubmitScope, actionFinalizeProject:
		return user.Role == "counselor_admin" || user.Role == "counselor"
	case actionUpdateProfile:
		return user.Role == "counselor_admin" || user.Role == "counselor" || user.Role == "org_admin" || user.Role == "assessor"
	case actionSaveResponse, actionSubmitResponse:
		return user.Role == "org_admin" || user.Role == "assessor"
	case actionReviewResponse:
		return user.Role == "reviewer"
	case actionInviteStakeholder:
		return user.Role == "counselor_admin" || user.Role == "counselor" || user.Role == "org_admin"
	default:
		return false
	}
}

func stakeholderCanReadProfile(user store.User, row store.ProfileRow) bool {
	if !row.Included {
		return false
	}
	switch user.Role {
	case "reviewer", "viewer":
		return true
	case "org_admin", "assessor":
		return row.AssignedUserID != nil && *row.AssignedUserID == user.ID
	default:
		return false
	}
}

func stakeholderCanEditProfile(user store.User, row store.ProfileRow) bool {
	if !row.Included || row.AssignedUserID == nil || *row.AssignedUserID != user.ID {
		return false
	}
	return user.Role == "org_admin" || user.Role == "assessor"
}

func profileFieldsAllowedForRole(user store.User, patch store.ProfilePatch) bool {
	if user.Role == "counselor_admin" || user.Role == "counselor" {
		return patch.CurrentPriority == nil &&
			patch.CurrentCoverageLevel == nil &&
			patch.CurrentStatusText == nil &&
			patch.CurrentPoliciesText == nil &&
			patch.TargetPriority == nil &&
			patch.TargetCoverageLevel == nil &&
			patch.TargetApproachText == nil &&
			patch.Notes == nil &&
			patch.Considerations == nil
	}
	if user.Role == "org_admin" || user.Role == "assessor" {
		return patch.Included == nil && patch.Rationale == nil && patch.AssignedUserID == nil
	}
	return false
}

type userContextKey struct{}

func withCurrentUser(r *http.Request, user store.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userContextKey{}, user))
}

func currentUser(r *http.Request) store.User {
	user, _ := r.Context().Value(userContextKey{}).(store.User)
	return user
}

func canAccessOrganization(user store.User, organizationID string) bool {
	if user.UserType == "counselor" {
		return true
	}
	return user.OrganizationID != nil && *user.OrganizationID == organizationID
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) (*http.Request, bool) {
	if h.Auth == nil {
		return r, true
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required")
		return r, false
	}
	user, err := h.Auth.Authenticate(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required")
		return r, false
	}
	return withCurrentUser(r, user), true
}

func (h *Handler) authorizeProject(w http.ResponseWriter, r *http.Request, projectID string, requested *action) bool {
	if h.Auth == nil {
		return true
	}
	project, err := h.Store.GetProject(r.Context(), projectID)
	if currentUser(r).UserType == "stakeholder" && project.Status == "setup" {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return false
	}
	if err != nil || !canAccessOrganization(currentUser(r), project.OrganizationID) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return false
	}
	if requested != nil && !can(currentUser(r), *requested) {
		writeError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return false
	}
	if requested != nil && project.Status == "closed" && *requested != actionDeleteProject && *requested != actionFinalizeProject {
		writeError(w, http.StatusConflict, "project_finalized", "Project is finalized and read-only")
		return false
	}
	return true
}

func (h *Handler) profileOutcome(ctx context.Context, projectID, subcategoryID string) (store.ProfileRow, bool, error) {
	rows, err := h.Store.ListProfile(ctx, projectID)
	if err != nil {
		return store.ProfileRow{}, false, err
	}
	for _, row := range rows {
		if row.SubcategoryID == subcategoryID {
			return row, true, nil
		}
	}
	return store.ProfileRow{}, false, nil
}

func (h *Handler) ensureOutcomeEditable(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string) bool {
	if h.Responses == nil {
		return true
	}
	responses, err := h.Responses.ListResponses(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not check response access")
		return false
	}
	for _, response := range responses {
		if response.SubcategoryID == subcategoryID && !domain.CanEditResponse(domain.ResponseStatus(response.Status)) {
			writeError(w, http.StatusConflict, "invalid_transition", "response cannot be changed after it is sent for review")
			return false
		}
	}
	return true
}

func outcomeMutation(requested *action) bool {
	return requested != nil && (*requested == actionSaveResponse || *requested == actionSubmitResponse)
}

func (h *Handler) authorizeProjectOutcome(w http.ResponseWriter, r *http.Request, projectID, subcategoryID string, requested *action) bool {
	if !h.authorizeProject(w, r, projectID, requested) {
		return false
	}
	if h.Auth == nil || currentUser(r).UserType != "stakeholder" {
		return true
	}
	row, found, err := h.profileOutcome(r.Context(), projectID, subcategoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not check outcome access")
		return false
	}
	if !found || !row.Included {
		writeError(w, http.StatusNotFound, "not_found", "outcome not found")
		return false
	}
	user := currentUser(r)
	if (user.Role == "org_admin" || user.Role == "assessor") && !stakeholderCanEditProfile(user, row) {
		if outcomeMutation(requested) {
			writeError(w, http.StatusForbidden, "forbidden", "Outcome is not assigned to this user")
		} else {
			writeError(w, http.StatusNotFound, "not_found", "outcome not found")
		}
		return false
	}
	if !stakeholderCanReadProfile(user, row) {
		writeError(w, http.StatusNotFound, "not_found", "outcome not found")
		return false
	}
	if outcomeMutation(requested) && !h.ensureOutcomeEditable(w, r, projectID, subcategoryID) {
		return false
	}
	return true
}
