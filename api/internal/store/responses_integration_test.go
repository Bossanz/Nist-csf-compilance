package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"compliance/api/internal/domain"
)

func TestResponseLifecycleAndDocuments(t *testing.T) {
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
	var counselorID, assessorID, reviewerID, organizationID, projectID, subcategoryID string
	mustScan := func(query string, args []any, destinations ...any) {
		t.Helper()
		if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil {
			t.Fatal(err)
		}
	}
	mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Response Counselor',$1,'counselor','counselor','active') RETURNING id`, []any{"response-counselor-" + suffix + "@example.com"}, &counselorID)
	mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Response Test " + suffix}, &organizationID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Response Assessor',$2,'stakeholder','assessor','active') RETURNING id`, []any{organizationID, "response-assessor-" + suffix + "@example.com"}, &assessorID)
	mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Response Reviewer',$2,'stakeholder','reviewer','active') RETURNING id`, []any{organizationID, "response-reviewer-" + suffix + "@example.com"}, &reviewerID)
	mustScan(`INSERT INTO projects(organization_id,counselor_id,name,slug) VALUES ($1,$2,$3,$4) RETURNING id`, []any{organizationID, counselorID, "Response Test", "response-test-" + suffix}, &projectID)
	mustScan(`SELECT id FROM subcategories ORDER BY code LIMIT 1`, nil, &subcategoryID)
	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, counselorID, assessorID, reviewerID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, organizationID)
		data.Close()
	})

	draft, err := data.SaveResponseDraft(ctx, projectID, subcategoryID, assessorID, "Quarterly access review is performed.")
	if err != nil || draft.Status != string(domain.ResponseDraft) {
		t.Fatalf("save draft: %#v err=%v", draft, err)
	}
	document, err := data.CreateResponseDocument(ctx, projectID, subcategoryID, assessorID, "review.pdf", "opaque-"+suffix, "application/pdf", 12)
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	responses, err := data.ListResponses(ctx, projectID)
	if err != nil || len(responses) != 1 || len(responses[0].Documents) != 1 || responses[0].Documents[0].ID != document.ID {
		t.Fatalf("list responses: %#v err=%v", responses, err)
	}
	if _, err := data.SubmitResponse(ctx, projectID, subcategoryID, assessorID); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := data.SaveResponseDraft(ctx, projectID, subcategoryID, assessorID, "cannot edit after submit"); err != domain.ErrInvalidResponseTransition {
		t.Fatalf("save after submit error: %v", err)
	}
	needsInfo, err := data.ReviewResponse(ctx, projectID, subcategoryID, reviewerID, string(domain.ResponseNeedsMoreInfo), "Please attach the review record.")
	if err != nil || needsInfo.Status != string(domain.ResponseNeedsMoreInfo) {
		t.Fatalf("needs more info: %#v err=%v", needsInfo, err)
	}
	if _, err := data.SaveResponseDraft(ctx, projectID, subcategoryID, assessorID, "Review record attached."); err != nil {
		t.Fatalf("save after needs more info: %v", err)
	}
	if _, err := data.SubmitResponse(ctx, projectID, subcategoryID, assessorID); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if reviewed, err := data.ReviewResponse(ctx, projectID, subcategoryID, reviewerID, string(domain.ResponseReviewed), "Accepted."); err != nil || reviewed.Status != string(domain.ResponseReviewed) {
		t.Fatalf("reviewed: %#v err=%v", reviewed, err)
	}
	deleted, err := data.DeleteResponseDocument(ctx, projectID, subcategoryID, document.ID)
	if err != nil || deleted.StorageKey != "opaque-"+suffix {
		t.Fatalf("delete document: %#v err=%v", deleted, err)
	}
}
