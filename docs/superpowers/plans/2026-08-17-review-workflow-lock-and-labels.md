# Review Workflow Lock and Labels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Reviewer the final gate for each outcome, lock stakeholder edits after review submission or approval, and use clearer UI labels for the internal response states.

**Architecture:** Keep the database/API response values `submitted` and `reviewed` stable to avoid a migration, but present them as `Reviewing` and `Approved` in the UI. Reuse the existing response lifecycle as the single editability rule: stakeholder profile fields and evidence are editable only when the response is `draft` or `needs_more_info`; Reviewer transitions `submitted` to `reviewed` or `needs_more_info`.

**Tech Stack:** Go HTTP API, PostgreSQL-backed store, Next.js/React/TypeScript, Vitest, Go tests.

## Global Constraints

- Do not change the persisted response enum values or require a database migration.
- Counselor scope and assignment editing remains available to Counselor roles.
- Assessor and assigned Organization Admin may edit only their assigned outcome while it is editable.
- Reviewer remains read-only for profile and evidence content and controls only the final review decision.
- Preserve the existing `needs_more_info` resubmission loop.

---

### Task 1: Add failing backend lifecycle-lock tests

**Files:**
- Modify: `api/internal/domain/response_test.go`
- Modify: `api/internal/httpapi/handler_test.go`
- Modify: `api/internal/httpapi/responses_handler_test.go`
- Modify: `api/internal/httpapi/documents_handler_test.go`

**Interfaces:**
- Consumes: existing response and document test fakes.
- Produces: regression tests proving that profile, response, and evidence mutations are rejected after `submitted` or `reviewed`.

- [ ] **Step 1: Write the failing tests**

Add lifecycle cases for editable states (`draft`, `needs_more_info`) and locked states (`submitted`, `reviewed`). Add HTTP tests where an assigned assessor receives `409 invalid_transition` and the fake store records no mutation when updating a profile, saving a response, uploading evidence, or deleting evidence after a locked state.

- [ ] **Step 2: Run the focused backend tests and verify they fail for the missing lock**

Run from `C:\Acuitmesh\NIST-CSF-Compliance\api`:

```powershell
go test ./internal/domain ./internal/httpapi
```

Expected: the new locked-state tests fail because the current authorization path permits the mutation.

### Task 2: Implement one lifecycle editability rule in the API

**Files:**
- Modify: `api/internal/domain/response.go`
- Modify: `api/internal/httpapi/authorization.go`
- Modify: `api/internal/httpapi/responses_handler.go`
- Modify: `api/internal/httpapi/documents_handler.go`
- Modify: `api/internal/store/responses.go`

**Interfaces:**
- Consumes: `domain.ResponseStatus`, `responseStore.ListResponses`, and existing outcome authorization.
- Produces: `domain.CanEditResponse`, returning `409 invalid_transition` for locked mutations while retaining the existing response transitions.

- [ ] **Step 1: Add `domain.CanEditResponse` and use it in the response store**

Return true only for `draft` and `needs_more_info`. Use it in `SaveResponseDraft` so the domain rule is not duplicated in Go code.

- [ ] **Step 2: Add an API helper that loads the outcome response status**

Use the existing response store list method, allow an outcome with no response yet, and return an internal error if the response store cannot be read.

- [ ] **Step 3: Apply the guard to stakeholder profile, response, and evidence mutations**

Reject mutations for `submitted` and `reviewed` with the existing `invalid_transition` error. Keep Reviewer review transitions handled by `ReviewResponse`, and keep Counselor scope changes outside this stakeholder editability guard.

- [ ] **Step 4: Run the focused backend tests and verify they pass**

Run:

```powershell
go test ./internal/domain ./internal/httpapi
```

Expected: all focused tests pass, including the new lock regressions.

### Task 3: Make the UI lifecycle lock match the API and rename labels

**Files:**
- Modify: `web/src/components/AssessmentCard.tsx`
- Modify: `web/src/components/StakeholderResponsePanel.tsx`
- Modify: `web/src/components/StakeholderResponsePanel.test.tsx`
- Modify: `web/src/components/AssessmentCard.test.tsx`

**Interfaces:**
- Consumes: `StakeholderResponse.status` and existing role-based workspace props.
- Produces: read-only Current/Target fields after `submitted` or `reviewed`, plus `Reviewing` and `Approved` display labels.

- [ ] **Step 1: Add failing UI assertions**

Assert that a stakeholder cannot see the assessment save button or editable Current/Target controls for `submitted` and `reviewed` responses, while `needs_more_info` remains editable. Assert that the visible labels are `Reviewing` and `Approved`.

- [ ] **Step 2: Run the focused frontend tests and verify they fail**

Run from `C:\Acuitmesh\NIST-CSF-Compliance\web`:

```powershell
npm test -- --run src/components/StakeholderResponsePanel.test.tsx src/components/AssessmentCard.test.tsx
```

Expected: the new assertions fail because profile editability is currently role-only and the old labels are still rendered.

- [ ] **Step 3: Implement the minimal UI changes**

Derive per-outcome profile editability from the response status, keep the response panel visible as read-only for Reviewer/Counselor/viewer, and change only presentation labels; keep API enum values unchanged.

- [ ] **Step 4: Run the focused frontend tests and verify they pass**

Run the same focused Vitest command and expect all tests to pass.

### Task 4: Align documentation and run the full verification suite

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the implemented response lifecycle and role rules.
- Produces: documentation using `Reviewing` and `Approved` as user-facing terms and explicitly describing the lock behavior.

- [ ] **Step 1: Update the workflow and role descriptions**

Document that Assessor sends work for review, Reviewer approves or sends it back, and approved outcomes are read-only for stakeholder editors.

- [ ] **Step 2: Run full frontend verification**

Run:

```powershell
npm test -- --run
npm run typecheck
```

- [ ] **Step 3: Run full backend verification**

Run from the API directory:

```powershell
go test ./...
```

- [ ] **Step 4: Review the diff and confirm no unrelated files changed**

Run:

```powershell
git status --short
git diff --check
```

