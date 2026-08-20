package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestInvitationLifecyclePersistsProjectAccessAndInvalidatesTokens(t *testing.T) {
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
	var counselorID, organizationID, orgAdminID, projectOneID, projectTwoID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Lifecycle Counselor',$1,'counselor','counselor_admin','active') RETURNING id`, []any{"lifecycle-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Lifecycle Org " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Lifecycle Admin',$2,'stakeholder','org_admin','active') RETURNING id`, []any{organizationID, "lifecycle-admin-" + suffix + "@example.com"}, &orgAdminID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,$3) RETURNING id`, []any{organizationID, counselorID, "Lifecycle Project One " + suffix}, &projectOneID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,$3) RETURNING id`, []any{organizationID, counselorID, "Lifecycle Project Two " + suffix}, &projectTwoID)

	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM invitations WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE organization_id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, counselorID)
	})

	first, err := data.CreateInvitation(ctx, Invitation{
		OrganizationID: &organizationID,
		Email:          "auditor-one-" + suffix + "@example.com",
		UserType:       "stakeholder",
		Role:           "auditor",
		TokenHash:      "lifecycle-token-one-" + suffix,
		InvitedBy:      orgAdminID,
		ExpiresAt:      time.Now().Add(time.Hour),
		ProjectIDs:     []string{projectOneID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != "pending" || len(first.ProjectIDs) != 1 || first.ProjectIDs[0] != projectOneID {
		t.Fatalf("unexpected first invitation: %#v", first)
	}

	acceptedUser, err := data.AcceptInvitation(ctx, first.TokenHash, "Auditor One", "password-hash", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var grantCount int
	mustScan(`SELECT count(*) FROM project_auditor_access WHERE user_id=$1 AND project_id=$2 AND revoked_at IS NULL`, []any{acceptedUser.ID, projectOneID}, &grantCount)
	if grantCount != 1 {
		t.Fatalf("expected accepted Auditor project grant, got %d", grantCount)
	}

	second, err := data.CreateInvitation(ctx, Invitation{
		OrganizationID: &organizationID,
		Email:          "auditor-two-" + suffix + "@example.com",
		UserType:       "stakeholder",
		Role:           "auditor",
		TokenHash:      "lifecycle-token-two-" + suffix,
		InvitedBy:      orgAdminID,
		ExpiresAt:      time.Now().Add(time.Hour),
		ProjectIDs:     []string{projectOneID, projectTwoID},
	})
	if err != nil {
		t.Fatal(err)
	}
	resendNow := time.Now()
	replacement, err := data.ResendInvitation(ctx, organizationID, second.ID, "lifecycle-token-two-replacement-"+suffix, orgAdminID, resendNow.Add(2*time.Hour), resendNow)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Status != "pending" || len(replacement.ProjectIDs) != 2 {
		t.Fatalf("unexpected replacement invitation: %#v", replacement)
	}
	var oldStatus, replacementAccessCount string
	mustScan(`SELECT CASE WHEN superseded_at IS NOT NULL THEN 'superseded' ELSE 'other' END FROM invitations WHERE id=$1`, []any{second.ID}, &oldStatus)
	mustScan(`SELECT count(*)::text FROM invitation_project_access WHERE invitation_id=$1`, []any{replacement.ID}, &replacementAccessCount)
	if oldStatus != "superseded" || replacementAccessCount != "2" {
		t.Fatalf("replacement did not preserve lifecycle/access: old=%s access=%s", oldStatus, replacementAccessCount)
	}
	if _, err := data.AcceptInvitation(ctx, second.TokenHash, "Should Not Work", "password-hash", time.Now()); !errors.Is(err, ErrInvalidInvitation) {
		t.Fatalf("superseded token accepted: %v", err)
	}

	third, err := data.CreateInvitation(ctx, Invitation{
		OrganizationID: &organizationID,
		Email:          "auditor-three-" + suffix + "@example.com",
		UserType:       "stakeholder",
		Role:           "auditor",
		TokenHash:      "lifecycle-token-three-" + suffix,
		InvitedBy:      orgAdminID,
		ExpiresAt:      time.Now().Add(time.Hour),
		ProjectIDs:     []string{projectTwoID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := data.CancelInvitation(ctx, organizationID, third.ID, orgAdminID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := data.AcceptInvitation(ctx, third.TokenHash, "Should Not Work", "password-hash", time.Now()); !errors.Is(err, ErrInvalidInvitation) {
		t.Fatalf("cancelled token accepted: %v", err)
	}
	var cancelledStatus string
	mustScan(`SELECT CASE WHEN cancelled_at IS NOT NULL THEN 'cancelled' ELSE 'other' END FROM invitations WHERE id=$1`, []any{third.ID}, &cancelledStatus)
	if cancelledStatus != "cancelled" {
		t.Fatalf("cancelled invitation status = %s", cancelledStatus)
	}
}
