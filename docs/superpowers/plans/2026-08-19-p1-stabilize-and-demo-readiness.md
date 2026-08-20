# P1 Stabilize and Demo Readiness Implementation Plan

> **For agentic workers:** Execute this plan inline with a red-green verification checkpoint for each behavior. Keep the existing local-development scope; do not add production deployment, HTTPS, secrets management, or PostgreSQL backup.

**Goal:** Make the current assessment and remediation workflow repeatable on existing Docker volumes, executable through one authenticated smoke test, and easy to demonstrate or audit locally.

**Architecture:** Add a small one-shot PostgreSQL migration service to Compose. It tracks only application migrations in `schema_migrations` and runs before the API, so a pre-existing volume receives new migrations without being deleted. Add a local-only Go password reset command that reuses the existing password hashing and session revocation logic. Replace the legacy unauthenticated smoke script with an authenticated PowerShell workflow that creates temporary users through the real invitation API, drives the assessment/review/finalization/remediation lifecycle, and optionally keeps the resulting sample data.

**Tech Stack:** Docker Compose, PostgreSQL 16, Go 1.24, pgx, bcrypt, PowerShell, existing Go and Vitest test suites.

**Spec:** `docs/superpowers/specs/2026-08-19-action-plan-remediation-design.md`

## Global Constraints

- Preserve the current role model: Counselor configures scope and remediation; Stakeholder fills assigned assessment work; Reviewer approves assessment responses.
- Keep `Project` finalization read-only for assessment data while allowing remediation updates after finalization.
- Do not add a production password-reset API or expose database credentials in application routes.
- Do not delete or recreate the existing PostgreSQL volume during verification.
- Preserve existing untracked `output/` and `tmp/` directories.

---

### Task 1: Automatic migration for existing Docker volumes

**Files:**
- Create: `db/migrations/010_remediation_actions.sql`
- Create: `db/migrations/run.sh`
- Create: `scripts/migration-smoke-test.ps1`
- Modify: `docker-compose.yml`
- Modify: `README.md`

**Interfaces:**
- `migrate` Compose service runs `/migrations/run.sh` after `postgres` is healthy.
- `schema_migrations(version text primary key, applied_at timestamptz)` records successful application migrations.
- `scripts/migration-smoke-test.ps1` exits non-zero when migration 010 is absent and zero when both remediation tables and its marker exist.

- [ ] **Step 1: Write the failing migration smoke test.**

The test must query the running Compose PostgreSQL service without deleting data:

```powershell
$ErrorActionPreference = 'Stop'
$marker = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT 1 FROM schema_migrations WHERE version='010_remediation_actions'"
if ($marker.Trim() -ne '1') { throw 'migration 010 is not recorded' }
$tables = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM pg_class WHERE relname IN ('remediation_actions','remediation_evidence')"
if ([int]$tables.Trim() -ne 2) { throw 'remediation tables are missing' }
Write-Output 'migration smoke test passed'
```

- [ ] **Step 2: Run it against the current Compose stack and confirm RED.**

Run `powershell -ExecutionPolicy Bypass -File scripts/migration-smoke-test.ps1`. It must fail because the current runtime has the tables from the manual apply but no migration tracking table.

- [ ] **Step 3: Add the idempotent migration and runner.**

`db/migrations/010_remediation_actions.sql` contains the existing remediation DDL and must be safe to run after `db/init/010_remediation_actions.sql`. `db/migrations/run.sh` must create `schema_migrations`, loop over `/migrations/*.sql`, skip versions already recorded, run each pending file with `ON_ERROR_STOP`, and insert its marker only after success.

- [ ] **Step 4: Add Compose ordering.**

Add a `migrate` service using `postgres:16-alpine`, mount `./db/migrations` read-only, depend on healthy `postgres`, and make `api` depend on `migrate` with `condition: service_completed_successfully`. Keep the existing `db/init` mount for fresh database initialization.

- [ ] **Step 5: Run the migration smoke test and verify GREEN.**

Run `docker compose up -d --build`, then `powershell -ExecutionPolicy Bypass -File scripts/migration-smoke-test.ps1`. Verify the existing rows remain intact and the test exits zero.

- [ ] **Step 6: Document the automatic path.**

Update README startup and migration notes so normal `docker compose up --build -d` applies pending migrations. Keep the manual `psql` command as a recovery procedure, not the primary startup path.

### Task 2: Local password reset and authenticated smoke workflow

**Files:**
- Create: `api/cmd/local-password/main.go`
- Create: `scripts/smoke-test.ps1`
- Modify: `README.md`
- Test: `api/internal/auth/password_reset_test.go` (reuse existing password/session behavior)

**Interfaces:**
- `go run ./cmd/local-password --database-url <url> --email <email> --password <password>` updates an active local user password and revokes all sessions.
- `scripts/smoke-test.ps1 -CounselorAdminEmail ... -CounselorAdminPassword ... [-KeepData]` runs the authenticated workflow and removes its temporary organization unless `-KeepData` is supplied.

- [ ] **Step 1: Add a CLI acceptance check before implementation.**

Run the documented local command with a temporary password against a disposable test user only after the command exists; first add a unit-level assertion around the existing `ChangePassword` contract that a password change revokes sessions, then run the focused test to confirm current behavior remains the baseline.

- [ ] **Step 2: Implement the local CLI with existing store/auth primitives.**

The command must:

1. Require a database URL, email, and password.
2. Load the active user with `FindActiveUserByEmail`.
3. Hash with `auth.HashPassword`.
4. Call `Store.ChangePassword`, which revokes sessions.
5. Print only the normalized email and success message; never print the password.

- [ ] **Step 3: Replace the unauthenticated smoke script.**

The script must use `WebRequestSession` and the real API to:

1. Log in as Counselor Admin.
2. Create a unique Organization.
3. Invite an assessor and reviewer, then accept both invitations.
4. Create a Project with metadata.
5. Include one Outcome and assign the assessor.
6. Submit Scope and assert `in_review`.
7. Log in as assessor, fill Current/Target and Response, and submit it.
8. Log in as reviewer and approve the Response.
9. Log in as Counselor Admin, finalize the Project, and assert `closed`.
10. Create a remediation Action for the coverage gap.
11. Log in as assessor, update progress, and submit the Action.
12. Log in as Counselor Admin, close the Action.
13. Read Final Report and Audit Package and assert the Action is present and closed.

Use a unique email suffix for invitations. Cleanup the created organization in a `finally`-style block unless `-KeepData` is set. Do not upload binary files in this smoke test; the existing Go integration test and browser tests cover evidence upload/preview.

- [ ] **Step 4: Run the smoke test RED before wiring the new command/config.**

Run it with the current local account. A failed login must report a useful reset-password command rather than a generic PowerShell exception.

- [ ] **Step 5: Run it GREEN with the local password reset command.**

Reset the documented local Counselor Admin password, run the smoke test, and verify all lifecycle assertions. Run once with `-KeepData` to leave a complete demo Project for manual review.

- [ ] **Step 6: Document the commands.**

README must include password reset, smoke test, cleanup behavior, and the `-KeepData` demo command. Credentials remain explicitly local-development-only.

### Task 3: Role-permission regression matrix

**Files:**
- Create: `api/internal/httpapi/role_matrix_test.go`
- Modify: `api/internal/httpapi/remediation_handler_test.go` only if shared helpers need a small extraction

**Interfaces:**
- Tests exercise the existing HTTP authorization boundary, not direct store calls.

- [ ] **Step 1: Add failing/behavioral matrix cases.**

Cover these exact rules:

```text
counselor_admin/counselor -> create, edit, review, close remediation
org_admin/assessor        -> update and submit only when assigned
reviewer/viewer           -> read remediation, no mutation
assessor/org_admin        -> no finalize or scope mutation
closed project            -> assessment mutation rejected; remediation mutation allowed
```

- [ ] **Step 2: Run focused HTTP tests and inspect any RED case.**

If an existing guard fails, add only the smallest authorization change needed; otherwise keep production code unchanged and treat the matrix as a regression lock.

- [ ] **Step 3: Run the focused and full Go test suites.**

Verify no role can bypass the API guard by calling the endpoint directly.

### Task 4: Demo/audit handoff and final verification

**Files:**
- Modify: `README.md`
- Modify: `scripts/verify-compose.ps1` if it needs to include `migrate`

**Interfaces:**
- README points to the generated `-KeepData` demo workflow and the Final Report/Audit Package routes.
- Compose verification checks `web`, `api`, `postgres`, and migration completion.

- [ ] **Step 1: Run the kept-data smoke workflow.**

Use the output Project/Organization slug to manually verify the assessment page, Action Plan, `/report`, and `/audit` routes.

- [ ] **Step 2: Run all automated verification.**

```powershell
cd api
go vet ./...
go test ./...

cd ..\web
npm test
npx tsc --noEmit --incremental false
npm run build

cd ..
docker compose ps --all
powershell -ExecutionPolicy Bypass -File scripts\migration-smoke-test.ps1
```

- [ ] **Step 3: Confirm the working tree scope.**

Run `git diff --check` and verify only P1 files changed; leave `output/` and `tmp/` untouched.
