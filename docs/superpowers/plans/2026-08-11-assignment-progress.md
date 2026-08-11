# Counselor Assignment Progress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Counselor-only assignment progress strip that counts included, assigned, and unassigned outcomes from the existing profile state.

**Architecture:** Keep the feature client-side and derive counts in `ProjectAssessmentWorkspace` from the profile rows already loaded by `ProjectPage`. Render a focused `AssignmentProgress` component so the display and count logic are independently testable; do not add an API or database field.

**Tech Stack:** Next.js, React, TypeScript, Vitest, Testing Library, existing CSS tokens.

## Global Constraints

- Preserve the current Counselor/Stakeholder/Reviewer/Viewer permission boundaries.
- Do not add a new API endpoint or database migration for display-only counts.
- Keep the existing white editorial layout and responsive behavior.
- Keep included-but-unassigned outcomes hidden from stakeholder users.

---

### Task 1: Add the failing assignment progress tests

**Files:**
- Create: `web/src/components/AssignmentProgress.test.tsx`
- Modify: `web/src/components/ProjectAssessmentWorkspace.test.tsx`

**Interfaces:**
- `AssignmentProgress({ included, assigned, unassigned })` renders the three labeled counts.
- `ProjectAssessmentWorkspace` renders `AssignmentProgress` only when `user.userType === "counselor"`.

- [ ] **Step 1: Write the failing component test**

```tsx
test("shows included, assigned, and waiting counts", () => {
  render(<AssignmentProgress included={12} assigned={5} unassigned={7} />);
  expect(screen.getByText("Included").parentElement).toHaveTextContent("12");
  expect(screen.getByText("Assigned").parentElement).toHaveTextContent("5");
  expect(screen.getByText("Waiting for assignment").parentElement).toHaveTextContent("7");
});
```

- [ ] **Step 2: Add workspace visibility/count assertions**

Use the existing `profile` fixture with two assigned included rows and one excluded row, then add one included unassigned row. Assert that a Counselor sees `Included 3`, `Assigned 2`, and `Waiting for assignment 1`; assert that an Assessor does not see the `Assignment progress` region.

- [ ] **Step 3: Run the focused tests and verify the expected failure**

Run from `web`:

```powershell
npm.cmd test -- AssignmentProgress ProjectAssessmentWorkspace
```

Expected: FAIL because `AssignmentProgress` and its workspace rendering do not exist yet.

### Task 2: Implement the display-only progress strip

**Files:**
- Create: `web/src/components/AssignmentProgress.tsx`
- Modify: `web/src/components/ProjectAssessmentWorkspace.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- `AssignmentProgress` accepts three non-negative numeric props and renders an accessible region named `Assignment progress`.
- The workspace derives counts with `useMemo` from `profile` and renders the component only for Counselor users.

- [ ] **Step 1: Implement the minimal component**

```tsx
export function AssignmentProgress({ included, assigned, unassigned }: Props) {
  return (
    <section className="assignment-progress" aria-label="Assignment progress">
      <strong>Assignment progress</strong>
      <span><b>{included}</b> Included</span>
      <span><b>{assigned}</b> Assigned</span>
      <span><b>{unassigned}</b> Waiting for assignment</span>
    </section>
  );
}
```

- [ ] **Step 2: Derive and render counts in the workspace**

```tsx
const assignmentProgress = useMemo(() => {
  const includedRows = profile.filter((row) => row.included);
  const assigned = includedRows.filter((row) => row.assignedUserID !== null).length;
  return { included: includedRows.length, assigned, unassigned: includedRows.length - assigned };
}, [profile]);
```

Render it after `<SummaryCards summary={summary} />` only when `isCounselor` is true.

- [ ] **Step 3: Add responsive styles**

Use existing border, surface, muted, and accent tokens. Make the strip wrap on small screens and keep the labels readable without changing the existing assessment card layout.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run:

```powershell
npm.cmd test -- AssignmentProgress ProjectAssessmentWorkspace
```

Expected: all focused tests pass.

### Task 3: Verify the workflow and commit

**Files:**
- Modify: `README.md`
- Modify: `api/internal/store/projects.go`
- Modify: `api/internal/store/profile_assignment_integration_test.go`

**Interfaces:**
- The existing bug fix remains part of this change: Counselors may include an outcome before assigning a stakeholder, while invalid stakeholder assignments remain rejected.

- [ ] **Step 1: Run the full web test suite**

```powershell
npm.cmd test
npm.cmd run typecheck -- --incremental false
npm.cmd run build
```

- [ ] **Step 2: Run the Go suite against local PostgreSQL**

```powershell
$env:TEST_DATABASE_URL = "postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable"
go test ./...
```

- [ ] **Step 3: Rebuild and health-check Docker**

```powershell
docker compose up --build -d
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:3000
```

Expected: API and web return HTTP 200 and all containers are healthy/running.

- [ ] **Step 4: Review the diff and commit**

```powershell
git diff --check
git add README.md api/internal/store/projects.go api/internal/store/profile_assignment_integration_test.go web/src/components/AssignmentProgress.tsx web/src/components/AssignmentProgress.test.tsx web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx web/src/app/globals.css docs/superpowers/specs/2026-08-11-assignment-progress-design.md docs/superpowers/plans/2026-08-11-assignment-progress.md
git commit -m "feat: show counselor assignment progress"
```
