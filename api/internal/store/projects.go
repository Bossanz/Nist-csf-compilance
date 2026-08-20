package store

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (s *Store) CreateProject(ctx context.Context, organizationID, name string) (Project, error) {
	return s.CreateScopedProject(ctx, organizationID, name)
}

func (s *Store) CreateScopedProject(ctx context.Context, organizationID, name string) (Project, error) {
	return s.CreateScopedProjectWithMetadata(ctx, organizationID, name, ProjectMetadata{})
}

func (s *Store) CreateScopedProjectWithMetadata(ctx context.Context, organizationID, name string, metadata ProjectMetadata) (Project, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Project{}, err
	}
	defer tx.Rollback(ctx)
	var projectID string
	slug, err := nextProjectSlug(ctx, tx, organizationID, Slugify(name))
	if err != nil {
		return Project{}, err
	}
	if err = tx.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,objective,assessment_period,target_completion_date,scope_boundary,compliance_driver) VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::date,$7,$8) RETURNING id`, organizationID, name, slug, metadata.Objective, metadata.AssessmentPeriod, metadata.TargetCompletionDate, metadata.ScopeBoundary, metadata.ComplianceDriver).Scan(&projectID); err != nil {
		return Project{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions`, projectID); err != nil {
		return Project{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO project_subcategory_profiles(project_id,subcategory_id) SELECT $1,id FROM subcategories`, projectID); err != nil {
		return Project{}, err
	}
	var p Project
	err = tx.QueryRow(ctx, projectSelect+` WHERE p.id=$1`, projectID).Scan(projectArgs(&p)...)
	if err != nil {
		return Project{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Project{}, err
	}
	return p, nil
}

func (s *Store) SubmitProjectScope(ctx context.Context, projectID string) (Project, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Project{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM projects WHERE id=$1 FOR UPDATE", projectID).Scan(&status); err != nil {
		return Project{}, err
	}
	if status != "setup" && status != "in_review" {
		return Project{}, ErrInvalidProjectTransition
	}
	if status == "setup" {
		var includedCount, invalidAssignmentCount int
		if err := tx.QueryRow(ctx, "SELECT count(*) FILTER (WHERE p.included), count(*) FILTER (WHERE p.included AND (u.id IS NULL OR u.user_type <> 'stakeholder' OR u.role NOT IN ('org_admin','assessor') OR u.status <> 'active')) FROM project_subcategory_profiles p LEFT JOIN users u ON u.id=p.assigned_user_id WHERE p.project_id=$1", projectID).Scan(&includedCount, &invalidAssignmentCount); err != nil {
			return Project{}, err
		}
		if includedCount == 0 || invalidAssignmentCount > 0 {
			return Project{}, ErrInvalidProjectTransition
		}
		if _, err := tx.Exec(ctx, "UPDATE projects SET status='in_review' WHERE id=$1", projectID); err != nil {
			return Project{}, err
		}
	}

	var project Project
	if err := tx.QueryRow(ctx, projectSelect+" WHERE p.id=$1", projectID).Scan(projectArgs(&project)...); err != nil {
		return Project{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) FinalizeProject(ctx context.Context, projectID, actorID string) (Project, int, int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Project{}, 0, 0, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM projects WHERE id=$1 FOR UPDATE`, projectID).Scan(&status); err != nil {
		return Project{}, 0, 0, err
	}
	if status == "closed" {
		return Project{}, 0, 0, ErrProjectFinalized
	}
	if status != "in_review" {
		return Project{}, 0, 0, ErrProjectNotReady
	}

	var includedCount, approvedCount int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE p.included),
		       count(*) FILTER (WHERE p.included AND r.status='reviewed')
		FROM project_subcategory_profiles p
		LEFT JOIN stakeholder_responses r
		  ON r.project_id=p.project_id AND r.subcategory_id=p.subcategory_id
		WHERE p.project_id=$1`, projectID).Scan(&includedCount, &approvedCount); err != nil {
		return Project{}, 0, 0, err
	}
	if includedCount == 0 || approvedCount != includedCount {
		return Project{}, approvedCount, includedCount, ErrProjectNotReady
	}

	if _, err := tx.Exec(ctx, `UPDATE projects SET status='closed', finalized_at=now(), finalized_by=$2 WHERE id=$1`, projectID, actorID); err != nil {
		return Project{}, 0, 0, err
	}

	var project Project
	if err := tx.QueryRow(ctx, projectSelect+` WHERE p.id=$1`, projectID).Scan(projectArgs(&project)...); err != nil {
		return Project{}, 0, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Project{}, 0, 0, err
	}
	return project, approvedCount, includedCount, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (Project, error) {
	var p Project
	err := s.DB.QueryRow(ctx, projectSelect+` WHERE p.id=$1`, id).Scan(projectArgs(&p)...)
	return p, err
}

func (s *Store) GetProjectBySlug(ctx context.Context, organizationID, slug string) (Project, error) {
	var p Project
	err := s.DB.QueryRow(ctx, projectSelect+` WHERE p.organization_id=$1 AND lower(p.slug)=lower($2)`, organizationID, slug).Scan(projectArgs(&p)...)
	return p, err
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.DB.Query(ctx, projectSelect+` ORDER BY p.created_at DESC,p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(projectArgs(&p)...); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const projectSelect = `SELECT p.id,p.organization_id,o.name,p.name,COALESCE(p.slug,''),p.status,p.created_at::text,p.objective,p.assessment_period,COALESCE(p.target_completion_date::text,''),p.scope_boundary,p.compliance_driver,p.finalized_at,p.finalized_by FROM projects p JOIN organizations o ON o.id=p.organization_id`

func projectArgs(p *Project) []any {
	return []any{&p.ID, &p.OrganizationID, &p.OrganizationName, &p.Name, &p.Slug, &p.Status, &p.CreatedAt, &p.Objective, &p.AssessmentPeriod, &p.TargetCompletionDate, &p.ScopeBoundary, &p.ComplianceDriver, &p.FinalizedAt, &p.FinalizedBy}
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM audit_logs WHERE project_id=$1`, id); err != nil {
			return err
		}
		command, err := tx.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
		if err != nil {
			return err
		}
		if command.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
}

func (s *Store) ListProjectEvidenceKeys(ctx context.Context, projectID string) ([]string, error) {
	rows, err := s.DB.Query(ctx, `SELECT d.storage_key FROM response_documents d JOIN stakeholder_responses r ON r.id=d.response_id WHERE r.project_id=$1
		UNION ALL
		SELECT e.storage_path FROM remediation_evidence e JOIN remediation_actions a ON a.id=e.action_id WHERE a.project_id=$1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *Store) ListProfile(ctx context.Context, projectID string) ([]ProfileRow, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT p.id,p.project_id,p.subcategory_id,f.code,c.code,sc.code,sc.description,p.included,p.rationale,p.current_priority,p.current_coverage_level,p.current_status_text,p.current_policies_text,p.current_tier,p.target_priority,p.target_coverage_level,p.target_approach_text,p.target_tier,p.notes,p.considerations,p.review_status,p.assigned_user_id,COALESCE(assigned_user.name,''),COALESCE(assigned_user.email,'') FROM project_subcategory_profiles p JOIN subcategories sc ON sc.id=p.subcategory_id JOIN categories c ON c.id=sc.category_id JOIN functions f ON f.id=c.function_id LEFT JOIN users assigned_user ON assigned_user.id=p.assigned_user_id WHERE p.project_id=$1 ORDER BY f.code,c.code,sc.code`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProfileRow{}
	for rows.Next() {
		var p ProfileRow
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.SubcategoryID, &p.FunctionCode, &p.CategoryCode, &p.SubcategoryCode, &p.Description, &p.Included, &p.Rationale, &p.CurrentPriority, &p.CurrentCoverageLevel, &p.CurrentStatusText, &p.CurrentPoliciesText, &p.CurrentTier, &p.TargetPriority, &p.TargetCoverageLevel, &p.TargetApproachText, &p.TargetTier, &p.Notes, &p.Considerations, &p.ReviewStatus, &p.AssignedUserID, &p.AssignedUserName, &p.AssignedUserEmail); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpdateProfile(ctx context.Context, projectID, subcategoryID string, patch ProfilePatch) (ProfileRow, error) {
	if _, err := uuid.Parse(projectID); err != nil {
		return ProfileRow{}, fmt.Errorf("invalid project id")
	}
	if _, err := uuid.Parse(subcategoryID); err != nil {
		return ProfileRow{}, fmt.Errorf("invalid subcategory id")
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return ProfileRow{}, err
	}
	defer tx.Rollback(ctx)

	var organizationID string
	var currentIncluded bool
	var currentAssignedUserID *string
	if err := tx.QueryRow(ctx, `SELECT p.included,p.assigned_user_id,pr.organization_id FROM project_subcategory_profiles p JOIN projects pr ON pr.id=p.project_id WHERE p.project_id=$1 AND p.subcategory_id=$2 FOR UPDATE`, projectID, subcategoryID).Scan(&currentIncluded, &currentAssignedUserID, &organizationID); err != nil {
		return ProfileRow{}, err
	}
	finalIncluded := currentIncluded
	if patch.Included != nil {
		finalIncluded = *patch.Included
	}
	finalAssignedUserID := currentAssignedUserID
	if patch.AssignedUserID != nil {
		finalAssignedUserID = patch.AssignedUserID
	}
	if !finalIncluded {
		finalAssignedUserID = nil
	}
	if finalAssignedUserID != nil {
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND organization_id=$2 AND user_type='stakeholder' AND role IN ('org_admin','assessor') AND status='active')`, *finalAssignedUserID, organizationID).Scan(&valid); err != nil {
			return ProfileRow{}, err
		}
		if !valid {
			return ProfileRow{}, ErrInvalidProfileAssignment
		}
	}

	sets, args := []string{}, []any{}
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)))
	}
	if patch.Included != nil {
		add("included", *patch.Included)
	}
	if patch.Rationale != nil {
		add("rationale", *patch.Rationale)
	}
	if patch.CurrentPriority != nil {
		add("current_priority", *patch.CurrentPriority)
	}
	if patch.CurrentCoverageLevel != nil {
		add("current_coverage_level", *patch.CurrentCoverageLevel)
	}
	if patch.CurrentStatusText != nil {
		add("current_status_text", *patch.CurrentStatusText)
	}
	if patch.CurrentPoliciesText != nil {
		add("current_policies_text", *patch.CurrentPoliciesText)
	}
	if patch.TargetPriority != nil {
		add("target_priority", *patch.TargetPriority)
	}
	if patch.TargetCoverageLevel != nil {
		add("target_coverage_level", *patch.TargetCoverageLevel)
	}
	if patch.TargetApproachText != nil {
		add("target_approach_text", *patch.TargetApproachText)
	}
	if patch.Notes != nil {
		add("notes", *patch.Notes)
	}
	if patch.Considerations != nil {
		add("considerations", *patch.Considerations)
	}
	if len(sets) == 0 && patch.AssignedUserID == nil {
		return ProfileRow{}, fmt.Errorf("no fields to update")
	}
	add("assigned_user_id", finalAssignedUserID)
	args = append(args, projectID, subcategoryID)
	_, err = tx.Exec(ctx, `UPDATE project_subcategory_profiles SET `+strings.Join(sets, ",")+` WHERE project_id=$`+fmt.Sprint(len(args)-1)+` AND subcategory_id=$`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return ProfileRow{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProfileRow{}, err
	}
	rows, err := s.ListProfile(ctx, projectID)
	if err != nil {
		return ProfileRow{}, err
	}
	for _, row := range rows {
		if row.SubcategoryID == subcategoryID {
			return row, nil
		}
	}
	return ProfileRow{}, fmt.Errorf("profile not found")
}
