# Action Plan and Remediation

## Context

The assessment workflow already lets a Counselor define scope, an Assessor document the current and target profiles with evidence, and a Reviewer approve or return each included outcome. Projects can then be finalized and exported as a Final Report or Audit Package.

The missing step is controlled follow-up for approved gaps. An approved outcome may show that current coverage is below target coverage, but the system cannot yet turn that gap into assigned, trackable remediation work.

This feature adds a small Action Plan module linked directly to approved outcomes. It deliberately remains separate from assessment data so a finalized assessment stays immutable while remediation can continue afterward.

## Goals

1. Let a Counselor create one or more remediation actions for an approved outcome whose current coverage is below its target coverage.
2. Let the Counselor assign each action to an active organization stakeholder and set its priority and due date.
3. Let the assigned stakeholder record progress, attach evidence, and submit the action for Counselor review.
4. Let the Counselor close the action or return it for more work.
5. Keep remediation editable after project finalization without reopening or changing the underlying assessment.
6. Surface remediation status in the project workspace, Final Report, Audit Package, and audit trail.
7. Keep v1 small by using one action record, evidence records, and the existing audit log rather than building a general ticketing system.

## Non-goals

- Project versioning or reopening finalized assessment data.
- Dependencies between actions, subtasks, recurring work, or workflow automation.
- Email reminders, scheduled escalation, or SLA configuration.
- Time tracking, cost tracking, comments/chat, or approval chains with multiple reviewers.
- Automatically creating actions without Counselor confirmation.
- A separate Auditor role.

## Recommended approach

Three implementation shapes were considered:

1. Store one remediation note directly on each profile row. This is small but cannot support several actions, owners, or due dates for one outcome.
2. Add dedicated remediation actions linked to project and outcome, plus action evidence. This supports the agreed workflow without becoming a generic work-management system.
3. Build a complete ticketing subsystem with comments, history tables, dependencies, and configurable states. This is flexible but unnecessary for v1.

Use approach 2. Each action has one current progress note and one Counselor review comment. Immutable state changes are already captured by the existing audit log.

## Eligibility and gap calculation

An action may be created only when all of the following are true:

- The outcome is included in the submitted project scope.
- Its stakeholder response has status `reviewed` (shown as **Approved**).
- Current coverage is lower than target coverage.

Coverage is ordered as:

```text
none < partial < substantial < full
```

The server is authoritative for this check. The UI may explain why creation is unavailable, but it cannot bypass the API guard.

One eligible outcome may have any number of actions because policy, process, and technical work may need different owners and due dates.

## Roles and permissions

### Counselor and Counselor Admin

- Read all project actions.
- Create actions for eligible outcomes.
- Edit title, description, desired result, priority, owner, and due date while an action is not closed.
- Return an action from `awaiting_review` to `in_progress` with a required review comment.
- Close an action from `awaiting_review`.
- Reopen a closed action only in a later release; v1 treats closed as final.

### Assigned Org Admin or Assessor

- Read actions in their organization according to existing project access.
- Update the progress note for actions assigned to them.
- Move their action from `open` to `in_progress` by saving progress.
- Upload and delete evidence while the action is `open`, `in_progress`, or `returned`.
- Submit their action for review when a non-empty progress note exists.

### Reviewer and Viewer

- Read actions and evidence only.
- Outcome Reviewer approval remains separate from remediation closure; Counselor is the final remediation gate.

Assignment choices are limited to active users in the project organization with role `org_admin` or `assessor`. The API validates organization, role, and active status.

## Lifecycle

```text
Approved outcome with a coverage gap
  -> Counselor creates action (open)
  -> Assigned stakeholder records progress (in_progress)
  -> Assigned stakeholder submits (awaiting_review)
  -> Counselor returns with comment (in_progress)
       or
     Counselor accepts and closes (closed)
```

The status values are:

- `open`: assigned but work has not been recorded.
- `in_progress`: owner is working or the Counselor returned the action.
- `awaiting_review`: owner submitted the work; owner editing is locked.
- `closed`: Counselor accepted the remediation; all mutation is locked.

Overdue is a derived display state, not a stored status. An action is overdue when its due date is before the current date and its status is not `closed`.

## Project finalization boundary

Finalization continues to lock assessment scope, profile fields, stakeholder responses, review decisions, and assessment evidence.

Remediation actions and remediation evidence remain mutable after the project is finalized. They use dedicated endpoints and are deliberately excluded from the existing finalized-project mutation guard. Every remediation endpoint still performs project access and role authorization.

The Final Report and Audit Package are live views rather than stored snapshots, so their remediation sections show the latest action state even after assessment finalization. The report must label assessment finalization separately from remediation completion to avoid implying that all actions were closed at finalization time.

## Data model

Add migration `db/init/010_remediation_actions.sql`.

### `remediation_actions`

- `id uuid primary key`
- `project_id uuid not null references projects(id) on delete cascade`
- `subcategory_id uuid not null references subcategories(id)`
- `title text not null`
- `description text not null default ''`
- `desired_result text not null default ''`
- `priority text not null` constrained to `low`, `medium`, `high`, `critical`
- `owner_user_id uuid not null references users(id)`
- `due_date date not null`
- `status text not null default 'open'` constrained to `open`, `in_progress`, `awaiting_review`, `closed`
- `progress_note text not null default ''`
- `review_comment text not null default ''`
- `created_by uuid not null references users(id)`
- `submitted_at timestamptz null`
- `closed_by uuid null references users(id)`
- `closed_at timestamptz null`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`

Indexes cover project/status, owner/status, and project/subcategory lookups.

### `remediation_evidence`

- `id uuid primary key`
- `action_id uuid not null references remediation_actions(id) on delete cascade`
- `original_name text not null`
- `storage_path text not null`
- `mime_type text not null`
- `size_bytes bigint not null`
- `uploaded_by uuid not null references users(id)`
- `created_at timestamptz not null default now()`

Evidence uses the same MIME type, size, safe filename, preview, and download rules as stakeholder response evidence.

## API surface

### List and create

- `GET /api/projects/{projectID}/remediation-actions`
- `POST /api/projects/{projectID}/remediation-actions`

The create body contains `subcategoryID`, `title`, `description`, `desiredResult`, `priority`, `ownerUserID`, and `dueDate`. Creation validates eligibility, assignment, required fields, and due-date format.

### Update and transition

- `PATCH /api/projects/{projectID}/remediation-actions/{actionID}` for Counselor-managed fields.
- `PATCH /api/projects/{projectID}/remediation-actions/{actionID}/progress` for the assigned owner's progress note.
- `POST /api/projects/{projectID}/remediation-actions/{actionID}/submit`
- `POST /api/projects/{projectID}/remediation-actions/{actionID}/review` with `decision: close | return` and `comment`.

Return requires a non-empty Counselor comment. Close may include an optional comment. Transitions use conditional updates or row locking so stale requests cannot skip lifecycle states.

### Evidence

- `POST /api/projects/{projectID}/remediation-actions/{actionID}/evidence`
- `GET /api/projects/{projectID}/remediation-actions/{actionID}/evidence/{documentID}`
- `GET /api/projects/{projectID}/remediation-actions/{actionID}/evidence/{documentID}/preview`
- `DELETE /api/projects/{projectID}/remediation-actions/{actionID}/evidence/{documentID}`

The evidence routes reuse existing upload limits and file handling where practical.

## Audit events

Write events to the existing `audit_logs` table:

- `remediation.created`
- `remediation.updated`
- `remediation.progress_updated`
- `remediation.evidence_uploaded`
- `remediation.evidence_deleted`
- `remediation.submitted`
- `remediation.returned`
- `remediation.closed`

Metadata contains changed fields or transition details but never file contents or secrets.

## Frontend behavior

Add an `Action Plan` workspace mode beside the existing assessment work modes.

### Summary

Show compact counts for:

- Open
- In progress
- Awaiting review
- Overdue
- Closed

Counts respect the user's existing project visibility. Overdue can overlap another non-closed status and is displayed as a warning metric.

### Counselor view

- Show approved gap outcomes and their Current → Target coverage.
- Provide `Create action` only for eligible outcomes.
- Present a short form with title, description, desired result, priority, owner, and due date.
- Group or filter actions by status, owner, priority, and outcome.
- Provide edit, return, and close controls according to state.

### Assigned stakeholder view

- Default to `My actions`.
- Show outcome code, rationale, current/target coverage, due date, and Counselor-defined desired result.
- Allow progress editing and evidence upload only when actionable.
- Require a progress note before `Send for review`.

### Read-only view

Reviewer and Viewer see status, ownership, notes, review comment, dates, and evidence without mutation controls.

All views need explicit loading, empty, success, validation, unauthorized, and stale-transition error states. Status must not rely on color alone.

## Reporting

Extend Final Report and Audit Package DTOs with remediation data.

### Final Report

- Overall action counts and overdue count.
- Action table grouped by outcome, including owner, priority, due date, status, and closed date.
- Clear distinction between `Assessment finalized` and `Remediation progress`.

### Audit Package

- Complete remediation register.
- Evidence metadata for each action.
- Remediation audit events in the existing chronological trail.
- CSV columns for action ID, outcome, title, owner, priority, due date, status, submitted date, closed date, review comment, and evidence names.

## Validation and errors

- `409 outcome_not_approved` when the linked response is not approved.
- `409 no_coverage_gap` when Current is equal to or above Target.
- `409 invalid_remediation_transition` for an invalid or stale lifecycle transition.
- `409 remediation_closed` for mutation after closure.
- `403 forbidden` for incorrect role, organization, or action owner.
- `422 validation_error` for missing title, invalid priority/date, invalid owner, or submit without progress.
- Evidence errors follow existing stable upload error behavior.

Deleting an action is intentionally omitted from v1 to preserve traceability. Incorrectly created actions can be corrected before closure; a cancel state can be added later if real usage requires it.

## Test strategy

### Store and domain tests

- Coverage ordering and gap eligibility for every coverage pair.
- Multiple actions may reference one outcome.
- Create rejects unapproved, excluded, and no-gap outcomes.
- Owner must be active, in the project organization, and have an allowed role.
- Lifecycle permits only valid transitions.
- Closed actions are immutable.
- Finalized projects still allow authorized remediation operations.
- Report and audit DTOs include actions and remediation evidence.

### HTTP authorization tests

- Counselor roles can create/edit/review actions.
- Assigned Org Admin/Assessor can update and submit only their own actions.
- Unassigned stakeholders, Reviewer, and Viewer cannot mutate actions.
- Direct endpoint calls cannot bypass eligibility, ownership, status, or project access guards.

### Frontend tests

- Eligible outcome exposes Create action; ineligible outcome explains why it is unavailable.
- Counselor form validates required fields and owner choices.
- Assigned stakeholder can update, attach evidence, and submit.
- Counselor can return with a required comment or close.
- Summary counts and overdue state are calculated correctly.
- Closed and awaiting-review actions hide mutation controls appropriately.
- Finalized projects keep Action Plan controls while assessment controls remain locked.
- Final Report and Audit Package render remediation content.

## Rollout sequence

1. Add domain types, migration, and store tests.
2. Implement store operations and lifecycle guards.
3. Add API authorization and endpoint tests, then handlers.
4. Add remediation evidence tests and handlers by reusing current evidence rules.
5. Add frontend types/API methods and component tests.
6. Add Action Plan workspace UI and role-specific controls.
7. Extend Final Report, Audit Package, and CSV.
8. Update README and run complete Go, web, build, and Docker verification.
