# Project Finalization and Audit Package Implementation Plan

> For agentic workers: implement this plan task-by-task with TDD and verification after each task.

**Goal:** Finalize an assessment only after every included outcome is Reviewer-approved, lock the project, and expose a print-friendly Final Report plus an Auditor-ready Audit Package.

**Architecture:** Reuse the existing projects.status='closed', profile, response, evidence, summary, and audit-log tables. Add finalization metadata and a reporting store layer; expose JSON/CSV endpoints and two Next.js reader routes. Browser Print/Save as PDF is the v1 export.

**Tech Stack:** Go net/http + pgx/PostgreSQL, Next.js App Router + React + TypeScript, existing CSS, Vitest/Testing Library, Go tests.

**Spec:** docs/superpowers/specs/2026-08-18-project-finalization-audit-package-design.md

## Global Constraints

- Finalization requires at least one included outcome and reviewed status for every included outcome.
- closed is shown as Finalized.
- The API enforces finalization readiness and read-only behavior.
- Existing response statuses remain unchanged; submitted is shown as Reviewing and reviewed as Approved.
- Report and Audit Package are read-only for authorized project readers.
- Do not add server-side PDF generation or a new storage service.
- Do not stage output/ or tmp/ artifacts.

---

### Task 1: Add finalization metadata and transactional transition

Files:
- Create db/init/008_project_finalization.sql
- Modify api/internal/store/models.go
- Modify api/internal/store/projects.go
- Create api/internal/store/finalization_test.go
- Modify api/internal/store/projects_integration_test.go

Interfaces:
- Add ErrProjectFinalized and ErrProjectNotReady.
- Add Store.FinalizeProject(ctx, projectID, actorID) (Project, approvedCount, includedCount, error).
- Extend Project with FinalizedAt *time.Time and FinalizedBy *string.

Steps:
- [ ] Write failing integration tests for success, no included outcomes, missing response, submitted, needs_more_info, and a second finalization.
- [ ] Run go test ./api/internal/store -run TestFinalizeProject -count=1 and confirm it fails because the method and fields are missing.
- [ ] Add finalized_at and finalized_by columns plus the partial index in migration 008.
- [ ] Update projectSelect and projectArgs in lockstep with the new Project fields.
- [ ] Implement a transaction that locks the project row, requires in_review, counts included profiles and reviewed responses, updates status='closed' with actor/time, and returns both counts.
- [ ] Run the focused tests and then go test ./api/internal/store -count=1.
- [ ] Commit with message feat: add project finalization transition.

### Task 2: Add API finalization endpoint and mutation lock

Files:
- Modify api/internal/httpapi/handler.go
- Modify api/internal/httpapi/authorization.go
- Modify api/internal/httpapi/responses_handler.go
- Modify api/internal/httpapi/documents_handler.go
- Modify api/internal/httpapi/handler_test.go
- Modify api/internal/httpapi/responses_handler_test.go
- Modify api/internal/httpapi/documents_handler_test.go

Interfaces:
- Add actionFinalizeProject.
- Add POST /api/projects/{id}/finalize.
- Use error codes project_not_ready and project_finalized.

Steps:
- [ ] Write failing HTTP tests for counselor success, assessor forbidden, not-ready conflict, and finalized profile/response/evidence mutation conflicts.
- [ ] Run go test ./api/internal/httpapi -run TestFinalizeProject -count=1 and confirm the expected route/action failure.
- [ ] Allow counselor and counselor_admin to finalize, dispatch the route, call Store.FinalizeProject, write project.finalized, and return the Project JSON.
- [ ] Map ErrProjectNotReady to 409/project_not_ready and ErrProjectFinalized to 409/project_finalized.
- [ ] Add a shared finalized-project check to authorizeProject or ensureOutcomeEditable and apply it to scope, profile, response, review, upload, and delete mutations while keeping GET/download/preview readable.
- [ ] Run go test ./api/internal/httpapi -count=1 and go test ./api/internal/store -count=1.
- [ ] Commit with message feat: finalize projects and lock mutations.

### Task 3: Add Final Report and Audit Package store models

Files:
- Create api/internal/store/reporting.go
- Create api/internal/store/reporting_test.go
- Modify api/internal/store/audit.go
- Modify api/internal/store/models.go

Interfaces:
- Add FinalReport, AuditPackage, ReportOutcome, ScopeRegisterEntry, EvidenceRegisterEntry, AuditTrailEntry, and report summary DTOs.
- Add Store.GetFinalReport(ctx, projectID).
- Add Store.GetAuditPackage(ctx, projectID).
- Add Store.ListProjectAuditEvents(ctx, projectID).

Steps:
- [ ] Write failing integration tests with included/excluded profiles, responses, evidence documents, and audit rows. Assert outcome ordering, excluded filtering, evidence counts, reviewer decisions, scope rationale/assignment, and oldest-first audit events.
- [ ] Run the focused reporting tests and confirm the DTO methods are missing.
- [ ] Implement report queries by reusing GetProject, ListProfile, ListResponses, and the existing coverage calculation. Return empty arrays rather than null.
- [ ] Join audit_logs to users for actor name/email, include action/entity/metadata/timestamp, and never expose storage_key.
- [ ] Build the Audit Package outcome trace from scope -> assignment -> response -> evidence -> review.
- [ ] Run focused tests and go test ./api/internal/store -count=1.
- [ ] Commit with message feat: add final report and audit package data.

### Task 4: Expose JSON and CSV report endpoints

Files:
- Create api/internal/httpapi/reporting_handler.go
- Create api/internal/httpapi/reporting_handler_test.go
- Modify api/internal/httpapi/handler.go
- Modify api/internal/httpapi/handler_test.go

Interfaces:
- GET /api/projects/{id}/final-report.
- GET /api/projects/{id}/audit-package.
- GET /api/projects/{id}/audit-package.csv.

Steps:
- [ ] Write failing tests for authorized readers, unauthorized access, JSON content types, CSV content type/disposition, stable header, and one CSV row per included outcome.
- [ ] Run the focused HTTP tests and verify they fail before implementation.
- [ ] Add a reporting store interface to Handler and JSON handlers that use existing project authorization.
- [ ] Implement CSV with encoding/csv, stable columns, standard escaping, Function/category/outcome ordering, and evidence names rather than private storage keys.
- [ ] Run go test ./api/internal/httpapi -count=1.
- [ ] Commit with message feat: expose final report and audit exports.

### Task 5: Add web API types and methods

Files:
- Modify web/src/lib/types.ts
- Modify web/src/lib/api.ts
- Create web/src/lib/reporting.test.ts

Interfaces:
- api.finalizeProject(id).
- api.getFinalReport(id).
- api.getAuditPackage(id).
- api.downloadAuditPackageCSV(id, signal?) returning Blob.

Steps:
- [ ] Write failing fetch tests for the four paths and methods.
- [ ] Run npm.cmd test -- --run src/lib/reporting.test.ts and confirm failure.
- [ ] Add strongly typed DTOs matching the Go JSON and use the existing request/download helpers.
- [ ] Run the focused test and npx.cmd tsc --noEmit --incremental false.
- [ ] Commit with message feat: add finalization and audit client APIs.

### Task 6: Add Final Report and Audit Package routes

Files:
- Create web/src/components/FinalReport.tsx
- Create web/src/components/AuditPackageView.tsx
- Create web/src/components/FinalReport.test.tsx
- Create web/src/components/AuditPackageView.test.tsx
- Create web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/report/page.tsx
- Create web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/audit/page.tsx
- Modify web/src/app/globals.css

Interfaces:
- FinalReport renders metadata, finalization, summary, Function coverage, included outcomes, responses, evidence, and Print / Save as PDF.
- AuditPackageView renders scope/assignment, evidence register, review history, audit trail, and Download CSV.

Steps:
- [ ] Write failing component tests for the required sections, empty states, print action, and CSV action.
- [ ] Run the focused component tests and confirm missing component failures.
- [ ] Implement both components using existing theme variables, table layouts, readable dates, and print-only rules.
- [ ] Implement both App Router pages with auth redirect, slug resolution, loading/error/retry state, and workspace links.
- [ ] Create a Blob URL for CSV download and trigger a project-slug filename.
- [ ] Add print rules that remove navigation/buttons and keep report tables readable.
- [ ] Run component tests and TypeScript checks.
- [ ] Commit with message feat: add final report and audit package views.

### Task 7: Add workspace readiness and finalized read-only state

Files:
- Modify web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.tsx
- Create web/src/components/ProjectFinalizationPanel.tsx
- Create web/src/components/ProjectFinalizationPanel.test.tsx
- Modify ProfileEditor and StakeholderResponsePanel only where needed for finalized read-only props
- Modify related workspace tests

Interfaces:
- ProjectFinalizationPanel receives included count, approved count, remaining outcome codes, project status, and onFinalize.
- Workspace shows Finalize Project only to counselor/counselor_admin while status is in_review.

Steps:
- [ ] Write failing UI tests for blocked readiness, successful Counselor confirmation, finalized banner, report links, and hidden mutation controls.
- [ ] Run the focused tests and confirm failure before implementation.
- [ ] Calculate approved responses by subcategory ID, distinguish missing/Reviewing/Returned, and call api.finalizeProject from the route page.
- [ ] Refresh project data after finalization and show Final Report and Audit Package links.
- [ ] Pass read-only state to editors for closed projects while preserving evidence preview/download.
- [ ] Run the focused tests, full npm.cmd test -- --run, and TypeScript.
- [ ] Commit with message feat: add finalization readiness to assessment workspace.

### Task 8: Document and verify the complete feature

Files:
- Modify README.md
- Modify DESIGN.md only if the UI map needs the new routes

Steps:
- [ ] Document the finalization guard, read-only behavior, report routes, Audit Package contents, CSV export, browser Print/Save as PDF, and the traceability chain.
- [ ] Run from web: npm.cmd test -- --run, npx.cmd tsc --noEmit --incremental false, npm.cmd run build.
- [ ] Run from repository root: go test ./... and git diff --check.
- [ ] Inspect git status and confirm output/ and tmp/ are not staged and report output has no storage keys or private data.
- [ ] Commit documentation with message docs: document final report and audit package workflow.

