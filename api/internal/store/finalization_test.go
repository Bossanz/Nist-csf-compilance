package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestFinalizeProjectRequiresEveryIncludedOutcomeToBeApproved(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	var counselorID, organizationID, assessorID, projectID, subcategoryID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Finalization Counselor',$1,'counselor','counselor','active') RETURNING id`, []any{"finalization-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Finalization Test " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Finalization Assessor',$2,'stakeholder','assessor','active') RETURNING id`, []any{organizationID, "finalization-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name,status) VALUES ($1,$2,'Finalization Test','in_review') RETURNING id`, []any{organizationID, counselorID}, &projectID)
	mustScan(`SELECT id FROM subcategories ORDER BY code LIMIT 1`, nil, &subcategoryID)
	if _, err := data.DB.Exec(ctx, `INSERT INTO project_subcategory_profiles(project_id,subcategory_id,included,assigned_user_id) VALUES ($1,$2,true,$3)`, projectID, subcategoryID, assessorID); err != nil {
		t.Fatal(err)
	}
	if _, err := data.DB.Exec(ctx, `INSERT INTO stakeholder_responses(project_id,subcategory_id,response_text,status,responded_by,submitted_at,reviewed_by,reviewed_at) VALUES ($1,$2,'Approved response','reviewed',$3,now(),$3,now())`, projectID, subcategoryID, assessorID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, counselorID, assessorID)
		data.Close()
	})

	project, approved, included, err := data.FinalizeProject(ctx, projectID, counselorID)
	if err != nil {
		t.Fatal(err)
	}
	if project.Status != "closed" || project.FinalizedBy == nil || project.FinalizedAt == nil {
		t.Fatalf("expected finalized project, got %#v", project)
	}
	if approved != 1 || included != 1 {
		t.Fatalf("expected one approved included outcome, got approved=%d included=%d", approved, included)
	}
}
