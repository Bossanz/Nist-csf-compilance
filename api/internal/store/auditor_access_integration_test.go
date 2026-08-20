package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestAuditorAccessSchemaSupportsRoleAndProjectGrant(t *testing.T) {
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
	var counselorID, organizationID, auditorID, projectID, invitationID string
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

	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Schema Counselor',$1,'counselor','counselor_admin','active') RETURNING id`, []any{"schema-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Auditor Schema " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Schema Auditor',$2,'stakeholder','auditor','active') RETURNING id`, []any{organizationID, "schema-auditor-" + suffix + "@example.com"}, &auditorID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,$3) RETURNING id`, []any{organizationID, counselorID, "Auditor Schema Project " + suffix}, &projectID)
	mustScan(`INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at) VALUES ($1,$2,'stakeholder','auditor',$3,$4,now()+interval '1 day') RETURNING id`, []any{organizationID, "schema-invite-" + suffix + "@example.com", "schema-token-" + suffix, counselorID}, &invitationID)
	mustExec(`INSERT INTO invitation_project_access(invitation_id,project_id) VALUES ($1,$2)`, invitationID, projectID)
	mustExec(`INSERT INTO project_auditor_access(project_id,user_id,granted_by) VALUES ($1,$2,$3)`, projectID, auditorID, counselorID)
	mustExec(`INSERT INTO audit_logs(actor_user_id,organization_id,project_id,actor_role,result,action,entity_type,metadata) VALUES ($1,$2,$3,'auditor','success','schema.test','project','{"safe":true}')`, auditorID, organizationID, projectID)

	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM audit_logs WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM invitations WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, auditorID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, counselorID)
	})

	var grantCount, auditCount int
	mustScan(`SELECT count(*) FROM project_auditor_access WHERE project_id=$1 AND user_id=$2 AND revoked_at IS NULL`, []any{projectID, auditorID}, &grantCount)
	mustScan(`SELECT count(*) FROM audit_logs WHERE project_id=$1 AND actor_role='auditor' AND result='success'`, []any{projectID}, &auditCount)
	if grantCount != 1 || auditCount != 1 {
		t.Fatalf("unexpected schema rows: grants=%d audit=%d", grantCount, auditCount)
	}
	allowed, err := data.HasActiveProjectAuditorAccess(ctx, projectID, auditorID)
	if err != nil || !allowed {
		t.Fatalf("expected active Auditor access, allowed=%v err=%v", allowed, err)
	}
	events, err := data.ListProjectAuditEvents(ctx, projectID)
	if err != nil || len(events) != 1 || events[0].ActorRole != "auditor" || events[0].Result != "success" || events[0].ProjectID == nil || *events[0].ProjectID != projectID {
		t.Fatalf("unexpected project audit event: events=%#v err=%v", events, err)
	}
	organizationEvents, err := data.ListOrganizationAuditEvents(ctx, organizationID, auditorID)
	if err != nil || len(organizationEvents) != 1 || organizationEvents[0].ID != events[0].ID {
		t.Fatalf("unexpected scoped organization audit events: events=%#v err=%v", organizationEvents, err)
	}
	mustExec(`UPDATE project_auditor_access SET revoked_at=now() WHERE project_id=$1 AND user_id=$2`, projectID, auditorID)
	allowed, err = data.HasActiveProjectAuditorAccess(ctx, projectID, auditorID)
	if err != nil || allowed {
		t.Fatalf("expected revoked Auditor access to be denied, allowed=%v err=%v", allowed, err)
	}
}
