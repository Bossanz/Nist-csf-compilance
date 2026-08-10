package store

import (
	"errors"
	"time"
)

var ErrInvalidProfileAssignment = errors.New("invalid profile assignment")

type AuditEvent struct {
	ActorUserID    string
	OrganizationID *string
	ProjectID      *string
	Action         string
	EntityType     string
	EntityID       *string
	Metadata       map[string]any
}

type Function struct {
	ID          string     `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Categories  []Category `json:"categories,omitempty"`
}
type Category struct {
	ID            string        `json:"id"`
	FunctionID    string        `json:"functionId"`
	Code          string        `json:"code"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Subcategories []Subcategory `json:"subcategories,omitempty"`
}
type Subcategory struct {
	ID          string `json:"id"`
	CategoryID  string `json:"categoryId"`
	Code        string `json:"code"`
	Description string `json:"description"`
}
type Project struct {
	ID               string `json:"id"`
	OrganizationID   string `json:"organizationID"`
	OrganizationName string `json:"organizationName"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
}
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}
type User struct {
	ID             string  `json:"id"`
	OrganizationID *string `json:"organizationID"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	UserType       string  `json:"userType"`
	Role           string  `json:"role"`
	Status         string  `json:"status"`
	PasswordHash   string  `json:"-"`
}
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userID"`
	TokenHash  string    `json:"-"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	CreatedAt  time.Time `json:"createdAt"`
}
type Invitation struct {
	ID             string     `json:"id"`
	OrganizationID *string    `json:"organizationID"`
	Email          string     `json:"email"`
	UserType       string     `json:"userType"`
	Role           string     `json:"role"`
	TokenHash      string     `json:"-"`
	InvitedBy      string     `json:"invitedBy"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}
type StakeholderResponse struct {
	ID            string             `json:"id"`
	ProjectID     string             `json:"projectID"`
	SubcategoryID string             `json:"subcategoryID"`
	ResponseText  string             `json:"responseText"`
	Status        string             `json:"status"`
	RespondedBy   *string            `json:"respondedBy"`
	SubmittedAt   *time.Time         `json:"submittedAt"`
	ReviewComment string             `json:"reviewComment"`
	ReviewedBy    *string            `json:"reviewedBy"`
	ReviewedAt    *time.Time         `json:"reviewedAt"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	Documents     []ResponseDocument `json:"documents"`
}
type ResponseDocument struct {
	ID           string    `json:"id"`
	ResponseID   string    `json:"responseID"`
	OriginalName string    `json:"originalName"`
	StorageKey   string    `json:"-"`
	MIMEType     string    `json:"mimeType"`
	SizeBytes    int64     `json:"sizeBytes"`
	UploadedBy   string    `json:"uploadedBy"`
	CreatedAt    time.Time `json:"createdAt"`
}
type ProjectFunction struct {
	ID         string `json:"id"`
	FunctionID string `json:"functionId"`
	Code       string `json:"code"`
	Applicable bool   `json:"applicable"`
	Reason     string `json:"reason"`
}
type ProfileRow struct {
	AssignedUserID       *string `json:"assignedUserID"`
	AssignedUserName     string  `json:"assignedUserName"`
	AssignedUserEmail    string  `json:"assignedUserEmail"`
	ID                   string  `json:"id"`
	ProjectID            string  `json:"projectID"`
	SubcategoryID        string  `json:"subcategoryID"`
	FunctionCode         string  `json:"functionCode"`
	CategoryCode         string  `json:"categoryCode"`
	SubcategoryCode      string  `json:"subcategoryCode"`
	Description          string  `json:"description"`
	Included             bool    `json:"included"`
	Rationale            string  `json:"rationale"`
	CurrentPriority      string  `json:"currentPriority"`
	CurrentCoverageLevel string  `json:"currentCoverageLevel"`
	CurrentStatusText    string  `json:"currentStatusText"`
	CurrentPoliciesText  string  `json:"currentPoliciesText"`
	CurrentTier          string  `json:"currentTier"`
	TargetPriority       string  `json:"targetPriority"`
	TargetCoverageLevel  string  `json:"targetCoverageLevel"`
	TargetApproachText   string  `json:"targetApproachText"`
	TargetTier           string  `json:"targetTier"`
	Notes                string  `json:"notes"`
	Considerations       string  `json:"considerations"`
	ReviewStatus         string  `json:"reviewStatus"`
}
type ProfilePatch struct {
	AssignedUserID       *string `json:"assignedUserID,omitempty"`
	Included             *bool   `json:"included,omitempty"`
	Rationale            *string `json:"rationale,omitempty"`
	CurrentPriority      *string `json:"currentPriority,omitempty"`
	CurrentCoverageLevel *string `json:"currentCoverageLevel,omitempty"`
	CurrentStatusText    *string `json:"currentStatusText,omitempty"`
	CurrentPoliciesText  *string `json:"currentPoliciesText,omitempty"`
	TargetPriority       *string `json:"targetPriority,omitempty"`
	TargetCoverageLevel  *string `json:"targetCoverageLevel,omitempty"`
	TargetApproachText   *string `json:"targetApproachText,omitempty"`
	Notes                *string `json:"notes,omitempty"`
	Considerations       *string `json:"considerations,omitempty"`
}
