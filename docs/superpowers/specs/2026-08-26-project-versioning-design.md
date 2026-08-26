# Project Versioning and Re-assessment

## Context

The current workflow treats each Project row as one assessment. A Project can
move from Scope setup through Stakeholder assessment and Reviewer approval to
Finalized. Finalization correctly locks assessment data, but the current model
does not provide a clean way to begin the next assessment cycle while keeping
the previous result, evidence, review decisions, Action Plan, and audit trail
available for reference.

Project Versioning adds a small re-assessment flow. A finalized Project can be
copied into the next version under the same Organization. The new version gets
its own assessment data and lifecycle while the previous version remains
read-only. Existing child tables already use `project_id`, so treating each
version as a Project row avoids a broad foreign-key migration.

## Goals

1. Let a Counselor or Counselor Admin start a new assessment from the latest
   Finalized version.
2. Keep every version's scope, responses, evidence, review decisions, Action
   Plan, reports, and audit events isolated by Project ID.
3. Carry forward useful setup context without copying stale Stakeholder
   assessment answers or evidence into the new cycle.
4. Provide a version history and navigation path from the Project workspace.
5. Enforce the version rules in the API and database-backed store, not only in
   the UI.
6. Keep the existing readable slug routing and current Project APIs working.

## Non-goals

- Reopening or editing a Finalized version.
- Comparing two versions field-by-field in v1.
- Copying Stakeholder responses, response evidence, Reviewer decisions, or
  Remediation Actions into a new version.
- Changing the existing Finalization gate.
- Adding a separate version service, event-sourcing model, or generic clone
  framework.
- Automatically sending invitations or changing existing Organization access.

## Recommended approach

Three approaches were considered:

1. **Create an unrelated new Project.** This reuses most existing code, but it
   loses the relationship between assessment cycles and makes history hard to
   understand.
2. **Add version metadata to the existing `projects` table and create a new
   Project row for each cycle.** Existing profiles, responses, evidence, Action
   Plans, reports, and audit logs remain naturally isolated by `project_id`.
   This is the recommended approach for v1.
3. **Create a parent Project and move every child table to an
   `assessment_version_id`.** This gives a pure domain model but requires a
   wide migration and changes every existing endpoint and query without adding
   value needed for the first release.

The implementation uses approach 2. Existing Projects become version 1 records
through a safe backfill. A new version is another Project row with the same
`version_group_id`, an incremented `version_number`, and a pointer to the
previous version.

## Version rules

### Version identity

Add these fields to `projects`:

- `version_group_id uuid not null`: identifies all versions of one assessment
  series.
- `version_number integer not null`: starts at 1 and increments by one.
- `previous_version_id uuid null references projects(id) on delete set null`:
  points to the immediately previous version.

For existing Projects, set `version_group_id` to the Project's own ID,
`version_number` to `1`, and `previous_version_id` to `NULL`. Enforce a unique
constraint on `(version_group_id, version_number)` and indexes for the version
group and previous-version lookup.

The new version keeps the same Project name in the API data model. Its slug is
generated from the source slug plus the version suffix, producing
`ru-registration-v2` unless that slug is already taken. The UI displays the
numeric version label beside the unchanged name, and the API response also
contains the numeric version fields, so the UI does not need to infer the
version from a name or slug.

### Who can create a version

Only `counselor` and `counselor_admin` can create a version. The API must first
authorize access to the source Project's Organization and then enforce the
role. Stakeholder roles and Auditor cannot create versions.

### Which source Project is valid

The source Project must:

- exist and be visible to the requesting Counselor;
- have status `closed` (shown as **Finalized**);
- be the latest version in its `version_group_id`.

The latest-version rule prevents branching from an old version. If the latest
version is still in `setup` or `in_review`, the API returns a conflict and the
user must finish or delete that draft before starting another cycle. The
operation locks the source series while selecting the next number so two
simultaneous requests cannot create the same version number.

### New version state

The new Project starts with status `setup`, no `finalized_at`, and no
`finalized_by`. The Counselor must submit its Scope before Stakeholder users
can see the included outcomes, exactly as in the first assessment cycle.

## Data copied into a new version

The clone operation is transactional. It copies:

- Organization association and Counselor association;
- Project metadata: objective, assessment period, target completion date, scope
  boundary, and compliance driver;
- every Function's `applicable` value and reason;
- every catalog Subcategory profile's `included` value, scope rationale, and
  assigned Stakeholder user.

The following fields are intentionally reset because they are new-cycle
assessment input:

- Current priority, coverage, status text, policies text, and tier;
- Target priority, coverage, approach text, and tier;
- notes and considerations;
- profile review status and review timestamps;
- Stakeholder responses and their response documents;
- Reviewer identity, review timestamps, and review comments.

No Remediation Action or Remediation Evidence is cloned. Existing Actions stay
with the previous version and remain available there after finalization. A
Counselor can create fresh Actions for approved gaps in the new version.

The clone preserves the previous version as the historical baseline. v1 does
not display a field-level comparison; the version history and separate reports
provide traceability.

## API surface

### Create the next version

```text
POST /api/projects/{projectID}/versions
```

Request body: empty JSON object or no body.

Response: `201 Created` with the new `Project`, including version metadata.

Errors:

- `403 forbidden` when the actor is not a Counselor or Counselor Admin;
- `404 not_found` when the Project is outside the actor's accessible scope;
- `409 project_not_finalized` when the source is not closed;
- `409 version_not_latest` when the source is not the latest version;
- `409 version_creation_conflict` when the version number cannot be reserved.

The operation writes one `project.version_created` audit event containing the
source Project ID, new Project ID, source version number, and new version
number. It must never include tokens, file paths, response contents, or
password data.

### List versions

```text
GET /api/projects/{projectID}/versions
```

Returns all versions in the same `version_group_id`, ordered by
`version_number DESC`. Each entry contains the Project ID, slug, name, status,
version number, creation time, finalization time, and whether it is the latest
version. Authorization follows the source Project's Organization/project
access rules. An Auditor receives only versions for Projects to which the
Auditor has explicit access; the endpoint must not reveal unrelated versions.

### Existing endpoints

Existing Project, profile, response, evidence, remediation, report, and audit
endpoints continue to receive a concrete Project ID. No endpoint silently
redirects from an old version to the latest version. This preserves stable
links in audit material.

## Store and transaction boundary

Add a store method with a focused interface:

```go
CreateNextProjectVersion(ctx context.Context, sourceProjectID, actorID string) (Project, error)
ListProjectVersions(ctx context.Context, projectID string) ([]Project, error)
```

`CreateNextProjectVersion` performs all validation, number allocation, Project
insert, Function copy, and profile copy in one transaction. It obtains a row
lock on the source Project and a lock or serializable lookup covering the
version group before calculating `max(version_number) + 1`. The returned
Project is loaded with the existing `projectSelect` shape plus version fields.

The Store returns typed errors for not-finalized, non-latest, and version
allocation conflicts so the HTTP handler can map them to stable status codes.

## Frontend behavior

### Project workspace

Add a compact Version History / Version Switcher near the Project Overview
heading. It shows:

- current version, such as `v2`;
- status badge, such as `Finalized` or `In Review`;
- other versions in descending order;
- a link to each version's existing slug route.

For a Finalized Project, Counselors see `Start new assessment`. The button is
hidden for Stakeholder, Reviewer, Viewer, and Auditor roles. Counselor Admin
and Counselor use the same confirmation dialog:

> Start a new assessment from v1? Scope assignments will be copied. Responses,
> evidence, reviews, and Action Plan items will start empty in v2.

After a successful creation, navigate to the new version's slug URL. The new
workspace opens on Overview and clearly shows `v2 · Setup` so users understand
that the assessment has not been submitted yet.

For a non-latest old version, the UI shows the version as historical and does
not show the create button. The existing Finalized read-only behavior remains
unchanged.

### Organization Project list

Project cards display the version label when `version_number > 1`. Versions
remain separate cards in v1 so existing links and delete behavior are not
silently changed. The card shows the version group context and links to the
selected version.

### Error and loading states

- Disable the create button while the request is in flight.
- On a conflict, keep the current page open and show the server message.
- On success, refresh the version list before navigating.
- If the version list fails, keep the main Project workspace usable and show a
  small retryable error beside the history control.

## Audit and reporting

The new version event is added to the existing project and organization audit
views according to their current authorization. Reports remain version-scoped:
the Final Report and Audit Package for v1 never include v2 responses or actions.
The Project context in both reports includes the version number and, when
available, the previous version label.

## Testing requirements

### Store tests

- Existing Project backfill values are version 1 with a self-contained group.
- Creating from a closed latest version creates v2 with a new ID and setup
  status.
- Project metadata, Function scope, inclusion, rationale, and assignment are
  copied.
- Current/Target assessment fields are reset.
- Responses, documents, Actions, and Action Evidence are not copied.
- Creating from setup or in-review returns the typed not-finalized error.
- Creating from an older version returns the typed non-latest error.
- Two concurrent creation attempts cannot produce duplicate version numbers.
- Version listing is ordered newest first and stays within one version group.

### HTTP/API tests

- Counselor and Counselor Admin receive `201`.
- Assessor, Reviewer, Viewer, and Auditor receive `403` or the existing
  project-not-found behavior when project scoping applies.
- A Stakeholder cannot read versions outside the Organization.
- The audit event records source and new Project IDs without sensitive data.
- Existing Project endpoints still operate against an explicitly selected
  version ID.

### Frontend tests

- Version history renders the current version and links to other versions.
- The Start New Assessment action is visible only to Counselor roles on a
  Finalized latest version.
- Confirmation and loading states work, and success navigates to the returned
  slug.
- Overview remains the first surface after opening a newly created version.
- A failed version request shows a retryable message without blanking the
  assessment workspace.

## Migration and rollout

Add `db/init/012_project_versions.sql` for fresh databases and the matching
`db/migrations/012_project_versions.sql` for existing local databases. The
migration must be idempotent and backfill existing rows before adding the
not-null and unique constraints.

The feature is additive. Existing version-1 Project URLs, child records, audit
events, and deletion cleanup keep their current IDs. No production deployment,
HTTPS, secrets, or backup work is part of this feature.

## Acceptance criteria

1. A finalized latest Project can produce exactly one next setup version at a
   time through the API.
2. The new version has copied scope context and empty assessment input.
3. The previous version remains finalized, read-only, and independently
   reportable.
4. Version history is visible to authorized users and does not leak unrelated
   Projects.
5. API role and lifecycle guards reject direct unauthorized or invalid calls.
6. Audit history identifies who created the version and which versions are
   related.
7. Existing assessment, finalization, remediation, reporting, and audit tests
   remain green.
