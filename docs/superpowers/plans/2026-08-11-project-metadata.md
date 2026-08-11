# Project Metadata Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend v1 project creation with useful assessment context without mixing project metadata with outcome-level scope and assessment fields.

**Architecture:** Store five metadata fields on `projects`, keep the existing `CreateScopedProject` API as a name-only compatibility wrapper, and add a focused metadata-aware store method for the organization project route. The organization workspace will submit the metadata in one compact form; organization, counselor, framework, slug, status, and created date remain system-managed.

**Tech Stack:** PostgreSQL migration, Go store/HTTP API, Next.js/React/TypeScript, Vitest, Docker Compose.

## Global Constraints

- Keep the implementation lean and limited to project-level metadata.
- Do not add stakeholder assignment, outcome scope, Current/Target, Priority/Coverage, response, evidence, or reviewer controls to the create form.
- `targetCompletionDate` is optional and uses `YYYY-MM-DD` at the API boundary.
- Existing projects receive empty metadata defaults through the migration.
- Follow RED → GREEN → REFACTOR for production behavior changes.

---

### Task 1: Persist project metadata

**Files:**
- Create: `db/init/007_project_metadata.sql`
- Modify: `api/internal/store/models.go`
- Modify: `api/internal/store/projects.go`
- Create: `api/internal/store/project_metadata_integration_test.go`

**Interfaces:**
- `store.Project` produces `objective`, `assessmentPeriod`, `targetCompletionDate`, `scopeBoundary`, and `complianceDriver`.
- `store.ProjectMetadata` carries the five create-time values.
- `Store.CreateScopedProjectWithMetadata(ctx, organizationID, name string, metadata ProjectMetadata) (Project, error)` persists the values.
- Existing `CreateProject` and `CreateScopedProject` remain name-only wrappers with empty metadata.

- [ ] **Step 1: Write the failing integration test**

Create a PostgreSQL-backed test that creates an organization, calls `CreateScopedProjectWithMetadata` with all five values, and asserts the returned project contains them. Skip only when `TEST_DATABASE_URL` is absent.

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```powershell
$env:TEST_DATABASE_URL = 'postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
Push-Location api
go test ./internal/store -run TestCreateScopedProjectPersistsMetadata -count=1
Pop-Location
```

Expected: compile failure because the metadata type/method does not exist yet.

- [ ] **Step 3: Add the migration and store model**

Add `objective`, `assessment_period`, `scope_boundary`, and `compliance_driver` as non-null text columns with empty defaults, plus nullable `target_completion_date date`. Add the matching fields to `Project` and define `ProjectMetadata`.

- [ ] **Step 4: Implement the metadata-aware store method**

Insert the five values in the existing project transaction, use `NULLIF($5, '')::date` for the optional date, include the metadata in the create return query, and update all project read queries to select the new fields with `COALESCE(p.target_completion_date::text, '')`.

- [ ] **Step 5: Run the focused test and verify GREEN**

Run the same focused command. Expected: PASS with the metadata round-tripping through PostgreSQL.

- [ ] **Step 6: Commit**

```powershell
git add db/init/007_project_metadata.sql api/internal/store/models.go api/internal/store/projects.go api/internal/store/project_metadata_integration_test.go
git commit -m "feat: persist project metadata"
```

### Task 2: Accept and validate metadata in the organization API

**Files:**
- Modify: `api/internal/httpapi/organizations_handler.go`
- Modify: `api/internal/httpapi/handler_test.go`
- Modify: `api/internal/httpapi/organizations_handler_test.go` if needed by existing test layout

**Interfaces:**
- Organization project POST accepts `name`, `objective`, `assessmentPeriod`, `targetCompletionDate`, `scopeBoundary`, and `complianceDriver`.
- The handler calls `CreateScopedProjectWithMetadata` when the store supports it and keeps the existing name-only fallback for test/legacy stores.

- [ ] **Step 1: Write the failing HTTP tests**

Add one test asserting all metadata is passed to the store and returned in the 201 JSON response. Add one test asserting a malformed non-empty `targetCompletionDate` returns 400 with `validation_error`.

- [ ] **Step 2: Run the focused HTTP tests and verify RED**

Run:

```powershell
Push-Location api
go test ./internal/httpapi -run 'TestCreateOrganizationProject(StoresMetadata|RejectsInvalidTargetDate)$' -count=1
Pop-Location
```

Expected: FAIL because the handler currently reads only `name`.

- [ ] **Step 3: Add input normalization and validation**

Trim all text fields, accept an empty target date, and parse non-empty dates with Go layout `2006-01-02`. Return `400 validation_error` for invalid dates.

- [ ] **Step 4: Call the metadata-aware store method**

Add a narrow optional store interface so existing fakes remain compatible, use the metadata method for the real store, and preserve the existing audit event.

- [ ] **Step 5: Run the focused HTTP tests and verify GREEN**

Run the same focused command and expect both tests to pass.

- [ ] **Step 6: Commit**

```powershell
git add api/internal/httpapi/organizations_handler.go api/internal/httpapi/handler_test.go api/internal/httpapi/organizations_handler_test.go
git commit -m "feat: accept project metadata in API"
```

### Task 3: Add the project creation form

**Files:**
- Modify: `web/src/lib/types.ts`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/app/organizations/[organizationSlug]/page.tsx`
- Modify: `web/src/components/OrganizationWorkspace.tsx`
- Modify: `web/src/components/OrganizationWorkspace.test.tsx`
- Modify: project fixtures in existing frontend tests as required by the expanded `Project` type

**Interfaces:**
- `Project` exposes the five metadata fields returned by the API.
- `ProjectCreateInput` exposes `name`, `objective`, `assessmentPeriod`, `targetCompletionDate`, `scopeBoundary`, and `complianceDriver`.
- The form submits trimmed values to `api.createOrganizationProject` and clears only after a successful create.

- [ ] **Step 1: Write the failing component test**

Extend the Counselor create test to fill objective, assessment period, target date, scope boundary, and compliance driver, submit, and assert the complete object passed to `onCreateProject`.

- [ ] **Step 2: Run the focused component test and verify RED**

Run:

```powershell
Push-Location web
npm.cmd test -- --run src/components/OrganizationWorkspace.test.tsx
Pop-Location
```

Expected: FAIL because the form renders and submits only `name`.

- [ ] **Step 3: Implement the compact metadata form**

Add a textarea for objective, a text input for assessment period, a date input for target completion, and textareas for scope boundary and compliance driver. Keep the form visible only to Counselor roles and use the existing editorial form styles.

- [ ] **Step 4: Wire page state and API payload**

Update the organization page callback and API method to accept the new input. Keep the selected organization and logged-in counselor out of the form payload.

- [ ] **Step 5: Run the focused component test and verify GREEN**

Run the same focused command and expect the updated create test to pass.

- [ ] **Step 6: Commit**

```powershell
git add web/src/lib/types.ts web/src/lib/api.ts web/src/app/organizations/[organizationSlug]/page.tsx web/src/components/OrganizationWorkspace.tsx web/src/components/OrganizationWorkspace.test.tsx
git commit -m "feat: collect project assessment context"
```

### Task 4: Document and verify the end-to-end flow

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Document the new project metadata and migration command**

Explain the six project creation fields, the system-managed fields, and the one-time `007_project_metadata.sql` command for existing PostgreSQL volumes.

- [ ] **Step 2: Run the full verification suite**

Run `go test ./...` with `TEST_DATABASE_URL`, `npm.cmd test`, `npm.cmd run typecheck`, `npm.cmd run build`, `git diff --check`, and `docker compose up --build -d`. Verify Web/API health endpoints and `docker compose ps`.

- [ ] **Step 3: Commit**

```powershell
git add README.md
git commit -m "docs: describe project metadata"
```
