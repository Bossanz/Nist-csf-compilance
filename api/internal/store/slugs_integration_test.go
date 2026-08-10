package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSlugPersistenceAndLookup(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	if err := data.EnsureSlugs(ctx); err != nil {
		t.Fatal(err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	organization, err := data.CreateOrganization(ctx, "Slug Test Organization "+suffix)
	if err != nil {
		t.Fatal(err)
	}
	project, err := data.CreateScopedProject(ctx, organization.ID, "Readiness")
	if err != nil {
		t.Fatal(err)
	}
	secondProject, err := data.CreateScopedProject(ctx, organization.ID, "Readiness")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE organization_id=$1`, organization.ID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organization.ID)
	})

	wantOrganizationSlug := "slug-test-organization-" + suffix
	if organization.Slug != wantOrganizationSlug {
		t.Fatalf("organization slug = %q, want %q", organization.Slug, wantOrganizationSlug)
	}
	if project.Slug != "readiness" {
		t.Fatalf("project slug = %q, want readiness", project.Slug)
	}
	if secondProject.Slug != "readiness-2" {
		t.Fatalf("second project slug = %q, want readiness-2", secondProject.Slug)
	}

	loadedOrganization, err := data.GetOrganizationBySlug(ctx, organization.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if loadedOrganization.ID != organization.ID {
		t.Fatalf("organization lookup returned %q, want %q", loadedOrganization.ID, organization.ID)
	}
	loadedProject, err := data.GetProjectBySlug(ctx, organization.ID, secondProject.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if loadedProject.ID != secondProject.ID {
		t.Fatalf("project lookup returned %q, want %q", loadedProject.ID, secondProject.ID)
	}
}
