package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDeleteProjectRemovesProjectAuditLogs(t *testing.T) {
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
	var counselorID, organizationID, projectID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Delete Project Counselor',$1,'counselor','counselor','active') RETURNING id`, []any{"delete-project-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Delete Project Test " + suffix}, &organizationID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,'Delete Project Test') RETURNING id`, []any{organizationID, counselorID}, &projectID)
	if _, err := data.DB.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,organization_id,project_id,action,entity_type) VALUES ($1,$2,$3,'test','project')`, counselorID, organizationID, projectID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM audit_logs WHERE project_id=$1 OR organization_id=$2`, projectID, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, counselorID)
		data.Close()
	})

	if err := data.DeleteProject(ctx, projectID); err != nil {
		t.Fatalf("delete project with audit logs: %v", err)
	}
	var auditCount int
	if err := data.DB.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE project_id=$1`, projectID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 {
		t.Fatalf("project audit logs remain: %d", auditCount)
	}
}

func TestSubmitProjectScopeMovesSetupProjectToInReview(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	projectID, _, _, assessorID, subcategoryID := setupScopeSubmissionFixture(t, data)
	included := true
	if _, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{Included: &included}); err != nil {
		t.Fatal(err)
	}
	if _, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{AssignedUserID: &assessorID}); err != nil {
		t.Fatal(err)
	}

	project, err := data.SubmitProjectScope(ctx, projectID)
	if err != nil {
		t.Fatalf("submit scope: %v", err)
	}
	if project.Status != "in_review" {
		t.Fatalf("expected in_review status, got %q", project.Status)
	}
}

func TestSubmitProjectScopeRejectsProjectWithoutAssignedIncludedOutcomes(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	projectID, _, _, _, _ := setupScopeSubmissionFixture(t, data)

	if _, err := data.SubmitProjectScope(ctx, projectID); err != ErrInvalidProjectTransition {
		t.Fatalf("expected invalid project transition, got %v", err)
	}
}

func setupScopeSubmissionFixture(t *testing.T, data *Store) (projectID, organizationID, counselorID, assessorID, subcategoryID string) {
	t.Helper()
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan("INSERT INTO users(name,email,user_type,role,status) VALUES ('Scope Counselor',$1,'counselor','counselor','active') RETURNING id",
		[]any{"scope-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan("INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id",
		[]any{"Scope Submission Test " + suffix}, &organizationID)
	mustScan("INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Scope Assessor',$2,'stakeholder','assessor','active') RETURNING id",
		[]any{organizationID, "scope-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan("INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,'Scope Submission Test Project') RETURNING id",
		[]any{organizationID, counselorID}, &projectID)
	mustScan("SELECT id FROM subcategories ORDER BY code LIMIT 1", nil, &subcategoryID)
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM projects WHERE id=$1", projectID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organizationID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2)", counselorID, assessorID)
		data.Close()
	})
	return
}
