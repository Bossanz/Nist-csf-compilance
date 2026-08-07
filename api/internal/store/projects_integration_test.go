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
