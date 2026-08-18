package store

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNotificationRecipientQueriesRespectOrganizationAndStatus(t *testing.T) {
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
	var organizationID, otherOrganizationID, projectID, subcategoryID, assessorID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan("INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id", []any{"Notification Test " + suffix}, &organizationID)
	mustScan("INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id", []any{"Other Notification Test " + suffix}, &otherOrganizationID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Active Reviewer',$2,'stakeholder','reviewer','active') RETURNING id", []any{organizationID, "reviewer-" + suffix + "@example.com"}, new(string))
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Disabled Reviewer',$2,'stakeholder','reviewer','disabled') RETURNING id", []any{organizationID, "disabled-reviewer-" + suffix + "@example.com"}, new(string))
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Organization Admin',$2,'stakeholder','org_admin','active') RETURNING id", []any{organizationID, "admin-" + suffix + "@example.com"}, new(string))
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Assigned Assessor',$2,'stakeholder','assessor','active') RETURNING id", []any{organizationID, "assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Other Reviewer',$2,'stakeholder','reviewer','active') RETURNING id", []any{otherOrganizationID, "other-reviewer-" + suffix + "@example.com"}, new(string))
	mustScan("INSERT INTO projects(organization_id,name,slug) VALUES ($1,$2,$3) RETURNING id", []any{organizationID, "Notification Project", "notification-" + suffix}, &projectID)
	mustScan("SELECT id FROM subcategories ORDER BY code LIMIT 1", nil, &subcategoryID)
	if _, err := data.DB.Exec(ctx, "INSERT INTO project_subcategory_profiles(project_id,subcategory_id,included,assigned_user_id) VALUES ($1,$2,true,$3)", projectID, subcategoryID, assessorID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM projects WHERE id=$1", projectID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM users WHERE organization_id IN ($1,$2)", organizationID, otherOrganizationID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id IN ($1,$2)", organizationID, otherOrganizationID)
		data.Close()
	})

	reviewers, err := data.ListProjectReviewerEmails(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if expected := []string{"reviewer-" + suffix + "@example.com"}; !reflect.DeepEqual(reviewers, expected) {
		t.Fatalf("project reviewers: got %#v want %#v", reviewers, expected)
	}

	assessorEmail, err := data.GetAssignedAssessorEmail(ctx, projectID, subcategoryID)
	if err != nil || assessorEmail != "assessor-"+suffix+"@example.com" {
		t.Fatalf("assigned assessor: got %q err=%v", assessorEmail, err)
	}

	organizationRecipients, err := data.ListOrganizationEmailsByRoles(ctx, organizationID, []string{"org_admin", "reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	expectedOrganizationRecipients := []string{"admin-" + suffix + "@example.com", "reviewer-" + suffix + "@example.com"}
	if !reflect.DeepEqual(organizationRecipients, expectedOrganizationRecipients) {
		t.Fatalf("organization recipients: got %#v want %#v", organizationRecipients, expectedOrganizationRecipients)
	}
}
