package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRemediationLifecycleAndEligibility(t *testing.T) {
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
	var counselorID, organizationID, assessorID, reviewerID, projectID, subcategoryID string
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

	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Remediation Counselor',$1,'counselor','counselor','active') RETURNING id`, []any{"remediation-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Remediation Test " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Remediation Assessor',$2,'stakeholder','assessor','active') RETURNING id`, []any{organizationID, "remediation-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Remediation Reviewer',$2,'stakeholder','reviewer','active') RETURNING id`, []any{organizationID, "remediation-reviewer-" + suffix + "@example.com"}, &reviewerID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name,status) VALUES ($1,$2,'Remediation Test','in_review') RETURNING id`, []any{organizationID, counselorID}, &projectID)
	mustScan(`SELECT id FROM subcategories ORDER BY code LIMIT 1`, nil, &subcategoryID)
	mustExec(`INSERT INTO project_subcategory_profiles(project_id,subcategory_id,included,current_coverage_level,target_coverage_level,assigned_user_id) VALUES ($1,$2,true,'partial','full',$3)`, projectID, subcategoryID, assessorID)
	mustExec(`INSERT INTO stakeholder_responses(project_id,subcategory_id,response_text,status,responded_by,submitted_at,reviewed_by,reviewed_at) VALUES ($1,$2,'Approved response','reviewed',$3,now(),$3,now())`, projectID, subcategoryID, assessorID)
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, counselorID, assessorID, reviewerID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		data.Close()
	})

	create := RemediationCreate{
		SubcategoryID: subcategoryID,
		Title:         "Implement centralized logging",
		Description:   "Forward application and API logs to the SIEM.",
		DesiredResult: "Security events are searchable and retained.",
		Priority:      "high",
		OwnerUserID:   assessorID,
		DueDate:       time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
	}
	action, err := data.CreateRemediationAction(ctx, projectID, counselorID, create)
	if err != nil || action.Status != "open" || action.OwnerUserID != assessorID || action.OutcomeCode == "" {
		t.Fatalf("unexpected action: %#v err=%v", action, err)
	}

	actions, err := data.ListRemediationActions(ctx, projectID)
	if err != nil || len(actions) != 1 || actions[0].ID != action.ID {
		t.Fatalf("unexpected actions: %#v err=%v", actions, err)
	}
	evidence, err := data.CreateRemediationEvidence(ctx, projectID, action.ID, assessorID, "deployment.pdf", "remediation-"+suffix, "application/pdf", 12)
	if err != nil {
		t.Fatalf("create remediation evidence: %v", err)
	}
	actions, err = data.ListRemediationActions(ctx, projectID)
	if err != nil || len(actions) != 1 || len(actions[0].Evidence) != 1 || actions[0].Evidence[0].ID != evidence.ID {
		t.Fatalf("unexpected action evidence: %#v err=%v", actions, err)
	}
	deletedEvidence, err := data.DeleteRemediationEvidence(ctx, projectID, action.ID, evidence.ID, assessorID)
	if err != nil || deletedEvidence.StoragePath != "remediation-"+suffix {
		t.Fatalf("delete remediation evidence: %#v err=%v", deletedEvidence, err)
	}

	updatedTitle := "Deploy and verify centralized logging"
	updated, err := data.UpdateRemediationAction(ctx, projectID, action.ID, counselorID, RemediationPatch{Title: &updatedTitle})
	if err != nil || updated.Title != updatedTitle {
		t.Fatalf("unexpected update: %#v err=%v", updated, err)
	}

	if _, err := data.UpdateRemediationProgress(ctx, projectID, action.ID, reviewerID, "Not the owner"); !errors.Is(err, ErrRemediationForbidden) {
		t.Fatalf("expected owner guard, got %v", err)
	}
	progressed, err := data.UpdateRemediationProgress(ctx, projectID, action.ID, assessorID, "SIEM forwarding is enabled in staging.")
	if err != nil || progressed.Status != "in_progress" {
		t.Fatalf("unexpected progress: %#v err=%v", progressed, err)
	}

	// Remediation remains mutable after assessment finalization.
	mustExec(`UPDATE projects SET status='closed',finalized_at=now(),finalized_by=$2 WHERE id=$1`, projectID, counselorID)
	submitted, err := data.SubmitRemediationAction(ctx, projectID, action.ID, assessorID)
	if err != nil || submitted.Status != "awaiting_review" || submitted.SubmittedAt == nil {
		t.Fatalf("unexpected submit: %#v err=%v", submitted, err)
	}
	returned, err := data.ReviewRemediationAction(ctx, projectID, action.ID, counselorID, "return", "Attach the production deployment record.")
	if err != nil || returned.Status != "in_progress" || returned.ReviewComment == "" {
		t.Fatalf("unexpected return: %#v err=%v", returned, err)
	}
	if _, err := data.UpdateRemediationProgress(ctx, projectID, action.ID, assessorID, "Production deployment record attached."); err != nil {
		t.Fatal(err)
	}
	if _, err := data.SubmitRemediationAction(ctx, projectID, action.ID, assessorID); err != nil {
		t.Fatal(err)
	}
	closed, err := data.ReviewRemediationAction(ctx, projectID, action.ID, counselorID, "close", "Verified.")
	if err != nil || closed.Status != "closed" || closed.ClosedAt == nil || closed.ClosedBy == nil {
		t.Fatalf("unexpected close: %#v err=%v", closed, err)
	}
	if _, err := data.UpdateRemediationAction(ctx, projectID, action.ID, counselorID, RemediationPatch{Title: &updatedTitle}); !errors.Is(err, ErrRemediationClosed) {
		t.Fatalf("expected closed action guard, got %v", err)
	}
}

func TestCreateRemediationRejectsInvalidOutcomeAndOwner(t *testing.T) {
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
	var counselorID, organizationID, assessorID, reviewerID, projectID, subcategoryID string
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

	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Guard Counselor',$1,'counselor','counselor','active') RETURNING id`, []any{"guard-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Remediation Guard " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Guard Assessor',$2,'stakeholder','assessor','active') RETURNING id`, []any{organizationID, "guard-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Guard Reviewer',$2,'stakeholder','reviewer','active') RETURNING id`, []any{organizationID, "guard-reviewer-" + suffix + "@example.com"}, &reviewerID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name,status) VALUES ($1,$2,'Remediation Guard','in_review') RETURNING id`, []any{organizationID, counselorID}, &projectID)
	mustScan(`SELECT id FROM subcategories ORDER BY code LIMIT 1`, nil, &subcategoryID)
	mustExec(`INSERT INTO project_subcategory_profiles(project_id,subcategory_id,included,current_coverage_level,target_coverage_level,assigned_user_id) VALUES ($1,$2,true,'partial','full',$3)`, projectID, subcategoryID, assessorID)
	mustExec(`INSERT INTO stakeholder_responses(project_id,subcategory_id,response_text,status,responded_by) VALUES ($1,$2,'Draft response','draft',$3)`, projectID, subcategoryID, assessorID)
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, counselorID, assessorID, reviewerID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		data.Close()
	})

	create := RemediationCreate{SubcategoryID: subcategoryID, Title: "Close the gap", Priority: "medium", OwnerUserID: assessorID, DueDate: time.Now().AddDate(0, 1, 0)}
	if _, err := data.CreateRemediationAction(ctx, projectID, counselorID, create); !errors.Is(err, ErrOutcomeNotApproved) {
		t.Fatalf("expected approval guard, got %v", err)
	}
	mustExec(`UPDATE stakeholder_responses SET status='reviewed' WHERE project_id=$1 AND subcategory_id=$2`, projectID, subcategoryID)
	mustExec(`UPDATE project_subcategory_profiles SET current_coverage_level='full',target_coverage_level='full' WHERE project_id=$1 AND subcategory_id=$2`, projectID, subcategoryID)
	if _, err := data.CreateRemediationAction(ctx, projectID, counselorID, create); !errors.Is(err, ErrNoCoverageGap) {
		t.Fatalf("expected gap guard, got %v", err)
	}
	mustExec(`UPDATE project_subcategory_profiles SET current_coverage_level='partial',target_coverage_level='full' WHERE project_id=$1 AND subcategory_id=$2`, projectID, subcategoryID)
	create.OwnerUserID = reviewerID
	if _, err := data.CreateRemediationAction(ctx, projectID, counselorID, create); !errors.Is(err, ErrInvalidRemediationOwner) {
		t.Fatalf("expected owner guard, got %v", err)
	}
}
