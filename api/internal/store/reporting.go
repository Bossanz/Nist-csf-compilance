package store

import (
	"context"
	"encoding/json"
	"time"

	"compliance/api/internal/domain"
)

type ReportFunctionSummary struct {
	Code           string  `json:"code"`
	CoveragePct    float64 `json:"coveragePct"`
	IncludedCount  int     `json:"includedCount"`
	ApprovedCount  int     `json:"approvedCount"`
	ReviewingCount int     `json:"reviewingCount"`
	ReturnedCount  int     `json:"returnedCount"`
	PendingCount   int     `json:"pendingCount"`
	EvidenceCount  int     `json:"evidenceCount"`
}

type ReportSummary struct {
	CoveragePct    float64                 `json:"coveragePct"`
	IncludedCount  int                     `json:"includedCount"`
	ApprovedCount  int                     `json:"approvedCount"`
	ReviewingCount int                     `json:"reviewingCount"`
	ReturnedCount  int                     `json:"returnedCount"`
	PendingCount   int                     `json:"pendingCount"`
	EvidenceCount  int                     `json:"evidenceCount"`
	Functions      []ReportFunctionSummary `json:"functions"`
}

type ReportResponse struct {
	ID            string     `json:"id"`
	ResponseText  string     `json:"responseText"`
	Status        string     `json:"status"`
	RespondedBy   *string    `json:"respondedBy"`
	SubmittedAt   *time.Time `json:"submittedAt"`
	ReviewComment string     `json:"reviewComment"`
	ReviewedBy    *string    `json:"reviewedBy"`
	ReviewedAt    *time.Time `json:"reviewedAt"`
}

type ReportEvidence struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"originalName"`
	MIMEType     string    `json:"mimeType"`
	SizeBytes    int64     `json:"sizeBytes"`
	UploadedBy   string    `json:"uploadedBy"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ReportOutcome struct {
	Profile  ProfileRow       `json:"profile"`
	Response *ReportResponse  `json:"response"`
	Evidence []ReportEvidence `json:"evidence"`
}

type ScopeRegisterEntry struct {
	Profile ProfileRow `json:"profile"`
}

type AuditTrailEntry struct {
	ID          string         `json:"id"`
	ActorUserID *string        `json:"actorUserID"`
	ActorName   string         `json:"actorName"`
	ActorEmail  string         `json:"actorEmail"`
	Action      string         `json:"action"`
	EntityType  string         `json:"entityType"`
	EntityID    *string        `json:"entityID"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type FinalReport struct {
	Project            Project             `json:"project"`
	Summary            ReportSummary       `json:"summary"`
	Outcomes           []ReportOutcome     `json:"outcomes"`
	RemediationSummary RemediationSummary  `json:"remediationSummary"`
	RemediationActions []RemediationAction `json:"remediationActions"`
}

type AuditPackage struct {
	Project            Project              `json:"project"`
	Summary            ReportSummary        `json:"summary"`
	Scope              []ScopeRegisterEntry `json:"scope"`
	Outcomes           []ReportOutcome      `json:"outcomes"`
	AuditTrail         []AuditTrailEntry    `json:"auditTrail"`
	RemediationSummary RemediationSummary   `json:"remediationSummary"`
	RemediationActions []RemediationAction  `json:"remediationActions"`
}

func (s *Store) GetFinalReport(ctx context.Context, projectID string) (FinalReport, error) {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return FinalReport{}, err
	}
	profiles, err := s.ListProfile(ctx, projectID)
	if err != nil {
		return FinalReport{}, err
	}
	responses, err := s.ListResponses(ctx, projectID)
	if err != nil {
		return FinalReport{}, err
	}
	remediationActions, err := s.ListRemediationActions(ctx, projectID)
	if err != nil {
		return FinalReport{}, err
	}
	responseBySubcategory := make(map[string]StakeholderResponse, len(responses))
	for _, response := range responses {
		responseBySubcategory[response.SubcategoryID] = response
	}

	report := FinalReport{Project: project, Outcomes: []ReportOutcome{}, RemediationActions: remediationActions, RemediationSummary: calculateRemediationSummary(remediationActions, time.Now())}
	report.Summary = calculateReportSummary(profiles, responseBySubcategory)
	for _, profile := range profiles {
		if !profile.Included {
			continue
		}
		response, found := responseBySubcategory[profile.SubcategoryID]
		outcome := ReportOutcome{Profile: profile, Evidence: []ReportEvidence{}}
		if found {
			outcome.Response = &ReportResponse{
				ID: response.ID, ResponseText: response.ResponseText, Status: response.Status,
				RespondedBy: response.RespondedBy, SubmittedAt: response.SubmittedAt,
				ReviewComment: response.ReviewComment, ReviewedBy: response.ReviewedBy, ReviewedAt: response.ReviewedAt,
			}
			for _, evidence := range response.Documents {
				outcome.Evidence = append(outcome.Evidence, ReportEvidence{
					ID: evidence.ID, OriginalName: evidence.OriginalName, MIMEType: evidence.MIMEType,
					SizeBytes: evidence.SizeBytes, UploadedBy: evidence.UploadedBy, CreatedAt: evidence.CreatedAt,
				})
			}
		}
		report.Outcomes = append(report.Outcomes, outcome)
	}
	return report, nil
}

func (s *Store) GetAuditPackage(ctx context.Context, projectID string) (AuditPackage, error) {
	report, err := s.GetFinalReport(ctx, projectID)
	if err != nil {
		return AuditPackage{}, err
	}
	profiles, err := s.ListProfile(ctx, projectID)
	if err != nil {
		return AuditPackage{}, err
	}
	auditTrail, err := s.ListProjectAuditEvents(ctx, projectID)
	if err != nil {
		return AuditPackage{}, err
	}
	scope := make([]ScopeRegisterEntry, 0, len(profiles))
	for _, profile := range profiles {
		scope = append(scope, ScopeRegisterEntry{Profile: profile})
	}
	return AuditPackage{Project: report.Project, Summary: report.Summary, Scope: scope, Outcomes: report.Outcomes, AuditTrail: auditTrail, RemediationSummary: report.RemediationSummary, RemediationActions: report.RemediationActions}, nil
}

func calculateRemediationSummary(actions []RemediationAction, now time.Time) RemediationSummary {
	var summary RemediationSummary
	today := now.UTC().Format("2006-01-02")
	for _, action := range actions {
		switch action.Status {
		case "open":
			summary.OpenCount++
		case "in_progress":
			summary.InProgressCount++
		case "awaiting_review":
			summary.AwaitingReviewCount++
		case "closed":
			summary.ClosedCount++
		}
		if action.Status != "closed" && action.DueDate.UTC().Format("2006-01-02") < today {
			summary.OverdueCount++
		}
	}
	return summary
}

func (s *Store) ListProjectAuditEvents(ctx context.Context, projectID string) ([]AuditTrailEntry, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id::text,a.actor_user_id::text,COALESCE(u.name,''),COALESCE(u.email,''),
		       a.action,a.entity_type,a.entity_id::text,a.metadata,a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id=a.actor_user_id
		WHERE a.project_id=$1
		ORDER BY a.created_at ASC,a.id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []AuditTrailEntry{}
	for rows.Next() {
		var event AuditTrailEntry
		var metadata []byte
		if err := rows.Scan(&event.ID, &event.ActorUserID, &event.ActorName, &event.ActorEmail, &event.Action, &event.EntityType, &event.EntityID, &metadata, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.Metadata = map[string]any{}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
				return nil, err
			}
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

type reportCounter struct {
	approved  int
	reviewing int
	returned  int
	pending   int
	evidence  int
}

func calculateReportSummary(profiles []ProfileRow, responses map[string]StakeholderResponse) ReportSummary {
	allScores := make([]domain.ProfileScore, 0, len(profiles))
	functionScores := map[string][]domain.ProfileScore{}
	functionOrder := []string{}
	functionCounters := map[string]*reportCounter{}
	var overall reportCounter
	for _, profile := range profiles {
		score := domain.ProfileScore{Included: profile.Included, Current: domain.CoverageLevel(profile.CurrentCoverageLevel), Target: domain.CoverageLevel(profile.TargetCoverageLevel)}
		allScores = append(allScores, score)
		if !profile.Included {
			continue
		}
		if _, ok := functionCounters[profile.FunctionCode]; !ok {
			functionOrder = append(functionOrder, profile.FunctionCode)
			functionCounters[profile.FunctionCode] = &reportCounter{}
		}
		functionScores[profile.FunctionCode] = append(functionScores[profile.FunctionCode], score)
		response, ok := responses[profile.SubcategoryID]
		counter := functionCounters[profile.FunctionCode]
		if !ok || response.Status == "draft" {
			overall.pending++
			counter.pending++
		} else {
			switch response.Status {
			case "reviewed":
				overall.approved++
				counter.approved++
			case "submitted":
				overall.reviewing++
				counter.reviewing++
			case "needs_more_info":
				overall.returned++
				counter.returned++
			default:
				overall.pending++
				counter.pending++
			}
		}
		if ok {
			overall.evidence += len(response.Documents)
			counter.evidence += len(response.Documents)
		}
	}
	base := domain.CalculateSummary(allScores)
	summary := ReportSummary{
		CoveragePct: base.CoveragePct, IncludedCount: base.IncludedCount,
		ApprovedCount: overall.approved, ReviewingCount: overall.reviewing,
		ReturnedCount: overall.returned, PendingCount: overall.pending, EvidenceCount: overall.evidence,
		Functions: []ReportFunctionSummary{},
	}
	for _, code := range functionOrder {
		functionSummary := domain.CalculateSummary(functionScores[code])
		counter := functionCounters[code]
		summary.Functions = append(summary.Functions, ReportFunctionSummary{
			Code: code, CoveragePct: functionSummary.CoveragePct, IncludedCount: functionSummary.IncludedCount,
			ApprovedCount: counter.approved, ReviewingCount: counter.reviewing, ReturnedCount: counter.returned,
			PendingCount: counter.pending, EvidenceCount: counter.evidence,
		})
	}
	return summary
}
