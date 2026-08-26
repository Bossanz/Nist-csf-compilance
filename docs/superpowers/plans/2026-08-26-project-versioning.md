# Project Versioning Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans (or execute the same tasks inline with the checkpoints below) to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Let Counselors start a new isolated assessment version from the latest Finalized Project while preserving every previous assessment, Action Plan, report, evidence record, and audit event.

**Architecture:** Keep the existing projects table as the assessment-version record. Add version-group metadata to each Project, create the next version as a transactional Project clone, and continue using the concrete Project ID for all existing child tables and routes. Add a compact version-history control to the existing workspace instead of introducing a separate version service.

**Tech Stack:** PostgreSQL 16 migrations, Go 1.24 with pgx store and HTTP handlers, Next.js 16 / React 19 / TypeScript, Vitest, Docker Compose.

**Spec:** docs/superpowers/specs/2026-08-26-project-versioning-design.md

## Global Constraints

- Only counselor and counselor_admin can create a new Project version.
- A source Project must be closed / Finalized and must be the latest version in its version group.
- A new version starts in setup and must submit Scope before Stakeholders can see included outcomes.
- Copy Project metadata, Function scope, included flags, Scope rationale, and Stakeholder assignments.
- Reset Current/Target assessment fields, notes, considerations, responses, evidence, review data, and all Remediation Actions.
- Existing version-specific URLs and Project APIs continue to use concrete Project IDs and never silently redirect.
- Existing version-1 rows receive version_group_id = id, version_number = 1, and previous_version_id = NULL.
- Keep migrations idempotent for fresh db/init databases and existing Docker volumes through db/migrations.
- Do not add production deployment, HTTPS, secrets management, PostgreSQL backup, server-side PDF generation, or DOCX/XLSX inline preview.
- Use the existing audit log and authorization patterns; do not add a second permission or event subsystem.
- For each behavior, follow red-green-refactor: write a failing test, run it and inspect the expected failure, implement the smallest change, rerun the focused test, then run the affected suite.

---

### Task 1: Add failing Store and data-contract tests

**Files:**
- Create: api/internal/store/project_versions_integration_test.go
- Test: api/internal/store/project_versions_integration_test.go

**Interfaces:**
- Consumes: the existing Store, Project, ProfileRow, response, document, remediation, and evidence tables.
- Produces: failing contracts for CreateNextProjectVersion and ListProjectVersions.

- [ ] **Step 1: Write the successful-clone integration test.**

Create a TEST_DATABASE_URL-gated fixture using the existing integration-test style. It must create a unique Counselor, client Organization, active Assessor, and Project; insert catalog Function/profile rows; include and assign one profile; populate its Current/Target fields; insert one Stakeholder response plus response document; insert one Remediation Action plus remediation evidence; and mark the source Project closed with finalized_at. Register cleanup for the temporary rows and Store.

Define a local mustCount helper in the new test file so every count assertion fails through testing.T with the query and argument; it should call data.DB.QueryRow(ctx, query, projectID).Scan(destination) and t.Fatal on database error.

Call the wished-for method and assert the actual database behavior:

~~~
created, err := data.CreateNextProjectVersion(ctx, sourceProjectID, counselorID)
if err != nil {
	t.Fatalf("create next version: %v", err)
}
if created.ID == sourceProjectID || created.VersionNumber != 2 || created.Status != "setup" {
	t.Fatalf("unexpected new version: %#v", created)
}
if created.VersionGroupID == "" || created.PreviousVersionID == nil || *created.PreviousVersionID != sourceProjectID {
	t.Fatalf("version linkage was not created: %#v", created)
}

rows, err := data.ListProfile(ctx, created.ID)
if err != nil {
	t.Fatal(err)
}
for _, row := range rows {
	if row.Included {
		if row.Rationale != "Scope rationale" || row.AssignedUserID == nil {
			t.Fatalf("scope assignment was not copied: %#v", row)
		}
		if row.CurrentPriority != "" || row.CurrentCoverageLevel != "none" ||
			row.TargetPriority != "" || row.TargetCoverageLevel != "none" ||
			row.ReviewStatus != "draft" {
			t.Fatalf("assessment input was copied: %#v", row)
		}
	}
}

var responseCount, documentCount, actionCount, actionEvidenceCount int
mustCount(ctx, data, "SELECT count(*) FROM stakeholder_responses WHERE project_id=$1", created.ID, &responseCount)
mustCount(ctx, data, "SELECT count(*) FROM response_documents d JOIN stakeholder_responses r ON r.id=d.response_id WHERE r.project_id=$1", created.ID, &documentCount)
mustCount(ctx, data, "SELECT count(*) FROM remediation_actions WHERE project_id=$1", created.ID, &actionCount)
mustCount(ctx, data, "SELECT count(*) FROM remediation_evidence e JOIN remediation_actions a ON a.id=e.action_id WHERE a.project_id=$1", created.ID, &actionEvidenceCount)
if responseCount != 0 || documentCount != 0 || actionCount != 0 || actionEvidenceCount != 0 {
	t.Fatalf("new version copied old data: responses=%d documents=%d actions=%d action evidence=%d", responseCount, documentCount, actionCount, actionEvidenceCount)
}
~~~

- [ ] **Step 2: Add invalid-source, ordering, and concurrency tests.**

Add tests named TestCreateNextProjectVersionRejectsUnfinalizedProject, TestCreateNextProjectVersionRejectsOlderVersion, TestCreateNextProjectVersionSerializesVersionNumbers, and TestListProjectVersionsReturnsNewestFirst. The first uses a setup source and expects ErrProjectVersionNotFinalized. The second creates v1 and v2 then calls with v1 and expects ErrProjectVersionNotLatest. The concurrency test runs two calls against the same closed latest source and asserts that any successful rows have distinct version numbers and that the unique version index is never violated. The list test asserts [v2, v1] and one version group.

- [ ] **Step 3: Run the focused tests and verify RED.**

~~~
Set-Location api
go test ./internal/store -run 'Test(CreateNextProjectVersion|ListProjectVersions)' -count=1
Set-Location ..
~~~

Expected result before production changes: compilation fails because Project has no version fields and Store has no version methods. Do not continue if the failure is caused by a test typo.

- [ ] **Step 4: Commit only the new test file.**

~~~
git add -- api/internal/store/project_versions_integration_test.go
git commit -m "test: define project version store behavior"
~~~

---

### Task 2: Add the migration, models, and transactional Store clone

**Files:**
- Create: db/init/012_project_versions.sql
- Create: db/migrations/012_project_versions.sql
- Modify: api/internal/store/models.go
- Modify: api/internal/store/projects.go
- Test: api/internal/store/project_versions_integration_test.go

**Interfaces:**
- Consumes: Task 1 failing tests, existing shared projectSelect/projectArgs, and slug allocation helpers.
- Produces: version fields on Project, CreateNextProjectVersion, and ListProjectVersions.

- [ ] **Step 1: Add identical idempotent SQL to both migration files.**

Use this order so existing rows are backfilled before required constraints:

~~~
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS version_group_id uuid,
  ADD COLUMN IF NOT EXISTS version_number integer,
  ADD COLUMN IF NOT EXISTS previous_version_id uuid;

UPDATE projects
SET version_group_id = COALESCE(version_group_id, id),
    version_number = COALESCE(version_number, 1)
WHERE version_group_id IS NULL OR version_number IS NULL;

ALTER TABLE projects
  ALTER COLUMN version_group_id SET DEFAULT gen_random_uuid(),
  ALTER COLUMN version_group_id SET NOT NULL,
  ALTER COLUMN version_number SET DEFAULT 1,
  ALTER COLUMN version_number SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_version_number_check') THEN
    ALTER TABLE projects ADD CONSTRAINT projects_version_number_check CHECK (version_number > 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_previous_version_fk') THEN
    ALTER TABLE projects ADD CONSTRAINT projects_previous_version_fk
      FOREIGN KEY (previous_version_id) REFERENCES projects(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS projects_version_group_number_unique
  ON projects(version_group_id, version_number);
CREATE INDEX IF NOT EXISTS idx_projects_version_group
  ON projects(version_group_id, version_number DESC);
CREATE INDEX IF NOT EXISTS idx_projects_previous_version
  ON projects(previous_version_id);
~~~

- [ ] **Step 2: Extend the Project model and shared SELECT.**

Add these fields to store.Project, preserving the existing JSON naming convention:

~~~
VersionGroupID    string
VersionNumber     int
PreviousVersionID *string
IsLatest          bool
~~~

Extend projectSelect after the slug with p.version_group_id, p.version_number, p.previous_version_id and a NOT EXISTS subquery that checks for a newer row in the same group. Update projectArgs in the same order. All existing Get/List/Create methods must keep using this shared select. Fresh Projects use migration defaults and remain version 1.

- [ ] **Step 3: Implement the Store methods with one transaction.**

Add:

~~~
var ErrProjectVersionNotFinalized = errors.New("project is not finalized for versioning")
var ErrProjectVersionNotLatest = errors.New("project is not the latest version")
var ErrProjectVersionConflict = errors.New("project version could not be created")

func (s *Store) CreateNextProjectVersion(ctx context.Context, sourceProjectID, actorID string) (Project, error)
func (s *Store) ListProjectVersions(ctx context.Context, projectID string) ([]Project, error)
~~~

CreateNextProjectVersion must begin a transaction, load the source group, acquire pg_advisory_xact_lock(hashtextextended(groupID, 0)), reload the source with FOR UPDATE, reject non-closed and non-latest sources, calculate max(version_number)+1, and derive the slug from the root version slug plus -vN through nextProjectSlug. Insert a setup Project with copied metadata, same Organization/Counselor, same group, next number, and previous_version_id. Copy project_functions applicable/reason rows. Copy only profile included, rationale, and assigned_user_id columns so assessment and review columns use defaults. Do not insert response, document, remediation action, or remediation evidence rows. Load the new Project through projectSelect, commit, and return it. Use _ = actorID because the current projects schema has no creator column; the HTTP layer records the actor.

ListProjectVersions must verify the requested Project exists, select all rows with the same version_group_id ordered by version_number DESC, and return the shared Project shape. A missing source returns pgx.ErrNoRows.

- [ ] **Step 4: Run migration and focused Store tests.**

~~~
docker compose up -d postgres
docker compose run --rm migrate
Set-Location api
go test ./internal/store -run 'Test(CreateNextProjectVersion|ListProjectVersions)' -count=1
Set-Location ..
~~~

Expected result: all focused version tests pass. If PostgreSQL is unavailable, run go test ./internal/store -run '^$' and report the integration runtime blocker separately.

- [ ] **Step 5: Commit the migration and Store slice.**

~~~
git add -- db/init/012_project_versions.sql db/migrations/012_project_versions.sql api/internal/store/models.go api/internal/store/projects.go api/internal/store/project_versions_integration_test.go
git commit -m "feat: add project version persistence"
~~~

---

### Task 3: Add API routes, role guards, and audit event

**Files:**
- Modify: api/internal/httpapi/authorization.go
- Modify: api/internal/httpapi/handler.go
- Modify: api/internal/httpapi/handler_test.go
- Modify: api/internal/httpapi/authorization_test.go

**Interfaces:**
- Consumes: Store version methods, authorizeProject, can, and the existing audit writer.
- Produces: POST /api/projects/{projectID}/versions and GET /api/projects/{projectID}/versions.

- [ ] **Step 1: Write failing authorization and handler tests.**

Add actionCreateProjectVersion and tests that assert only counselor and counselor_admin pass can(user, actionCreateProjectVersion). Add an authenticated Counselor POST test with a fake version store that records source and actor, returns v2, and expects 201 plus a project.version_created event containing sourceProjectID, sourceVersion, and newVersion. Add an ErrProjectVersionNotFinalized mapping test expecting 409/project_not_finalized and an Assessor 403 test that never calls the store. Add a GET test for two versions and an Auditor test that filters out a version without active project access.

- [ ] **Step 2: Run API tests and verify RED.**

~~~
Set-Location api
go test ./internal/httpapi -run 'Test(CanCreateProjectVersion|CreateProjectVersion|ListProjectVersions)' -count=1
Set-Location ..
~~~

Expected result: the new tests fail because the action, route, handler, and fake-store support do not exist.

- [ ] **Step 3: Implement the separate projectVersionStore interface and route.**

Add:

~~~
type projectVersionStore interface {
  CreateNextProjectVersion(context.Context, string, string) (store.Project, error)
  ListProjectVersions(context.Context, string) ([]store.Project, error)
}
~~~

Allow actionCreateProjectVersion only for Counselor roles. Add it to the finalized-project exceptions in authorizeProject. Route versions before generic Project GET/DELETE:

~~~
if len(parts) == 4 && parts[3] == "versions" {
  if r.Method == http.MethodGet {
    if !h.authorizeProject(w, r, id, nil) { return }
    h.listProjectVersions(w, r, id)
    return
  }
  if r.Method == http.MethodPost {
    requested := actionCreateProjectVersion
    if !h.authorizeProject(w, r, id, &requested) { return }
    h.createProjectVersion(w, r, id)
    return
  }
}
~~~

createProjectVersion calls the Store with currentUser(r).ID, maps pgx.ErrNoRows to 404, ErrProjectVersionNotFinalized to 409/project_not_finalized, ErrProjectVersionNotLatest to 409/version_not_latest, and ErrProjectVersionConflict to 409/version_creation_conflict. On success it writes one project.version_created event for the new Project with only source/new IDs and version numbers, then returns 201.

listProjectVersions authorizes the source first. Counselors see all versions. Stakeholder users cannot see setup versions; Auditors additionally require active project access for each returned version.

- [ ] **Step 4: Run focused and existing authorization tests.**

~~~
Set-Location api
go test ./internal/httpapi -run 'Test(CanCreateProjectVersion|CreateProjectVersion|ListProjectVersions|Authorization|Role)' -count=1
Set-Location ..
~~~

- [ ] **Step 5: Commit the API slice.**

~~~
git add -- api/internal/httpapi/authorization.go api/internal/httpapi/authorization_test.go api/internal/httpapi/handler.go api/internal/httpapi/handler_test.go api/internal/httpapi/role_matrix_test.go
git commit -m "feat: expose project version APIs"
~~~

---

### Task 4: Add frontend types, API methods, and contract tests

**Files:**
- Modify: web/src/lib/types.ts
- Modify: web/src/lib/api.ts
- Modify: web/src/lib/api.test.ts

**Interfaces:**
- Consumes: backend JSON contracts from Tasks 2–3.
- Produces: Project version fields, api.getProjectVersions, and api.createProjectVersion.

- [ ] **Step 1: Write failing API tests.**

~~~
test("gets Project versions", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("[]", { status: 200 })));
  await api.getProjectVersions("project-1");
  expect(fetch).toHaveBeenCalledWith("/api/projects/project-1/versions", expect.objectContaining({ headers: expect.any(Object) }));
});

test("creates the next Project version", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ id: "project-2", versionNumber: 2 }), { status: 201 })));
  const created = await api.createProjectVersion("project-1");
  expect(created.versionNumber).toBe(2);
  expect(fetch).toHaveBeenCalledWith("/api/projects/project-1/versions", expect.objectContaining({ method: "POST", body: "{}" }));
});
~~~

- [ ] **Step 2: Run focused tests and verify RED.**

~~~
Set-Location web
npm.cmd test -- --run src/lib/api.test.ts
Set-Location ..
~~~

Expected result: the methods and version field are missing.

- [ ] **Step 3: Add types and API methods.**

Extend Project with optional versionGroupID, versionNumber, previousVersionID, and isLatest fields. Add:

~~~
getProjectVersions: (id: string) => request<Project[]>("/api/projects/" + id + "/versions"),
createProjectVersion: (id: string) => request<Project>("/api/projects/" + id + "/versions", { method: "POST", body: JSON.stringify({}) }),
~~~

Keep fields optional for existing test fixtures; migrated production responses include them.

- [ ] **Step 4: Run focused tests and typecheck.**

~~~
Set-Location web
npm.cmd test -- --run src/lib/api.test.ts
npx.cmd tsc --noEmit --incremental false
Set-Location ..
~~~

- [ ] **Step 5: Commit the frontend contract.**

~~~
git add -- web/src/lib/types.ts web/src/lib/api.ts web/src/lib/api.test.ts
git commit -m "feat: add frontend project version API"
~~~

---

### Task 5: Build Version History UI with component tests

**Files:**
- Create: web/src/components/ProjectVersionHistory.tsx
- Create: web/src/components/ProjectVersionHistory.test.tsx
- Modify: web/src/app/globals.css

**Interfaces:**
- Consumes: Organization, Project, projectPath, and useDialogFocus.
- Produces: accessible version links, current marker, Finalized-only create action, confirmation, loading, retry, and error states.

- [ ] **Step 1: Write failing component tests.**

Test that current v2 and a v1 link render; that Start new assessment appears only when canCreate is true, Project is latest, and status is closed; that confirmation calls onCreate once; that creating disables the action; and that an error keeps history visible and calls onRetry.

~~~
test("shows current and historical versions", () => {
  render(<ProjectVersionHistory organization={organization} project={v2} versions={[v2, v1]} loading={false} error="" creating={false} canCreate={false} onCreate={vi.fn()} onRetry={vi.fn()} />);
  expect(screen.getByText("v2")).toBeTruthy();
  expect(screen.getByRole("link", { name: /v1.*finalized/i })).toHaveAttribute("href", "/organizations/acme/projects/readiness");
});

test("confirms before starting a new assessment", () => {
  const onCreate = vi.fn();
  render(<ProjectVersionHistory organization={organization} project={v1} versions={[v1]} loading={false} error="" creating={false} canCreate onCreate={onCreate} onRetry={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: /start new assessment/i }));
  fireEvent.click(screen.getByRole("button", { name: /confirm start/i }));
  expect(onCreate).toHaveBeenCalledTimes(1);
});
~~~

- [ ] **Step 2: Run the component tests and verify RED.**

~~~
Set-Location web
npm.cmd test -- --run src/components/ProjectVersionHistory.test.tsx
Set-Location ..
~~~

Expected result: the missing component import fails.

- [ ] **Step 3: Implement the compact component and QID styles.**

Render a section labelled Version history, show v{versionNumber || 1}, status labels Setup/Reviewing/Finalized, and links made with projectPath. Use the existing useDialogFocus pattern for a role=dialog confirmation with this copy: “Start a new assessment from v1? Scope assignments will be copied. Responses, evidence, reviews, and Action Plan items will start empty in v2.” Keep retry/error inside the panel. Add only QID variable-based borders, spacing, current marker, responsive stacking, and disabled states to globals.css.

- [ ] **Step 4: Run component tests and commit.**

~~~
Set-Location web
npm.cmd test -- --run src/components/ProjectVersionHistory.test.tsx
Set-Location ..
git add -- web/src/components/ProjectVersionHistory.tsx web/src/components/ProjectVersionHistory.test.tsx web/src/app/globals.css
git commit -m "feat: add project version history control"
~~~

---

### Task 6: Wire version loading and creation into the workspace

**Files:**
- Modify: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx
- Modify: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.test.tsx

**Interfaces:**
- Consumes: Task 4 API methods and Task 5 Version History component.
- Produces: non-blocking history loading, Finalized Counselor creation flow, new-slug navigation, and Overview as the first surface.

- [ ] **Step 1: Add failing page/workspace tests.**

Mock getProjectVersions and createProjectVersion. Test that Project heading renders before version history resolves, that history calls getProjectVersions(project-1), that a Finalized latest Counselor confirms creation and navigates to /organizations/acme-corporation/projects/readiness-v2, that history failure shows retry without blanking Overview, and that Stakeholder/Reviewer/Auditor get no create button.

- [ ] **Step 2: Run focused tests and verify RED.**

~~~
Set-Location web
npm.cmd test -- --run "src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx" src/components/ProjectAssessmentWorkspace.test.tsx
Set-Location ..
~~~

- [ ] **Step 3: Add page state and non-blocking loading.**

Add projectVersions, versionHistoryLoading, versionHistoryError, and versionCreating state. Fetch history after critical project/profile/response data is available and outside the critical Promise.all. Reset these states in initialize and retryLoad.

~~~
async function loadProjectVersions(projectID: string, active = true) {
  setVersionHistoryLoading(true);
  setVersionHistoryError("");
  try {
    const rows = await api.getProjectVersions(projectID);
    if (active) setProjectVersions(rows);
  } catch (cause) {
    if (active) setVersionHistoryError(messageOf(cause));
  } finally {
    if (active) setVersionHistoryLoading(false);
  }
}
~~~

createProjectVersion sets versionCreating, calls the API, refreshes history for the created ID, and router.push(projectPath(organization, created)). On failure it keeps the current workspace open and stores the message.

- [ ] **Step 4: Pass props and render the control.**

Add versions, versionHistoryLoading, versionHistoryError, versionCreating, onCreateProjectVersion, and onRetryProjectVersions to ProjectAssessmentWorkspace. Set canCreate to counselor user type, closed status, and project.isLatest !== false. Render the history control near the Project Overview/header. Do not change the existing useState("overview") default.

- [ ] **Step 5: Run focused tests and commit.**

~~~
Set-Location web
npm.cmd test -- --run "src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx" src/components/ProjectAssessmentWorkspace.test.tsx
Set-Location ..
git add -- "web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx" "web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx" web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx
git commit -m "feat: start reassessments from finalized projects"
~~~

---

### Task 7: Add version labels to lists/reports and update migration docs

**Files:**
- Modify: web/src/components/OrganizationWorkspace.tsx
- Modify: web/src/components/OrganizationWorkspace.test.tsx
- Modify: web/src/components/ProjectDashboard.tsx
- Modify: web/src/components/ProjectDashboard.test.tsx
- Modify: web/src/components/FinalReport.tsx
- Modify: web/src/components/FinalReport.test.tsx
- Modify: web/src/components/AuditPackageView.tsx
- Modify: web/src/components/AuditPackageView.test.tsx
- Modify: README.md
- Modify: scripts/migration-smoke-test.ps1

**Interfaces:**
- Consumes: version fields on Project.
- Produces: visible v2/v3 labels, report context, migration verification, and documented API/workflow.

- [ ] **Step 1: Add failing presentation and migration assertions.**

Assert that OrganizationWorkspace and ProjectDashboard show v2 for versionNumber 2, and that FinalReport and AuditPackageView show Assessment version v2. Extend migration smoke verification to require the 012_project_versions marker and all three project columns.

~~~
$marker = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT 1 FROM schema_migrations WHERE version='012_project_versions'"
if ($marker.Trim() -ne '1') { throw 'project version migration is not recorded' }
$columns = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM information_schema.columns WHERE table_name='projects' AND column_name IN ('version_group_id','version_number','previous_version_id')"
if ([int]$columns.Trim() -ne 3) { throw 'project version columns are missing' }
~~~

- [ ] **Step 2: Run focused tests and verify RED.**

~~~
Set-Location web
npm.cmd test -- --run src/components/OrganizationWorkspace.test.tsx src/components/ProjectDashboard.test.tsx src/components/FinalReport.test.tsx src/components/AuditPackageView.test.tsx
Set-Location ..
.\scripts\migration-smoke-test.ps1
~~~

Expected result: new labels fail until implemented; migration check fails until 012 is applied. Do not delete volumes.

- [ ] **Step 3: Implement labels, report context, and documentation.**

Use versionNumber || 1 for display. Keep Project cards separate in v1 and preserve open/delete behavior. Add Assessment version and Previous version to report context without merging data across Project IDs. Update README with the two version endpoints, Finalized-latest-only rule, copied/reset data, stable old links, and docker compose run --rm migrate.

- [ ] **Step 4: Run focused checks and commit.**

~~~
Set-Location web
npm.cmd test -- --run src/components/OrganizationWorkspace.test.tsx src/components/ProjectDashboard.test.tsx src/components/FinalReport.test.tsx src/components/AuditPackageView.test.tsx
Set-Location ..
.\scripts\migration-smoke-test.ps1
git add -- web/src/components/OrganizationWorkspace.tsx web/src/components/OrganizationWorkspace.test.tsx web/src/components/ProjectDashboard.tsx web/src/components/ProjectDashboard.test.tsx web/src/components/FinalReport.tsx web/src/components/FinalReport.test.tsx web/src/components/AuditPackageView.tsx web/src/components/AuditPackageView.test.tsx README.md scripts/migration-smoke-test.ps1
git commit -m "feat: show project assessment versions"
~~~

---

### Task 8: Full verification and authenticated runtime flow

**Files:**
- Modify: scripts/smoke-test.ps1 only if one focused version assertion is needed.
- Test: complete Go/web suites, migration smoke, Compose health checks, and authenticated workflow.

**Interfaces:**
- Consumes: Tasks 1–7.
- Produces: evidence that version creation is isolated, authorized, routable, and compatible with assessment/finalization/remediation.

- [ ] **Step 1: Run Go tests and vet.**

~~~
Set-Location api
go test ./...
go vet ./...
Set-Location ..
~~~

- [ ] **Step 2: Run web tests and TypeScript.**

~~~
Set-Location web
npm.cmd test -- --run
npx.cmd tsc --noEmit --incremental false
Set-Location ..
~~~

- [ ] **Step 3: Apply migration without deleting local data.**

~~~
docker compose up -d postgres
docker compose run --rm migrate
powershell -ExecutionPolicy Bypass -File scripts/migration-smoke-test.ps1
docker compose run --rm migrate
~~~

Confirm a pre-existing Project is version 1 with version_group_id equal to its own ID, and that the second migrate run skips 012 cleanly.

- [ ] **Step 4: Run Compose and authenticated smoke flow.**

~~~
docker compose up --build -d
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:3000
.\scripts\smoke-test.ps1 -CounselorAdminEmail admin@example.com -CounselorAdminPassword LocalAdmin!2026 -KeepData
~~~

The flow must create/finish a source Project, create v2 as Counselor, verify the v2 slug opens on Overview/setup, verify v1 remains Finalized/read-only, and verify v1 report/audit data contains no v2 responses or Actions. Add only these checks if the existing smoke script lacks them; keep its cleanup behavior.

- [ ] **Step 5: Run the production build and record the actual result.**

~~~
Set-Location web
npm.cmd run build
Set-Location ..
~~~

If the environment again reports EPERM while opening web/.next/trace, report that exact cache/runtime blocker and do not call the build successful. Keep tests/typecheck separate from build status.

- [ ] **Step 6: Review the final diff and acceptance criteria.**

~~~
git diff --check
git status --short
git log -8 --oneline
~~~

Re-read the acceptance criteria in docs/superpowers/specs/2026-08-26-project-versioning-design.md. Confirm only intended versioning files are staged in each commit, preserve unrelated worktree changes, and record any runtime limitation before handoff.
