# NIST CSF Compliance Web App Design

**Date:** 2026-08-05  
**Status:** Approved direction; pending written-spec review  
**Goal:** Build a fast, runnable Docker Compose vertical slice for NIST CSF 2.0 project assessment, Current/Target Profile editing, and coverage reporting.

## Scope

Version 1 includes:

- Next.js web app with Dashboard and Project Profile screens
- Go REST API for health, catalog, project, profile, and summary endpoints
- PostgreSQL schema for the catalog, projects, and profile data needed by the vertical slice
- Docker Compose services: `web`, `api`, and `postgres`
- Catalog seed from the supplied workbook: 6 Functions, 22 Categories, and 106 Subcategories
- Coverage and gap calculations from the system specification
- Unit and service tests for important behavior

Version 1 does not implement real authentication, email delivery, evidence object storage, PDF reports, or the full action-plan UI. Those areas remain explicit future boundaries rather than fake completed features.

## Reference Inputs

- `C:\Users\CHAWAKAN\Downloads\compliance_webapp_spec_en.md`
- `C:\Users\CHAWAKAN\Downloads\CSF 2.0 Organizational Profile Template Thai R1.xlsx`
- `C:\Users\CHAWAKAN\Downloads\compliance_full_flow_v2.mermaid`
- `C:\Users\CHAWAKAN\Downloads\compliance_er_diagram.mermaid`

The workbook contains `Column Descriptions` and `Current and Target Profile` sheets. The profile has 16 columns and the complete 6/22/106 CSF hierarchy. Thai text from the workbook must be stored as UTF-8 and returned without loss.

## Architecture

```text
Browser
  |
  v
Next.js web (port 3000)
  |
  | REST/JSON
  v
Go API (port 8080)
  |
  v
PostgreSQL (port 5432)
```

Use a lean monorepo:

- `web/`: UI, loading/error states, and REST client
- `api/`: HTTP handlers, validation, service logic, repositories, and calculations
- `db/`: migrations and seed data
- Root: Docker Compose, environment example, and run documentation

The Go API uses the standard library HTTP server and a PostgreSQL driver. No ORM or heavy framework is required. Handler/service/repository boundaries are introduced only where they improve testability.

## Version 1 Data Model

Core tables:

- `functions(id, code, name, description)`
- `categories(id, function_id, code, name, description)`
- `subcategories(id, category_id, code, description)`
- `organizations(id, name, type)`
- `users(id, organization_id, name, email, user_type)`
- `projects(id, organization_id, counselor_id, name, status, created_at)`
- `project_functions(id, project_id, function_id, applicable, reason)`
- `project_subcategory_profiles(id, project_id, subcategory_id, included, rationale, current_priority, current_coverage_level, current_status_text, current_policies_text, current_tier, target_priority, target_coverage_level, target_approach_text, target_tier, notes, considerations, review_status, submitted_at, reviewed_by, reviewed_at)`

Constraints:

- Catalog codes are unique.
- `(project_id, function_id)` is unique.
- `(project_id, subcategory_id)` is unique.
- Coverage levels are `none`, `partial`, `substantial`, or `full`.
- Review status is `draft`, `submitted`, `approved`, or `rejected`.
- Project status is `setup`, `in_review`, `gap_analysis`, `reporting`, or `closed`.

Creating a project creates all 6 project-function rows and all 106 project-profile rows in one transaction.

The v1 schema does not need responsible-person, evidence, action-plan, notification, or report tables yet, but its IDs and boundaries must allow those tables to be added later.

## API Contract

All responses are JSON. Errors use:

```json
{
  "error": {
    "code": "validation_error",
    "message": "project name is required"
  }
}
```

Endpoints:

- `GET /healthz` → `{ "status": "ok" }`
- `GET /api/functions` → functions with categories and subcategories
- `POST /api/projects` with `{ "name": "...", "organizationName": "..." }`
- `GET /api/projects/:id` → project and project-function setup
- `GET /api/projects/:id/profile` → profile rows with hierarchy context
- `PUT /api/projects/:id/profile/:subcategoryId` → partial update of one profile row
- `GET /api/projects/:id/summary` → project percentage, function percentages, and counts

Profile updates validate every supplied field and enum. Clients cannot change `subcategory_id` or move a row to another project.

## Calculation Rules

```text
none = 0
partial = 1
substantial = 2
full = 3
```

For included subcategories:

```text
coverage_pct = average(current_score / 3) * 100
gap = target_score - current_score
```

Function percentage uses only included subcategories within that Function. Project percentage uses only included subcategories across the project. If there are no included rows, return percentage 0 and count 0 rather than divide by zero. Excluded rows do not enter any denominator or pending count.

## Data Flow

1. Browser posts a new project.
2. API transaction creates the organization/project and project-scoped rows from the catalog.
3. Browser loads the project profile and renders Function → Category → Subcategory navigation.
4. User edits one subcategory and sends a PUT request.
5. API validates, updates, and returns the latest row.
6. Dashboard refreshes the summary after save.
7. Future review, evidence, action-plan, and report features reuse the same project/profile IDs.

## UI Direction

Keep the first UI functional and compact:

- Function sidebar with progress status
- Searchable/filterable profile table or form
- Summary cards for overall coverage, included count, pending count, and rejected count
- Saving, success, and error states for every mutation
- Basic desktop/tablet responsiveness

No specific visual theme is required by the source spec.

## Error Handling

- Malformed JSON → HTTP 400 `invalid_json`
- Invalid values or enums → HTTP 400 `validation_error`
- Missing resource → HTTP 404 `not_found`
- Database conflict → HTTP 409 `conflict`
- Internal/database failure → HTTP 500 `internal_error`, with server logs and safe client text
- API requests have timeouts; rows and transactions are closed on every path
- Web errors are readable and do not discard unsaved form values

## Testing Strategy

Use TDD for new behavior:

- Go unit tests for coverage mapping, gap, rollups, zero-included behavior, and invalid enums
- Go handler/service tests for project creation and profile validation
- Docker smoke tests for health, catalog, and project creation
- Web tests for summary rendering, profile save state, and API error state
- Final verification commands: `go test ./...`, web lint/typecheck/build, `docker compose config`, and service smoke tests

Each logic behavior must have a failing test observed before production implementation.

## Local Run

Target command:

```bash
docker compose up --build
```

Expected URLs:

- Web: `http://localhost:3000`
- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

Local connection settings belong in `.env.example`; real secrets must not be committed.

## Future Extension Boundaries

- Authentication middleware around API handlers
- Responsible-person join tables from the ER diagram
- Evidence adapter backed by object storage, not PostgreSQL binary columns
- Notification service and `notification_log`
- Many-to-many action plan links
- Report service using a project snapshot

Future features must keep calculation rules in the service layer and preserve the summary contract.

