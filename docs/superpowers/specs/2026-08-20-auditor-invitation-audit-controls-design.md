# Auditor Access, Invitation Lifecycle, and Audit Controls

## Goal

Add accountable, read-only Auditor access to selected Projects; give Organization Admins control over invitation lifecycle; and expand the append-only audit trail so material security and assessment actions are traceable without changing the existing assessment workflow.

## Scope

This increment includes:

- A new `auditor` stakeholder role.
- Project-scoped Auditor access.
- Organization Admin invitation, resend, cancel, and expiry handling.
- Invitation project assignments carried into Auditor access after acceptance.
- Read-only Auditor access to assessment data, evidence, reports, and project audit history.
- Organization and Project audit-log read views for authorized internal users.
- Audit events for authentication, invitations, access changes, assessment changes, evidence operations, reports, finalization, and remediation.
- API authorization and UI tests for the new role and lifecycle.

This increment does not include SSO, MFA, external identity providers, production deployment hardening, report versioning, or a general-purpose notification inbox.

## Approved decisions

### Auditor access

An Auditor is a stakeholder account with role `auditor`. An Organization Admin invites the Auditor and selects one or more Projects. The Auditor can read only those Projects. Access is read-only and includes Projects in `setup`, `in_review`, and `closed` states so an audit can happen before or after finalization.

An Auditor can read:

- Project context, scope rationale, and assignments.
- Included outcomes and Current/Target profiles.
- Stakeholder responses and reviewer decisions.
- Evidence metadata and authorized evidence preview/download.
- Final Report and Audit Package.
- The chronological audit trail for an authorized Project.

An Auditor cannot:

- Change scope, assignments, profiles, responses, reviews, or evidence.
- Create, edit, submit, return, or close remediation Actions.
- Finalize, delete, or create Projects.
- Invite, disable, or change users.
- Read another Project without an active Project access grant.

The existing Counselor access model remains unchanged so Counselors can support every client Organization. The Company-side `org_admin` is the only role that may create, resend, or cancel an Auditor invitation.

### Invitation lifecycle

Invitation status is derived from stored timestamps and token state:

```text
Pending -> Accepted
   |\
   | +-> Cancelled
   | +-> Expired
   | +-> Superseded (after resend)
```

- New invitations expire 72 hours after creation, matching the current behavior.
- Resend creates a new one-time token and marks the previous invitation `superseded`; the old token cannot be accepted.
- Cancel marks a pending invitation `cancelled`; it cannot be accepted or resent without creating a new invitation.
- Expired invitations remain visible in history and cannot be accepted.
- Accepted, cancelled, superseded, and expired invitations are not considered active duplicates.
- An Auditor invitation must include at least one Project belonging to the Organization.
- Accepting an Auditor invitation creates the user and its Project access grants in one transaction.

### Auditability

Audit records are append-only from the application surface. No API or UI may update or delete an audit record. Each event records the actor, actor role when known, Organization/Project context, action, entity, result, timestamp, and structured metadata. Request correlation and safe request metadata may be included; passwords, session tokens, invitation tokens, and binary file contents must never be stored.

Automatic expiry is a derived state, not a background job. The system records an `invitation.expired` event when an expired invitation is observed during listing or acceptance, without repeatedly creating duplicate events for the same invitation.

## Data model

### Users

Extend the existing role and ownership constraints to allow:

```text
user_type = stakeholder
role = auditor
organization_id IS NOT NULL
```

### Invitations

Keep the existing `invitations` table and add cancellation/supersession fields:

- `cancelled_at timestamptz NULL`
- `cancelled_by uuid NULL REFERENCES users(id)`
- `superseded_at timestamptz NULL`
- `superseded_by uuid NULL REFERENCES invitations(id)`

The token remains hashed. The API never returns the raw token except in the existing local invitation URL response used for manual development activation.

### Invitation Project Access

Add `invitation_project_access`:

```text
invitation_id uuid REFERENCES invitations(id) ON DELETE CASCADE
project_id uuid REFERENCES projects(id) ON DELETE CASCADE
PRIMARY KEY (invitation_id, project_id)
```

Only Auditor invitations require rows in this table. Other stakeholder invitations retain their current organization-level behavior.

### Project Auditor Access

Add `project_auditor_access`:

```text
project_id uuid REFERENCES projects(id) ON DELETE CASCADE
user_id uuid REFERENCES users(id) ON DELETE CASCADE
granted_by uuid REFERENCES users(id)
created_at timestamptz NOT NULL DEFAULT now()
revoked_at timestamptz NULL
PRIMARY KEY (project_id, user_id)
```

The first version does not add a generic membership abstraction. This focused table keeps the Auditor boundary explicit and avoids changing existing Stakeholder assignment semantics.

### Audit logs

Extend the existing `audit_logs` table with structured fields needed by the timeline and filters:

- `actor_role text NULL`
- `result text NOT NULL DEFAULT 'success'`
- `request_id text NULL`
- `ip_address inet NULL`
- `user_agent text NULL`

Existing metadata remains JSON and is used for before/after values, safe labels, counts, and reason/comment fields. Sensitive values are redacted before insertion.

## API contract

### Auditor access

- Existing project GET/report/evidence read routes accept an Auditor only when an active `project_auditor_access` row exists.
- All mutation routes reject Auditor with `403 forbidden`.
- `GET /api/projects/{projectID}/audit-log` returns the chronological audit events for an authorized Project.
- `GET /api/organizations/{organizationID}/audit-log` is available to the Counselor, Counselor Admin, and Organization Admin for that Organization. It is not available to Auditor.

### Invitations

Extend the existing organization invitation request with an optional `projectIDs` array. For `auditor`, it is required and every Project must belong to the target Organization.

Add:

```text
GET  /api/organizations/{organizationID}/invitations
POST /api/organizations/{organizationID}/invitations/{invitationID}/resend
POST /api/organizations/{organizationID}/invitations/{invitationID}/cancel
```

Organization Admins may manage invitations for their Organization. Counselors may read invitation state for support but may not create Auditor invitations. Existing Counselor and Organization Admin invitation behavior for non-Auditor stakeholder roles remains compatible unless the role guard explicitly requires the Company-side Organization Admin.

Invitation responses expose derived status and assigned Project summaries, never token hashes.

## Audit event catalog

The implementation must emit, at minimum, these events:

| Area | Events |
| --- | --- |
| Authentication | `auth.login_succeeded`, `auth.login_failed`, `auth.logout`, `auth.password_changed`, `auth.password_reset` |
| Invitations | `invitation.created`, `invitation.resent`, `invitation.cancelled`, `invitation.expired`, `invitation.accepted` |
| Access | `user.role_changed`, `user.status_changed`, `project.auditor_granted`, `project.auditor_revoked` |
| Organization/Project | `organization.created`, `organization.deleted`, `project.created`, `project.deleted`, `project.finalized` |
| Scope/Assessment | `scope.updated`, `scope.submitted`, `outcome.assignment_changed`, `profile.updated`, `response.draft_saved`, `response.submitted`, `response.reviewed` |
| Evidence | `response.evidence_uploaded`, `response.evidence_previewed`, `response.evidence_downloaded`, `response.evidence_deleted`, `remediation.evidence_uploaded`, `remediation.evidence_previewed`, `remediation.evidence_downloaded`, `remediation.evidence_deleted` |
| Reporting | `report.viewed`, `audit_package.viewed`, `audit_package.downloaded` |
| Remediation | `remediation.created`, `remediation.updated`, `remediation.progress_updated`, `remediation.submitted`, `remediation.returned`, `remediation.closed` |

Read events are recorded only for the report, audit package, audit-log, and evidence preview/download endpoints; ordinary catalog and health reads remain unlogged to avoid noise.

## UI behavior

### Organization workspace

Add an Invitations section that shows email, role, derived status, expiry, invited-by, and assigned Projects. Organization Admins see Resend and Cancel controls. Auditor creation presents a required multi-select Project list. The existing one-time invitation URL remains available for local development.

### Auditor workspace

Reuse the existing Project workspace and report routes with a clear `Auditor / Read only` mode. Hide mutation controls in the UI, while API guards remain authoritative. The Audit page includes an Activity Timeline backed by the Project audit-log endpoint.

### Internal audit view

Counselors and Organization Admins can open the organization or Project audit timeline, filter by actor/action/date, and inspect structured metadata. The first version does not add bulk export beyond the existing Audit Package CSV.

## Authorization rules

1. An Auditor must have `status = active` and an active Project access grant.
2. Organization Admins may invite Auditor accounts only for their own Organization.
3. Project access grants may reference only Projects in the inviter's Organization.
4. A disabled Auditor immediately loses access without deleting historical logs.
5. Revoking a grant immediately blocks Project reads but preserves audit history.
6. A Project or Organization deletion cascades access grants and evidence as current deletion behavior already does.
7. Audit-log reads are subject to the same Organization/Project boundary as the underlying data.

## Error handling

- Invalid or expired invitation acceptance returns the existing `invalid_invitation` response and never creates a partial user/access record.
- Resend on accepted, cancelled, or superseded invitation returns `409 invitation_not_pending`.
- Cancel on accepted or already cancelled invitation returns `409 invitation_not_pending`.
- Auditor invitation without Project IDs returns `400 invalid_project_access`.
- Auditor access to an ungranted Project returns `404 not_found` to avoid leaking Project existence.
- Auditor mutation attempts return `403 forbidden`.
- Email delivery remains best-effort and does not roll back a successful invitation state change.

## Acceptance criteria

- A Company Organization Admin can invite an Auditor with one or more selected Projects.
- The Auditor can accept, set a password, log in, and see only the selected Projects.
- The Auditor can read the complete selected Project assessment and audit trail but cannot mutate any assessment, evidence, report, or remediation record.
- Resend invalidates the old invitation token; Cancel blocks acceptance; expiry is visible and blocks acceptance.
- Every event in the catalog is written once with actor/context/result metadata and no secret values.
- Counselors and authorized Organization Admins can view the appropriate organization/project audit timeline.
- Existing Counselor, Assessor, Reviewer, Viewer, Finalization, Remediation, Report, and evidence workflows continue to pass their existing tests.
- The implementation remains compatible with existing PostgreSQL volumes through an idempotent migration.
