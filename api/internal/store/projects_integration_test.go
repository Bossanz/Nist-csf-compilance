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

func TestUpdateFunctionScopeChangesOnlyTheSelectedFunction(t *testing.T) {
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

	var functionCode, otherFunctionCode string
	if err := data.DB.QueryRow(ctx, "SELECT code FROM functions ORDER BY code LIMIT 1").Scan(&functionCode); err != nil {
		t.Fatal(err)
	}
	if err := data.DB.QueryRow(ctx, "SELECT code FROM functions ORDER BY code OFFSET 1 LIMIT 1").Scan(&otherFunctionCode); err != nil {
		t.Fatal(err)
	}

	rows, err := data.UpdateFunctionScope(ctx, projectID, functionCode, true)
	if err != nil {
		t.Fatalf("update Function scope: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected updated profile rows")
	}

	var selectedIncluded, otherIncluded int
	if err := data.DB.QueryRow(ctx, `SELECT count(*) FROM project_subcategory_profiles p JOIN subcategories sc ON sc.id=p.subcategory_id JOIN categories c ON c.id=sc.category_id JOIN functions f ON f.id=c.function_id WHERE p.project_id=$1 AND f.code=$2 AND p.included`, projectID, functionCode).Scan(&selectedIncluded); err != nil {
		t.Fatal(err)
	}
	if err := data.DB.QueryRow(ctx, `SELECT count(*) FROM project_subcategory_profiles p JOIN subcategories sc ON sc.id=p.subcategory_id JOIN categories c ON c.id=sc.category_id JOIN functions f ON f.id=c.function_id WHERE p.project_id=$1 AND f.code=$2 AND p.included`, projectID, otherFunctionCode).Scan(&otherIncluded); err != nil {
		t.Fatal(err)
	}
	if selectedIncluded != len(rows) || otherIncluded != 0 {
		t.Fatalf("unexpected scope counts: selected=%d rows=%d other=%d", selectedIncluded, len(rows), otherIncluded)
	}
}

func TestUpdateFunctionScopeRejectsUnknownFunctionWithoutChangingProfiles(t *testing.T) {
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

	if _, err := data.UpdateFunctionScope(ctx, projectID, "UNKNOWN", true); err != ErrInvalidFunctionScope {
		t.Fatalf("expected ErrInvalidFunctionScope, got %v", err)
	}
	var includedCount int
	if err := data.DB.QueryRow(ctx, "SELECT count(*) FROM project_subcategory_profiles WHERE project_id=$1 AND included", projectID).Scan(&includedCount); err != nil {
		t.Fatal(err)
	}
	if includedCount != 0 {
		t.Fatalf("unknown Function changed %d profile rows", includedCount)
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
	if _, err := data.DB.Exec(ctx, "INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions", projectID); err != nil {
		t.Fatal(err)
	}
	if _, err := data.DB.Exec(ctx, "INSERT INTO project_subcategory_profiles(project_id,subcategory_id) SELECT $1,id FROM subcategories", projectID); err != nil {
		t.Fatal(err)
	}
	mustScan("SELECT id FROM subcategories ORDER BY code LIMIT 1", nil, &subcategoryID)
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, "DELETE FROM projects WHERE id=$1", projectID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM organizations WHERE id=$1", organizationID)
		_, _ = data.DB.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2)", counselorID, assessorID)
		data.Close()
	})
	return
}
