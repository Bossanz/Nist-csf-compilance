# Delete Project Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Permanently delete one project from the dashboard after explicit confirmation.

**Architecture:** Add a transactional delete method behind the existing Go store interface and expose it as `DELETE /api/projects/:id`. Extend the existing dashboard callbacks and Home state; keep confirmation in the browser and remove a card only after the API succeeds.

**Tech Stack:** Go standard-library HTTP, pgx/PostgreSQL, Next.js 16, React 19, TypeScript, Vitest, Testing Library, Docker Compose.

## Global Constraints

- No schema migration or dependency is added.
- Child project rows use existing database cascades.
- An organization is removed only when no project or user references it.
- Cancel and API failure never remove a dashboard card.
- Use red-green-refactor for each behavior.

---

### Task 1: DELETE Project API

**Files:**
- Modify: `api/internal/store/projects.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/handler_test.go`

**Interfaces:**
- Produces: `DeleteProject(context.Context, string) error`
- Produces: `DELETE /api/projects/:id -> 204`
- Uses: `pgx.ErrNoRows` as the store-level not-found signal

- [ ] **Step 1: Write failing handler tests**

Extend `fakeStore` with `deleteErr error` and `deletedID *string`. Its value-receiver `DeleteProject` writes through the optional pointer and returns the configured error, preserving compatibility with existing tests that pass `fakeStore` by value. Add tests for 204, `pgx.ErrNoRows -> 404`, a malformed UUID -> 404, and a generic error -> 500.

```go
func TestDeleteProject(t *testing.T) {
    var deletedID string
    fake := fakeStore{deletedID: &deletedID}
    projectID := "11111111-1111-1111-1111-111111111111"
    r := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID, nil)
    w := httptest.NewRecorder()
    (&Handler{Store: fake}).ServeHTTP(w, r)
    if w.Code != http.StatusNoContent || w.Body.Len() != 0 || deletedID != projectID {
        t.Fatalf("unexpected response: %d %s id=%s", w.Code, w.Body.String(), deletedID)
    }
}
```

- [ ] **Step 2: Run RED**

Run: `cd api; go test ./internal/httpapi -run DeleteProject -v`

Expected: compilation fails because the store interface and DELETE route do not exist.

- [ ] **Step 3: Implement minimal delete behavior**

Validate the path ID with `uuid.Parse` and return 404 when malformed. In one transaction, delete the project with `RETURNING organization_id`; then run:

```sql
DELETE FROM organizations o
WHERE o.id=$1
  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.organization_id=o.id)
  AND NOT EXISTS (SELECT 1 FROM users u WHERE u.organization_id=o.id)
```

Map `errors.Is(err, pgx.ErrNoRows)` to 404 and other errors to 500. Return 204 without calling `writeJSON`.

- [ ] **Step 4: Run GREEN and commit**

Run: `cd api; go test ./...`

```powershell
git add api/internal/store/projects.go api/internal/httpapi/handler.go api/internal/httpapi/handler_test.go
git commit -m "feat: delete compliance projects"
```

---

### Task 2: Dashboard Delete Confirmation

**Files:**
- Modify: `web/src/components/ProjectDashboard.tsx`
- Modify: `web/src/components/ProjectDashboard.test.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Adds: `onDelete(project: Project): void`
- Reuses: `openingID` to disable all project actions while deleting

- [ ] **Step 1: Write failing component tests**

Stub `window.confirm`. Verify confirmation text includes the project name, confirmation calls `onDelete(project)`, cancellation does not call it, and an active `openingID` disables Open and Delete.

```tsx
test("confirms before deleting a project", () => {
  vi.spyOn(window, "confirm").mockReturnValue(true);
  const onDelete = vi.fn();
  render(<ProjectDashboard {...baseProps} projects={[project]} onDelete={onDelete} />);
  fireEvent.click(screen.getByRole("button", { name: /delete readiness review/i }));
  expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining("Readiness Review"));
  expect(onDelete).toHaveBeenCalledWith(project);
});
```

- [ ] **Step 2: Run RED**

Run: `cd web; npm.cmd test -- src/components/ProjectDashboard.test.tsx`

Expected: tests fail because `onDelete` and the Delete button do not exist.

- [ ] **Step 3: Implement the minimal UI**

Add a `Delete` button with accessible name `Delete ${project.name}`. Its click handler calls `window.confirm("Delete ${project.name}? Its assessment data will be permanently deleted.")`; invoke `onDelete` only when true. Group Open and Delete in `.project-actions`; style delete as a restrained danger action using the existing `--error` token.

- [ ] **Step 4: Run GREEN and commit**

Run: `cd web; npm.cmd test; npm.cmd run typecheck`

```powershell
git add web/src/components/ProjectDashboard.tsx web/src/components/ProjectDashboard.test.tsx web/src/app/globals.css
git commit -m "feat: confirm project deletion"
```

---

### Task 3: Home State and Real Database Verification

**Files:**
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/app/page.tsx`
- Modify: `web/src/app/page.test.tsx`
- Modify: `scripts/smoke-test.ps1`

**Interfaces:**
- Adds: `api.deleteProject(id: string): Promise<void>`
- Adds: Home `deleteProject(project: Project)` callback

- [ ] **Step 1: Write failing Home workflow test**

Render two projects, confirm deletion, click Delete on the first, and verify only that card disappears after the mocked DELETE resolves. Use a project-ID-checking fake so a wrong ID rejects.

- [ ] **Step 2: Run RED**

Run: `cd web; npm.cmd test -- src/app/page.test.tsx`

Expected: test fails because `api.deleteProject` and the Home callback do not exist.

- [ ] **Step 3: Implement API and state coordination**

Update the shared request helper to return `undefined` for HTTP 204 before parsing JSON. Add `deleteProject`. In Home, set `openingID`, call the API, filter the matching project from state only after success, preserve it on error, and pass the callback to `ProjectDashboard`.

- [ ] **Step 4: Extend smoke verification**

After existing profile and summary checks, call DELETE, require HTTP 204, verify the project no longer appears in `GET /api/projects`, and require the deleted project endpoint to return 404.

- [ ] **Step 5: Run full verification**

Run Go tests, all web tests, typecheck, production build, `docker compose config`, Docker build/up, smoke test, and browser desktop/mobile checks. Then run `docker compose down`.

- [ ] **Step 6: Commit integrated deletion**

```powershell
git add web/src/lib/api.ts web/src/app/page.tsx web/src/app/page.test.tsx scripts/smoke-test.ps1
git commit -m "feat: remove deleted projects from dashboard"
```
