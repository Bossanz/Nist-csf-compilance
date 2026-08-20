# Auditor Access, Invitation Lifecycle, and Audit Controls Implementation Plan

> **For agentic workers:** Execute this plan task-by-task with a test-first checkpoint for every behavior change.

**Goal:** Add project-scoped Auditor accounts, company-managed invitation lifecycle controls, and complete append-only audit visibility without changing the existing assessment workflow.

**Architecture:** Keep the existing stakeholder users model and add auditor as a read-only role. Store Auditor Project grants in project_auditor_access; copy selected Project IDs from an Auditor invitation into those grants during invitation acceptance. Extend the existing audit writer and audit_logs table instead of introducing a second event system.

**Tech Stack:** Go, net/http, PostgreSQL 16, pgx/v5, Next.js App Router, React, TypeScript, Vitest, Docker Compose, PowerShell smoke tests.

**Spec:** docs/superpowers/specs/2026-08-20-auditor-invitation-audit-controls-design.md

## Global Constraints

- Auditor access is project-scoped and read-only.
- Only an Organization Admin may create, resend, or cancel an Auditor invitation.
- Existing Counselor, Assessor, Reviewer, Viewer, finalization, remediation, report, and evidence behavior remains compatible.
- Invitation tokens are stored hashed; passwords, tokens, and binary evidence contents never enter audit metadata.
- Audit records have no application update or delete endpoint.
- PostgreSQL migrations are idempotent and support fresh databases and existing Docker volumes.
- No SSO, MFA, external identity provider, notification inbox, report versioning, or production deployment work is included.
- Use apply_patch for source edits and keep output/ and tmp/ untracked.

## Current boundaries to preserve

- Role validation lives in api/internal/auth/invitations.go, api/internal/httpapi/authorization.go, api/internal/httpapi/accounts_handler.go, and PostgreSQL constraints in db/init/003_auth_rbac.sql.
- Invitation creation and acceptance use api/internal/auth/invitations.go, api/internal/store/invitations.go, and api/internal/httpapi/invitations_handler.go.
- Project-level authorization is centralized in api/internal/httpapi/authorization.go.
- Audit writes use api/internal/httpapi/audit.go and api/internal/store/audit.go; report history is hydrated by api/internal/store/reporting.go.
- The Organization UI is composed by web/src/app/organizations/[organizationSlug]/page.tsx and web/src/components/OrganizationWorkspace.tsx.
- The Project UI already hides mutation controls for read-only roles; add Auditor to that existing path.

---

### Task 1: Add the Auditor and invitation-access database migration

**Files:**
- Create: db/init/011_auditor_invitation_audit.sql
- Create: db/migrations/011_auditor_invitation_audit.sql
- Modify: scripts/migration-smoke-test.ps1
- Test: api/internal/store/auditor_access_integration_test.go

**Interfaces:**
- Produces the auditor role constraint, invitation lifecycle columns, invitation Project join table, Project Auditor grant table, and structured audit columns.
- Migration runner version: 011_auditor_invitation_audit.

- [ ] **Step 1: Write the failing integration test**

Add a schema-focused integration test that inserts an active stakeholder with role auditor, creates an invitation_project_access row, creates a project_auditor_access row, and asserts the new audit columns are writable. The test must not call the invitation acceptance workflow yet; that behavior is implemented in Task 2.

~~~go
func TestAuditorAccessSchemaSupportsRoleAndProjectGrant(t *testing.T) {
    // Use the existing integration database helper and unique suffix pattern.
    // Insert organization, project, active auditor, invitation access, project grant, and audit metadata.
}
~~~

- [ ] **Step 2: Run the focused test and verify the expected schema failure**

Run from api:

~~~powershell
$env:TEST_DATABASE_URL = 'postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
go test ./internal/store -run TestAuditorAccessSchemaSupportsRoleAndProjectGrant -count=1
~~~

Expected: FAIL because the auditor role, invitation Project join table, and Project grant table do not exist.

- [ ] **Step 3: Add both idempotent migration files**

Create the same SQL in db/init/011_auditor_invitation_audit.sql and db/migrations/011_auditor_invitation_audit.sql:

~~~sql
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
  CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer','auditor'));

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_organization_ownership_check;
ALTER TABLE users ADD CONSTRAINT users_organization_ownership_check CHECK (
  (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
  (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer','auditor'))
);

ALTER TABLE invitations ADD COLUMN IF NOT EXISTS cancelled_at timestamptz;
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS cancelled_by uuid REFERENCES users(id);
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS superseded_at timestamptz;
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS superseded_by uuid REFERENCES invitations(id);
ALTER TABLE invitations DROP CONSTRAINT IF EXISTS invitations_role_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_role_check
  CHECK (role IN ('counselor_admin','counselor','org_admin','assessor','reviewer','viewer','auditor'));
ALTER TABLE invitations DROP CONSTRAINT IF EXISTS invitations_scope_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_scope_check CHECK (
  (user_type = 'counselor' AND organization_id IS NULL AND role IN ('counselor_admin','counselor')) OR
  (user_type = 'stakeholder' AND organization_id IS NOT NULL AND role IN ('org_admin','assessor','reviewer','viewer','auditor'))
);

CREATE TABLE IF NOT EXISTS invitation_project_access (
  invitation_id uuid NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  PRIMARY KEY (invitation_id, project_id)
);

CREATE TABLE IF NOT EXISTS project_auditor_access (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  granted_by uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  revoked_at timestamptz,
  PRIMARY KEY (project_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_project_auditor_access_user
  ON project_auditor_access(user_id, revoked_at);

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_role text;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS result text NOT NULL DEFAULT 'success';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id text;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS ip_address inet;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_agent text;
~~~

- [ ] **Step 4: Verify migration idempotency**

Run the migration service twice against the same database. The second run must skip version 011 without changing existing rows or failing on already-present constraints, columns, tables, or indexes.

- [ ] **Step 5: Run focused store and migration checks**

~~~powershell
$env:TEST_DATABASE_URL = 'postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
go test ./internal/store -run TestAuditorAccessSchemaSupportsRoleAndProjectGrant -count=1
cd ..
.\scripts\migration-smoke-test.ps1
~~~

Expected: PASS, with migration version 011_auditor_invitation_audit recorded and both access tables present.

- [ ] **Step 6: Commit the migration slice**

~~~powershell
git add db/init/011_auditor_invitation_audit.sql db/migrations/011_auditor_invitation_audit.sql scripts/migration-smoke-test.ps1 api/internal/store/auditor_access_integration_test.go
git commit -m "feat: add auditor access schema"
~~~

### Task 2: Implement invitation lifecycle and Project assignment in the domain/store layer

**Files:**
- Modify: api/internal/store/models.go
- Modify: api/internal/auth/invitations.go
- Modify: api/internal/auth/invitations_test.go
- Modify: api/internal/store/invitations.go
- Create: api/internal/store/invitation_lifecycle_integration_test.go

**Interfaces:**
- store.Invitation exposes derived Status, lifecycle timestamps, and ProjectIDs []string.
- auth.InvitationService exposes Invite, Resend, Cancel, and Accept.
- Auditor creation requires an Organization Admin in the target Organization and at least one Project ID.

- [ ] **Step 1: Write failing service tests**

Add tests with these names and behaviors:

~~~go
func TestInviteAuditorRequiresOrganizationAdminAndProjectIDs(t *testing.T) {}
func TestResendSupersedesPendingInvitationAndReturnsNewToken(t *testing.T) {}
func TestCancelRejectsAcceptanceOfInvitation(t *testing.T) {}
func TestExpiredInvitationIsReportedOnceWhenObserved(t *testing.T) {}
func TestAuditorInvitationAcceptanceCreatesProjectGrant(t *testing.T) {}
~~~

Assert the raw token is returned only to the caller, the stored token is hashed, resend supersedes the old row, cancel records the actor, and acceptance creates Project grants in the same transaction as the user.

- [ ] **Step 2: Run focused auth tests and verify failure**

~~~powershell
go test ./internal/auth -run 'TestInviteAuditor|TestResend|TestCancel|TestExpired' -count=1
~~~

Expected: FAIL because the role and lifecycle methods do not exist.

- [ ] **Step 3: Extend repository and service interfaces**

Add repository operations with these signatures:

~~~go
ListInvitations(ctx context.Context, organizationID string) ([]store.Invitation, error)
GetInvitation(ctx context.Context, organizationID, invitationID string) (store.Invitation, error)
CreateInvitationProjectAccess(ctx context.Context, invitationID string, projectIDs []string) error
ResendInvitation(ctx context.Context, organizationID, invitationID, newTokenHash, actorID string, expiresAt time.Time) (store.Invitation, error)
CancelInvitation(ctx context.Context, organizationID, invitationID, actorID string, now time.Time) (store.Invitation, error)
~~~

Load Project IDs with the invitation row so acceptance can remain one transaction.

- [ ] **Step 4: Implement one derived status helper**

Implement and reuse:

~~~go
func InvitationStatus(inv Invitation, now time.Time) string
~~~

Return accepted, cancelled, superseded, expired, or pending in that precedence order. Do not add a second mutable status column.

- [ ] **Step 5: Implement resend and cancel transactions**

Resend locks the current row, requires pending, creates a new hashed token and invitation, copies Project access rows, then marks the old row superseded. Cancel locks the row, requires pending, sets cancelled_at and cancelled_by, and is safe when called twice.

- [ ] **Step 6: Run tests and commit**

~~~powershell
go test ./internal/auth ./internal/store -count=1
git add api/internal/store/models.go api/internal/auth/invitations.go api/internal/auth/invitations_test.go api/internal/store/invitations.go api/internal/store/invitation_lifecycle_integration_test.go
git commit -m "feat: add invitation lifecycle controls"
~~~

### Task 3: Add project-scoped Auditor authorization

**Files:**
- Modify: api/internal/httpapi/authorization.go
- Modify: api/internal/httpapi/authorization_test.go
- Modify: api/internal/httpapi/handler.go
- Modify: api/internal/store/projects.go
- Create: api/internal/store/auditor_access.go
- Create: api/internal/httpapi/auditor_authorization_test.go

**Interfaces:**
- store.HasActiveProjectAuditorAccess(ctx, projectID, userID string) (bool, error) is the single read guard.
- authorizeProject returns 404 not_found for an Auditor without an active grant and permits reads for a granted Auditor.
- can returns false for every mutation action when user.Role == auditor.

- [ ] **Step 1: Write failing HTTP authorization tests**

Add:

~~~go
func TestAuditorCanReadGrantedProject(t *testing.T) {}
func TestAuditorCannotReadUnassignedProject(t *testing.T) {}
func TestAuditorCannotMutateGrantedProject(t *testing.T) {}
func TestRevokedAuditorGrantStopsProjectAccess(t *testing.T) {}
~~~

Assert 200 for a granted read, 404 for an ungranted read, and 403 for profile/response/review/finalize/remediation mutations.

- [ ] **Step 2: Run focused tests and verify failure**

~~~powershell
go test ./internal/httpapi -run 'TestAuditor|TestRevokedAuditor' -count=1
~~~

Expected: FAIL because the current authorization branch does not recognize Auditor grants.

- [ ] **Step 3: Implement the Auditor read-only branch**

In authorizeProject, after the Organization boundary check, query HasActiveProjectAuditorAccess for an Auditor. Return 404 when false. Keep the existing setup hiding rule for other stakeholders, but allow a granted Auditor to read setup Projects.

In stakeholderCanReadProfile, allow Auditor after the Project guard passes. Do not add Auditor to stakeholderCanEditProfile or mutation actions.

- [ ] **Step 4: Add the store query**

Implement:

~~~go
func (s *Store) HasActiveProjectAuditorAccess(ctx context.Context, projectID, userID string) (bool, error) {
    // SELECT EXISTS(... WHERE project_id=$1 AND user_id=$2 AND revoked_at IS NULL)
}
~~~

Use the existing not-found response shape so an ungranted Auditor cannot discover Project existence.

- [ ] **Step 5: Run and commit**

~~~powershell
go test ./internal/httpapi ./internal/store -count=1
git add api/internal/httpapi/authorization.go api/internal/httpapi/authorization_test.go api/internal/httpapi/handler.go api/internal/httpapi/auditor_authorization_test.go api/internal/store/projects.go api/internal/store/auditor_access.go
git commit -m "feat: enforce project-scoped auditor access"
~~~

### Task 4: Expand append-only audit logging and read endpoints

**Files:**
- Modify: api/internal/store/models.go
- Modify: api/internal/store/audit.go
- Modify: api/internal/httpapi/audit.go
- Modify: api/internal/httpapi/handler.go
- Create: api/internal/store/audit_queries.go
- Create: api/internal/httpapi/audit_handler_test.go
- Modify: endpoint tests for auth, invitations, accounts, documents, responses, remediation, reporting, and finalization.

**Interfaces:**
- store.AuditEvent gains ActorRole, Result, RequestID, IPAddress, and UserAgent.
- store.ListProjectAuditLogs(ctx, projectID string) and store.ListOrganizationAuditLogs(ctx, organizationID string) return []store.AuditLogEntry in descending timestamp order.
- Add GET /api/projects/{projectID}/audit-log and GET /api/organizations/{organizationID}/audit-log.

- [ ] **Step 1: Write failing audit tests**

Add:

~~~go
func TestWriteAuditPersistsActorRoleAndResult(t *testing.T) {}
func TestProjectAuditLogIsReadableByGrantedAuditor(t *testing.T) {}
func TestAuditorCannotReadOrganizationAuditLog(t *testing.T) {}
func TestAuditLogHasNoMutationRoute(t *testing.T) {}
~~~

Assert sensitive keys such as password, token, storagePath, and storageKey are absent from metadata.

- [ ] **Step 2: Run focused tests and verify failure**

~~~powershell
go test ./internal/httpapi ./internal/store -run 'TestWriteAudit|TestProjectAuditLog|TestAuditorCannotReadOrganization|TestAuditLogHasNoMutation' -count=1
~~~

Expected: FAIL because structured fields, queries, and routes are not present.

- [ ] **Step 3: Extend the audit model and insert query**

Update store.AuditEvent, store.AuditLogEntry, and WriteAudit to insert structured fields. Add a redaction helper that removes keys matching password, token, secret, storagePath, and storageKey before JSON insertion.

- [ ] **Step 4: Add list queries and routes**

Implement project and organization list queries with actor display joins. Add routes before generic Project GET:

~~~text
GET /api/projects/{id}/audit-log
GET /api/organizations/{id}/audit-log
~~~

The Project endpoint uses project authorization. The Organization endpoint permits Counselor, Counselor Admin, and Organization Admin, but rejects Auditor with 403.

- [ ] **Step 5: Instrument the event catalog**

Use one event per successful material mutation and one safe event for failed login. Add event writes for authentication, invitation lifecycle, user role/status, Auditor grants, Organization/Project changes, scope/profile/assignment/response changes, evidence operations, finalization, remediation, reports, Audit Package, and audit-log reads. Do not log catalog reads, health checks, request bodies, passwords, tokens, or binary bodies.

- [ ] **Step 6: Run all Go tests and commit**

~~~powershell
go vet ./...
$env:TEST_DATABASE_URL = 'postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
go test ./...
git add api/internal/store/models.go api/internal/store/audit.go api/internal/store/audit_queries.go api/internal/httpapi/audit.go api/internal/httpapi/handler.go api/internal/httpapi/audit_handler_test.go api/internal/httpapi api/internal/store
git commit -m "feat: expand audit trail and audit log access"
~~~

### Task 5: Add invitation APIs and Organization invitation management UI

**Files:**
- Modify: api/internal/httpapi/invitations_handler.go
- Modify: api/internal/httpapi/handler.go
- Modify: api/internal/httpapi/invitations_handler_test.go
- Modify: web/src/lib/types.ts
- Modify: web/src/lib/api.ts
- Modify: web/src/app/organizations/[organizationSlug]/page.tsx
- Modify: web/src/components/OrganizationWorkspace.tsx
- Modify: web/src/components/OrganizationWorkspace.test.tsx
- Create: web/src/components/InvitationList.tsx
- Create: web/src/components/InvitationList.test.tsx

**Interfaces:**
- api.getOrganizationInvitations(organizationID) returns derived status and Project summaries.
- api.resendInvitation(organizationID, invitationID) returns the new local invitation URL.
- api.cancelInvitation(organizationID, invitationID) returns the cancelled invitation.
- api.createInvitation(organizationID, { email, role, projectIDs }) requires Project IDs only for Auditor.

- [ ] **Step 1: Write failing API and component tests**

Add:

~~~go
func TestOrgAdminCreatesAuditorInvitationWithProjectIDs(t *testing.T) {}
func TestNonOrgAdminCannotCreateAuditorInvitation(t *testing.T) {}
func TestResendAndCancelInvitationEndpoints(t *testing.T) {}
~~~

~~~tsx
test("requires a project when inviting an auditor", () => {})
test("renders invitation status and resend/cancel actions", () => {})
~~~

- [ ] **Step 2: Run focused tests and verify failure**

~~~powershell
go test ./internal/httpapi -run 'TestOrgAdminCreatesAuditor|TestNonOrgAdminCannotCreateAuditor|TestResendAndCancel' -count=1
cd ../web
npm test -- --run src/components/InvitationList.test.tsx src/components/OrganizationWorkspace.test.tsx
~~~

Expected: FAIL because request fields, routes, and UI component do not exist.

- [ ] **Step 3: Implement routes and serialization**

Add organization invitation list, resend, and cancel routes. Enforce Organization Admin for Auditor lifecycle actions. Emit invitation lifecycle events after successful changes. Preserve local invitation URL responses.

- [ ] **Step 4: Add typed client methods and invitation UI**

Create InvitationList with email, role, derived status, expiry, invited-by, assigned Projects, and Organization Admin controls. Auditor creation requires one or more Project selections. Other role invitation behavior remains compatible.

- [ ] **Step 5: Run tests and commit**

~~~powershell
npm test -- --run
npx tsc --noEmit --incremental false
cd ..
git add api/internal/httpapi/invitations_handler.go api/internal/httpapi/handler.go api/internal/httpapi/invitations_handler_test.go web/src/lib/types.ts web/src/lib/api.ts web/src/app/organizations/[organizationSlug]/page.tsx web/src/components/OrganizationWorkspace.tsx web/src/components/OrganizationWorkspace.test.tsx web/src/components/InvitationList.tsx web/src/components/InvitationList.test.tsx
git commit -m "feat: manage organization invitations"
~~~

### Task 6: Add Auditor read-only workspace and audit timeline UI

**Files:**
- Modify: web/src/lib/types.ts
- Modify: web/src/lib/api.ts
- Modify: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx
- Modify: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/audit/page.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.tsx
- Modify: web/src/components/AssessmentCard.tsx
- Modify: web/src/components/AuditPackageView.tsx
- Modify: web/src/components/ProjectAssessmentWorkspace.test.tsx
- Create: web/src/components/AuditTimeline.tsx
- Create: web/src/components/AuditTimeline.test.tsx

**Interfaces:**
- api.getProjectAuditLog(projectID) returns entries from the Project audit-log endpoint.
- AuditTimeline renders actor, role, action, result, timestamp, entity, and safe metadata.

- [ ] **Step 1: Write failing UI tests**

Add:

~~~tsx
test("Auditor sees the project in read-only mode", () => {})
test("Auditor does not see mutation controls", () => {})
test("Audit timeline renders safe event metadata", () => {})
~~~

- [ ] **Step 2: Run focused tests and verify failure**

~~~powershell
npm test -- --run src/components/ProjectAssessmentWorkspace.test.tsx src/components/AuditTimeline.test.tsx
~~~

Expected: FAIL because auditor is not in the UI role union and the timeline component does not exist.

- [ ] **Step 3: Implement read-only Auditor mode**

Add auditor to the TypeScript Role union and existing read-only conditions. Show an Auditor / Read only mode marker. Keep all mutation controls hidden while API guards remain authoritative.

- [ ] **Step 4: Implement the audit timeline**

Load the Project audit log on the Audit route and render it below the existing package sections. Add loading, empty, and error states with stable labels for invitation, access, assessment, report, and remediation actions.

- [ ] **Step 5: Run the web suite and commit**

~~~powershell
npm test -- --run
npx tsc --noEmit --incremental false
npm run build
git add web/src/lib/types.ts web/src/lib/api.ts web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/audit/page.tsx web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/AssessmentCard.tsx web/src/components/AuditPackageView.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx web/src/components/AuditTimeline.tsx web/src/components/AuditTimeline.test.tsx
git commit -m "feat: add auditor read-only workspace"
~~~

### Task 7: Add end-to-end Auditor smoke coverage and documentation

**Files:**
- Modify: scripts/smoke-test.ps1
- Modify: scripts/migration-smoke-test.ps1
- Modify: scripts/verify-compose.ps1
- Modify: README.md

**Interfaces:**
- The smoke script creates an Organization, two Projects, invites an Auditor for exactly one Project, accepts the invitation, verifies selected/unselected access, tests a mutation rejection, and checks expected events in the Project audit log.

- [ ] **Step 1: Add smoke assertions before implementation changes**

Add assertions for selected Project 200, unselected Project 404, Auditor mutation 403, old-token acceptance failure, and audit actions including invitation.accepted.

- [ ] **Step 2: Run the smoke test and verify the new assertions fail**

~~~powershell
.\scripts\smoke-test.ps1 -CounselorAdminEmail admin@example.com -CounselorAdminPassword LocalAdmin!2026
~~~

Expected: FAIL at Auditor invitation before Tasks 1–6 are implemented.

- [ ] **Step 3: Implement the complete smoke flow and safe cleanup**

Use unique emails and Organization names. Accept the Auditor invitation in a separate session. Delete temporary Organizations in a try/finally cleanup block unless -KeepData is supplied. Never print passwords or token hashes.

- [ ] **Step 4: Update README**

Document Auditor permissions, Project-scoped access, invitation lifecycle, audit timeline, migration/runtime checks, and the existing out-of-scope list.

- [ ] **Step 5: Run complete verification**

~~~powershell
cd api
go vet ./...
$env:TEST_DATABASE_URL = 'postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
go test ./...
cd ../web
npm test -- --run
npx tsc --noEmit --incremental false
npm run build
cd ..
.\scripts\migration-smoke-test.ps1
.\scripts\verify-compose.ps1
.\scripts\smoke-test.ps1 -CounselorAdminEmail admin@example.com -CounselorAdminPassword LocalAdmin!2026
git diff --check
~~~

Expected: Go vet/tests, all web tests, TypeScript, build, migration smoke, compose verification, and authenticated Auditor smoke pass. If Docker is stopped, report the runtime check as blocked instead of treating static checks as runtime verification.

- [ ] **Step 6: Commit the verification/documentation slice**

~~~powershell
git add scripts/smoke-test.ps1 scripts/migration-smoke-test.ps1 scripts/verify-compose.ps1 README.md
git commit -m "test: verify auditor access and audit workflow"
~~~

## Final review checklist

- [ ] auditor exists in PostgreSQL, Go, and TypeScript role definitions.
- [ ] Auditor Project grants are created atomically on invitation acceptance.
- [ ] An ungranted Auditor receives 404 for Project reads and cannot infer Project existence.
- [ ] Auditor mutation attempts receive 403 and are not persisted.
- [ ] Resend supersedes the old invitation token; cancel prevents acceptance; expiry is visible.
- [ ] Audit events include actor/context/result and redact sensitive values.
- [ ] No audit update/delete route exists.
- [ ] Existing assessment workflow remains green.
- [ ] Docker runtime smoke is green or explicitly reported as blocked by Docker daemon state.
- [ ] output/ and tmp/ remain untracked.
