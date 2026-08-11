package store

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestCreateScopedProjectPersistsMetadata(t *testing.T) {
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

	suffix := uuid.NewString()
	organization, err := data.CreateOrganization(ctx, "Project Metadata Test "+suffix)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM projects WHERE organization_id=$1", organization.ID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organization.ID)
	})

	metadata := ProjectMetadata{
		Objective:            "Prepare for RU registration",
		AssessmentPeriod:     "2026",
		TargetCompletionDate: "2026-09-30",
		ScopeBoundary:        "RU registration systems and supporting processes",
		ComplianceDriver:     "Regulatory requirement",
	}
	project, err := data.CreateScopedProjectWithMetadata(ctx, organization.ID, "RU Registration", metadata)
	if err != nil {
		t.Fatal(err)
	}

	if project.Objective != metadata.Objective || project.AssessmentPeriod != metadata.AssessmentPeriod || project.TargetCompletionDate != metadata.TargetCompletionDate || project.ScopeBoundary != metadata.ScopeBoundary || project.ComplianceDriver != metadata.ComplianceDriver {
		t.Fatalf("metadata did not round-trip: got %#v", project)
	}
}
