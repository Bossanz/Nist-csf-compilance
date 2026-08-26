package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

type projectVersionFixture struct {
	data            *Store
	ctx             context.Context
	organizationID  string
	counselorID     string
	assessorID      string
	sourceProjectID string
	subcategoryID   string
}

func newProjectVersionFixture(t *testing.T, finalized bool) projectVersionFixture {
	t.Helper()
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
	fixture := projectVersionFixture{data: data, ctx: ctx}

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

	mustScan(
		`INSERT INTO users(name,email,user_type,role,status)
		 VALUES ('Version Counselor',$1,'counselor','counselor','active')
		 RETURNING id`,
		[]any{"version-counselor-" + suffix + "@example.com"},
		&fixture.counselorID,
	)
	mustScan(
		`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`,
		[]any{"Project Version Test " + suffix},
		&fixture.organizationID,
	)
	mustScan(
		`INSERT INTO users(organization_id,name,email,user_type,role,status)
		 VALUES ($1,'Version Assessor',$2,'stakeholder','assessor','active')
		 RETURNING id`,
		[]any{fixture.organizationID, "version-assessor-" + suffix + "@example.com"},
		&fixture.assessorID,
	)

	status := "setup"
	if finalized {
		status = "closed"
	}
	mustScan(
		`INSERT INTO projects(
			organization_id,counselor_id,name,slug,status,objective,assessment_period,
			target_completion_date,scope_boundary,compliance_driver
		) VALUES ($1,$2,'Versioned RU Registration',$3,$4,
			'Assess the cybersecurity readiness of the registration service',
			'Q3 2026','2026-09-30',
			'Production application, API, database, and operations team',
			'Customer assurance and NIST CSF 2.0 readiness'
		) RETURNING id`,
		[]any{fixture.organizationID, fixture.counselorID, "versioned-ru-registration-" + suffix, status},
		&fixture.sourceProjectID,
	)
	mustExec(`INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions`, fixture.sourceProjectID)
	mustExec(`INSERT INTO project_subcategory_profiles(project_id,subcategory_id) SELECT $1,id FROM subcategories`, fixture.sourceProjectID)
	mustScan(`SELECT id FROM subcategories ORDER BY code LIMIT 1`, nil, &fixture.subcategoryID)

	mustExec(`
		UPDATE project_subcategory_profiles
		SET included=true,
			rationale='Scope rationale',
			assigned_user_id=$2,
			current_priority='medium',
			current_coverage_level='partial',
			current_status_text='Current state evidence',
			current_policies_text='Current policy evidence',
			current_tier='1',
			target_priority='high',
			target_coverage_level='full',
			target_approach_text='Target state approach',
			target_tier='2',
			notes='Source notes',
			considerations='Source considerations',
			review_status='approved',
			submitted_at=now(),
			reviewed_by=$2,
			reviewed_at=now()
		WHERE project_id=$1 AND subcategory_id=$3`,
		fixture.sourceProjectID, fixture.assessorID, fixture.subcategoryID)

	var responseID string
	mustScan(`
		INSERT INTO stakeholder_responses(
			project_id,subcategory_id,response_text,status,responded_by,submitted_at,
			review_comment,reviewed_by,reviewed_at
		) VALUES ($1,$2,'Source response','reviewed',$3,now(),'Source review',$4,now())
		RETURNING id`,
		[]any{fixture.sourceProjectID, fixture.subcategoryID, fixture.assessorID, fixture.counselorID},
		&responseID,
	)
	mustExec(`
		INSERT INTO response_documents(response_id,original_name,storage_key,mime_type,size_bytes,uploaded_by)
		VALUES ($1,'source-evidence.pdf',$2,'application/pdf',128,$3)`,
		responseID, "version-source-evidence-"+suffix+".pdf", fixture.assessorID)

	var actionID string
	mustScan(`
		INSERT INTO remediation_actions(
			project_id,subcategory_id,title,description,desired_result,priority,
			owner_user_id,due_date,status,progress_note,review_comment,created_by
		) VALUES ($1,$2,'Source remediation','Source remediation description',
			'Source desired result','high',$3,'2026-10-01','in_progress',
			'Source progress','Source review',$4)
		RETURNING id`,
		[]any{fixture.sourceProjectID, fixture.subcategoryID, fixture.assessorID, fixture.counselorID},
		&actionID,
	)
	mustExec(`
		INSERT INTO remediation_evidence(action_id,original_name,storage_path,mime_type,size_bytes,uploaded_by)
		VALUES ($1,'source-remediation.pdf',$2,'application/pdf',256,$3)`,
		actionID, "version-source-remediation-"+suffix+".pdf", fixture.assessorID)

	t.Cleanup(func() {
		_, _ = data.DB.Exec(ctx, `DELETE FROM audit_logs WHERE organization_id=$1`, fixture.organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM projects WHERE organization_id=$1`, fixture.organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, fixture.organizationID)
		_, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, fixture.counselorID, fixture.assessorID)
		data.Close()
	})
	return fixture
}

func mustCount(t *testing.T, data *Store, ctx context.Context, query, projectID string, destination *int) {
	t.Helper()
	if err := data.DB.QueryRow(ctx, query, projectID).Scan(destination); err != nil {
		t.Fatal(err)
	}
}

func TestCreateNextProjectVersionClonesScopeAndResetsAssessmentData(t *testing.T) {
	fixture := newProjectVersionFixture(t, true)

	created, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID)
	if err != nil {
		t.Fatalf("create next version: %v", err)
	}
	if created.ID == fixture.sourceProjectID || created.VersionNumber != 2 || created.Status != "setup" {
		t.Fatalf("unexpected new version: %#v", created)
	}
	if created.VersionGroupID == "" || created.PreviousVersionID == nil || *created.PreviousVersionID != fixture.sourceProjectID {
		t.Fatalf("version linkage was not created: %#v", created)
	}
	if !created.IsLatest {
		t.Fatal("new version should be the latest version")
	}

	rows, err := fixture.data.ListProfile(fixture.ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	foundIncluded := false
	for _, row := range rows {
		if row.Included {
			foundIncluded = true
			if row.Rationale != "Scope rationale" || row.AssignedUserID == nil || *row.AssignedUserID != fixture.assessorID {
				t.Fatalf("scope assignment was not copied: %#v", row)
			}
			if row.CurrentPriority != "" || row.CurrentCoverageLevel != "none" ||
				row.TargetPriority != "" || row.TargetCoverageLevel != "none" ||
				row.CurrentStatusText != "" || row.TargetApproachText != "" ||
				row.Notes != "" || row.Considerations != "" || row.ReviewStatus != "draft" {
				t.Fatalf("assessment input was copied: %#v", row)
			}
		} else if row.Rationale != "" || row.AssignedUserID != nil {
			t.Fatalf("non-included scope data was copied: %#v", row)
		}
	}
	if !foundIncluded {
		t.Fatal("expected the included scope row to be copied")
	}

	var responseCount, documentCount, actionCount, actionEvidenceCount int
	mustCount(t, fixture.data, fixture.ctx, `SELECT count(*) FROM stakeholder_responses WHERE project_id=$1`, created.ID, &responseCount)
	mustCount(t, fixture.data, fixture.ctx, `SELECT count(*) FROM response_documents d JOIN stakeholder_responses r ON r.id=d.response_id WHERE r.project_id=$1`, created.ID, &documentCount)
	mustCount(t, fixture.data, fixture.ctx, `SELECT count(*) FROM remediation_actions WHERE project_id=$1`, created.ID, &actionCount)
	mustCount(t, fixture.data, fixture.ctx, `SELECT count(*) FROM remediation_evidence e JOIN remediation_actions a ON a.id=e.action_id WHERE a.project_id=$1`, created.ID, &actionEvidenceCount)
	if responseCount != 0 || documentCount != 0 || actionCount != 0 || actionEvidenceCount != 0 {
		t.Fatalf("new version copied old data: responses=%d documents=%d actions=%d action evidence=%d", responseCount, documentCount, actionCount, actionEvidenceCount)
	}
}

func TestCreateNextProjectVersionRejectsUnfinalizedProject(t *testing.T) {
	fixture := newProjectVersionFixture(t, false)

	if _, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID); !errors.Is(err, ErrProjectVersionNotFinalized) {
		t.Fatalf("expected ErrProjectVersionNotFinalized, got %v", err)
	}
}

func TestCreateNextProjectVersionRejectsOlderVersion(t *testing.T) {
	fixture := newProjectVersionFixture(t, true)
	if _, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID); err != nil {
		t.Fatalf("create version 2: %v", err)
	}

	if _, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID); !errors.Is(err, ErrProjectVersionNotLatest) {
		t.Fatalf("expected ErrProjectVersionNotLatest, got %v", err)
	}
}

func TestCreateNextProjectVersionSerializesVersionNumbers(t *testing.T) {
	fixture := newProjectVersionFixture(t, true)
	results := make(chan struct {
		project Project
		err     error
	}, 2)
	var waitGroup sync.WaitGroup
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			project, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID)
			results <- struct {
				project Project
				err     error
			}{project: project, err: err}
		}()
	}
	waitGroup.Wait()
	close(results)

	successes := []Project{}
	for result := range results {
		if result.err == nil {
			successes = append(successes, result.project)
			continue
		}
		if !errors.Is(result.err, ErrProjectVersionNotLatest) {
			t.Fatalf("unexpected concurrent version error: %v", result.err)
		}
	}
	if len(successes) != 1 || successes[0].VersionNumber != 2 {
		t.Fatalf("expected exactly one version 2, got %#v", successes)
	}

	versions, err := fixture.data.ListProjectVersions(fixture.ctx, fixture.sourceProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0].VersionNumber != 2 || versions[1].VersionNumber != 1 {
		t.Fatalf("unexpected versions after concurrent create: %#v", versions)
	}
}

func TestListProjectVersionsReturnsNewestFirst(t *testing.T) {
	fixture := newProjectVersionFixture(t, true)
	created, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID)
	if err != nil {
		t.Fatalf("create version 2: %v", err)
	}

	versions, err := fixture.data.ListProjectVersions(fixture.ctx, fixture.sourceProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected two project versions, got %d", len(versions))
	}
	if versions[0].ID != created.ID || versions[0].VersionNumber != 2 || versions[1].ID != fixture.sourceProjectID || versions[1].VersionNumber != 1 {
		t.Fatalf("versions are not newest first: %#v", versions)
	}
	if versions[0].VersionGroupID == "" || versions[0].VersionGroupID != versions[1].VersionGroupID {
		t.Fatalf("versions do not share a group: %#v", versions)
	}
}

func TestCreateNextProjectVersionKeepsTheVersionSlugRoot(t *testing.T) {
	fixture := newProjectVersionFixture(t, true)
	sourceProject, err := fixture.data.GetProject(fixture.ctx, fixture.sourceProjectID)
	if err != nil {
		t.Fatal(err)
	}
	versionTwo, err := fixture.data.CreateNextProjectVersion(fixture.ctx, fixture.sourceProjectID, fixture.counselorID)
	if err != nil {
		t.Fatalf("create version 2: %v", err)
	}
	if _, err := fixture.data.DB.Exec(fixture.ctx, `UPDATE projects SET status='closed' WHERE id=$1`, versionTwo.ID); err != nil {
		t.Fatal(err)
	}

	versionThree, err := fixture.data.CreateNextProjectVersion(fixture.ctx, versionTwo.ID, fixture.counselorID)
	if err != nil {
		t.Fatalf("create version 3: %v", err)
	}
	if versionThree.Slug != sourceProject.Slug+"-v3" {
		t.Fatalf("version 3 slug should use the original root: got %q, want %q", versionThree.Slug, sourceProject.Slug+"-v3")
	}
}
