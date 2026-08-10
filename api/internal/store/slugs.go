package store

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
)

// Slugify creates a stable, URL-safe name while preserving Unicode letters.
func Slugify(value string) string {
	var slug strings.Builder
	pendingSeparator := false
	wroteValue := false

	for _, r := range strings.TrimSpace(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || (unicode.IsMark(r) && wroteValue) {
			if pendingSeparator && wroteValue {
				slug.WriteRune('-')
			}
			slug.WriteRune(unicode.ToLower(r))
			pendingSeparator = false
			wroteValue = true
			continue
		}
		if wroteValue {
			pendingSeparator = true
		}
	}

	if slug.Len() == 0 {
		return "item"
	}
	return slug.String()
}

func nextOrganizationSlug(ctx context.Context, tx pgx.Tx, base string) (string, error) {
	return nextSlug(ctx, tx, base, func(candidate string) (bool, error) {
		var exists bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug=$1)`, candidate).Scan(&exists)
		return exists, err
	})
}

func nextProjectSlug(ctx context.Context, tx pgx.Tx, organizationID, base string) (string, error) {
	return nextSlug(ctx, tx, base, func(candidate string) (bool, error) {
		var exists bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE organization_id=$1 AND slug=$2)`, organizationID, candidate).Scan(&exists)
		return exists, err
	})
}

func nextSlug(ctx context.Context, tx pgx.Tx, base string, exists func(string) (bool, error)) (string, error) {
	for suffix := 1; ; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		used, err := exists(candidate)
		if err != nil {
			return "", err
		}
		if !used {
			return candidate, nil
		}
	}
}

func (s *Store) EnsureSlugs(ctx context.Context) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		statements := []string{
			`ALTER TABLE organizations ADD COLUMN IF NOT EXISTS slug text`,
			`ALTER TABLE projects ADD COLUMN IF NOT EXISTS slug text`,
			`CREATE UNIQUE INDEX IF NOT EXISTS organizations_slug_unique ON organizations(slug) WHERE slug IS NOT NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS projects_organization_slug_unique ON projects(organization_id, slug) WHERE slug IS NOT NULL`,
		}
		for _, statement := range statements {
			if _, err := tx.Exec(ctx, statement); err != nil {
				return err
			}
		}

		type organizationRow struct{ id, name string }
		organizationRows, err := tx.Query(ctx, `SELECT id,name FROM organizations WHERE slug IS NULL OR btrim(slug)='' ORDER BY id`)
		if err != nil {
			return err
		}
		organizations := []organizationRow{}
		for organizationRows.Next() {
			var row organizationRow
			if err := organizationRows.Scan(&row.id, &row.name); err != nil {
				organizationRows.Close()
				return err
			}
			organizations = append(organizations, row)
		}
		if err := organizationRows.Err(); err != nil {
			organizationRows.Close()
			return err
		}
		organizationRows.Close()
		for _, row := range organizations {
			slug, err := nextOrganizationSlug(ctx, tx, Slugify(row.name))
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE organizations SET slug=$2 WHERE id=$1`, row.id, slug); err != nil {
				return err
			}
		}

		type projectRow struct{ id, organizationID, name string }
		projectRows, err := tx.Query(ctx, `SELECT id,organization_id,name FROM projects WHERE slug IS NULL OR btrim(slug)='' ORDER BY organization_id,id`)
		if err != nil {
			return err
		}
		projects := []projectRow{}
		for projectRows.Next() {
			var row projectRow
			if err := projectRows.Scan(&row.id, &row.organizationID, &row.name); err != nil {
				projectRows.Close()
				return err
			}
			projects = append(projects, row)
		}
		if err := projectRows.Err(); err != nil {
			projectRows.Close()
			return err
		}
		projectRows.Close()
		for _, row := range projects {
			slug, err := nextProjectSlug(ctx, tx, row.organizationID, Slugify(row.name))
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE projects SET slug=$2 WHERE id=$1`, row.id, slug); err != nil {
				return err
			}
		}
		return nil
	})
}
