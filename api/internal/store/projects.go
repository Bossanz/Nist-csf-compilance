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
	if err = tx.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug) VALUES ($1,$2,$3) RETURNING id`, organizationID, name, slug).Scan(&projectID); err != nil {
		return Project{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions`, projectID); err != nil {
		return Project{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO project_subcategory_profiles(project_id,subcategory_id) SELECT $1,id FROM subcategories`, projectID); err != nil {
		return Project{}, err
	}
	var p Project
	err = tx.QueryRow(ctx, `SELECT p.id,p.organization_id,o.name,p.name,p.slug,p.status,p.created_at::text FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.id=$1`, projectID).Scan(&p.ID, &p.OrganizationID, &p.OrganizationName, &p.Name, &p.Slug, &p.Status, &p.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Project{}, err
	}
	return p, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (Project, error) {
	var p Project
	err := s.DB.QueryRow(ctx, `SELECT p.id,p.organization_id,o.name,p.name,p.slug,p.status,p.created_at::text FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.id=$1`, id).Scan(&p.ID, &p.OrganizationID, &p.OrganizationName, &p.Name, &p.Slug, &p.Status, &p.CreatedAt)
	return p, err
}

func (s *Store) GetProjectBySlug(ctx context.Context, organizationID, slug string) (Project, error) {
	var p Project
	err := s.DB.QueryRow(ctx, `SELECT p.id,p.organization_id,o.name,p.name,p.slug,p.status,p.created_at::text FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.organization_id=$1 AND lower(p.slug)=lower($2)`, organizationID, slug).Scan(&p.ID, &p.OrganizationID, &p.OrganizationName, &p.Name, &p.Slug, &p.Status, &p.CreatedAt)
	return p, err
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.DB.Query(ctx, `SELECT p.id,p.organization_id,o.name,p.name,p.slug,p.status,p.created_at::text FROM projects p JOIN organizations o ON o.id=p.organization_id ORDER BY p.created_at DESC,p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.OrganizationName, &p.Name, &p.Slug, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
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
	rows, err := s.DB.Query(ctx, `SELECT d.storage_key FROM response_documents d JOIN stakeholder_responses r ON r.id=d.response_id WHERE r.project_id=$1`, projectID)
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
	rows, err := s.DB.Query(ctx, `SELECT p.id,p.project_id,p.subcategory_id,f.code,c.code,sc.code,sc.description,p.included,p.rationale,p.current_priority,p.current_coverage_level,p.current_status_text,p.current_policies_text,p.current_tier,p.target_priority,p.target_coverage_level,p.target_approach_text,p.target_tier,p.notes,p.considerations,p.review_status FROM project_subcategory_profiles p JOIN subcategories sc ON sc.id=p.subcategory_id JOIN categories c ON c.id=sc.category_id JOIN functions f ON f.id=c.function_id WHERE p.project_id=$1 ORDER BY f.code,c.code,sc.code`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProfileRow{}
	for rows.Next() {
		var p ProfileRow
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.SubcategoryID, &p.FunctionCode, &p.CategoryCode, &p.SubcategoryCode, &p.Description, &p.Included, &p.Rationale, &p.CurrentPriority, &p.CurrentCoverageLevel, &p.CurrentStatusText, &p.CurrentPoliciesText, &p.CurrentTier, &p.TargetPriority, &p.TargetCoverageLevel, &p.TargetApproachText, &p.TargetTier, &p.Notes, &p.Considerations, &p.ReviewStatus); err != nil {
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
	if len(sets) == 0 {
		return ProfileRow{}, fmt.Errorf("no fields to update")
	}
	args = append(args, projectID, subcategoryID)
	_, err := s.DB.Exec(ctx, `UPDATE project_subcategory_profiles SET `+strings.Join(sets, ",")+` WHERE project_id=$`+fmt.Sprint(len(args)-1)+` AND subcategory_id=$`+fmt.Sprint(len(args)), args...)
	if err != nil {
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
