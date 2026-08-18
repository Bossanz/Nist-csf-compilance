# Scope Submission Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use inline execution with TDD checkpoints for each task.

**Goal:** Let a Counselor configure Project scope in draft, submit the selected outcomes once, and keep Stakeholder response/profile/evidence work hidden until submission.

**Architecture:** Reuse the existing Project status: setup is scope draft and in_review is scope submitted. Add one protected POST endpoint that validates included outcomes and assignments, then update the existing status. Filter setup Projects for Stakeholders at the API boundary and pass the submitted state into the existing assessment components so the Counselor draft view renders only scope controls.

**Tech Stack:** Go 1.24, PostgreSQL 16, Next.js 16, React, TypeScript, Vitest, Docker Compose.

## Global Constraints

- Keep the implementation V1-sized; do not add a second scope-status column.
- Keep response enum values submitted/reviewed unchanged; this feature controls Project scope visibility only.
- Only counselor and counselor_admin can submit a scope.
- A submitted scope must contain at least one included outcome and every included outcome must have an active Assessor assignment.
- Stakeholders must not receive setup Projects or setup Project profile/response data through direct API calls.

---

### Task 1: Add the Project scope submission store operation

**Files:**
- Modify: api/internal/store/models.go
- Modify: api/internal/store/projects.go
- Test: api/internal/store/projects_integration_test.go

**Interfaces:**
- Produce func (s *Store) SubmitProjectScope(ctx context.Context, projectID string) (Project, error).
- Reuse store.Project.Status and return store.ErrInvalidProjectTransition for a Project that is not in setup or in_review.

- [ ] **Step 1: Write the failing integration tests**

Add tests named TestSubmitProjectScopeMovesSetupProjectToInReview and TestSubmitProjectScopeRejectsProjectWithoutAssignedIncludedOutcomes. Create a Project with CreateScopedProject, update one profile to included=true and assign an active Assessor, call SubmitProjectScope, and assert the returned status is in_review. Create a second Project with no included assigned outcome and assert the call returns store.ErrInvalidProjectTransition.

- [ ] **Step 2: Run the store tests to verify the expected failure**

Run: go test ./internal/store -run 'TestSubmitProjectScope'

Expected: FAIL because SubmitProjectScope and ErrInvalidProjectTransition do not exist.

- [ ] **Step 3: Implement the minimal store operation**

Add ErrInvalidProjectTransition to api/internal/store/models.go. In SubmitProjectScope, lock the Project row, return the existing Project when its status is already in_review, reject statuses other than setup, query for at least one included profile and any included profile without an active assigned user, update projects.status to in_review, and return the updated Project. Keep the operation in one transaction.

- [ ] **Step 4: Run the store tests to verify green**

Run: go test ./internal/store -run 'TestSubmitProjectScope'

Expected: PASS when TEST_DATABASE_URL is configured; otherwise the repository's existing integration-test skip behavior is acceptable.

### Task 2: Add and protect the scope submission API

**Files:**
- Modify: api/internal/httpapi/authorization.go
- Modify: api/internal/httpapi/handler.go
- Modify: api/internal/httpapi/organizations_handler.go
- Modify: api/internal/httpapi/handler_test.go
- Modify: api/internal/httpapi/organizations_handler_test.go

**Interfaces:**
- Add actionSubmitScope.
- Add type scopeStore interface { SubmitProjectScope(context.Context, string) (store.Project, error) }.
- Add POST /api/projects/{projectID}/scope/submit.
- Add func (h *Handler) submitProjectScope(w http.ResponseWriter, r *http.Request, projectID string).

- [ ] **Step 1: Write failing API tests**

Add TestCounselorCanSubmitProjectScope, TestStakeholderCannotSubmitProjectScope, TestStakeholderCannotReadSetupProject, and TestStakeholderCannotResolveSetupProjectSlug. The Counselor test uses a fake scope store that records the Project ID and returns status in_review; it expects HTTP 200 and the submitted status. The Stakeholder tests use a same-organization viewer and a Project with status setup; they expect HTTP 403 for submission and HTTP 404 for Project access.

- [ ] **Step 2: Run the focused API tests to verify red**

Run: go test ./internal/httpapi -run 'Test(CounselorCanSubmitProjectScope|StakeholderCannotSubmitProjectScope|StakeholderCannotReadSetupProject|StakeholderCannotResolveSetupProjectSlug)'

Expected: FAIL because the route, permission, and setup visibility checks do not exist.

- [ ] **Step 3: Implement the API route and visibility checks**

Allow actionSubmitScope only for counselor and counselor_admin. Route the POST request through authorizeProject. Type-assert the store to scopeStore, call SubmitProjectScope, map store.ErrInvalidProjectTransition to HTTP 409, audit project.scope_submitted, and return the Project JSON. In authorizeProject, organizationProjects, organizationProjectBySlug, and the Stakeholder branch of projects, hide Projects whose status is setup.

- [ ] **Step 4: Run the focused API tests to verify green**

Run: go test ./internal/httpapi -run 'Test(CounselorCanSubmitProjectScope|StakeholderCannotSubmitProjectScope|StakeholderCannotReadSetupProject|StakeholderCannotResolveSetupProjectSlug)'

Expected: PASS.

### Task 3: Add the frontend scope submission action

**Files:**
- Modify: web/src/lib/api.ts
- Modify: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.tsx
- Test: web/src/components/ProjectAssessmentWorkspace.test.tsx

**Interfaces:**
- Add api.submitProjectScope(projectID: string): Promise<Project>.
- Pass onSubmitScope: () => Promise<void> to ProjectAssessmentWorkspace.
- Derive scopeSubmitted from project.status !== setup.

- [ ] **Step 1: Write failing UI tests**

Add a draft Counselor test that renders a setup Project, opens an outcome, asserts the rationale and Responsible stakeholder controls exist, and asserts the Assessment profile and Stakeholder response heading do not exist. Assert a Submit scope button is present and calls onSubmitScope. Add a submitted Project test that asserts the scope button is absent and the existing read-only response/profile content can render.

- [ ] **Step 2: Run the focused UI tests to verify red**

Run: npm test -- --run src/components/ProjectAssessmentWorkspace.test.tsx

Expected: FAIL because the workspace has no scope submission prop/state and still renders the Counselor response/profile panels in draft.

- [ ] **Step 3: Implement the minimal frontend flow**

Add the API method and page callback that updates project with the returned Project. Add onSubmitScope to the workspace, show a single submit action for Counselor users while project.status === setup, and pass scopeSubmitted to each AssessmentCard. In the draft Counselor state, keep scope fields visible while suppressing profile reference and response panels. For non-Counselor users, render no included outcomes until the Project is submitted.

- [ ] **Step 4: Run the focused UI tests to verify green**

Run: npm test -- --run src/components/ProjectAssessmentWorkspace.test.tsx

Expected: PASS.

### Task 4: Finish copy, regression coverage, and verification

**Files:**
- Modify: web/src/components/AssessmentCard.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.tsx
- Modify: README.md
- Test: web/src/components/AssessmentCard.test.tsx
- Test: web/src/components/ProjectAssessmentWorkspace.test.tsx

- [ ] **Step 1: Add the draft-state regression assertion**

Keep a direct AssessmentCard test proving that scopeSubmitted=false hides the profile reference and response panel for a Counselor while preserving scope inclusion, rationale, and assignment controls.

- [ ] **Step 2: Update user-facing workflow copy**

Document setup as Scope draft, in_review as Scope submitted, and explain that Stakeholder fields appear only after Counselor submission. Keep response labels Reviewing and Approved unchanged.

- [ ] **Step 3: Run all verification commands**

Run:

    gofmt -w api/internal/store/models.go api/internal/store/projects.go api/internal/httpapi/authorization.go api/internal/httpapi/handler.go api/internal/httpapi/organizations_handler.go api/internal/store/projects_integration_test.go api/internal/httpapi/handler_test.go api/internal/httpapi/organizations_handler_test.go
    go test ./...
    npm test -- --run
    npm run typecheck
    git diff --check

