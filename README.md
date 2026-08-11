# NIST CSF Compliance Web App

Lean vertical slice for NIST CSF 2.0 compliance work using Next.js, Go, and PostgreSQL.

## Stack

- **Web:** Next.js 16, React, TypeScript
- **API:** Go
- **Database:** PostgreSQL 16
- **Run:** Docker Compose

## Quick start

Requirements: Docker Desktop must be running.

In PowerShell:

```powershell
Copy-Item .env.example .env
docker compose up --build -d
```

สำหรับ local admin ตามข้อมูลด้านล่าง ให้ตรวจสอบว่า `.env` มีค่าเหล่านี้ก่อนรัน:

```dotenv
BOOTSTRAP_ADMIN_EMAIL=admin@example.com
BOOTSTRAP_ADMIN_PASSWORD=LocalAdmin!2026
```

Open the app at <http://localhost:3000>.

Services:

- Web: <http://localhost:3000>
- API health check: <http://localhost:8080/healthz>
- PostgreSQL: `localhost:5432`

To see logs:

```powershell
docker compose logs -f api
```

To stop the application:

```powershell
docker compose down
```

`docker compose down` stops and removes containers but keeps the PostgreSQL and evidence volumes.

## Local admin login

Use this account to enter the local development environment:

| Field | Value |
| --- | --- |
| Login URL | <http://localhost:3000/login> |
| Email | `admin@example.com` |
| Password | `LocalAdmin!2026` |
| Role | `counselor_admin` |

This is a local development credential only. Do not use it in production or a shared environment.

The bootstrap account is created from `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` in `.env`, and only when the database does not already contain a `counselor_admin`. If an existing PostgreSQL volume has another admin, changing `.env` will not replace that account.

## URLs and routing

The API uses UUIDs internally, while browser URLs use readable slugs:

- Organizations: `/organizations/{organization-slug}`
- Projects: `/organizations/{organization-slug}/projects/{project-slug}`
- Invitation acceptance: `/invite/{token}`

Example:

```text
/organizations/versotis-co-ltd/projects/ru-registration
```

## Version 1 workflow

1. A Counselor Admin signs in and creates a client organization.
2. A Counselor opens the organization and creates an assessment project.
3. The Counselor selects which outcomes are included in the project, records the rationale, and assigns each included outcome to one active stakeholder.
4. Stakeholders activate their accounts from the invitation link. The assigned stakeholder fills Current Profile, Target Profile, Priority, Coverage, the response, and supporting evidence.
5. A Reviewer checks submitted responses. `Reviewed` completes the outcome; `Needs more information` returns it to the assigned stakeholder.
6. Viewers can read the included outcomes and their review status without changing assessment data.
7. Counselors can read all project progress and results, while managing project scope, rationale, and outcome assignments.

### Project setup fields

When a Counselor creates a project, the form captures the assessment context:

- Project name (required)
- Objective / purpose
- Assessment period (for example, `2026`)
- Target completion date
- Scope boundary
- Compliance driver

Organization, Counselor, framework (`NIST CSF 2.0`), slug, status, and created date are managed by the system. Outcome-level scope, rationale, assignments, Current/Target values, evidence, and review decisions are completed in the assessment workflow after the project is created.

Roles:

- `counselor_admin`: manages organizations and counselors.
- `counselor`: creates and manages client projects, scope, rationale, and stakeholder assignments.
- `org_admin`: manages users inside one organization.
- `assessor`: fills assigned outcomes and uploads evidence.
- `reviewer`: reads all included outcomes and is the only final review gate.
- `viewer`: read-only access to included outcomes.

Assigned `org_admin` and `assessor` users only see and edit their assigned included outcomes. Reviewers and viewers can see all included outcomes. Counselor-only scope controls remain unavailable to stakeholder roles.

Evidence files can be previewed inline when they are PDF or PNG/JPG/JPEG. DOCX and XLSX files remain download-only in v1; every evidence file still keeps its Download action.

## Database and migrations

The database initializes the NIST CSF 2.0 catalog from the supplied workbook with 6 Functions, 22 Categories, and 106 Subcategories.

Fresh databases run the SQL files in `db/init` in filename order. The API also ensures slug columns and backfilled slugs on startup for existing databases.

If a PostgreSQL volume existed before authentication was added, apply the auth migration once:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/003_auth_rbac.sql
docker compose restart api
```

If it existed before stakeholder responses were added, apply this migration once as well:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/004_stakeholder_responses.sql
docker compose restart api
```

If it existed before outcome assignments were added, apply this migration once as well:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/006_outcome_assignments.sql
docker compose restart api
```

If it existed before project setup metadata was added, apply this migration once as well:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/007_project_metadata.sql
docker compose restart api
```

Permanent organization deletion is restricted to Counselor Admin and removes organization-owned projects, assessments, stakeholders, invitations, and response metadata after exact-name confirmation. Evidence files are stored in the Docker volume `evidence_data` and cleaned up when their project or organization is deleted.

## Verify locally

```powershell
docker compose config

Set-Location api
go test ./...
Set-Location ..\web
npm install
npm run typecheck
npm test -- --run
npm run build
Set-Location ..
```

Version 1 includes email/password authentication, server-side sessions, organization-scoped projects, slug routing, invitation-based account creation, role access control, counselor profile editing, stakeholder responses, local evidence storage, review status, and coverage summaries. Password reset, notifications, reports, and full action planning are intentionally deferred.
