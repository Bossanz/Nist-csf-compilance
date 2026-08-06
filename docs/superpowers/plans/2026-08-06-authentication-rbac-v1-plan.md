# Authentication and RBAC V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add secure email/password sessions, organization-first navigation, invitations, and V1 role enforcement for counselors and customer stakeholders.

**Architecture:** The Go API owns authentication and authorization with opaque sessions stored as hashes in PostgreSQL. Next.js remains a same-origin client and renders login, organization, project, and invitation views. Existing project/profile routes are retained but protected and scoped from the authenticated user.

**Tech Stack:** Go 1.24, `golang.org/x/crypto/bcrypt`, PostgreSQL 16, Next.js 16, React 19, Vitest, Docker Compose

## Global Constraints

- Roles are exactly `counselor_admin`, `counselor`, `org_admin`, `assessor`, `reviewer`, and `viewer`.
- Stakeholders belong to exactly one organization; counselors belong to none and can access every organization in V1.
- The Go API is the authorization boundary; hidden UI controls are not security controls.
- V1 uses copied one-time invitation links and does not send email.
- Do not add SSO, MFA, password reset, multi-organization stakeholders, counselor assignment, reviewer approval workflow, per-project stakeholder permissions, Redis, or a separate session service.
- Existing organization and project data must survive the migration.

---

### Task 1: Authentication schema and security primitives

**Files:**
- Create: `db/init/003_auth_rbac.sql`
- Create: `api/internal/auth/password.go`
- Create: `api/internal/auth/password_test.go`
- Create: `api/internal/auth/token.go`
- Create: `api/internal/auth/token_test.go`
- Modify: `api/internal/store/models.go`
- Modify: `api/go.mod`
- Modify: `api/go.sum`

**Interfaces:**
- Produces: `auth.HashPassword(string) (string, error)`, `auth.CheckPassword(string, string) bool`, `auth.NewToken() (raw string, hash string, err error)`, and store models `User`, `Session`, `Organization`, `Invitation`.

- [ ] **Step 1: Write failing primitive tests**

```go
func TestPasswordHashDoesNotStorePlaintext(t *testing.T) {
    hash, err := HashPassword("correct horse battery staple")
    if err != nil || hash == "correct horse battery staple" || !CheckPassword(hash, "correct horse battery staple") {
        t.Fatalf("password hash verification failed")
    }
    if CheckPassword(hash, "wrong") { t.Fatal("wrong password matched") }
}

func TestNewTokenReturnsRawTokenAndOnlyItsHash(t *testing.T) {
    raw, hash, err := NewToken()
    if err != nil || raw == "" || hash == "" || raw == hash { t.Fatal("invalid token") }
    if HashToken(raw) != hash { t.Fatal("token hash mismatch") }
}
```

- [ ] **Step 2: Run RED**

Run: `go test ./internal/auth -v`

Expected: FAIL because the package and functions do not exist.

- [ ] **Step 3: Add the idempotent migration**

`003_auth_rbac.sql` must:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (lower(email));
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'invited';
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
```

Backfill existing users to valid roles based on `user_type`, then add named checks for role, status, and counselor/stakeholder organization ownership. Create `sessions`, `invitations`, and `audit_logs` with UUID primary keys, token-hash uniqueness, expiry columns, foreign keys, and indexes on session token hash and invitation token hash.

- [ ] **Step 4: Implement bcrypt and random tokens**

Use `bcrypt.GenerateFromPassword`, `bcrypt.CompareHashAndPassword`, 32 random bytes from `crypto/rand`, URL-safe base64 for raw tokens, and SHA-256 hex for stored token hashes.

- [ ] **Step 5: Run GREEN and full Go tests**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add db/init/003_auth_rbac.sql api/internal/auth api/internal/store/models.go
git commit -m "feat: add authentication schema and primitives"
```

---

### Task 2: Users, sessions, bootstrap, and authentication endpoints

**Files:**
- Create: `api/internal/store/auth.go`
- Create: `api/internal/auth/service.go`
- Create: `api/internal/auth/service_test.go`
- Create: `api/internal/httpapi/auth_handler.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/handler_test.go`
- Modify: `api/cmd/server/main.go`
- Modify: `docker-compose.yml`
- Create: `.env.example`

**Interfaces:**
- Consumes: password/token primitives and store auth models.
- Produces: `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/auth/me`, session middleware, and idempotent counselor-admin bootstrap.

- [ ] **Step 1: Write failing service tests**

Use a fake `auth.Repository` and assert:

```go
func TestLoginReturnsSameErrorForUnknownAndWrongPassword(t *testing.T)
func TestLoginCreatesHashedSessionForActiveUser(t *testing.T)
func TestAuthenticateRejectsExpiredOrDisabledSession(t *testing.T)
func TestBootstrapCreatesAdminOnlyWhenNoneExists(t *testing.T)
```

Each test uses a repository fake that records session hashes and administrator inserts. Assert `errors.Is` against the named auth errors, assert the recorded session value equals `HashToken(rawToken)`, and assert bootstrap insert count remains one after two calls.

- [ ] **Step 2: Run RED**

Run: `go test ./internal/auth ./internal/httpapi -run 'Test(Login|Authenticate|Bootstrap)' -v`

Expected: FAIL because the service, repository methods, and endpoints are absent.

- [ ] **Step 3: Implement the auth repository and service**

Define:

```go
type Repository interface {
    FindUserByEmail(context.Context, string) (store.User, error)
    FindUserBySessionHash(context.Context, string) (store.User, store.Session, error)
    CreateSession(context.Context, string, string, time.Time) error
    RevokeSession(context.Context, string) error
    HasCounselorAdmin(context.Context) (bool, error)
    CreateCounselorAdmin(context.Context, string, string) (store.User, error)
}
```

Normalize email with `strings.ToLower(strings.TrimSpace(email))`. Use a 12-hour absolute session lifetime and update `last_seen_at` without extending expiry.

- [ ] **Step 4: Implement auth HTTP behavior**

Cookie name: `compliance_session`. Login returns `401` and `{"error":{"code":"invalid_credentials","message":"Invalid email or password"}}` for every credential failure. `me` returns the safe user model without `password_hash`. Logout revokes the token hash and clears the cookie.

Add a small mutex-protected login throttle keyed by normalized email plus client address: five failed attempts in five minutes returns `429`; successful login clears the key.

- [ ] **Step 5: Implement bootstrap configuration**

At startup, call bootstrap only when both `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` are present. If no counselor admin exists, create one with role `counselor_admin`; never overwrite an existing admin. Add placeholders without secrets to `.env.example` and Compose variable forwarding:

```yaml
BOOTSTRAP_ADMIN_EMAIL: ${BOOTSTRAP_ADMIN_EMAIL:-}
BOOTSTRAP_ADMIN_PASSWORD: ${BOOTSTRAP_ADMIN_PASSWORD:-}
APP_ENV: ${APP_ENV:-development}
APP_ORIGIN: ${APP_ORIGIN:-http://localhost:3000}
```

- [ ] **Step 6: Run GREEN and full Go tests**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add api .env.example docker-compose.yml
git commit -m "feat: add session authentication"
```

---

### Task 3: Organization-first data model and protected project routes

**Files:**
- Create: `api/internal/store/organizations.go`
- Modify: `api/internal/store/projects.go`
- Create: `api/internal/httpapi/organizations_handler.go`
- Create: `api/internal/httpapi/authorization.go`
- Create: `api/internal/httpapi/authorization_test.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/handler_test.go`

**Interfaces:**
- Produces: organization CRUD/list methods, `CreateProject(ctx, organizationID, name)`, scoped project reads/writes, and role guards.

- [ ] **Step 1: Write failing authorization tests**

Cover these table cases using authenticated request context:

```go
{role:"counselor_admin", action:"create_organization", allowed:true}
{role:"counselor", action:"create_project", allowed:true}
{role:"org_admin", action:"create_project", allowed:false}
{role:"org_admin", action:"update_profile", allowed:false}
{role:"assessor", action:"update_profile", allowed:true}
{role:"reviewer", action:"update_profile", allowed:false}
{role:"viewer", action:"update_profile", allowed:false}
```

Also assert a stakeholder receives `404` for another organization's project and an unauthenticated request receives `401` for every route except health, login, and invitation acceptance.

- [ ] **Step 2: Run RED**

Run: `go test ./internal/httpapi -run 'TestAuthorization|TestOrganizationScope' -v`

Expected: FAIL because routes are currently public and unscoped.

- [ ] **Step 3: Implement organization and project storage**

Add `ListOrganizationsForUser`, `CreateOrganization`, `GetOrganizationForUser`, `ListProjectsForOrganization`, and `CreateProject(ctx, organizationID, name)`. Remove implicit organization creation from `CreateProject`. Stop deleting an organization when its final project is deleted.

Reject a case-insensitive trimmed organization-name duplicate with `409 organization_exists`. Add `DELETE /api/organizations/:id` for `counselor_admin`; deletion succeeds only when no projects or stakeholder users reference the organization.

- [ ] **Step 4: Protect all routes**

Authenticate before routing protected `/api` endpoints. Counselor roles can access all organizations. Stakeholders can access only matching `organization_id`. Enforce role checks for create/delete/update actions and write audit records for organization creation, project creation/deletion, and profile updates.

For non-GET/HEAD/OPTIONS requests, require an `Origin` matching the configured `APP_ORIGIN` and `Content-Type: application/json`; return `403` for an invalid origin.

- [ ] **Step 5: Run GREEN and full Go tests**

Run: `go test ./...`

Expected: PASS, including existing project/profile tests updated with authenticated context.

- [ ] **Step 6: Commit**

```powershell
git add api/internal/store api/internal/httpapi
git commit -m "feat: enforce organization and role access"
```

---

### Task 4: Account invitations and role management

**Files:**
- Create: `api/internal/store/invitations.go`
- Create: `api/internal/auth/invitations.go`
- Create: `api/internal/auth/invitations_test.go`
- Create: `api/internal/httpapi/invitations_handler.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/handler_test.go`

**Interfaces:**
- Produces: `POST /api/organizations/:id/invitations`, `GET /api/organizations/:id/users`, `PATCH /api/organizations/:id/users/:userID`, `GET /api/counselors`, `PATCH /api/counselors/:userID`, `POST /api/counselor-invitations`, and `POST /api/invitations/:token/accept`.

- [ ] **Step 1: Write failing invitation tests**

Assert that counselor roles may invite the first `org_admin`; `org_admin` may invite only stakeholder roles; only `counselor_admin` may invite counselor roles; counselor invitations have no organization; duplicate active/pending email returns conflict; expired/used/unknown tokens share one error; acceptance stores a password hash, activates the user, marks the invitation used, and cannot run twice.

- [ ] **Step 2: Run RED**

Run: `go test ./internal/auth ./internal/httpapi -run 'TestInvitation' -v`

Expected: FAIL because invitation behavior does not exist.

- [ ] **Step 3: Implement invitation transactions**

Create invitations with a 72-hour expiry and stored token hash. Accept in one transaction that locks the invitation row, checks validity, creates or activates the stakeholder, and sets `accepted_at`. Return `${APP_ORIGIN}/invite/${rawToken}` only at creation time.

- [ ] **Step 4: Add endpoints and audit events**

Return `409 duplicate_invitation`, generic `400 invalid_invitation`, and safe stakeholder/counselor lists without password/session data. The stakeholder patch endpoint accepts only `role` and `status`, only for users in the caller's organization, and never accepts counselor roles. The counselor patch endpoint is restricted to `counselor_admin`. Disabling a user revokes all their sessions. Record `user.invited`, `invitation.accepted`, `user.role_changed`, and `user.disabled` events.

- [ ] **Step 5: Run GREEN and full Go tests**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add api/internal/store/invitations.go api/internal/auth/invitations* api/internal/httpapi
git commit -m "feat: invite organization stakeholders"
```

---

### Task 5: Login and organization-first frontend

**Files:**
- Modify: `web/src/lib/types.ts`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/lib/api.test.ts`
- Create: `web/src/components/LoginForm.tsx`
- Create: `web/src/components/LoginForm.test.tsx`
- Create: `web/src/components/OrganizationDashboard.tsx`
- Create: `web/src/components/OrganizationDashboard.test.tsx`
- Create: `web/src/components/OrganizationWorkspace.tsx`
- Create: `web/src/components/OrganizationWorkspace.test.tsx`
- Modify: `web/src/components/ProjectDashboard.tsx`
- Modify: `web/src/app/page.tsx`
- Modify: `web/src/app/page.test.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Consumes: auth, organization, scoped project, user-list, and invitation endpoints.
- Produces: login gate and `Organizations -> Organization workspace -> Project assessment` navigation.

- [ ] **Step 1: Write failing frontend tests**

Tests must prove:

```ts
test("shows login when session restoration returns 401", restoreSessionScenario);
test("lists only organizations returned for the current user", organizationListScenario);
test("creates a project inside the selected organization without an organization name", scopedProjectScenario);
test("hides create controls from reviewer and viewer roles", readOnlyRoleScenario);
test("logs out and returns to login", logoutScenario);
```

Implement each named scenario in the test file using real rendered components, a stubbed global `fetch`, and Testing Library interactions. Assert visible headings and exact request URLs/bodies rather than asserting mock call counts alone.

- [ ] **Step 2: Run RED**

Run: `npm.cmd test -- src/app/page.test.tsx src/components/LoginForm.test.tsx src/components/OrganizationDashboard.test.tsx src/components/OrganizationWorkspace.test.tsx`

Expected: FAIL because the components and API methods are missing.

- [ ] **Step 3: Add typed API methods**

Add `User`, `Organization`, `Invitation`, and `Role` types and methods `login`, `logout`, `me`, `getOrganizations`, `createOrganization`, `getOrganizationProjects`, `createOrganizationProject`, `getOrganizationUsers`, `updateOrganizationUser`, `createInvitation`, `getCounselors`, `updateCounselor`, and `createCounselorInvitation`. Preserve same-origin requests and include JSON headers.

- [ ] **Step 4: Implement the view state**

On startup call `me`. A `401` renders `LoginForm`; success loads organizations. Selecting an organization loads its projects and users. Selecting a project reuses the existing assessment editor. Back navigation returns Project -> Organization -> Organizations. Project creation accepts only a project name because organization context is already selected.

- [ ] **Step 5: Mirror role permissions in controls**

Show organization create/delete and counselor-team management only to `counselor_admin`; project creation to counselor roles; stakeholder invitation and stakeholder role/status controls to counselor roles and `org_admin`; assessment editing to counselor roles and `assessor`. `org_admin`, reviewer, and viewer receive read-only assessment cards.

- [ ] **Step 6: Run GREEN and frontend verification**

Run:

```powershell
npm.cmd test
npm.cmd run typecheck
npm.cmd run build
```

Expected: all commands exit 0.

- [ ] **Step 7: Commit**

```powershell
git add web
git commit -m "feat: add login and organization workspace"
```

---

### Task 6: Invitation acceptance page and Docker verification

**Files:**
- Create: `web/src/app/invite/[token]/page.tsx`
- Create: `web/src/components/AcceptInvitationForm.tsx`
- Create: `web/src/components/AcceptInvitationForm.test.tsx`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/app/globals.css`
- Modify: `.env.example`
- Modify: `README.md`

**Interfaces:**
- Consumes: `POST /api/invitations/:token/accept`.
- Produces: one-time password setup, activation confirmation, and documented local bootstrap flow.

- [ ] **Step 1: Write the failing acceptance test**

Assert matching passwords are required, a successful submission calls the token endpoint once, and invalid/expired/used tokens display the generic API message.

- [ ] **Step 2: Run RED**

Run: `npm.cmd test -- src/components/AcceptInvitationForm.test.tsx`

Expected: FAIL because the component does not exist.

- [ ] **Step 3: Implement invitation acceptance**

Read the route token from params, submit `{name,password}` to the acceptance endpoint, never store the token or password in browser storage, and link to `/` after activation.

- [ ] **Step 4: Document and configure local bootstrap**

Document copying `.env.example` to `.env`, setting non-placeholder credentials, applying the existing-volume migration with:

```powershell
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/003_auth_rbac.sql
```

Then rebuild with `docker compose up --build -d`.

- [ ] **Step 5: Run complete automated verification**

Run:

```powershell
Set-Location api
go test ./...
Set-Location ..\web
npm.cmd test
npm.cmd run typecheck
npm.cmd run build
Set-Location ..
docker compose config
docker compose up --build -d
docker compose ps
```

Expected: all tests and builds pass; PostgreSQL and API are healthy; Web is running.

- [ ] **Step 6: Verify the complete browser workflow**

In the in-app browser:

1. Log in with the bootstrapped counselor admin.
2. Create an organization.
3. Create a project inside it and confirm no duplicate organization is created.
4. Create and copy an `org_admin` invitation link.
5. Open the invitation, activate the stakeholder, and log in.
6. Confirm the org admin sees only their organization and cannot create a project.
7. Confirm a viewer cannot edit an assessment.
8. Confirm no `Failed to fetch`, console error, cross-organization data, or password/token browser storage exists.

- [ ] **Step 7: Commit**

```powershell
git add web README.md .env.example
git commit -m "feat: activate invited stakeholders"
```
