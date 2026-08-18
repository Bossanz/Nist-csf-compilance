# NIST CSF Compliance Web App

A lean NIST CSF 2.0 assessment workspace for Counselors and customer Stakeholders. The application is built with Next.js, Go, and PostgreSQL and is designed to keep the assessment workflow short and easy to follow.

Current status: Version 3 local-development workflow slice. The core flow covers Counselor scope setup, Stakeholder assessment, Reviewer approval, evidence preview, and light/dark themes.

## Stack

- Frontend: Next.js 16, React 19, TypeScript
- Backend: Go 1.24
- Database: PostgreSQL 16
- Runtime: Docker Compose
- Authentication: Email/password with server-side sessions
- Evidence storage: Local Docker volume

## Quick start

### Requirements

- Windows and PowerShell
- Docker Desktop with the Docker Engine running

### Start the application

Run these commands from the repository root:

```powershell
Copy-Item .env.example .env
docker compose up --build -d
```

Open:

- Web app: <http://localhost:3000>
- Login: <http://localhost:3000/login>
- API health check: <http://localhost:8080/healthz>
- PostgreSQL: `localhost:5432`

Check container status:

```powershell
docker compose ps
```

Follow API logs:

```powershell
docker compose logs -f api
```

### Stop the application

```powershell
docker compose down
```

This removes containers but keeps the PostgreSQL and evidence volumes.

To reset all local data:

```powershell
docker compose down -v
```

> `docker compose down -v` permanently removes local organizations, projects, users, assessments, and evidence files.

## Local test accounts

### Counselor Admin

The bootstrap account is created from `.env`:

| Field | Value |
| --- | --- |
| Email | `admin@example.com` |
| Password | `LocalAdmin!2026` |
| Role | `counselor_admin` |

The corresponding local configuration is:

```dotenv
BOOTSTRAP_ADMIN_EMAIL=admin@example.com
BOOTSTRAP_ADMIN_PASSWORD=LocalAdmin!2026
```

The bootstrap account is created only when the database does not already contain a `counselor_admin`. Changing `.env` does not replace an existing account.

### Counselor

This account exists in the current local database:

| Field | Value |
| --- | --- |
| Email | `counselor@example.com` |
| Password | `Counselor!2026` |
| Role | `counselor` |

On a fresh database, use the supported invitation flow or create a test account before using this login.

> These credentials are for local development only. Do not use them in production or shared environments.

## System workflow

```text
Counselor Admin
  -> Create client Organization
  -> Open the Organization

Counselor
  -> Create an assessment Project
  -> Set project context
  -> Select which NIST CSF outcomes are in scope
  -> Add a reason for each selected outcome
  -> Assign each included outcome to one Stakeholder assessor
  -> Submit the Project scope

Stakeholder
  -> Activate the invited account
  -> See only included outcomes after scope submission
  -> Complete Current and Target profile fields
  -> Add the response and supporting evidence
  -> Submit the outcome (UI status: Reviewing)

Reviewer
  -> Read the response and evidence
  -> Mark Approved or Needs more information

Counselor
  -> Read progress, reviewer decisions, and final assessment results
```

Project setup starts in `setup`. The Project moves to `in_review` only after the Counselor submits a valid scope. Every included outcome must have one active Stakeholder assessor before submission. An included but unassigned outcome remains hidden from Stakeholder users.

Response status is stored with internal values but shown in reader-friendly UI language:

| Internal status | UI label | Meaning |
| --- | --- | --- |
| `draft` | Draft | Assessor is still preparing the response |
| `submitted` | Reviewing | Assessor sent the response to the Reviewer |
| `reviewed` | Approved | Reviewer accepted the response |
| `needs_more_info` | Needs more information | Reviewer returned it for clarification |

## Roles and permissions

| Role | Main permissions |
| --- | --- |
| `counselor_admin` | Create/delete Organizations and manage Counselors |
| `counselor` | Create/delete Projects, configure scope, add rationale, assign Stakeholders, and read all results |
| `org_admin` | Manage users inside one Organization and complete assigned outcomes |
| `assessor` | Complete assigned Current/Target fields, responses, and evidence |
| `reviewer` | Read included outcomes and perform the final review gate |
| `viewer` | Read included outcomes without editing assessment data |

Important rules:

- Counselors decide which outcomes belong to a Project.
- Counselors can use `Include all` for the currently selected Function.
- Counselors provide the scope rationale and assign one active Stakeholder assessor per included outcome.
- Current/Target profile fields, responses, and evidence belong to the Stakeholder assessor.
- Stakeholders see only included outcomes after the Counselor submits the Project scope.
- Included but unassigned outcomes remain hidden from Stakeholder users.
- Stakeholders cannot change scope, rationale, or assignments.
- Reviewer is the only final review gate.
- Outcomes in Reviewing or Approved status lock Stakeholder profile and evidence edits.

## Coverage calculation

Coverage is calculated from the **Current coverage** of included outcomes only. Target coverage, evidence count, and review status do not change the percentage directly.

```text
None = 0
Partial = 1
Substantial = 2
Full = 3

Coverage % = sum(Current coverage scores)
              / (included outcomes × 3)
              × 100
```

The workspace shows the overall percentage in the assessment summary and a separate percentage for each NIST CSF Function. A Function with no included outcomes displays `0%`.

## Implemented features

- Email/password login and logout
- Server-side session authentication
- Role-based access control
- Organization dashboard
- Project dashboard
- Create and delete Organizations
- Create and delete Projects
- Project context fields:
  - Objective / purpose
  - Assessment period
  - Target completion date
  - Scope boundary
  - Compliance driver
- Readable slug routing
- NIST CSF 2.0 catalog:
  - 6 Functions
  - 22 Categories
  - 106 Subcategories
- Counselor scope configuration
- Include all outcomes within a Function
- Per-Function and overall coverage percentages
- Assignment progress counts:
  - Included
  - Assigned
  - Waiting for assignment
- Outcome-level Stakeholder assignment
- Current Profile and Target Profile
- Priority and Coverage fields
- Stakeholder response workflow
- Evidence upload and delete
- Inline preview for PDF and PNG/JPG/JPEG
- Download fallback for DOCX/XLSX
- Light/dark theme toggle with system-theme fallback
- Reviewer decisions:
  - Approved
  - Needs more information
- Coverage summary
- Audit log
- Invitation-based account activation
- Organization and Project deletion with evidence cleanup

## Routes

The API uses UUIDs internally. Browser URLs use readable slugs:

```text
/login
/organizations
/organizations/{organization-slug}
/organizations/{organization-slug}/projects/{project-slug}
/invite/{token}
```

Example:

```text
/organizations/versotis-co-ltd/projects/ru-registration
```

## Project setup fields

When a Counselor creates a Project, the form captures:

- Project name (required)
- Objective / purpose
- Assessment period
- Target completion date
- Scope boundary
- Compliance driver

Organization, Counselor, framework (`NIST CSF 2.0`), slug, status, and created date are managed by the system. Outcome scope, rationale, assignment, Current/Target values, evidence, and review status are completed in the assessment workflow.

### Scope and assignment example

For a project such as `RU Registration Readiness`, the Counselor can include `GV.OC-01`, `DE.AE-02`, `DE.AE-08`, and `DE.CM-09`, explain why each outcome is relevant, and assign each one to a Stakeholder assessor. The assessor then completes the Current/Target profile and evidence for the assigned outcomes; the Reviewer makes the final decision.

## Repository structure

```text
.
|-- api/                         # Go API
|   |-- cmd/server/              # API entrypoint
|   `-- internal/
|       |-- auth/                # Login, password, sessions, invitations
|       |-- domain/              # Business rules and calculations
|       |-- httpapi/             # HTTP handlers and authorization
|       `-- store/               # PostgreSQL queries and models
|-- db/init/                     # Schema, seed data, and migrations
|-- web/
|   `-- src/
|       |-- app/                 # Next.js routes
|       |-- components/          # UI components
|       `-- lib/                 # API client, types, and route helpers
|-- docker-compose.yml
|-- .env.example
`-- README.md
```

## Database and persistent data

Fresh databases run the files in `db/init` in filename order:

```text
001_schema.sql
002_seed.sql
003_auth_rbac.sql
004_stakeholder_responses.sql
005_slug_routing.sql
006_outcome_assignments.sql
007_project_metadata.sql
```

Docker volumes:

- `pgdata` stores PostgreSQL data.
- `evidence_data` stores uploaded evidence files.

For an existing database volume, apply any migration that was added after the volume was created, then restart the API. For example:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/006_outcome_assignments.sql
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/007_project_metadata.sql
docker compose restart api
```

## Verify before committing

### Go API

```powershell
Set-Location api
go test ./...
Set-Location ..
```

### Web

```powershell
Set-Location web
npm.cmd ci
npx.cmd tsc --noEmit --incremental false
npm.cmd test -- --run
npm.cmd run build
Set-Location ..
```

The design reference is maintained in [DESIGN.md](DESIGN.md), with the Impeccable sidecar at `.impeccable/design.json`.

### Docker health check

```powershell
docker compose up --build -d
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:3000
```

Expected results:

- API health check returns HTTP `200`.
- Web returns HTTP `200`.
- PostgreSQL and API are healthy.

## Troubleshooting

### Docker API connection error

If you see an error such as:

```text
failed to connect to the docker API
Docker daemon is not running
```

Open Docker Desktop, wait until the Docker Engine is running, then retry:

```powershell
docker compose up --build -d
```

### Failed to fetch

Check the services in this order:

```powershell
docker compose ps
docker compose logs -f api
Invoke-WebRequest http://localhost:8080/healthz
```

If the API is not healthy, check PostgreSQL and migrations first.

### `invalid profile assignment`

The intended workflow is:

1. Counselor includes the outcome.
2. Counselor saves the assessment.
3. Counselor selects a Responsible stakeholder.
4. Counselor saves the assessment again.

The selected user must be active, belong to the same Organization, and have role `org_admin` or `assessor`.

## Not included in the current local slice

- Password reset
- Real email notifications
- Report export
- Full action planning
- Production deployment configuration
- Inline DOCX/XLSX preview

These can be added later without changing the core assessment workflow.
