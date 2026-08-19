# Action Plan and Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add outcome-linked remediation actions that Counselors assign, stakeholders execute with evidence, and Counselors close, while assessment data remains immutable after finalization.

**Architecture:** Add a focused remediation domain lifecycle, two PostgreSQL tables, a store/API surface with server-side authorization, and one Action Plan workspace component. Reuse local evidence storage and audit logs; extend live Final Report and Audit Package DTOs instead of adding snapshots or a ticket engine.

**Tech Stack:** Go 1.24, net/http, pgx v5, PostgreSQL, Next.js 16, React 19, TypeScript 5.8, Vitest, Testing Library, Docker Compose.

**Spec:** `docs/superpowers/specs/2026-08-19-action-plan-remediation-design.md`

## Global Constraints

- One approved gap outcome may contain multiple actions.
- Create only when the outcome is included, its response is `reviewed`, and `currentCoverageLevel < targetCoverageLevel` using `none < partial < substantial < full`.
- Counselor roles create, edit, return, and close; assigned active `org_admin` or `assessor` users update progress, evidence, and submit.
- Reviewer and Viewer are read-only.
- Assessment remains locked after project finalization; remediation remains mutable until each action is closed.
- Statuses are exactly `open`, `in_progress`, `awaiting_review`, and `closed`; overdue is derived.
- Evidence keeps the existing 20 MB and allowed MIME-type rules.
- No action deletion, comments subsystem, dependencies, reminders, snapshot service, or new dependency.
- Every production behavior is implemented test-first and every task ends with a focused commit.

---

### Task 1: Remediation lifecycle and persistence contract

**Files:**
- Create: `api/internal/domain/remediation.go`
- Test: `api/internal/domain/remediation_test.go`
- Create: `db/init/010_remediation_actions.sql`
- Modify: `api/internal/store/models.go`

**Interfaces:**
- Produces: `domain.RemediationStatus`, `domain.RemediationPriority`, `domain.HasCoverageGap(current, target CoverageLevel) bool`, `domain.CanTransitionRemediation(from, to RemediationStatus) bool`.
- Produces: `store.RemediationAction`, `store.RemediationEvidence`, `store.RemediationCreate`, `store.RemediationPatch`, and `store.RemediationSummary`.

- [ ] **Step 1: Write failing domain tests**

```go
func TestHasCoverageGapUsesCoverageOrder(t *testing.T) {
    tests := []struct{ current, target CoverageLevel; want bool }{
        {CoverageNone, CoveragePartial, true},
        {CoveragePartial, CoverageSubstantial, true},
        {CoverageSubstantial, CoverageFull, true},
        {CoverageFull, CoverageFull, false},
        {CoverageFull, CoveragePartial, false},
    }
    for _, tt := range tests {
        if got := HasCoverageGap(tt.current, tt.target); got != tt.want {
            t.Fatalf("HasCoverageGap(%q,%q)=%v want %v", tt.current, tt.target, got, tt.want)
        }
    }
}

func TestCanTransitionRemediation(t *testing.T) {
    allowed := [][2]RemediationStatus{{RemediationOpen, RemediationInProgress}, {RemediationInProgress, RemediationAwaitingReview}, {RemediationAwaitingReview, RemediationInProgress}, {RemediationAwaitingReview, RemediationClosed}}
    for _, pair := range allowed {
        if !CanTransitionRemediation(pair[0], pair[1]) { t.Fatalf("expected %s -> %s", pair[0], pair[1]) }
    }
    if CanTransitionRemediation(RemediationClosed, RemediationInProgress) { t.Fatal("closed action must be immutable") }
}
```

- [ ] **Step 2: Run the domain tests and verify RED**

Run: `cd api; go test ./internal/domain -run 'Test(HasCoverageGap|CanTransitionRemediation)'`

Expected: FAIL because remediation symbols do not exist.

- [ ] **Step 3: Implement the minimal domain lifecycle**

```go
type RemediationStatus string
const (
    RemediationOpen RemediationStatus = "open"
    RemediationInProgress RemediationStatus = "in_progress"
    RemediationAwaitingReview RemediationStatus = "awaiting_review"
    RemediationClosed RemediationStatus = "closed"
)
type RemediationPriority string
const (
    PriorityLow RemediationPriority = "low"
    PriorityMedium RemediationPriority = "medium"
    PriorityHigh RemediationPriority = "high"
    PriorityCritical RemediationPriority = "critical"
)
func HasCoverageGap(current, target CoverageLevel) bool {
    rank := map[CoverageLevel]int{CoverageNone: 0, CoveragePartial: 1, CoverageSubstantial: 2, CoverageFull: 3}
    return rank[current] < rank[target]
}
func CanTransitionRemediation(from, to RemediationStatus) bool {
    return (from == RemediationOpen && to == RemediationInProgress) ||
        (from == RemediationInProgress && to == RemediationAwaitingReview) ||
        (from == RemediationAwaitingReview && (to == RemediationInProgress || to == RemediationClosed))
}
```

- [ ] **Step 4: Add the migration and store DTOs**

Create the two tables and constraints exactly as specified. Add indexes:

```sql
CREATE INDEX idx_remediation_actions_project_status ON remediation_actions(project_id,status);
CREATE INDEX idx_remediation_actions_owner_status ON remediation_actions(owner_user_id,status);
CREATE INDEX idx_remediation_actions_outcome ON remediation_actions(project_id,subcategory_id);
```

Define JSON DTOs with `time.Time`/`*time.Time`, and include `OutcomeCode`, `OutcomeDescription`, `CurrentCoverageLevel`, `TargetCoverageLevel`, `OwnerName`, `OwnerEmail`, and `Evidence []RemediationEvidence` on the action read model.

- [ ] **Step 5: Run focused and package tests**

Run: `cd api; go test ./internal/domain ./internal/store`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/internal/domain/remediation.go api/internal/domain/remediation_test.go api/internal/store/models.go db/init/010_remediation_actions.sql
git commit -m "feat: define remediation lifecycle and schema"
```

### Task 2: Remediation store operations and guards

**Files:**
- Create: `api/internal/store/remediation.go`
- Test: `api/internal/store/remediation_integration_test.go`

**Interfaces:**
- Consumes: domain lifecycle and store DTOs from Task 1.
- Produces: `ListRemediationActions`, `CreateRemediationAction`, `UpdateRemediationAction`, `UpdateRemediationProgress`, `SubmitRemediationAction`, and `ReviewRemediationAction` methods on `*store.Store`.

- [ ] **Step 1: Write failing integration tests**

Use the existing integration-test database helper to seed one project, an included profile with `current_coverage_level='partial'` and `target_coverage_level='full'`, a `reviewed` response, a Counselor, and an active Assessor. Assert:

```go
action, err := data.CreateRemediationAction(ctx, projectID, counselorID, store.RemediationCreate{
    SubcategoryID: subcategoryID, Title: "Implement centralized logging",
    Description: "Forward application and API logs to the SIEM.", DesiredResult: "Security events are searchable and retained.",
    Priority: "high", OwnerUserID: assessorID, DueDate: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
})
if err != nil || action.Status != "open" || action.OwnerUserID != assessorID { t.Fatalf("unexpected action: %#v err=%v", action, err) }
```

Add separate tests proving `ErrOutcomeNotApproved`, `ErrNoCoverageGap`, `ErrInvalidRemediationOwner`, `ErrInvalidRemediationTransition`, and `ErrRemediationClosed`; include a test that creates and advances an action after setting the project to `closed`.

- [ ] **Step 2: Run and verify RED**

Run: `cd api; go test ./internal/store -run Remediation -count=1`

Expected: FAIL because the methods and sentinel errors are missing.

- [ ] **Step 3: Implement queries and transactional guards**

Create sentinel errors and implement creation in one transaction. Lock/read the profile, response, project organization, and owner eligibility before insert:

```sql
SELECT p.included,p.current_coverage_level,p.target_coverage_level,r.status
FROM project_subcategory_profiles p
LEFT JOIN stakeholder_responses r ON r.project_id=p.project_id AND r.subcategory_id=p.subcategory_id
WHERE p.project_id=$1 AND p.subcategory_id=$2
FOR UPDATE OF p
```

Validate the owner with:

```sql
SELECT EXISTS(
  SELECT 1 FROM users u JOIN projects p ON p.organization_id=u.organization_id
  WHERE p.id=$1 AND u.id=$2 AND u.status='active' AND u.role IN ('org_admin','assessor')
)
```

List actions ordered by `closed` last, due date, creation date and hydrate evidence. Use conditional `UPDATE ... WHERE status=... RETURNING` for transitions. `UpdateRemediationProgress` must verify owner, reject empty notes, set `status='in_progress'`, clear stale review comments, and update `updated_at`. Submit requires `in_progress` and a non-empty stored progress note. Review requires `awaiting_review`; return requires a trimmed comment and close records `closed_by/closed_at`.

- [ ] **Step 4: Run store tests and verify GREEN**

Run: `cd api; go test ./internal/store -run Remediation -count=1`

Expected: PASS (or SKIP only when the established integration database is unavailable).

- [ ] **Step 5: Run all API tests**

Run: `cd api; go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/internal/store/remediation.go api/internal/store/remediation_integration_test.go
git commit -m "feat: persist guarded remediation actions"
```

### Task 3: Remediation HTTP routes, authorization, and audit events

**Files:**
- Create: `api/internal/httpapi/remediation_handler.go`
- Test: `api/internal/httpapi/remediation_handler_test.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/authorization.go`
- Test: `api/internal/httpapi/authorization_test.go`

**Interfaces:**
- Consumes: Task 2 store methods.
- Produces: list/create/update/progress/submit/review HTTP routes and stable errors from the spec.

- [ ] **Step 1: Write failing handler and authorization tests**

Create a fake `remediationStore` and test these representative requests:

```go
request := authenticatedRequest(http.MethodPost, "/api/projects/project-1/remediation-actions", `{
  "subcategoryID":"outcome-1","title":"Centralize logs","description":"Forward logs",
  "desiredResult":"Searchable security events","priority":"high","ownerUserID":"assessor-1","dueDate":"2026-09-30"
}`, counselor)
handler.ServeHTTP(response, request)
if response.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", response.Code, response.Body.String()) }
```

Also assert Assessor creation is 403, an unassigned Assessor progress update is 403, the assigned Assessor can update/submit, Reviewer mutation is 403, Counselor can return/close, and store errors map to the documented `409`/`422` codes.

- [ ] **Step 2: Run and verify RED**

Run: `cd api; go test ./internal/httpapi -run 'Remediation|Authorization' -count=1`

Expected: FAIL because routes/actions do not exist.

- [ ] **Step 3: Add focused interfaces and routes**

Add to `handler.go`:

```go
type remediationStore interface {
    ListRemediationActions(context.Context, string) ([]store.RemediationAction, error)
    CreateRemediationAction(context.Context, string, string, store.RemediationCreate) (store.RemediationAction, error)
    UpdateRemediationAction(context.Context, string, string, string, store.RemediationPatch) (store.RemediationAction, error)
    UpdateRemediationProgress(context.Context, string, string, string, string) (store.RemediationAction, error)
    SubmitRemediationAction(context.Context, string, string, string) (store.RemediationAction, error)
    ReviewRemediationAction(context.Context, string, string, string, string, string) (store.RemediationAction, error)
}
```

Route under `/api/projects/{projectID}/remediation-actions`, passing the authenticated actor ID to every mutation. Do not call the generic finalized-project mutation guard for these routes; call project access authorization and remediation-specific role/ownership checks instead.

- [ ] **Step 4: Implement validation, error mapping, and audit writes**

Decode dates with `time.Parse("2006-01-02", value)`. Map errors exactly:

```go
case errors.Is(err, store.ErrOutcomeNotApproved): writeError(w, 409, "outcome_not_approved", "Outcome must be approved before creating an action")
case errors.Is(err, store.ErrNoCoverageGap): writeError(w, 409, "no_coverage_gap", "Current coverage must be below target coverage")
case errors.Is(err, store.ErrInvalidRemediationTransition): writeError(w, 409, "invalid_remediation_transition", "Action cannot change state from its current status")
case errors.Is(err, store.ErrRemediationClosed): writeError(w, 409, "remediation_closed", "Closed actions are read-only")
case errors.Is(err, store.ErrInvalidRemediationOwner): writeError(w, 422, "validation_error", "Owner must be an active organization Admin or Assessor")
```

Write the exact audit event names from the spec with project ID, action ID, outcome ID, transition, and changed field names.

- [ ] **Step 5: Run handler and full API tests**

Run: `cd api; go test ./internal/httpapi -run 'Remediation|Authorization' -count=1`

Run: `cd api; go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/internal/httpapi/remediation_handler.go api/internal/httpapi/remediation_handler_test.go api/internal/httpapi/handler.go api/internal/httpapi/authorization.go api/internal/httpapi/authorization_test.go
git commit -m "feat: expose remediation workflow API"
```

### Task 4: Remediation evidence

**Files:**
- Modify: `api/internal/store/remediation.go`
- Create: `api/internal/httpapi/remediation_evidence_handler.go`
- Test: `api/internal/httpapi/remediation_evidence_handler_test.go`
- Modify: `api/internal/httpapi/handler.go`

**Interfaces:**
- Produces store methods `CreateRemediationEvidence`, `GetRemediationEvidence`, `DeleteRemediationEvidence`.
- Produces upload, preview/download, and delete routes under each remediation action.

- [ ] **Step 1: Write failing evidence tests**

Adapt the existing response document tests to assert:

```go
POST /api/projects/project-1/remediation-actions/action-1/evidence
GET /api/projects/project-1/remediation-actions/action-1/evidence/document-1
GET /api/projects/project-1/remediation-actions/action-1/evidence/document-1/preview
DELETE /api/projects/project-1/remediation-actions/action-1/evidence/document-1
```

The assigned owner may upload/delete in `open` or `in_progress`, Counselor may read, Reviewer may read, unassigned users cannot mutate, `awaiting_review`/`closed` reject mutation, unsupported types return 415, and files above 20 MB return 413.

- [ ] **Step 2: Run and verify RED**

Run: `cd api; go test ./internal/httpapi -run RemediationEvidence -count=1`

Expected: FAIL because evidence routes are missing.

- [ ] **Step 3: Implement store methods and handlers**

Reuse `validateEvidenceFile`, `safeDownloadName`, `Evidence.Save/Open/Remove`, and inline preview headers from `documents_handler.go`. Store methods must join through `remediation_actions` and scope every query by both project ID and action ID. Return the deleted storage key so physical cleanup occurs only after database deletion succeeds.

Extend project and organization evidence-key queries to union response and remediation storage keys so deletion cleanup remains complete.

- [ ] **Step 4: Run evidence and API tests**

Run: `cd api; go test ./internal/httpapi -run 'RemediationEvidence|Evidence' -count=1`

Run: `cd api; go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/store/remediation.go api/internal/httpapi/remediation_evidence_handler.go api/internal/httpapi/remediation_evidence_handler_test.go api/internal/httpapi/handler.go
git commit -m "feat: attach evidence to remediation actions"
```

### Task 5: Frontend client and Action Plan component

**Files:**
- Modify: `web/src/lib/types.ts`
- Modify: `web/src/lib/api.ts`
- Test: `web/src/lib/api.test.ts`
- Create: `web/src/components/ActionPlan.tsx`
- Test: `web/src/components/ActionPlan.test.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Produces TypeScript types `RemediationAction`, `RemediationEvidence`, `RemediationCreateInput`, `RemediationStatus`, and `RemediationSummary`.
- Produces `ActionPlan` props containing project, current user, profile rows, organization users, action rows, and async callbacks matching API methods.

- [ ] **Step 1: Write failing API serialization tests**

Assert request method, path, and body for create, progress, submit, review, and evidence upload. Example:

```ts
await api.reviewRemediationAction("project-1", "action-1", { decision: "return", comment: "Attach the deployment record." });
expect(fetch).toHaveBeenCalledWith("/api/projects/project-1/remediation-actions/action-1/review", expect.objectContaining({
  method: "POST", body: JSON.stringify({ decision: "return", comment: "Attach the deployment record." })
}));
```

- [ ] **Step 2: Run and verify RED**

Run: `cd web; npm test -- src/lib/api.test.ts`

Expected: FAIL because API methods/types do not exist.

- [ ] **Step 3: Add types and API methods**

Add methods:

```ts
getRemediationActions(projectID)
createRemediationAction(projectID, input)
updateRemediationAction(projectID, actionID, patch)
updateRemediationProgress(projectID, actionID, progressNote)
submitRemediationAction(projectID, actionID)
reviewRemediationAction(projectID, actionID, { decision, comment })
uploadRemediationEvidence(projectID, actionID, file)
downloadRemediationEvidence(projectID, actionID, evidenceID, signal?)
deleteRemediationEvidence(projectID, actionID, evidenceID)
```

- [ ] **Step 4: Write failing component behavior tests**

Test Counselor creation only for an approved gap, owner-only progress controls, required progress before submit, required return comment, read-only Reviewer/Viewer, closed-action locking, summary counts, and overdue calculation using a fixed clock.

```tsx
render(<ActionPlan {...props} user={assessor} actions={[assignedOpenAction]} />);
expect(screen.getByRole("button", { name: "Send for review" })).toBeDisabled();
await userEvent.type(screen.getByLabelText("Progress update"), "SIEM forwarding enabled in staging.");
expect(screen.getByRole("button", { name: "Send for review" })).toBeEnabled();
```

- [ ] **Step 5: Run and verify RED**

Run: `cd web; npm test -- src/components/ActionPlan.test.tsx`

Expected: FAIL because the component is missing.

- [ ] **Step 6: Implement the minimal Action Plan UI**

Use one responsive summary row, filters, action cards, and an inline create/edit form. Derive eligible rows with approved response IDs supplied by the page, coverage rank, and included state; server errors remain authoritative. Use the current Versotis light/dark tokens, explicit labels/icons, focus-visible states, and top-aligned content.

- [ ] **Step 7: Run component, type, and accessibility-oriented tests**

Run: `cd web; npm test -- src/lib/api.test.ts src/components/ActionPlan.test.tsx`

Run: `cd web; npm run typecheck`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add web/src/lib/types.ts web/src/lib/api.ts web/src/lib/api.test.ts web/src/components/ActionPlan.tsx web/src/components/ActionPlan.test.tsx web/src/app/globals.css
git commit -m "feat: add remediation action plan interface"
```

### Task 6: Integrate Action Plan into the project workspace

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx`
- Test: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`
- Modify: `web/src/components/ProjectAssessmentWorkspace.tsx`
- Test: `web/src/components/ProjectAssessmentWorkspace.test.tsx`

**Interfaces:**
- Consumes: Action Plan component and API methods from Task 5.
- Produces: project workspace mode switch `Assessment | Action Plan`, loading/remutation state, and role-aware default behavior.

- [ ] **Step 1: Write failing page/workspace tests**

Assert the workspace exposes `Action Plan`, loads actions and organization users, supplies only active `org_admin`/`assessor` owner options, and refreshes actions after each mutation. Assert a finalized project still renders Action Plan controls for authorized users while Assessment remains read-only.

- [ ] **Step 2: Run and verify RED**

Run: `cd web; npm test -- 'src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx' src/components/ProjectAssessmentWorkspace.test.tsx`

Expected: FAIL because Action Plan is not integrated.

- [ ] **Step 3: Implement workspace state and callbacks**

Load `getRemediationActions(project.id)` with existing project/profile/response data. Add callbacks that update only the affected action or refresh the list. Pass `responseRows` to eligibility logic so only `reviewed` outcomes can show Counselor creation. Keep all assessment mutation callbacks unchanged.

- [ ] **Step 4: Run focused and full web tests**

Run: `cd web; npm test -- 'src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx' src/components/ProjectAssessmentWorkspace.test.tsx`

Run: `cd web; npm test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add 'web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx' 'web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx' web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx
git commit -m "feat: integrate action plan with project workspace"
```

### Task 7: Final Report, Audit Package, and CSV remediation sections

**Files:**
- Modify: `api/internal/store/reporting.go`
- Test: `api/internal/store/reporting_test.go`
- Modify: `api/internal/httpapi/reporting_handler.go`
- Test: `api/internal/httpapi/reporting_handler_test.go`
- Modify: `web/src/lib/types.ts`
- Modify: `web/src/components/FinalReport.tsx`
- Test: `web/src/components/FinalReport.test.tsx`
- Modify: `web/src/components/AuditPackageView.tsx`
- Test: `web/src/components/AuditPackageView.test.tsx`

**Interfaces:**
- Extends `FinalReport` and `AuditPackage` with `RemediationSummary` and `[]RemediationAction`.
- Extends audit CSV with the remediation register rows defined by the spec.

- [ ] **Step 1: Write failing API report tests**

Seed open, overdue, awaiting-review, and closed actions. Assert report summary counts, evidence hydration, action/outcome linkage, and CSV headers:

```text
record_type,action_id,outcome_code,action_title,owner,priority,due_date,action_status,submitted_at,closed_at,review_comment,evidence_names
```

Use `record_type=assessment_outcome` for existing rows and `record_type=remediation_action` for action rows so one export remains machine-readable.

- [ ] **Step 2: Run and verify RED**

Run: `cd api; go test ./internal/store ./internal/httpapi -run 'Report|AuditPackageCSV' -count=1`

Expected: FAIL because reporting DTOs omit remediation.

- [ ] **Step 3: Extend store reporting and CSV**

Call `ListRemediationActions` from both report builders. Calculate counts from action statuses and derive overdue with `due_date < CURRENT_DATE AND status <> 'closed'`. Keep assessment coverage unchanged. Append remediation CSV rows without removing current assessment rows.

- [ ] **Step 4: Write failing frontend report tests**

Assert Final Report shows assessment finalization separately from remediation progress, and Audit Package shows owner, due date, status, evidence metadata, and remediation audit events.

- [ ] **Step 5: Run and verify frontend RED**

Run: `cd web; npm test -- src/components/FinalReport.test.tsx src/components/AuditPackageView.test.tsx`

Expected: FAIL because remediation sections are absent.

- [ ] **Step 6: Implement report sections and update DTO types**

Render a compact remediation summary and a grouped action table. Preserve existing print styles, use text status labels, and show `No remediation actions recorded` when empty.

- [ ] **Step 7: Run report and full test suites**

Run: `cd api; go test ./...`

Run: `cd web; npm test; npm run typecheck`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add api/internal/store/reporting.go api/internal/store/reporting_test.go api/internal/httpapi/reporting_handler.go api/internal/httpapi/reporting_handler_test.go web/src/lib/types.ts web/src/components/FinalReport.tsx web/src/components/FinalReport.test.tsx web/src/components/AuditPackageView.tsx web/src/components/AuditPackageView.test.tsx
git commit -m "feat: include remediation in audit reporting"
```

### Task 8: Documentation and release verification

**Files:**
- Modify: `README.md`
- Modify: `docs/superpowers/plans/2026-08-19-action-plan-remediation.md`

**Interfaces:**
- Documents roles, lifecycle, API surface, finalized-project behavior, migration initialization, and local test commands.

- [ ] **Step 1: Update README**

Move full action planning out of the not-included list. Add the exact lifecycle, eligible owner roles, create eligibility, post-finalization behavior, evidence support, reporting output, and stable API status labels.

- [ ] **Step 2: Run formatting and static verification**

Run: `cd api; gofmt -w internal/domain/remediation.go internal/domain/remediation_test.go internal/store/remediation.go internal/store/remediation_integration_test.go internal/httpapi/remediation_handler.go internal/httpapi/remediation_handler_test.go internal/httpapi/remediation_evidence_handler.go internal/httpapi/remediation_evidence_handler_test.go`

Run: `cd api; go vet ./...; go test ./...`

Run: `cd web; npm test; npm run typecheck; npm run build`

Expected: every command exits 0 with no test failure or TypeScript/build error.

- [ ] **Step 3: Verify Docker migration and smoke paths**

Run: `docker compose up --build -d`

Run: `docker compose ps`

Run: `docker compose exec api go test ./...`

Verify with authenticated requests or browser flow: create an approved-gap action, update as assigned Assessor, upload evidence, submit, return, resubmit, close, finalize assessment independently, and confirm Action Plan remains readable/editable as designed.

- [ ] **Step 4: Check the diff and repository cleanliness**

Run: `git diff --check`

Run: `git status --short`

Expected: only intended source/docs files are modified; existing `output/` and `tmp/` remain untracked and unstaged.

- [ ] **Step 5: Commit documentation**

```bash
git add README.md docs/superpowers/plans/2026-08-19-action-plan-remediation.md
git commit -m "docs: document remediation action plan"
```
