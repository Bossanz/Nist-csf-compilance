# Delete Organization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a Counselor Admin permanently delete an organization and all organization-owned data after exact-name confirmation.

**Architecture:** Keep the existing `DELETE /api/organizations/{id}` contract and replace its empty-only store operation with an explicit PostgreSQL transaction. Add deletion confirmation to the existing organization card and let the page own API calls, list state, and user-visible errors.

**Tech Stack:** Go 1.24, pgx v5, PostgreSQL 16, Next.js 16, React 19, TypeScript, Vitest, Testing Library.

## Global Constraints

- Only `counselor_admin` can see and use organization deletion.
- Deletion is permanent and includes projects, assessments, stakeholder accounts, sessions, invitations, and organization audit logs.
- Exact organization name is required before confirmation.
- Do not add archive, restore, bulk deletion, or new dependencies.
- API failures must appear in the page UI and must not become unhandled promises.

---

### Task 1: Transactional Organization Deletion

**Files:**
- Modify: `api/internal/store/organizations.go`
- Create: `api/internal/store/organizations_integration_test.go`
- Modify: `api/internal/httpapi/handler_test.go`
- Modify: `api/internal/httpapi/organizations_handler_test.go`

**Interfaces:**
- Consumes: `Store.DB *pgxpool.Pool`, existing `DeleteOrganization(context.Context, string) error` interface.
- Produces: the same `DeleteOrganization` signature with dependent deletion in one transaction and `pgx.ErrNoRows` for an unknown ID.

- [ ] **Step 1: Write failing handler tests for the existing contract**

Extend `fakeStore.DeleteOrganization` to capture `deletedID` and return `deleteErr`. Add tests proving a Counselor Admin gets `204` and a Counselor gets `403` without calling the store:

```go
func TestCounselorAdminDeletesOrganization(t *testing.T) {
    var deletedID string
    handler := authenticatedHandler(
        store.User{UserType: "counselor", Role: "counselor_admin", Status: "active"},
        fakeStore{deletedID: &deletedID},
    )
    response := httptest.NewRecorder()
    handler.ServeHTTP(response, authenticatedRequest(http.MethodDelete, "/api/organizations/org-1", ""))
    if response.Code != http.StatusNoContent || deletedID != "org-1" {
        t.Fatalf("unexpected response: %d id=%s", response.Code, deletedID)
    }
}
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `cd api && go test ./internal/httpapi -run 'Delete.*Organization'`

Expected: FAIL because the fake store does not capture the organization ID.

- [ ] **Step 3: Implement the transactional store deletion**

Use `pgx.BeginFunc` and execute deletes in this order after locking the organization with `SELECT id ... FOR UPDATE`:

```go
return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
    var lockedID string
    if err := tx.QueryRow(ctx, `SELECT id FROM organizations WHERE id=$1 FOR UPDATE`, id).Scan(&lockedID); err != nil {
        return err
    }
    statements := []string{
        `DELETE FROM audit_logs WHERE organization_id=$1 OR project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
        `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE organization_id=$1)`,
        `DELETE FROM invitations WHERE organization_id=$1`,
        `DELETE FROM project_subcategory_profiles WHERE project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
        `DELETE FROM project_functions WHERE project_id IN (SELECT id FROM projects WHERE organization_id=$1)`,
        `DELETE FROM projects WHERE organization_id=$1`,
        `DELETE FROM users WHERE organization_id=$1`,
        `DELETE FROM organizations WHERE id=$1`,
    }
    for _, statement := range statements {
        if _, err := tx.Exec(ctx, statement, id); err != nil { return err }
    }
    return nil
})
```

Update the fake store method:

```go
func (f fakeStore) DeleteOrganization(_ context.Context, id string) error {
    if f.deletedID != nil { *f.deletedID = id }
    return f.deleteErr
}
```

- [ ] **Step 4: Run backend verification**

Create an opt-in integration test using `TEST_DATABASE_URL`. It must insert uniquely named Counselor, Organization, Stakeholder, Project, project function/profile, session, invitation, and audit records; call `DeleteOrganization`; assert every organization-owned row count is zero; and assert the Counselor row count remains one. Use `uuid.NewString()` for unique email and token values, and clean up the Counselor in `t.Cleanup`:

```go
func TestDeleteOrganizationRemovesOwnedData(t *testing.T) {
    databaseURL := os.Getenv("TEST_DATABASE_URL")
    if databaseURL == "" { t.Skip("TEST_DATABASE_URL is not set") }
    ctx := context.Background()
    data, err := New(ctx, databaseURL)
    if err != nil { t.Fatal(err) }
    defer data.Close()

    suffix := uuid.NewString()
    var counselorID, organizationID, stakeholderID, projectID string
    mustScan := func(query string, args []any, destinations ...any) {
        if err := data.DB.QueryRow(ctx, query, args...).Scan(destinations...); err != nil { t.Fatal(err) }
    }
    mustScan(`INSERT INTO users(name,email,user_type,role,status) VALUES ('Test Counselor',$1,'counselor','counselor_admin','active') RETURNING id`, []any{"counselor-" + suffix + "@example.com"}, &counselorID)
    t.Cleanup(func() { _, _ = data.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, counselorID) })
    mustScan(`INSERT INTO organizations(name,type) VALUES ($1,'client') RETURNING id`, []any{"Delete Test " + suffix}, &organizationID)
    mustScan(`INSERT INTO users(organization_id,name,email,user_type,role,status) VALUES ($1,'Stakeholder',$2,'stakeholder','viewer','active') RETURNING id`, []any{organizationID, "stakeholder-" + suffix + "@example.com"}, &stakeholderID)
    mustScan(`INSERT INTO projects(organization_id,counselor_id,name) VALUES ($1,$2,'Delete Test') RETURNING id`, []any{organizationID, counselorID}, &projectID)
    _, err = data.DB.Exec(ctx, `
      INSERT INTO project_functions(project_id,function_id) SELECT $1,id FROM functions LIMIT 1;
      INSERT INTO project_subcategory_profiles(project_id,subcategory_id,reviewed_by) SELECT $1,id,$2 FROM subcategories LIMIT 1;
      INSERT INTO sessions(user_id,token_hash,expires_at) VALUES ($2,$3,now()+interval '1 hour');
      INSERT INTO invitations(organization_id,email,user_type,role,token_hash,invited_by,expires_at) VALUES ($4,$5,'stakeholder','viewer',$6,$7,now()+interval '1 day');
      INSERT INTO audit_logs(actor_user_id,organization_id,project_id,action,entity_type) VALUES ($7,$4,$1,'test','organization')`,
      projectID, stakeholderID, "session-"+suffix, organizationID, "invite-"+suffix+"@example.com", "invite-"+suffix, counselorID)
    if err != nil { t.Fatal(err) }
    if err := data.DeleteOrganization(ctx, organizationID); err != nil { t.Fatal(err) }
    for table, column := range map[string]string{"organizations":"id", "projects":"organization_id", "users":"organization_id", "invitations":"organization_id", "audit_logs":"organization_id"} {
        var count int
        if err := data.DB.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE `+column+`=$1`, organizationID).Scan(&count); err != nil || count != 0 { t.Fatalf("%s remains: count=%d err=%v", table, count, err) }
    }
}
```

The implementation may split the multi-statement fixture into separate `Exec` calls if pgx rejects multiple statements; preserve the same records and assertions.

- [ ] **Step 5: Run backend verification**

Run: `cd api && gofmt -w internal/store/organizations.go internal/store/organizations_integration_test.go internal/httpapi/handler_test.go internal/httpapi/organizations_handler_test.go && go test ./...`

Run integration verification against the local Docker database:

`cd api && $env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'; go test ./internal/store -run TestDeleteOrganizationRemovesOwnedData`

Expected: all Go tests PASS.

- [ ] **Step 6: Commit backend deletion**

```bash
git add api/internal/store/organizations.go api/internal/store/organizations_integration_test.go api/internal/httpapi/handler_test.go api/internal/httpapi/organizations_handler_test.go
git commit -m "feat: delete organization data transactionally"
```

### Task 2: Exact-Name Delete Confirmation

**Files:**
- Modify: `web/src/components/OrganizationDashboard.tsx`
- Modify: `web/src/components/OrganizationDashboard.test.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Consumes: `Organization`, `User.role`.
- Produces: `onDelete: (organization: Organization) => void` prop and a role-gated inline confirmation panel.

- [ ] **Step 1: Write failing component tests**

Add `onDelete={vi.fn()}` to existing renders. Add one test that opens deletion, verifies the final button remains disabled until `Acme` is typed exactly, then asserts `onDelete(organization)`. Add another render using role `counselor` and assert no `Delete Acme` button exists.

```tsx
fireEvent.click(screen.getByRole("button", { name: "Delete Acme" }));
const confirm = screen.getByLabelText(/type acme to confirm/i);
expect((screen.getByRole("button", { name: /delete permanently/i }) as HTMLButtonElement).disabled).toBe(true);
fireEvent.change(confirm, { target: { value: "Acme" } });
fireEvent.click(screen.getByRole("button", { name: /delete permanently/i }));
expect(onDelete).toHaveBeenCalledWith(organization);
```

- [ ] **Step 2: Run the component test and verify RED**

Run: `cd web && npm.cmd test -- src/components/OrganizationDashboard.test.tsx`

Expected: FAIL because the dashboard has no delete action or `onDelete` prop.

- [ ] **Step 3: Implement minimal inline confirmation**

Add local `pendingDelete: Organization | null` and `confirmation: string` state. Show `Delete {name}` only when `user.role === "counselor_admin"`. Render an inline `role="dialog"` panel with Cancel and `Delete permanently`; disable final deletion unless `confirmation === pendingDelete.name` or `loading` is true. Reset both local states after submission or cancellation.

- [ ] **Step 4: Add focused destructive-action styling**

Add only the required `.danger`, `.delete-confirmation`, and `.organization-actions` rules to `globals.css`, reusing existing color, border, spacing, and responsive variables.

- [ ] **Step 5: Run component tests and verify GREEN**

Run: `cd web && npm.cmd test -- src/components/OrganizationDashboard.test.tsx`

Expected: all dashboard tests PASS.

- [ ] **Step 6: Commit confirmation UI**

```bash
git add web/src/components/OrganizationDashboard.tsx web/src/components/OrganizationDashboard.test.tsx web/src/app/globals.css
git commit -m "feat: confirm organization deletion"
```

### Task 3: Page Integration, Error Handling, and Runtime Verification

**Files:**
- Modify: `web/src/app/page.tsx`
- Modify: `web/src/app/page.test.tsx`
- Modify: `README.md`

**Interfaces:**
- Consumes: `api.deleteOrganization(id: string): Promise<void>` and `OrganizationDashboard.onDelete`.
- Produces: immediate list removal on success and page-level error text on failure.

- [ ] **Step 1: Write failing page tests**

Add `deleteOrganization: vi.fn()` to the mocked API. Use a `counselor_admin` fixture and add tests for success and failure:

```tsx
vi.mocked(api.deleteOrganization).mockResolvedValue(undefined);
// render, open Delete Acme, type Acme, confirm
await waitFor(() => expect(api.deleteOrganization).toHaveBeenCalledWith("org-1"));
expect(screen.queryByText("Acme")).toBeNull();
```

For failure, reject with `new APIError("Could not delete organization", 500)` and assert an element with `role="alert"` shows that message while `Acme` remains visible.

- [ ] **Step 2: Run page tests and verify RED**

Run: `cd web && npm.cmd test -- src/app/page.test.tsx`

Expected: FAIL because `Home` does not pass or handle `onDelete`.

- [ ] **Step 3: Implement page deletion handling**

Add this state-owning handler and pass it as `onDelete`:

```tsx
async function deleteOrganization(item: Organization) {
  setLoading(true);
  setError("");
  try {
    await api.deleteOrganization(item.id);
    setOrganizations((rows) => rows.filter((row) => row.id !== item.id));
  } catch (cause) {
    setError(messageOf(cause));
  } finally {
    setLoading(false);
  }
}
```

- [ ] **Step 4: Update operator documentation**

Add one sentence to `README.md`: permanent organization deletion is restricted to Counselor Admin and removes all organization-owned data after exact-name confirmation.

- [ ] **Step 5: Run full verification**

Run:

```bash
cd api && go test ./...
cd web && npx.cmd tsc --noEmit --incremental false && npm.cmd test && npm.cmd run build
docker compose up --build -d
docker compose ps
```

Expected: Go tests PASS; all frontend tests PASS; Next production build exits 0; API and PostgreSQL report healthy.

- [ ] **Step 6: Browser verification**

As `admin@example.com`, create a temporary organization and project, delete the organization by typing its exact name, and verify it disappears from the list. Verify browser console has no warnings or unhandled promise errors.

- [ ] **Step 7: Commit integration and docs**

```bash
git add web/src/app/page.tsx web/src/app/page.test.tsx README.md
git commit -m "feat: integrate organization deletion"
```
