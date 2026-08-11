# Bulk Include Outcomes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Counselor-only checkbox that bulk-updates the included flag for all outcomes in the selected Function.

**Architecture:** Keep the current Function selection and profile update API. The project page performs one existing profile PATCH per outcome and updates local profile state plus the summary after all calls finish. The workspace owns the checkbox presentation and loading/error state.

**Tech Stack:** Next.js, React, TypeScript, Vitest, Testing Library, Go API.

## Global Constraints

- Only update `included`; preserve rationale and assigned stakeholder values.
- Only Counselor users can see or trigger the control.
- Do not add a new endpoint or database schema for v1.
- Follow TDD: write a failing test, verify RED, implement the smallest change, verify GREEN.

---

### Task 1: Add the bulk control to the assessment workspace

**Files:**
- Modify: `web/src/components/ProjectAssessmentWorkspace.tsx`
- Test: `web/src/components/ProjectAssessmentWorkspace.test.tsx`

- [ ] **Step 1: Write the failing test**

  Render the Counselor with one included and one excluded outcome in the selected Function. Assert that the checkbox named `Include all outcomes in this Function` is visible and unchecked, then click it and expect `onSetFunctionIncluded("GV", true)`. Add a read-only-role assertion that the control is absent.

- [ ] **Step 2: Run the focused test and verify RED**

  ```powershell
  npm.cmd test -- --run src/components/ProjectAssessmentWorkspace.test.tsx
  ```

  Expected: the new checkbox cannot be found because the workspace has no bulk control.

- [ ] **Step 3: Implement the workspace control**

  Add optional prop `onSetFunctionIncluded: (functionCode: string, included: boolean) => Promise<void>`. Derive `functionRows` from the full profile, calculate `allIncluded`, and render a Counselor-only checkbox in the assessment header. Track `bulkState` and show a small saving/error status while the callback runs.

- [ ] **Step 4: Run the focused test and verify GREEN**

  Re-run the same command and expect all workspace tests to pass.

### Task 2: Wire the existing profile API to the bulk control

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx`
- Test: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`

- [ ] **Step 1: Write the failing page test**

  Provide two outcomes in Function `GV`, click the bulk checkbox, and assert that `api.updateProfile` is called for both subcategories with `{ included: true }`.

- [ ] **Step 2: Run the page test and verify RED**

  ```powershell
  npm.cmd test -- --run "src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx"
  ```

  Expected: the checkbox is absent because the page does not pass a bulk callback.

- [ ] **Step 3: Implement the page callback**

  Add `setFunctionIncluded(functionCode, included)`, call `api.updateProfile(project.id, row.subcategoryID, { included })` for each matching profile row, then update local rows and refresh the summary. Pass the callback into `ProjectAssessmentWorkspace`.

- [ ] **Step 4: Run full verification**

  ```powershell
  npm.cmd test
  npm.cmd run typecheck -- --incremental false
  Set-Location ..\api
  $env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
  go test ./...
  Set-Location ..\web
  npm.cmd run build
  ```

  Expected: all tests, typecheck, Go tests, and production build pass.

### Task 3: Update documentation and run Docker checks

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Document the Counselor bulk control**

  Note that a Counselor can include or exclude all outcomes in the selected Function and can still adjust individual outcomes.

- [ ] **Step 2: Rebuild and verify Docker**

  ```powershell
  docker compose up --build -d
  docker compose ps
  ```

  Expected: PostgreSQL and API are healthy and the web container is running.
