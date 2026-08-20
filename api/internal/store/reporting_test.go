package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCalculateRemediationSummaryCountsLifecycleAndOverdueActions(t *testing.T) {
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	actions := []RemediationAction{
		{Status: "open", DueDate: time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC)},
		{Status: "in_progress", DueDate: time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC)},
		{Status: "awaiting_review", DueDate: time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)},
		{Status: "closed", DueDate: time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)},
	}

	summary := calculateRemediationSummary(actions, now)

	if summary.OpenCount != 1 || summary.InProgressCount != 1 || summary.AwaitingReviewCount != 1 || summary.ClosedCount != 1 || summary.OverdueCount != 1 {
		t.Fatalf("unexpected remediation summary: %#v", summary)
	}
}

func TestGetFinalReportIncludesIncludedOutcomeAndSummary(t *testing.T) {
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
	if _, err := data.DB.Exec(ctx, `INSERT INTO stakeholder_responses(project_id,subcategory_id,response_text,status,responded_by,submitted_at,reviewed_by,reviewed_at) VALUES ($1,$2,'Quarterly review','reviewed',$3,now(),$3,now())`, projectID, subcategoryID, assessorID); err != nil {
		t.Fatal(err)
	}

	report, err := data.GetFinalReport(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.IncludedCount != 1 || len(report.Outcomes) != 1 || report.Outcomes[0].Response == nil || report.Outcomes[0].Response.Status != "reviewed" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestGetAuditPackageOrdersAuditTrail(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	data, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	projectID, organizationID, counselorID, _, _ := setupScopeSubmissionFixture(t, data)
	if _, err := data.DB.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,organization_id,project_id,action,entity_type) VALUES ($1,$2,$3,'first.event','project'),($1,$2,$3,'second.event','profile')`, counselorID, organizationID, projectID); err != nil {
		t.Fatal(err)
	}

	packageData, err := data.GetAuditPackage(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(packageData.AuditTrail) < 2 || packageData.AuditTrail[0].CreatedAt.After(packageData.AuditTrail[1].CreatedAt) {
		t.Fatalf("audit trail is not ordered: %#v", packageData.AuditTrail)
	}
}
