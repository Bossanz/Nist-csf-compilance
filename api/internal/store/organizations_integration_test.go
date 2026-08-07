package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDeleteOrganizationRemovesOwnedData(t *testing.T) {
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

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	var counselorID, organizationID, stakeholderID, projectID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := data.DB.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}

	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Test Counselor',$1,'counselor','counselor_admin','active') RETURNING id`, []any{"counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Delete Test " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Stakeholder',$2,'stakeholder','viewer','active') RETURNING id`, []any{organizationID, "stakeholder-" + suffix + "@example.com"}, &stakeholderID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,'Delete Test') RETURNING id`, []any{organizationID, counselorID}, &projectID)
	mustExec(`INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions LIMIT 1`, projectID)
	mustExec(`INSERT INTO project_subcategory_profiles(project_id,subcategory_id,reviewed_by) SELECT $1,id,$2 FROM subcategories LIMIT 1`, projectID, stakeholderID)
	mustExec(`INSERT INTO sessions(user_id,token_hash,expires_at) VALUES ($1,$2,now()+interval '1 hour')`, stakeholderID, "session-"+suffix)
	mustExec(`INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at) VALUES ($1,$2,'stakeholder','viewer',$3,$4,now()+interval '1 day')`, organizationID, "invite-"+suffix+"@example.com", "invite-"+suffix, counselorID)
	mustExec(`INSERT INTO audit_logs(actor_user_id,organization_id,project_id,action,entity_type) VALUES ($1,$2,$3,'test','organization')`, counselorID, organizationID, projectID)

	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM audit_logs WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, stakeholderID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM invitations WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM project_subcategory_profiles WHERE project_id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM project_functions WHERE project_id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, stakeholderID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, counselorID)
	})

	if err := data.DeleteOrganization(ctx, organizationID); err != nil {
		t.Fatalf("delete organization with owned data: %v", err)
	}

	checks := []struct {
		name  string
		query string
		id    string
	}{
		{"organization", `SELECT count(*) FROM organizations WHERE id=$1`, organizationID},
		{"project", `SELECT count(*) FROM projects WHERE id=$1`, projectID},
		{"stakeholder", `SELECT count(*) FROM users WHERE id=$1`, stakeholderID},
		{"session", `SELECT count(*) FROM sessions WHERE user_id=$1`, stakeholderID},
		{"invitation", `SELECT count(*) FROM invitations WHERE organization_id=$1`, organizationID},
		{"profile", `SELECT count(*) FROM project_subcategory_profiles WHERE project_id=$1`, projectID},
		{"project function", `SELECT count(*) FROM project_functions WHERE project_id=$1`, projectID},
		{"audit log", `SELECT count(*) FROM audit_logs WHERE organization_id=$1`, organizationID},
	}
	for _, check := range checks {
		var count int
		if err := data.DB.QueryRow(ctx, check.query, check.id).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s remains: count=%d err=%v", check.name, count, err)
		}
	}
	var counselorCount int
	if err := data.DB.QueryRow(ctx, `SELECT count(*) FROM users WHERE id=$1`, counselorID).Scan(&counselorCount); err != nil || counselorCount != 1 {
		t.Fatalf("counselor should remain: count=%d err=%v", counselorCount, err)
	}
}
