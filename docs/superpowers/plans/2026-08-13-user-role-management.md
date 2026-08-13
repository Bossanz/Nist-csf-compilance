# User and Role Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose Counselor and Stakeholder role/status management through the existing Next.js screens and account APIs.

**Architecture:** Keep account state in the existing page components. Add compact controls to the existing organization/member sections rather than adding new routes or a global state layer. Reuse the existing API methods and server authorization rules, with row-level save/error state in the UI.

**Tech Stack:** Next.js, React, TypeScript, Vitest, Testing Library, existing Go API and PostgreSQL account APIs.

## Global Constraints

- Do not add a database migration or new API endpoint.
- Preserve the existing role boundaries and backend authorization as the source of truth.
- Stakeholder roles are limited to `org_admin`, `assessor`, `reviewer`, and `viewer`.
- Counselor roles are limited to `counselor` and `counselor_admin`.
- A signed-in user must not be able to disable their own account.
- Keep the current white editorial layout and existing invitation workflow.

---

### Task 1: Add failing component tests for account controls

**Files:**
- Modify: `web/src/components/OrganizationDashboard.test.tsx`
- Modify: `web/src/components/OrganizationWorkspace.test.tsx`

**Interfaces:**
- `OrganizationDashboard` accepts `counselors`, `onInviteCounselor`, and `onUpdateCounselor` callbacks.
- `OrganizationWorkspace` accepts `onUpdateUser` callback.

- [ ] **Step 1: Add the Counselor Admin management test**

Render `OrganizationDashboard` with one counselor and assert that:

```tsx
expect(screen.getByRole("region", { name: /counselors/i })).toBeTruthy();
expect(screen.getByDisplayValue("counselor@example.com")).toBeTruthy();
```

Change the role select to `counselor_admin`, change status to `disabled`, click the row save button, and assert:

```tsx
expect(onUpdateCounselor).toHaveBeenCalledWith("user-2", { role: "counselor_admin", status: "disabled" });
```

- [ ] **Step 2: Add the Counselor invitation test**

Fill the Counselor email and role, click `Create counselor invitation`, and assert that `onInviteCounselor` receives the normalized email and selected role.

- [ ] **Step 3: Add the Organization stakeholder management test**

Render `OrganizationWorkspace` with an active assessor, change its role to `reviewer`, set status to `disabled`, save the row, and assert:

```tsx
expect(onUpdateUser).toHaveBeenCalledWith("user-2", { role: "reviewer", status: "disabled" });
```

Also assert that the current user's status control cannot select `disabled`.

- [ ] **Step 4: Run the focused tests and verify the expected failure**

Run from `web`:

```powershell
npm.cmd test -- OrganizationDashboard OrganizationWorkspace
```

Expected: FAIL because the new props, sections, controls, and callbacks do not exist yet.

### Task 2: Implement Counselor Admin account management

**Files:**
- Modify: `web/src/app/organizations/page.tsx`
- Modify: `web/src/components/OrganizationDashboard.tsx`
- Modify: `web/src/app/organizations/page.test.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- `OrganizationsPage` loads counselors only for `counselor_admin` users using `api.getCounselors()`.
- `OrganizationsPage` updates counselor rows using `api.updateCounselor()` and stores the returned user.
- `OrganizationsPage` creates invitations using `api.createCounselorInvitation()` and stores the returned `invitationURL`.

- [ ] **Step 1: Extend the page test mocks and add a failing load assertion**

Add `getCounselors`, `updateCounselor`, and `createCounselorInvitation` to the mocked API. Render the page as `counselor_admin` and assert that `api.getCounselors()` is called.

- [ ] **Step 2: Implement page state and handlers**

Add counselor state and conditional loading:

```tsx
const [counselors, setCounselors] = useState<User[]>([]);
const [counselorInvitationURL, setCounselorInvitationURL] = useState("");

const nextCounselors = currentUser.role === "counselor_admin"
  ? await api.getCounselors()
  : [];
```

Add handlers that update one row from the API response and preserve the existing page error handling.

- [ ] **Step 3: Render the Counselor management section**

Add a `Counselors` region only for `counselor_admin`. Include an email/role invitation form and a compact role/status/save control for each existing Counselor. Disable the current user's status option for `disabled`.

- [ ] **Step 4: Add focused styles**

Use the existing `people-list`, `person-row`, `role-chip`, `status-chip`, `panel`, and form tokens. Add only the grid/control styles needed for readable row-level account controls and mobile wrapping.

- [ ] **Step 5: Run the focused tests and verify they pass**

```powershell
npm.cmd test -- OrganizationDashboard OrganizationPage
```

Expected: Counselor list, invitation, and update tests pass.

### Task 3: Implement Organization Stakeholder account management

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/page.tsx`
- Modify: `web/src/components/OrganizationWorkspace.tsx`
- Modify: `web/src/components/OrganizationWorkspace.test.tsx`

**Interfaces:**
- `OrganizationPage` exposes `updateUser(userID, input)` and replaces the matching member from `api.updateOrganizationUser()`.
- `OrganizationWorkspace` renders controls for `counselor_admin`, `counselor`, and `org_admin` users.

- [ ] **Step 1: Add a failing page test for the update callback**

Add `updateOrganizationUser` to the API mock, render the page with an active member, change the member role/status, save, and assert the API call uses the organization and user IDs.

- [ ] **Step 2: Implement the page handler**

```tsx
async function updateUser(userID: string, input: { role: Role; status: "active" | "disabled" }) {
  if (!organization) return;
  const updated = await api.updateOrganizationUser(organization.id, userID, input);
  setMembers((rows) => rows.map((row) => row.id === userID ? updated : row));
}
```

- [ ] **Step 3: Add row-level role/status controls**

Keep the existing name/email display. For managers, add role and status selects plus `Save access`. For the current user, keep the `disabled` option unavailable. Use local saving/error state keyed by user ID so one row does not affect the others.

- [ ] **Step 4: Run the focused tests and verify they pass**

```powershell
npm.cmd test -- OrganizationWorkspace OrganizationPage
```

Expected: stakeholder role/status update and self-disable prevention tests pass.

### Task 4: Verify, document, and commit

**Files:**
- Modify: `README.md`

**Interfaces:**
- The README documents the new Counselor and Stakeholder account-management controls.

- [ ] **Step 1: Update the workflow and role sections in README**

Document that Counselor Admin manages Counselors and that Counselor/Organization Admin can change Stakeholder roles/status from the Organization page.

- [ ] **Step 2: Run all verification commands**

```powershell
Set-Location web
npm.cmd test
npm.cmd run typecheck -- --incremental false
npm.cmd run build
Set-Location ..
Set-Location api
$env:TEST_DATABASE_URL = "postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable"
go test ./...
Set-Location ..
git diff --check
```

- [ ] **Step 3: Rebuild and health-check Docker**

```powershell
docker compose up --build -d
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:3000
```

Expected: all tests/builds pass and both HTTP endpoints return `200`.

- [ ] **Step 4: Review the working tree**

```powershell
git status --short
git diff --stat
```

Confirm that only the intended account-management files and docs changed.
