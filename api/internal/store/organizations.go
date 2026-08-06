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
	command, err := s.DB.Exec(ctx, `DELETE FROM organizations o WHERE o.id=$1
		AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.organization_id=o.id)
		AND NOT EXISTS (SELECT 1 FROM users u WHERE u.organization_id=o.id)`, id)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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
