# Scope Submission Design

## Goal

Let the Counselor configure the complete Project scope first, then publish all selected outcomes to Stakeholders with one action.

## Approved workflow

    Project status: setup
      Counselor selects outcomes
      Counselor adds rationale
      Counselor assigns an active Assessor
      Stakeholder response/profile/evidence fields stay hidden

    Submit scope
      Project status changes to in_review
      Selected outcomes become available to Stakeholders
      Counselor can read progress and final results

The existing Project status is reused:

- setup means Scope draft.
- in_review means Scope submitted and available to Stakeholders.

This avoids adding a second project-level status column for the v1 workflow.

## Options considered

1. Reuse projects.status (setup to in_review) — selected. It uses the existing schema, needs one endpoint, and keeps the implementation small.
2. Add scope_status or scope_submitted_at — rejected for now because it duplicates the existing lifecycle field and requires a database migration.
3. Client-only visibility — rejected because a Stakeholder could still call the profile and response endpoints directly.

## API behavior

- Add POST /api/projects/{projectID}/scope/submit.
- Only counselor and counselor_admin can submit.
- The operation is idempotent when the Project is already in_review.
- The server rejects direct Stakeholder access to Projects still in setup.
- Stakeholder Project lists and slug resolution hide Projects still in setup.
- The server records project.scope_submitted in the existing audit log.

## UI behavior

- Counselor draft view shows only scope inclusion, rationale, and Assessor assignment for each outcome.
- Counselor sees a single Submit scope action while the Project is in setup.
- Counselor can still see Stakeholder response/profile/evidence content after the scope is submitted.
- Stakeholder-facing outcome rows and response/profile/evidence controls render only after the Project is in_review.
- Existing response statuses remain unchanged; this feature only controls when Stakeholder work becomes visible.

## Validation

- Scope submission requires at least one included outcome.
- Every included outcome must have an active Assessor assignment.
- Rationale remains editable text and is not made mandatory in this iteration.
- Existing per-outcome profile assignment validation remains the source of truth for valid Assessor accounts.

## Testing
