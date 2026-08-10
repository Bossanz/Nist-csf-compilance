package store

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestProfileAssignmentLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	suffix := uuid.NewString()
	var counselorID, organizationID, assessorID, disabledAssessorID, reviewerID, projectID, subcategoryID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}

	mustScan("INSERT INTO users(name,email,user_type,role,status) VALUES ('Assignment Counselor',$1,'counselor','counselor','active') RETURNING id",
		[]any{"assignment-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan("INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id",
		[]any{"Assignment Test " + suffix}, &organizationID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Assignment Assessor',$2,'stakeholder','assessor','active') RETURNING id",
		[]any{organizationID, "assignment-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Disabled Assessor',$2,'stakeholder','assessor','disabled') RETURNING id",
		[]any{organizationID, "assignment-disabled-" + suffix + "@example.com"}, &disabledAssessorID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Assignment Reviewer',$2,'stakeholder','reviewer','active') RETURNING id",
		[]any{organizationID, "assignment-reviewer-" + suffix + "@example.com"}, &reviewerID)
	mustScan("INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,$3) RETURNING id",
		[]any{organizationID, counselorID, "Assignment Test Project"}, &projectID)
	mustScan("SELECT id FROM subcategories ORDER BY code LIMIT 1", nil, &subcategoryID)
	if _, err := data.DB.Exec(ctx,
		"INSERT INTO project_subcategory_profiles(project_id,subcategory_id) VALUES ($1,$2)",
		projectID, subcategoryID); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM projects WHERE id=$1", projectID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2,$3,$4)", counselorID, assessorID, disabledAssessorID, reviewerID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organizationID)
		data.Close()
	})

	assigned := assessorID
	included := true
	row, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{
		Included:       &included,
		AssignedUserID: &assigned,
	})
	if err != nil || row.AssignedUserID == nil || *row.AssignedUserID != assessorID || row.AssignedUserName != "Assignment Assessor" {
		t.Fatalf("assigned profile: %#v err=%v", row, err)
	}

	disabled := disabledAssessorID
	if _, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{AssignedUserID: &disabled}); err != ErrInvalidProfileAssignment {
		t.Fatalf("disabled assignee error: %v", err)
	}

	excluded := false
	cleared, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{Included: &excluded})
	if err != nil || cleared.AssignedUserID != nil {
		t.Fatalf("cleared assignment: %#v err=%v", cleared, err)
	}
}
