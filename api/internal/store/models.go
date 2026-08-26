package store

import (
	"errors"
	"time"
)

var ErrInvalidProfileAssignment = errors.New("invalid profile assignment")
var ErrInvalidFunctionScope = errors.New("invalid function scope")
var ErrInvalidProjectTransition = errors.New("invalid project transition")
var ErrProjectFinalized = errors.New("project is finalized")
var ErrProjectNotReady = errors.New("project is not ready to finalize")
var ErrProjectVersionNotFinalized = errors.New("project is not finalized for versioning")
var ErrProjectVersionNotLatest = errors.New("project is not the latest version")
var ErrProjectVersionConflict = errors.New("project version could not be created")
var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
var ErrInvitationNotPending = errors.New("invitation is not pending")
var ErrInvalidProjectAccess = errors.New("invalid project access")

type AuditEvent struct {
	ActorUserID    string
	ActorRole      string
	OrganizationID *string
	ProjectID      *string
	Action         string
	EntityType     string
	EntityID       *string
	Metadata       map[string]any
	Result         string
	RequestID      string
	IPAddress      string
	UserAgent      string
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
	ID                   string     `json:"id"`
	OrganizationID       string     `json:"organizationID"`
	OrganizationName     string     `json:"organizationName"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	VersionGroupID       string     `json:"versionGroupID"`
	VersionNumber        int        `json:"versionNumber"`
	PreviousVersionID    *string    `json:"previousVersionID"`
	IsLatest             bool       `json:"isLatest"`
	Status               string     `json:"status"`
	CreatedAt            string     `json:"createdAt"`
	Objective            string     `json:"objective"`
	AssessmentPeriod     string     `json:"assessmentPeriod"`
	TargetCompletionDate string     `json:"targetCompletionDate"`
	ScopeBoundary        string     `json:"scopeBoundary"`
	ComplianceDriver     string     `json:"complianceDriver"`
	FinalizedAt          *time.Time `json:"finalizedAt"`
	FinalizedBy          *string    `json:"finalizedBy"`
}
type ProjectMetadata struct {
	Objective            string
	AssessmentPeriod     string
	TargetCompletionDate string
	ScopeBoundary        string
	ComplianceDriver     string
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
	CancelledAt    *time.Time `json:"cancelledAt"`
	CancelledBy    *string    `json:"cancelledBy"`
	SupersededAt   *time.Time `json:"supersededAt"`
	SupersededBy   *string    `json:"supersededBy"`
	Status         string     `json:"status"`
	ProjectIDs     []string   `json:"projectIDs"`
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
type RemediationEvidence struct {
	ID           string    `json:"id"`
	ActionID     string    `json:"actionID"`
	OriginalName string    `json:"originalName"`
	StoragePath  string    `json:"-"`
	MIMEType     string    `json:"mimeType"`
	SizeBytes    int64     `json:"sizeBytes"`
	UploadedBy   string    `json:"uploadedBy"`
	CreatedAt    time.Time `json:"createdAt"`
}
type RemediationAction struct {
	ID                   string                `json:"id"`
	ProjectID            string                `json:"projectID"`
	SubcategoryID        string                `json:"subcategoryID"`
	OutcomeCode          string                `json:"outcomeCode"`
	OutcomeDescription   string                `json:"outcomeDescription"`
	CurrentCoverageLevel string                `json:"currentCoverageLevel"`
	TargetCoverageLevel  string                `json:"targetCoverageLevel"`
	Title                string                `json:"title"`
	Description          string                `json:"description"`
	DesiredResult        string                `json:"desiredResult"`
	Priority             string                `json:"priority"`
	OwnerUserID          string                `json:"ownerUserID"`
	OwnerName            string                `json:"ownerName"`
	OwnerEmail           string                `json:"ownerEmail"`
	DueDate              time.Time             `json:"dueDate"`
	Status               string                `json:"status"`
	ProgressNote         string                `json:"progressNote"`
	ReviewComment        string                `json:"reviewComment"`
	CreatedBy            string                `json:"createdBy"`
	SubmittedAt          *time.Time            `json:"submittedAt"`
	ClosedBy             *string               `json:"closedBy"`
	ClosedAt             *time.Time            `json:"closedAt"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	Evidence             []RemediationEvidence `json:"evidence"`
}
type RemediationCreate struct {
	SubcategoryID string    `json:"subcategoryID"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	DesiredResult string    `json:"desiredResult"`
	Priority      string    `json:"priority"`
	OwnerUserID   string    `json:"ownerUserID"`
	DueDate       time.Time `json:"dueDate"`
}
type RemediationPatch struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	DesiredResult *string    `json:"desiredResult,omitempty"`
	Priority      *string    `json:"priority,omitempty"`
	OwnerUserID   *string    `json:"ownerUserID,omitempty"`
	DueDate       *time.Time `json:"dueDate,omitempty"`
}
type RemediationSummary struct {
	OpenCount           int `json:"openCount"`
	InProgressCount     int `json:"inProgressCount"`
	AwaitingReviewCount int `json:"awaitingReviewCount"`
	OverdueCount        int `json:"overdueCount"`
	ClosedCount         int `json:"closedCount"`
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
