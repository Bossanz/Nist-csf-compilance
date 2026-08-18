# Project Finalization, Final Report, and Audit Package

## Context

The assessment workspace already supports Counselor scope selection and assignment, Stakeholder/Assessor responses and evidence, and Reviewer decisions. The database already allows the project status `closed`, but there is no project-level finalization action, immutable read-only state, final report, or audit export.

This feature closes the assessment workflow and produces two related outputs:

- **Final Report**: a human-readable, print-friendly summary for the client, Counselor, and management.
- **Audit Package**: a traceable evidence index and activity history that lets an Auditor follow each included outcome from scope through response, evidence, review, and finalization.

## Goals

1. Allow the assigned Counselor (and Counselor Admin) to finalize a project only after every included outcome has been approved by a Reviewer.
2. Make a finalized project read-only through both the UI and API.
3. Provide a print-friendly report route with a browser `Print / Save as PDF` action.
4. Provide an audit view/export containing scope decisions, assignments, responses, evidence metadata, reviewer decisions, and audit events.
5. Preserve the existing internal statuses while presenting readable labels in the UI: `closed` is shown as `Finalized`.
6. Keep the v1 implementation small: reuse current project, profile, response, document, summary, and audit-log data rather than introducing a report-generation service.

## Non-goals

- Server-side PDF generation or permanent PDF storage.
- A separate report snapshot/versioning service.
- Action-plan/remediation management.
- Notifications or scheduled audit delivery.
- Changing the existing outcome review workflow.

## Workflow

```text
setup
  -> Counselor selects included outcomes, writes rationale, assigns Assessor
  -> Counselor submits scope
in_review
  -> Assessor saves/submits response and evidence
  -> Reviewer approves or returns each included outcome
  -> Counselor reviews final readiness
closed (shown as Finalized)
  -> Final Report and Audit Package are available read-only
```

The finalization guard requires:

- Project status is `in_review`.
- At least one outcome is included.
- Every included outcome has a response with status `reviewed`.
- No included outcome is missing a response or still has `submitted`, `draft`, or `needs_more_info` status.

The API performs these checks inside one transaction. The UI readiness summary is informative only and cannot replace the server guard.

## Authorization

- `counselor` and `counselor_admin` may finalize a project they can access. The project Counselor is the expected finalizer; Counselor Admin retains administrative access.
- `counselor`, `counselor_admin`, `org_admin`, `assessor`, `reviewer`, and `viewer` may read the Final Report and Audit Package according to existing project visibility rules.
- Only `reviewer` may approve or return responses, as in the current workflow.
- After finalization, all profile, response, review, evidence upload/delete, and scope mutation endpoints return a conflict response with a stable `project_finalized` error code.

## Data model

Add a migration after the current migrations:

```sql
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS finalized_at timestamptz,
  ADD COLUMN IF NOT EXISTS finalized_by uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_projects_finalized_at
  ON projects(finalized_at)
  WHERE finalized_at IS NOT NULL;
```

Extend the API `Project` model with:

- `finalizedAt: string | null`
- `finalizedBy: string | null`

The existing `audit_logs` table is the source of finalization and mutation history. Finalization writes `project.finalized` with the finalizer and included/approved counts in metadata.

## API surface

### Finalize

`POST /api/projects/{projectID}/finalize`

Response: the updated `Project`.

Success behavior:

- Lock the project row.
- Recheck readiness and included outcome count.
- Set `status='closed'`, `finalized_at=NOW()`, and `finalized_by=current user`.
- Write the `project.finalized` audit event.
- Return the updated project.

Failure behavior:

- `409 project_not_ready` when any included outcome is not approved.
- `409 project_finalized` when the project is already closed.
- `403 forbidden` when the role cannot finalize.

### Final report

`GET /api/projects/{projectID}/final-report`

Returns a report DTO composed from current immutable project data:

- Project and organization metadata.
- Finalization identity and timestamp.
- Overall summary and per-Function coverage.
- Included outcomes with profile values.
- Stakeholder response and reviewer decision.
- Evidence references and counts.

The endpoint is read-only and can serve a report before finalization for Counselor readiness review, but the UI labels it `Draft report` until the project is finalized. The finalized report shows `Finalized`.

### Audit package

`GET /api/projects/{projectID}/audit-package`

Returns a structured JSON DTO with:

- `project`: metadata, status, finalized-by/date.
- `scope`: included/excluded decisions, rationales, assignments.
- `outcomes`: profile, response, evidence register, and reviewer decision for each included outcome.
- `summary`: overall and per-Function coverage/counts.
- `auditTrail`: ordered audit events with actor, action, entity, metadata, and timestamp.

`GET /api/projects/{projectID}/audit-package.csv`

Returns a CSV evidence/outcome register suitable for an Auditor to filter. Each row contains project, Function, category, outcome code, inclusion, rationale, assignee, response status, response timestamp, reviewer, review timestamp, reviewer comment, evidence count, and evidence names/types/sizes.

The v1 package does not copy binary evidence files. The evidence register links each file to its response and exposes the existing preview/download action to authorized users.

## Audit traceability

The audit package must make this chain visible for every included outcome:

```text
Scope decision
  -> Assigned Assessor
  -> Stakeholder response
  -> Evidence metadata
  -> Reviewer decision/comment
  -> Finalization
```

Existing audit events are retained. New finalization and relevant existing mutation events should be surfaced in chronological order. No event is deleted when a project is finalized.

## Frontend behavior

### Workspace

- Add a Counselor-only `Finalize Project` action when the project is in `in_review`.
- Show a readiness panel with included count, approved count, remaining outcomes, and a clear blocked reason.
- Confirm before finalizing.
- On success, refresh project data and show links to `Final Report` and `Audit Package`.
- For `closed`, replace editing controls with a finalized banner and read-only content.

### Final Report route

Add:

`/organizations/{organizationSlug}/projects/{projectSlug}/report`

The route uses a clean, print-friendly layout and includes:

- Report title and project metadata.
- Finalization status, finalizer, and date.
- Executive summary cards.
- Function coverage table.
- Included outcome result table.
- Response/evidence summary.
- Reviewer decision summary.
- `Print / Save as PDF` button hidden in print media.

### Audit Package route

Add:

`/organizations/{organizationSlug}/projects/{projectSlug}/audit`

The route provides:

- Audit readiness and finalization summary.
- Scope and assignment register.
- Evidence register.
- Review history.
- Chronological audit trail.
- `Download CSV` button for the evidence register.
- Print support for the complete audit view.

Both routes remain readable after finalization. The workspace itself remains the editing surface before finalization and a read-only reader after it.

## Error and empty states

- Finalize button is disabled while a request is in progress.
- If outcomes remain, show the outcome codes and missing status in the readiness panel.
- If there are no included outcomes, explain that scope must contain at least one outcome.
- If the project is already finalized, redirect or link to the Final Report rather than showing an editable form.
- If an evidence file is unavailable, retain its register row and show an unavailable preview/download state.
- If no audit events exist, show a clear empty state rather than omitting the section.

## Test strategy

### Store/API tests

- Finalization succeeds only when all included outcomes are `reviewed`.
- Finalization rejects missing, draft, submitted, or needs-more-info outcomes.
- Finalization is idempotently rejected after the project is closed.
- Finalization records `finalized_at`, `finalized_by`, and an audit event.
- Mutations are rejected after finalization.
- Report DTO contains included outcomes, coverage, responses, and evidence metadata.
- Audit DTO orders events and exposes the complete outcome trace.
- CSV output includes a row per included outcome.

### Frontend tests

- Counselor sees readiness and Finalize action.
- Counselor cannot finalize while outcomes remain.
- Finalized workspace is read-only and links to both reports.
- Final Report renders metadata, coverage, outcomes, and print action.
- Audit Package renders evidence register and review history.
- CSV download action targets the project audit endpoint.

## Rollout

1. Add migration and store/report DTOs.
2. Add failing tests, then implement finalize guard and mutation lock.
3. Add report and audit endpoints.
4. Add frontend routes and workspace actions.
5. Run API and web tests, TypeScript, Go tests, and production build.

