package store

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrOrganizationExists = errors.New("organization already exists")

func (s *Store) ListOrganizations(ctx context.Context, organizationID *string) ([]Organization, error) {
	query := `SELECT id,name,type FROM organizations`
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
		if err := rows.Scan(&organization.ID, &organization.Name, &organization.Type); err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}
	return organizations, rows.Err()
}

func (s *Store) GetOrganization(ctx context.Context, id string) (Organization, error) {
	var organization Organization
	err := s.DB.QueryRow(ctx, `SELECT id,name,type FROM organizations WHERE id=$1`, id).Scan(&organization.ID, &organization.Name, &organization.Type)
	return organization, err
}

func (s *Store) CreateOrganization(ctx context.Context, name string) (Organization, error) {
	name = strings.TrimSpace(name)
	var exists bool
	if err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE lower(btrim(name))=lower($1))`, name).Scan(&exists); err != nil {
		return Organization{}, err
	}
	if exists {
		return Organization{}, ErrOrganizationExists
	}
	var organization Organization
	err := s.DB.QueryRow(ctx, `INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id,name,type`, name).Scan(&organization.ID, &organization.Name, &organization.Type)
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

func (s *Store) ListProjectsByOrganization(ctx context.Context, organizationID string) ([]Project, error) {
	rows, err := s.DB.Query(ctx, `SELECT p.id,p.organization_id,o.name,p.name,p.status,p.created_at::text
		FROM projects p JOIN organizations o ON o.id=p.organization_id WHERE p.organization_id=$1
		ORDER BY p.created_at DESC,p.id DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []Project{}
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.OrganizationID, &project.OrganizationName, &project.Name, &project.Status, &project.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}
