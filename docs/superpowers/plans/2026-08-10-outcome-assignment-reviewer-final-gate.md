# Outcome Assignment and Reviewer Final Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Counselor-controlled outcome scope and stakeholder assignment while allowing the assigned Stakeholder to complete the profile and evidence, with Reviewer as the only final gate.

**Architecture:** Extend the existing project_subcategory_profiles row with one nullable assigned_user_id. Keep Current/Target Profile, response, and evidence in the existing tables and endpoints. Enforce organization, active-status, role, inclusion, and assignment rules at the Go HTTP boundary and store transaction, then pass explicit role capabilities to the existing Next.js assessment components.

**Tech Stack:** Next.js 16 + React 19 + TypeScript, Go 1.24 + pgx/v5, PostgreSQL 16, Vitest, Go testing.

## Global Constraints

- Keep the existing lean Next.js + Go + PostgreSQL architecture and avoid a separate task-management subsystem.
- Counselor edits only included, rationale, and one assignee per outcome.
- Assigned org_admin or assessor edits Current/Target Profile, Priority, Coverage, response, and evidence.
- Reviewer is the only final review gate; reviewed completes the outcome and needs_more_info returns it to the assigned Stakeholder.
- Excluded outcomes never appear in Stakeholder profile or response lists.
- The API repeats all client-side assignment validation and never trusts filtered UI options.
- Existing response transitions remain draft -> submitted, submitted -> reviewed or needs_more_info, and needs_more_info -> submitted.
- Existing Go tests, frontend tests, typecheck, production build, and Docker build remain green.

## File map

Create:

- db/init/006_outcome_assignments.sql — additive assignment column and index.
- api/internal/store/profile_assignment_integration_test.go — opt-in PostgreSQL assignment lifecycle coverage.
- web/src/components/ProjectAssessmentWorkspace.test.tsx — role-specific outcome visibility coverage.

Modify:

- api/internal/store/models.go and api/internal/store/projects.go — assignment fields, joins, transaction validation, clearing, and persistence.
- api/internal/httpapi/authorization.go and api/internal/httpapi/handler.go — role and assignment checks for profile and outcome access.
- api/internal/httpapi/responses_handler.go — assigned response list filtering.
- API authorization, handler, response, and document tests — role and assignment cases.
- web/src/lib/types.ts, web/src/lib/api.ts, and the project page — assignment contract and organization-user loading.
- web/src/components/ProjectAssessmentWorkspace.tsx, ProfileEditor.tsx, AssessmentCard.tsx — role-specific capabilities and controls.
- web/src/components/ProfileEditor.test.tsx and StakeholderResponsePanel.test.tsx — UI behavior.
- web/src/components/StakeholderResponsePanel.tsx — explicit Reviewer final-decision copy.
- README.md — workflow, role boundaries, and existing-database migration command.

---

### Task 1: Persist one eligible assignee per outcome

**Files:**

- Create db/init/006_outcome_assignments.sql
- Create api/internal/store/profile_assignment_integration_test.go
- Modify api/internal/store/models.go
- Modify api/internal/store/projects.go

**Interfaces:**

- ProfileRow exposes AssignedUserID *string, AssignedUserName string, and AssignedUserEmail string.
- ProfilePatch exposes AssignedUserID *string.
- Existing ListProfile and UpdateProfile method signatures stay unchanged.
- UpdateProfile validates the final state in one transaction: included rows require an active same-organization stakeholder with role org_admin or assessor; excluded rows always clear the assignment.

- [ ] Step 1: Write the failing integration test

Create a TEST_DATABASE_URL-guarded TestProfileAssignmentLifecycle. Insert a counselor, client organization, active assessor, disabled assessor, reviewer, and project. Select one seeded subcategory, insert one project_subcategory_profiles row for that project and subcategory, then assert this sequence:

~~~go
assigned := assessorID
included := true
row, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{Included: &included, AssignedUserID: &assigned})
if err != nil || row.AssignedUserID == nil || *row.AssignedUserID != assessorID || row.AssignedUserName != "Assignment Assessor" {
    t.Fatalf("assigned profile: %#v err=%v", row, err)
}

disabled := disabledAssessorID
if _, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{AssignedUserID: &disabled}); err != ErrInvalidProfileAssignment {
    t.Fatalf("disabled assignee error: %v", err)
}

excluded := false
cleared, err := data.UpdateProfile(ctx, projectID, subcategoryID, ProfilePatch{Included: &excluded})
if err != nil || cleared.AssignedUserID != nil {
    t.Fatalf("cleared assignment: %#v err=%v", cleared, err)
}
~~~

Use unique emails with uuid.NewString, skip only when TEST_DATABASE_URL is absent, and clean up all inserted rows in t.Cleanup.

- [ ] Step 2: Run the test to verify it fails

~~~powershell
go test ./internal/store -run TestProfileAssignmentLifecycle -count=1
~~~

Expected: compile failure because assignment fields and ErrInvalidProfileAssignment do not exist.

- [ ] Step 3: Add the additive migration

Create db/init/006_outcome_assignments.sql:

~~~sql
ALTER TABLE project_subcategory_profiles
  ADD COLUMN IF NOT EXISTS assigned_user_id uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_profiles_assigned_user
  ON project_subcategory_profiles(assigned_user_id)
  WHERE assigned_user_id IS NOT NULL;
~~~

Do not add a cross-table database constraint for active status or role; the store transaction validates those values against the project organization.

- [ ] Step 4: Implement model fields and transactional store validation

In ListProfile, left join users as assigned_user and select the nullable ID plus COALESCE name and email. In UpdateProfile, lock the row, read the current included value, assignment, and project organization, then calculate the final values:

~~~go
if !finalIncluded {
    finalAssignedUserID = nil
}
if finalIncluded && finalAssignedUserID == nil {
    return ProfileRow{}, ErrInvalidProfileAssignment
}
~~~

When an assignment is present, require the user to match the project organization, user_type stakeholder, role org_admin or assessor, and status active. Update the requested profile columns plus assigned_user_id, commit, and return the updated row. Define ErrInvalidProfileAssignment as a store sentinel error.

- [ ] Step 5: Run formatting and the focused test

~~~powershell
gofmt -w internal/store/models.go internal/store/projects.go internal/store/profile_assignment_integration_test.go
go test ./internal/store -run TestProfileAssignmentLifecycle -count=1
~~~

Expected: it compiles and skips without TEST_DATABASE_URL; with 006 applied, it passes.

- [ ] Step 6: Commit

~~~powershell
git add db/init/006_outcome_assignments.sql api/internal/store/models.go api/internal/store/projects.go api/internal/store/profile_assignment_integration_test.go
git commit -m "feat: persist outcome assignees"
~~~

### Task 2: Enforce profile roles and outcome visibility in the API

**Files:**

- Modify api/internal/httpapi/authorization.go
- Modify api/internal/httpapi/handler.go
- Modify api/internal/httpapi/authorization_test.go
- Modify api/internal/httpapi/handler_test.go

**Interfaces:**

- actionUpdateProfile is allowed for Counselor roles and org_admin or assessor; the profile handler applies assignment checks before allowing stakeholder fields.
- Add helpers equivalent to stakeholderCanReadProfile, stakeholderCanEditProfile, and profileFieldsAllowedForRole.
- Counselor reads all outcomes and updates only included, rationale, and assigned user.
- Assigned active org_admin or assessor reads and updates only assigned included outcomes and only Current/Target/Profile fields.
- Reviewer and Viewer read all included outcomes but cannot update profiles.

- [ ] Step 1: Write failing HTTP tests

Extend fakeStore to capture the patch and add TestCounselorCanUpdateScopeOnly, TestAssignedAssessorCanUpdateProfileOnly, TestUnassignedAssessorCannotUpdateProfile, and TestStakeholderProfileOnlyReturnsAssignedIncludedOutcomes. Add a positive org_admin case. Use ProfileRow.Included true with AssignedUserID pointing to the current assessor for positive cases.

- [ ] Step 2: Run the focused tests to verify they fail

~~~powershell
go test ./internal/httpapi -run 'Test(CounselorCanUpdateScopeOnly|AssignedAssessorCanUpdateProfileOnly|UnassignedAssessorCannotUpdateProfile|StakeholderProfileOnlyReturnsAssignedIncludedOutcomes)' -count=1
~~~

Expected: the tests fail because profile updates are currently Counselor-only and stakeholder filtering checks only included.

- [ ] Step 3: Implement role-specific profile authorization

Let existing project authorization verify organization access, then make updateProfile load the target row and enforce this matrix:

~~~text
counselor_admin or counselor -> included, rationale, assignedUserID
assigned org_admin or assessor -> current/target fields, notes, considerations
reviewer or viewer -> no profile mutation
~~~

Reject fields outside the matrix with 403 forbidden. For an assigned stakeholder, require included true and AssignedUserID equal to currentUser.ID. Map ErrInvalidProfileAssignment to 400 validation_error with a message that an included outcome needs one active Organization Admin or Assessor.

Update profile GET so Counselor users keep all rows, Reviewer and Viewer receive included rows, and org_admin or assessor receive only included rows assigned to their own user ID. Preserve no-auth test behavior when h.Auth is nil.

- [ ] Step 4: Format and run focused tests

~~~powershell
gofmt -w internal/httpapi/authorization.go internal/httpapi/handler.go internal/httpapi/authorization_test.go internal/httpapi/handler_test.go
go test ./internal/httpapi -run 'Test(AuthorizationMatrix|Profile|StakeholderProfile)' -count=1
~~~

- [ ] Step 5: Commit

~~~powershell
git add api/internal/httpapi/authorization.go api/internal/httpapi/handler.go api/internal/httpapi/authorization_test.go api/internal/httpapi/handler_test.go
git commit -m "feat: enforce outcome assignment permissions"
~~~

### Task 3: Apply assignment checks to response and evidence endpoints

**Files:**

- Modify api/internal/httpapi/authorization.go
- Modify api/internal/httpapi/responses_handler.go
- Modify api/internal/httpapi/responses_handler_test.go
- Modify api/internal/httpapi/documents_handler_test.go

**Interfaces:**

- Keep response and document store method signatures unchanged.
- authorizeProjectOutcome uses the profile row, not only an included-ID set.
- org_admin or assessor may save, submit, upload, or delete only when assigned.
- Reviewer and Viewer may read included outcomes; only Reviewer may review.
- GET responses uses the same visibility rules as GET profile.

- [ ] Step 1: Write failing response/evidence tests

Add TestUnassignedAssessorCannotSaveResponse, TestUnassignedAssessorCannotUploadEvidence, TestUnassignedAssessorCannotDeleteEvidence, TestReviewerCanReviewAnyIncludedOutcome, and TestAssignedOrgAdminCanSubmitResponse. Assert 403 and no fake store/storage mutation for unassigned cases. Keep excluded 404, viewer 403, and reviewer submitted-to-reviewed tests.

- [ ] Step 2: Run focused tests to verify they fail

~~~powershell
go test ./internal/httpapi -run 'Test(UnassignedAssessor|ReviewerCanReviewAnyIncludedOutcome|AssignedOrgAdminCanSubmitResponse|AssessorCan|ReviewerCan|ViewerCannot)' -count=1
~~~

Expected: unassigned assessor requests currently pass because the API checks inclusion only.

- [ ] Step 3: Implement shared outcome access

Refactor includedOutcomeIDs into a row lookup/filter helper. Reject missing or excluded rows with 404. For save, submit, upload, and delete actions require the current org_admin or assessor ID to equal the row assignment. For reviewer review require only included access and the existing reviewer role.

Filter stakeholder response lists so an assessor sees assigned outcomes while Reviewer and Viewer see every included outcome. Leave Counselor visibility broad and read-only.

- [ ] Step 4: Format and run response/evidence tests

~~~powershell
gofmt -w internal/httpapi/authorization.go internal/httpapi/responses_handler.go internal/httpapi/responses_handler_test.go internal/httpapi/documents_handler_test.go
go test ./internal/httpapi -run 'Test(Response|Assessor|Reviewer|Viewer|Unassigned)' -count=1
~~~

- [ ] Step 5: Commit

~~~powershell
git add api/internal/httpapi/authorization.go api/internal/httpapi/responses_handler.go api/internal/httpapi/responses_handler_test.go api/internal/httpapi/documents_handler_test.go
git commit -m "feat: scope responses to assigned stakeholders"
~~~

### Task 4: Expose assignment data to the Next.js workspace

**Files:**

- Modify web/src/lib/types.ts
- Modify web/src/lib/api.ts
- Modify web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx
- Modify web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx

**Interfaces:**

- ProfileRow gains assignedUserID: string | null, assignedUserName?: string, and assignedUserEmail?: string.
- ProfilePatch gains assignedUserID?: string | null.
- api.updateProfile keeps its route and accepts the expanded ProfilePatch.
- The project page owns organizationUsers: User[], loads users only for a Counselor session, and passes them to the workspace.

- [ ] Step 1: Write failing TypeScript/page tests

Extend the page API mock with getOrganizationUsers, add active org_admin and assessor fixtures, and assert a Counselor page requests users. Add an api.test assertion that updateProfile sends included, rationale, and assignedUserID in the JSON body.

- [ ] Step 2: Run focused tests to verify they fail

~~~powershell
npm test -- --run src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx src/lib/api.test.ts
~~~

Expected: the new fixture/API assertion fails because assignment fields and organization-user loading are not represented.

- [ ] Step 3: Implement the contract and page load

Add the nullable fields to the TypeScript types. In initialize, call getOrganizationUsers(nextOrganization.id) only when currentUser.userType is counselor; otherwise use an empty list. Store the result and pass it into the workspace. Keep API routes unchanged.

- [ ] Step 4: Run focused tests and typecheck

~~~powershell
npm test -- --run src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx src/lib/api.test.ts
npm run typecheck
~~~

- [ ] Step 5: Commit

~~~powershell
git add web/src/lib/types.ts web/src/lib/api.ts web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx
git commit -m "feat: load outcome assignment options"
~~~

### Task 5: Split Counselor scope editing from Stakeholder profile editing

**Files:**

- Modify web/src/components/ProjectAssessmentWorkspace.tsx
- Modify web/src/components/ProfileEditor.tsx
- Modify web/src/components/AssessmentCard.tsx
- Modify web/src/components/ProfileEditor.test.tsx
- Create web/src/components/ProjectAssessmentWorkspace.test.tsx

**Interfaces:**

- The workspace derives:

~~~ts
const isCounselor = user.userType === "counselor";
const canEditScope = isCounselor;
const canEditProfile = user.role === "org_admin" || user.role === "assessor";
~~~

- Visible rows are all for Counselor, included for Reviewer or Viewer, and included plus assigned to user.id for org_admin or assessor.
- ProfileEditor accepts canEditScope, canEditProfile, and assigneeOptions: User[].
- AssessmentCard sends only scope fields for Counselor and only profile fields for an assigned Stakeholder.

- [ ] Step 1: Write failing component tests

Add a Counselor test that finds an enabled include checkbox and responsible stakeholder select but a disabled Current priority input. Add an assigned Assessor test that changes Current priority, saves, and asserts the patch contains currentPriority but not included or assignedUserID. Add a workspace test asserting Counselor, assigned Assessor, and Reviewer show all, assigned-only, and included-only rows.

- [ ] Step 2: Run component tests to verify they fail

~~~powershell
npm test -- --run src/components/ProfileEditor.test.tsx src/components/ProjectAssessmentWorkspace.test.tsx
~~~

Expected: capability props and the stakeholder select are missing, and the single readOnly fieldset currently lets Counselor edit profile fields.

- [ ] Step 3: Implement explicit capabilities

In the workspace derive capabilities and filter rows. Pass organizationUsers as assigneeOptions only for Counselor. In ProfileEditor pass the capability props to every AssessmentCard.

In AssessmentCard add assignedUserID to the draft, render a Responsible stakeholder select using active org_admin and assessor options, disable scope controls unless canEditScope, disable only Current/Target/supporting fields unless canEditProfile, show Save when either capability is true, and build the narrow patch for the active capability. Keep StakeholderResponsePanel rendered for all roles so Counselor and Reviewer can read.

- [ ] Step 4: Run component tests and typecheck

~~~powershell
npm test -- --run src/components/ProfileEditor.test.tsx src/components/ProjectAssessmentWorkspace.test.tsx src/components/StakeholderResponsePanel.test.tsx
npm run typecheck
~~~

- [ ] Step 5: Commit

~~~powershell
git add web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx web/src/components/ProfileEditor.tsx web/src/components/ProfileEditor.test.tsx web/src/components/AssessmentCard.tsx
git commit -m "feat: split scope and stakeholder assessment editing"
~~~

### Task 6: Make Reviewer completion explicit in the response panel

**Files:**

- Modify web/src/components/StakeholderResponsePanel.tsx
- Modify web/src/components/StakeholderResponsePanel.test.tsx

**Interfaces:**

- Keep onReview(status, comment) and the existing response status union.
- Reviewer sees review controls only for submitted responses.
- reviewed is presented as the final outcome decision; needs_more_info returns the response to the assigned Stakeholder.
- Counselor has no review controls.

- [ ] Step 1: Write the failing final-gate assertion

Render a submitted response as Reviewer, choose reviewed, save, and assert the callback receives reviewed plus the comment. Render the panel as Counselor and assert there is no Save review control.

- [ ] Step 2: Run the panel test to verify the new assertion fails

~~~powershell
npm test -- --run src/components/StakeholderResponsePanel.test.tsx
~~~

- [ ] Step 3: Update copy without adding a second review path

Change the review heading/help text to say Reviewer final decision. Keep the current status select and callback endpoints; do not add Counselor approval state, endpoint, or button.

- [ ] Step 4: Run the panel tests

~~~powershell
npm test -- --run src/components/StakeholderResponsePanel.test.tsx
~~~

- [ ] Step 5: Commit

~~~powershell
git add web/src/components/StakeholderResponsePanel.tsx web/src/components/StakeholderResponsePanel.test.tsx
git commit -m "feat: show reviewer as final outcome gate"
~~~

### Task 7: Document migration and the approved workflow

**Files:**

- Modify README.md

**Interfaces:**

- Document Counselor scope and assignment, assigned Stakeholder profile/response/evidence, Reviewer acceptance or return, and Reviewed as outcome completion.
- Document role boundaries for Counselor, assigned Stakeholder, Reviewer, and Viewer.
- Add the existing-database migration command:

~~~powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/006_outcome_assignments.sql
~~~

- [ ] Step 1: Update README workflow and migration notes

Keep existing login, Docker, and test instructions. Add the command after the 004 stakeholder response migration note and state that fresh databases run 006 automatically in filename order.

- [ ] Step 2: Check the documentation diff

~~~powershell
git diff --check -- README.md
~~~

Expected: no whitespace errors and no Counselor final-review wording.

- [ ] Step 3: Commit

~~~powershell
git add README.md
git commit -m "docs: describe outcome assignment workflow"
~~~

### Task 8: Run complete verification and Docker smoke check

**Files:**

- Verify api, web, db/init, README.md, and Docker Compose configuration.

**Interfaces:**

- No new application interface; this task proves the previous interfaces work together.

- [ ] Step 1: Run Go and frontend checks

~~~powershell
Push-Location api
go test ./...
Pop-Location
Push-Location web
npm test
npm run typecheck
npm run build
Pop-Location
~~~

Expected: Go tests pass, integration tests skip only without TEST_DATABASE_URL, and frontend checks pass.

- [ ] Step 2: Apply migration to the local running database when needed

~~~powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/006_outcome_assignments.sql
~~~

Expected: no SQL errors and assigned_user_id exists on project_subcategory_profiles.

- [ ] Step 3: Run database-backed tests

~~~powershell
$env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
Push-Location api
go test ./internal/store ./internal/httpapi -count=1
Pop-Location
~~~

- [ ] Step 4: Build and inspect Docker services

~~~powershell
docker compose config
docker compose up --build -d
docker compose ps
~~~

Expected: postgres, api, and web are healthy; API health responds; the project page loads with role-specific views.

- [ ] Step 5: Verify the three user paths

1. Counselor includes or excludes an outcome, enters rationale, chooses an active Organization Admin or Assessor, and cannot edit Current/Target.
2. Assigned Stakeholder sees only assigned included outcomes and edits Priority/Coverage, response, and evidence.
3. Reviewer sees included outcomes, marks a submitted outcome Reviewed or Needs more information, and has no Counselor approval step.

- [ ] Step 6: Run final diff checks

~~~powershell
git diff --check HEAD~8..HEAD
git status --short
git log --oneline -8
~~~

Expected: no whitespace errors, a clean worktree, and only intended feature changes in the commits.
