package store

import (
	"context"
	"errors"
	"strings"

	"compliance/api/internal/domain"
	"github.com/jackc/pgx/v5"
)

var (
	ErrOutcomeNotApproved           = errors.New("outcome is not approved")
	ErrNoCoverageGap                = errors.New("outcome has no coverage gap")
	ErrInvalidRemediationOwner      = errors.New("invalid remediation owner")
	ErrInvalidRemediationTransition = errors.New("invalid remediation transition")
	ErrRemediationClosed            = errors.New("remediation action is closed")
	ErrRemediationForbidden         = errors.New("remediation action is not assigned to actor")
	ErrInvalidRemediationInput      = errors.New("invalid remediation input")
)

const remediationColumns = `
	a.id,a.project_id,a.subcategory_id,sc.code,sc.description,
	p.current_coverage_level,p.target_coverage_level,
	a.title,a.description,a.desired_result,a.priority,a.owner_user_id,
	u.name,u.email,a.due_date,a.status,a.progress_note,a.review_comment,
	a.created_by,a.submitted_at,a.closed_by,a.closed_at,a.created_at,a.updated_at`

func (s *Store) ListRemediationActions(ctx context.Context, projectID string) ([]RemediationAction, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+remediationColumns+`
		FROM remediation_actions a
		JOIN project_subcategory_profiles p ON p.project_id=a.project_id AND p.subcategory_id=a.subcategory_id
		JOIN subcategories sc ON sc.id=a.subcategory_id
		JOIN users u ON u.id=a.owner_user_id
		WHERE a.project_id=$1
		ORDER BY (a.status='closed'),a.due_date,a.created_at,a.id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	actions := []RemediationAction{}
	for rows.Next() {
		action, err := scanRemediationAction(rows)
		if err != nil {
			return nil, err
		}
		action.Evidence, err = s.listRemediationEvidence(ctx, action.ID)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, rows.Err()
}

func (s *Store) CreateRemediationAction(ctx context.Context, projectID, actorID string, input RemediationCreate) (RemediationAction, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.DesiredResult = strings.TrimSpace(input.DesiredResult)
	if input.Title == "" || input.SubcategoryID == "" || input.OwnerUserID == "" || input.DueDate.IsZero() || !domain.ValidRemediationPriority(domain.RemediationPriority(input.Priority)) {
		return RemediationAction{}, ErrInvalidRemediationInput
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return RemediationAction{}, err
	}
	defer tx.Rollback(ctx)

	var included bool
	var currentCoverage, targetCoverage, responseStatus string
	err = tx.QueryRow(ctx, `SELECT p.included,p.current_coverage_level,p.target_coverage_level,COALESCE(r.status,'')
		FROM project_subcategory_profiles p
		LEFT JOIN stakeholder_responses r ON r.project_id=p.project_id AND r.subcategory_id=p.subcategory_id
		WHERE p.project_id=$1 AND p.subcategory_id=$2
		FOR UPDATE OF p`, projectID, input.SubcategoryID).Scan(&included, &currentCoverage, &targetCoverage, &responseStatus)
	if err != nil {
		return RemediationAction{}, err
	}
	if !included || responseStatus != string(domain.ResponseReviewed) {
		return RemediationAction{}, ErrOutcomeNotApproved
	}
	if !domain.HasCoverageGap(domain.CoverageLevel(currentCoverage), domain.CoverageLevel(targetCoverage)) {
		return RemediationAction{}, ErrNoCoverageGap
	}
	validOwner, err := validRemediationOwner(ctx, tx, projectID, input.OwnerUserID)
	if err != nil {
		return RemediationAction{}, err
	}
	if !validOwner {
		return RemediationAction{}, ErrInvalidRemediationOwner
	}

	var actionID string
	err = tx.QueryRow(ctx, `INSERT INTO remediation_actions(
		project_id,subcategory_id,title,description,desired_result,priority,owner_user_id,due_date,created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		projectID, input.SubcategoryID, input.Title, input.Description, input.DesiredResult,
		input.Priority, input.OwnerUserID, input.DueDate, actorID).Scan(&actionID)
	if err != nil {
		return RemediationAction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RemediationAction{}, err
	}
	return s.getRemediationAction(ctx, projectID, actionID)
}

func (s *Store) UpdateRemediationAction(ctx context.Context, projectID, actionID, _ string, patch RemediationPatch) (RemediationAction, error) {
	current, err := s.getRemediationAction(ctx, projectID, actionID)
	if err != nil {
		return RemediationAction{}, err
	}
	if current.Status == string(domain.RemediationClosed) {
		return RemediationAction{}, ErrRemediationClosed
	}
	if patch.Title != nil {
		trimmed := strings.TrimSpace(*patch.Title)
		if trimmed == "" {
			return RemediationAction{}, ErrInvalidRemediationInput
		}
		patch.Title = &trimmed
	}
	if patch.Description != nil {
		trimmed := strings.TrimSpace(*patch.Description)
		patch.Description = &trimmed
	}
	if patch.DesiredResult != nil {
		trimmed := strings.TrimSpace(*patch.DesiredResult)
		patch.DesiredResult = &trimmed
	}
	if patch.Priority != nil && !domain.ValidRemediationPriority(domain.RemediationPriority(*patch.Priority)) {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	if patch.DueDate != nil && patch.DueDate.IsZero() {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	if patch.OwnerUserID != nil {
		valid, err := validRemediationOwner(ctx, s.DB, projectID, *patch.OwnerUserID)
		if err != nil {
			return RemediationAction{}, err
		}
		if !valid {
			return RemediationAction{}, ErrInvalidRemediationOwner
		}
	}
	command, err := s.DB.Exec(ctx, `UPDATE remediation_actions SET
		title=COALESCE($3,title),description=COALESCE($4,description),desired_result=COALESCE($5,desired_result),
		priority=COALESCE($6,priority),owner_user_id=COALESCE($7,owner_user_id),due_date=COALESCE($8,due_date),updated_at=now()
		WHERE project_id=$1 AND id=$2 AND status<>'closed'`, projectID, actionID, patch.Title, patch.Description,
		patch.DesiredResult, patch.Priority, patch.OwnerUserID, patch.DueDate)
	if err != nil {
		return RemediationAction{}, err
	}
	if command.RowsAffected() == 0 {
		return RemediationAction{}, ErrInvalidRemediationTransition
	}
	return s.getRemediationAction(ctx, projectID, actionID)
}

func (s *Store) UpdateRemediationProgress(ctx context.Context, projectID, actionID, actorID, progressNote string) (RemediationAction, error) {
	progressNote = strings.TrimSpace(progressNote)
	if progressNote == "" {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	current, err := s.getRemediationAction(ctx, projectID, actionID)
	if err != nil {
		return RemediationAction{}, err
	}
	if current.Status == string(domain.RemediationClosed) {
		return RemediationAction{}, ErrRemediationClosed
	}
	if current.OwnerUserID != actorID {
		return RemediationAction{}, ErrRemediationForbidden
	}
	command, err := s.DB.Exec(ctx, `UPDATE remediation_actions
		SET progress_note=$4,status='in_progress',review_comment='',updated_at=now()
		WHERE project_id=$1 AND id=$2 AND owner_user_id=$3 AND status IN ('open','in_progress')`,
		projectID, actionID, actorID, progressNote)
	if err != nil {
		return RemediationAction{}, err
	}
	if command.RowsAffected() == 0 {
		return RemediationAction{}, ErrInvalidRemediationTransition
	}
	return s.getRemediationAction(ctx, projectID, actionID)
}

func (s *Store) SubmitRemediationAction(ctx context.Context, projectID, actionID, actorID string) (RemediationAction, error) {
	current, err := s.getRemediationAction(ctx, projectID, actionID)
	if err != nil {
		return RemediationAction{}, err
	}
	if current.Status == string(domain.RemediationClosed) {
		return RemediationAction{}, ErrRemediationClosed
	}
	if current.OwnerUserID != actorID {
		return RemediationAction{}, ErrRemediationForbidden
	}
	if strings.TrimSpace(current.ProgressNote) == "" {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	command, err := s.DB.Exec(ctx, `UPDATE remediation_actions
		SET status='awaiting_review',submitted_at=now(),updated_at=now()
		WHERE project_id=$1 AND id=$2 AND owner_user_id=$3 AND status='in_progress'`, projectID, actionID, actorID)
	if err != nil {
		return RemediationAction{}, err
	}
	if command.RowsAffected() == 0 {
		return RemediationAction{}, ErrInvalidRemediationTransition
	}
	return s.getRemediationAction(ctx, projectID, actionID)
}

func (s *Store) ReviewRemediationAction(ctx context.Context, projectID, actionID, actorID, decision, comment string) (RemediationAction, error) {
	comment = strings.TrimSpace(comment)
	if decision != "return" && decision != "close" {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	if decision == "return" && comment == "" {
		return RemediationAction{}, ErrInvalidRemediationInput
	}
	current, err := s.getRemediationAction(ctx, projectID, actionID)
	if err != nil {
		return RemediationAction{}, err
	}
	if current.Status == string(domain.RemediationClosed) {
		return RemediationAction{}, ErrRemediationClosed
	}

	if decision == "return" {
		result, err := s.DB.Exec(ctx, `UPDATE remediation_actions
			SET status='in_progress',review_comment=$3,updated_at=now()
			WHERE project_id=$1 AND id=$2 AND status='awaiting_review'`, projectID, actionID, comment)
		if err != nil {
			return RemediationAction{}, err
		}
		if result.RowsAffected() == 0 {
			return RemediationAction{}, ErrInvalidRemediationTransition
		}
	} else {
		result, err := s.DB.Exec(ctx, `UPDATE remediation_actions
			SET status='closed',review_comment=$3,closed_by=$4,closed_at=now(),updated_at=now()
			WHERE project_id=$1 AND id=$2 AND status='awaiting_review'`, projectID, actionID, comment, actorID)
		if err != nil {
			return RemediationAction{}, err
		}
		if result.RowsAffected() == 0 {
			return RemediationAction{}, ErrInvalidRemediationTransition
		}
	}
	return s.getRemediationAction(ctx, projectID, actionID)
}

func (s *Store) CreateRemediationEvidence(ctx context.Context, projectID, actionID, actorID, originalName, storagePath, mimeType string, sizeBytes int64) (RemediationEvidence, error) {
	var ownerID, status string
	err := s.DB.QueryRow(ctx, `SELECT owner_user_id,status FROM remediation_actions WHERE project_id=$1 AND id=$2`, projectID, actionID).Scan(&ownerID, &status)
	if err != nil {
		return RemediationEvidence{}, err
	}
	if status == string(domain.RemediationClosed) {
		return RemediationEvidence{}, ErrRemediationClosed
	}
	if ownerID != actorID {
		return RemediationEvidence{}, ErrRemediationForbidden
	}
	if status != string(domain.RemediationOpen) && status != string(domain.RemediationInProgress) {
		return RemediationEvidence{}, ErrInvalidRemediationTransition
	}
	var evidence RemediationEvidence
	err = s.DB.QueryRow(ctx, `INSERT INTO remediation_evidence(action_id,original_name,storage_path,mime_type,size_bytes,uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id,action_id,original_name,storage_path,mime_type,size_bytes,uploaded_by,created_at`,
		actionID, originalName, storagePath, mimeType, sizeBytes, actorID).Scan(remediationEvidenceArgs(&evidence)...)
	return evidence, err
}

func (s *Store) GetRemediationEvidence(ctx context.Context, projectID, actionID, evidenceID string) (RemediationEvidence, error) {
	var evidence RemediationEvidence
	err := s.DB.QueryRow(ctx, `SELECT e.id,e.action_id,e.original_name,e.storage_path,e.mime_type,e.size_bytes,e.uploaded_by,e.created_at
		FROM remediation_evidence e
		JOIN remediation_actions a ON a.id=e.action_id
		WHERE a.project_id=$1 AND a.id=$2 AND e.id=$3`, projectID, actionID, evidenceID).Scan(remediationEvidenceArgs(&evidence)...)
	return evidence, err
}

func (s *Store) DeleteRemediationEvidence(ctx context.Context, projectID, actionID, evidenceID, actorID string) (RemediationEvidence, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return RemediationEvidence{}, err
	}
	defer tx.Rollback(ctx)
	var ownerID, status string
	if err := tx.QueryRow(ctx, `SELECT owner_user_id,status FROM remediation_actions WHERE project_id=$1 AND id=$2 FOR UPDATE`, projectID, actionID).Scan(&ownerID, &status); err != nil {
		return RemediationEvidence{}, err
	}
	if status == string(domain.RemediationClosed) {
		return RemediationEvidence{}, ErrRemediationClosed
	}
	if ownerID != actorID {
		return RemediationEvidence{}, ErrRemediationForbidden
	}
	if status != string(domain.RemediationOpen) && status != string(domain.RemediationInProgress) {
		return RemediationEvidence{}, ErrInvalidRemediationTransition
	}
	var evidence RemediationEvidence
	if err := tx.QueryRow(ctx, `DELETE FROM remediation_evidence WHERE action_id=$1 AND id=$2
		RETURNING id,action_id,original_name,storage_path,mime_type,size_bytes,uploaded_by,created_at`, actionID, evidenceID).Scan(remediationEvidenceArgs(&evidence)...); err != nil {
		return RemediationEvidence{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RemediationEvidence{}, err
	}
	return evidence, nil
}

func (s *Store) listRemediationEvidence(ctx context.Context, actionID string) ([]RemediationEvidence, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,action_id,original_name,storage_path,mime_type,size_bytes,uploaded_by,created_at
		FROM remediation_evidence WHERE action_id=$1 ORDER BY created_at,id`, actionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	evidence := []RemediationEvidence{}
	for rows.Next() {
		var item RemediationEvidence
		if err := rows.Scan(remediationEvidenceArgs(&item)...); err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	return evidence, rows.Err()
}

func remediationEvidenceArgs(evidence *RemediationEvidence) []any {
	return []any{&evidence.ID, &evidence.ActionID, &evidence.OriginalName, &evidence.StoragePath, &evidence.MIMEType, &evidence.SizeBytes, &evidence.UploadedBy, &evidence.CreatedAt}
}

func (s *Store) getRemediationAction(ctx context.Context, projectID, actionID string) (RemediationAction, error) {
	row := s.DB.QueryRow(ctx, `SELECT `+remediationColumns+`
		FROM remediation_actions a
		JOIN project_subcategory_profiles p ON p.project_id=a.project_id AND p.subcategory_id=a.subcategory_id
		JOIN subcategories sc ON sc.id=a.subcategory_id
		JOIN users u ON u.id=a.owner_user_id
		WHERE a.project_id=$1 AND a.id=$2`, projectID, actionID)
	action, err := scanRemediationAction(row)
	if err != nil {
		return RemediationAction{}, err
	}
	action.Evidence, err = s.listRemediationEvidence(ctx, action.ID)
	return action, err
}

type remediationRow interface {
	Scan(dest ...any) error
}

func scanRemediationAction(row remediationRow) (RemediationAction, error) {
	var action RemediationAction
	err := row.Scan(
		&action.ID, &action.ProjectID, &action.SubcategoryID, &action.OutcomeCode, &action.OutcomeDescription,
		&action.CurrentCoverageLevel, &action.TargetCoverageLevel, &action.Title, &action.Description,
		&action.DesiredResult, &action.Priority, &action.OwnerUserID, &action.OwnerName, &action.OwnerEmail,
		&action.DueDate, &action.Status, &action.ProgressNote, &action.ReviewComment, &action.CreatedBy,
		&action.SubmittedAt, &action.ClosedBy, &action.ClosedAt, &action.CreatedAt, &action.UpdatedAt,
	)
	action.Evidence = []RemediationEvidence{}
	return action, err
}

type remediationOwnerQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func validRemediationOwner(ctx context.Context, query remediationOwnerQuerier, projectID, ownerID string) (bool, error) {
	var valid bool
	err := query.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM users u JOIN projects p ON p.organization_id=u.organization_id
		WHERE p.id=$1 AND u.id=$2 AND u.status='active' AND u.role IN ('org_admin','assessor')
	)`, projectID, ownerID).Scan(&valid)
	return valid, err
}
