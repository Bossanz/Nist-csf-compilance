package store

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrOrganizationExists = errors.New("organization already exists")

func (s *Store) ListOrganizations(ctx context.Context, organizationID *string) ([]Organization, error) {
	query := `SELECT id,name,slug,type FROM organizations`
	args := []any{}
	if organizationID != nil {
		query += ` WHERE id=$1`
		args = append(args, *organizationID)
	}
	query += ` ORDER BY lower(name),id`
	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	organizations := []Organization{}
	for rows.Next() {
		var organization Organization
		if err := rows.Scan(&organization.ID, &organization.Name, &organization.Slug, &organization.Type); err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}
	return organizations, rows.Err()
}

func (s *Store) GetOrganization(ctx context.Context, id string) (Organization, error) {
	var organization Organization
	err := s.DB.QueryRow(ctx, `SELECT id,name,slug,type FROM organizations WHERE id=$1`, id).Scan(&organization.ID, &organization.Name, &organization.Slug, &organization.Type)
	return organization, err
}

func (s *Store) GetOrganizationBySlug(ctx context.Context, slug string) (Organization, error) {
	var organization Organization
	err := s.DB.QueryRow(ctx, `SELECT id,name,slug,type FROM organizations WHERE lower(slug)=lower($1)`, slug).Scan(&organization.ID, &organization.Name, &organization.Slug, &organization.Type)
	return organization, err
}

func (s *Store) CreateOrganization(ctx context.Context, name string) (Organization, error) {
	name = strings.TrimSpace(name)
	var organization Organization
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE lower(btrim(name))=lower($1))`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return ErrOrganizationExists
		}
		slug, err := nextOrganizationSlug(ctx, tx, Slugify(name))
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `INSERT INTO organizations(name,slug,type) VALUES ($1,$2,'client') RETURNING id,name,slug,type`, name, slug).Scan(&organization.ID, &organization.Name, &organization.Slug, &organization.Type)
	})
	return organization, err
}

func (s *Store) DeleteOrganization(ctx context.Context, id string) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var lockedID string
		if err := tx.QueryRow(ctx, `SELECT id FROM organizations WHERE id=$1 FOR UPDATE`, id).Scan(&lockedID); err != nil {
			return err
		}

		statements := []string{
			`DELETE FROM audit_logs WHERE organization_id=$1 OR project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
			`DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE organization_id=$1)`,
			`DELETE FROM invitations WHERE organization_id=$1`,
			`DELETE FROM project_subcategory_profiles WHERE project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
			`DELETE FROM project_functions WHERE project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
			`DELETE FROM projects WHERE organization_id=$1`,
			`DELETE FROM users WHERE organization_id=$1`,
			`DELETE FROM organizations WHERE id=$1`,
		}
		for _, statement := range statements {
			if _, err := tx.Exec(ctx, statement, id); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ListOrganizationEvidenceKeys(ctx context.Context, organizationID string) ([]string, error) {
	rows, err := s.DB.Query(ctx, `SELECT d.storage_key
		FROM response_documents d
		JOIN stakeholder_responses r ON r.id=d.response_id
		JOIN projects p ON p.id=r.project_id
		WHERE p.organization_id=$1`, organizationID)
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

func (s *Store) ListProjectsByOrganization(ctx context.Context, organizationID string) ([]Project, error) {
	rows, err := s.DB.Query(ctx, projectSelect+` WHERE p.organization_id=$1
		ORDER BY p.created_at DESC,p.id DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []Project{}
	for rows.Next() {
		var project Project
		if err := rows.Scan(projectArgs(&project)...); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}
