# Vanta-inspired Compliance Workspace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (\`- [ ]\`) syntax for tracking.

**Goal:** Recompose the existing V3 organization and project surfaces into a status-first compliance workspace with work registers and role-aware next actions, without changing the backend workflow.

**Architecture:** Keep the current page containers and callbacks. Add small pure view helpers/components for derived portfolio and project overview data, then replace card-heavy compositions with semantic registers and an overview work queue. Use the existing QID variables and class contracts, adding a scoped visual layer in \`globals.css\` for the new register geometry.

**Tech Stack:** Next.js 16, React 19, TypeScript, Vitest, Testing Library, existing CSS custom properties.

**Spec:** \`docs/superpowers/specs/2026-09-03-vanta-inspired-compliance-workspace-design.md\`

## Current implementation status

- [x] Approved Evidence Workbench direction recorded in the root design contract.
- [x] Organization portfolio and organization project lists use readable status-first registers.
- [x] Project Overview has posture metrics, role-aware `Next up`, and a Function register.
- [x] Existing Overview-only summary and finalization behavior remains available without duplicate metric panels.
- [x] Focused component, integration, full frontend tests, typecheck, and production build verified.
- [x] Impeccable detector review passed with no mechanical findings.
- [ ] Commit remains optional until explicitly requested.

## Global Constraints

- Preserve existing API, workflow, role permissions, calculations, and content terminology.
- Use the QiD/Versotis palette and 60/30/10 rule: neutral work surfaces dominate, light navy/graphite structure supports, magenta-purple marks action and orientation.
- Keep Overview as the initial project surface and keep Assignment, Action Plan, and Log behavior unchanged.
- Do not add dependencies, routes, API fields, database changes, AI, integrations, or new status values.
- Status and permissions must always be understandable from text, not color alone.
- UI production code is written only after a test has demonstrated the intended behavior fails.

---

### Task 1: Lock the approved direction and derived portfolio helpers

**Files:**
- Modify: \`web/src/app/layout.tsx\`
- Create: \`web/src/lib/workspaceMetrics.ts\`
- Test: \`web/src/lib/workspaceMetrics.test.ts\`

**Interfaces:**
- \`getOrganizationPortfolioMetrics(organizations: Organization[], projectCounts?: Record<string, number>): { total: number; active: number; finalized: number; totalProjects: number }\`
- \`getProjectStatusLabel(status: string): string\`
- \`getAttentionLabel(role: User["role"]): string\`

- [ ] **Step 1: Write the failing tests**

~~~ts
import { expect, test } from "vitest";
import { getAttentionLabel, getOrganizationPortfolioMetrics, getProjectStatusLabel } from "./workspaceMetrics";

test("derives portfolio workspace counts from available project totals", () => {
  expect(getOrganizationPortfolioMetrics(
    [{ id: "org-1", name: "Acme", slug: "acme", type: "client" }, { id: "org-2", name: "Beta", slug: "beta", type: "client" }],
    { "org-1": 2, "org-2": 1 },
  )).toEqual({ total: 2, active: 2, finalized: 0, totalProjects: 3 });
});

test("formats workflow statuses for the register", () => {
  expect(getProjectStatusLabel("in_review")).toBe("Reviewing");
  expect(getProjectStatusLabel("closed")).toBe("Finalized");
});

test("names the role-aware next action", () => {
  expect(getAttentionLabel("reviewer")).toBe("To review");
  expect(getAttentionLabel("assessor")).toBe("Needs your input");
});
~~~

- [ ] **Step 2: Run the focused test and verify it fails**

Run: \`npm test -- --run src/lib/workspaceMetrics.test.ts\` from \`web\`.

Expected: FAIL because \`workspaceMetrics.ts\` does not exist.

- [ ] **Step 3: Implement the minimal pure helpers**

Create the named helpers. Count a project as finalized only when its status is \`closed\`; active is every non-finalized project. Keep the project count optional so the \`/organizations\` page can use the data it already has without a new request.

- [ ] **Step 4: Run the focused test and verify it passes**

Run: \`npm test -- --run src/lib/workspaceMetrics.test.ts\` from \`web\`.

Expected: 3 tests pass.

- [ ] **Step 5: Update the direction contract**

Update the existing contract string and source contract comment in \`layout.tsx\` to name the Evidence Workbench direction and seed \`250b41bb\`; keep the existing product truth and finish line. Add \`data-design-direction="evidence-workbench-250b41bb"\` to the body so the direction survives the production build.

- [ ] **Step 6: Commit**

~~~bash
git add web/src/app/layout.tsx web/src/lib/workspaceMetrics.ts web/src/lib/workspaceMetrics.test.ts
git commit -m "chore: lock compliance workspace direction"
~~~

### Task 2: Recompose the organization portfolio into a work register

**Files:**
- Modify: \`web/src/components/OrganizationDashboard.tsx\`
- Modify: \`web/src/components/OrganizationDashboard.test.tsx\`
- Modify: \`web/src/app/globals.css\`

**Interfaces:**
- Keep the existing \`OrganizationDashboard\` props and callbacks unchanged.
- Render a \`.portfolio-summary\` region with four text-labelled metrics.
- Render each organization as an \`.organization-register-row\` with name, type, project/people context if available, and existing open/delete actions.

- [ ] **Step 1: Write the failing tests**

Add these tests to \`OrganizationDashboard.test.tsx\`:

~~~tsx
test("shows a status-first client portfolio summary", () => {
  render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} />);

  expect(screen.getByRole("region", { name: /portfolio posture/i })).toBeTruthy();
  expect(screen.getByText("Client workspaces")).toBeTruthy();
  expect(screen.getByText("Ready to open")).toBeTruthy();
});

test("presents the client workspace as a register row", () => {
  render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onCreate={vi.fn()} onLogout={vi.fn()} />);

  expect(screen.getByRole("region", { name: /client workspace register/i })).toBeTruthy();
  expect(screen.getByRole("button", { name: /open acme/i })).toBeTruthy();
});
~~~

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: \`npm test -- --run src/components/OrganizationDashboard.test.tsx\` from \`web\`.

Expected: FAIL because the new region and register labels are not rendered.

- [ ] **Step 3: Implement the portfolio composition**

Add the posture summary after the header. Replace only the organization card markup with a semantic register while keeping the \`onSelect\`, protected delete trigger, empty state, loading state, error state, counselor controls, and auth links. Use the helper for labels; do not invent project/people counts that are not in the current props.

- [ ] **Step 4: Add scoped register CSS**

Create the \`.portfolio-summary\`, \`.organization-register\`, \`.organization-register-row\`, \`.register-primary\`, \`.register-meta\`, and responsive rules in the final application layer of \`globals.css\`. Use borders and tonal surfaces; no new gradients or hard offset shadows.

- [ ] **Step 5: Run the focused tests and verify they pass**

Run: \`npm test -- --run src/components/OrganizationDashboard.test.tsx\` from \`web\`.

Expected: all organization dashboard tests pass.

- [ ] **Step 6: Commit**

~~~bash
git add web/src/components/OrganizationDashboard.tsx web/src/components/OrganizationDashboard.test.tsx web/src/app/globals.css
git commit -m "feat: make organization portfolio status-first"
~~~

### Task 3: Recompose the organization project list into a workspace register

**Files:**
- Modify: \`web/src/components/OrganizationWorkspace.tsx\`
- Modify: \`web/src/components/OrganizationWorkspace.test.tsx\`
- Modify: \`web/src/app/globals.css\`

**Interfaces:**
- Keep all \`OrganizationWorkspace\` props and callbacks unchanged.
- Render a \`.workspace-register\` region with one \`.workspace-register-row\` per project.
- Preserve project creation and all stakeholder/invitation sections below the active workspace register.

- [ ] **Step 1: Write the failing test**

Add:

~~~tsx
test("shows projects as an organization workspace register", () => {
  render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()} />);

  expect(screen.getByRole("region", { name: /project workspace register/i })).toBeTruthy();
  expect(screen.getByText("Assessment v1")).toBeTruthy();
  expect(screen.getByRole("button", { name: /open readiness/i })).toBeTruthy();
});
~~~

- [ ] **Step 2: Run the focused test and verify it fails**

Run: \`npm test -- --run src/components/OrganizationWorkspace.test.tsx\` from \`web\`.

Expected: FAIL because the project register region is not rendered.

- [ ] **Step 3: Implement the project register**

Replace the project card grid markup with rows that keep the existing status chip, version label, creation date, project name, organization context, open callback, and counselor-only delete callback. Leave form fields and access controls untouched.

- [ ] **Step 4: Add responsive register CSS**

Reuse the existing QID variables and add desktop columns for project identity, status/version, date, and actions. Stack these regions at the mobile breakpoint and keep buttons full-width where necessary.

- [ ] **Step 5: Run the focused test and verify it passes**

Run: \`npm test -- --run src/components/OrganizationWorkspace.test.tsx\` from \`web\`.

Expected: all organization workspace tests pass.

- [ ] **Step 6: Commit**

~~~bash
git add web/src/components/OrganizationWorkspace.tsx web/src/components/OrganizationWorkspace.test.tsx web/src/app/globals.css
git commit -m "feat: turn organization projects into a work register"
~~~

### Task 4: Add the project overview workbench and role-aware next-up queue

**Files:**
- Create: \`web/src/components/ProjectOverviewWorkbench.tsx\`
- Create: \`web/src/components/ProjectOverviewWorkbench.test.tsx\`
- Modify: \`web/src/components/ProjectAssessmentWorkspace.tsx\`
- Modify: \`web/src/components/ProjectAssessmentWorkspace.test.tsx\`
- Modify: \`web/src/app/globals.css\`

**Interfaces:**
- \`ProjectOverviewWorkbench\` accepts the existing derived data only:

~~~ts
type ProjectOverviewWorkbenchProps = {
  user: User;
  project: Project;
  summary: Summary;
  functions: FunctionNode[];
  functionProgress: Record<string, FunctionProgress>;
  scopeSubmitted: boolean;
  isCounselor: boolean;
  onOpenAssignment: () => void;
};
~~~

- It renders \`Project posture\`, \`Next up\`, and \`Function register\` regions and does not own API calls or mutation state.

- [ ] **Step 1: Write the failing component tests**

~~~tsx
import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ProjectOverviewWorkbench } from "./ProjectOverviewWorkbench";

test("shows role-aware next work and function posture", () => {
  const onOpenAssignment = vi.fn();
  render(<ProjectOverviewWorkbench user={{ id: "a", organizationID: "o", name: "Assessor", email: "a@example.com", userType: "stakeholder", role: "assessor", status: "active" }} project={{ id: "p", organizationID: "o", organizationName: "Acme", name: "Readiness", slug: "readiness", status: "in_review", createdAt: "2026-08-01T00:00:00Z" }} summary={{ coveragePct: 50, includedCount: 2, pendingCount: 1, rejectedCount: 0, functions: [{ code: "GV", coveragePct: 50, includedCount: 2 }] }} functions={[{ id: "f", code: "GV", name: "Govern", description: "", categories: [] }]} functionProgress={{ GV: { value: 2, label: "included", coveragePct: 50, includedCount: 2, attention: 1, attentionLabel: "open" } }} scopeSubmitted={true} isCounselor={false} onOpenAssignment={onOpenAssignment} />);

  expect(screen.getByRole("region", { name: /project posture/i })).toBeTruthy();
  expect(screen.getByRole("region", { name: /next up/i }).textContent).toContain("Needs your input");
  expect(screen.getByRole("region", { name: /function register/i }).textContent).toContain("50%");
  fireEvent.click(screen.getByRole("button", { name: /open assignment/i }));
  expect(onOpenAssignment).toHaveBeenCalledOnce();
});
~~~

- [ ] **Step 2: Run the focused test and verify it fails**

Run: \`npm test -- --run src/components/ProjectOverviewWorkbench.test.tsx\` from \`web\`.

Expected: FAIL because the component does not exist.

- [ ] **Step 3: Implement the presentational workbench**

Create a pure component. Use the role to choose a short next-up label: counselors see scope/assignment or review readiness; assessors/org admins see their input; reviewers see responses to review; viewers/auditors see read-only posture. Show only data passed in. Keep a single primary “Open Assignment” action.

- [ ] **Step 4: Run the focused component test and verify it passes**

Run: \`npm test -- --run src/components/ProjectOverviewWorkbench.test.tsx\` from \`web\`.

Expected: 1 test passes.

- [ ] **Step 5: Integrate the workbench into Overview**

Render it immediately after \`SummaryCards\` and before Version History. Pass the already-derived \`functionProgress\`, existing role flags, and a callback that sets the current surface to \`assignment\`. Do not move or duplicate finalization/remediation behavior.

- [ ] **Step 6: Add overview register CSS**

Style \`.project-overview-workbench\`, \`.project-posture\`, \`.next-up\`, \`.function-register\`, \`.function-register-row\`, and their mobile rules. Use one wide register with columns for Function, Coverage, Included, and Attention. Use text labels for all statuses and preserve the current/target semantic colors elsewhere.

- [ ] **Step 7: Update integration expectations**

Add assertions to \`ProjectAssessmentWorkspace.test.tsx\` that Overview contains \`Next up\`, \`Function register\`, and the existing summary/finalization regions while Assignment does not contain these overview-only regions.

- [ ] **Step 8: Run focused integration tests and verify they pass**

Run: \`npm test -- --run src/components/ProjectAssessmentWorkspace.test.tsx\` from \`web\`.

Expected: all project workspace tests pass.

- [ ] **Step 9: Commit**

~~~bash
git add web/src/components/ProjectOverviewWorkbench.tsx web/src/components/ProjectOverviewWorkbench.test.tsx web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProjectAssessmentWorkspace.test.tsx web/src/app/globals.css
git commit -m "feat: add role-aware project overview workbench"
~~~

### Task 5: Finish the visual system, responsive states, and verification

**Files:**
- Modify: \`web/src/app/globals.css\`
- Modify: \`web/src/app/responsive-layout.test.tsx\`
- Modify: \`web/src/app/theme.test.ts\`
- Modify: \`DESIGN.md\`

- [ ] **Step 1: Add failing CSS contract assertions**

Extend the existing theme/responsive tests with assertions that changed UI selectors use QID variables, that the project register has a mobile layout rule, and that the overview workbench is present in both themes.

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: \`npm test -- --run src/app/theme.test.ts src/app/responsive-layout.test.tsx\` from \`web\`.

Expected: FAIL until the new selectors and rules are added.

- [ ] **Step 3: Add the scoped final CSS layer**

Make the registers readable at desktop and mobile, preserve visible focus, keep body copy within the reading measure, and remove reliance on obsolete grid/card presentation for the changed surfaces. Add no new external fonts or dependencies. Keep dark and light token branches symmetric.

- [ ] **Step 4: Update durable design documentation**

Update \`DESIGN.md\` to record the Evidence Workbench direction, the portfolio/project register anatomy, role-aware Next up block, and restrained color strategy while retaining all existing product commitments.

- [ ] **Step 5: Run the full frontend verification**

Run from \`web\`:

~~~bash
npm test
npm run typecheck
npm run build
~~~

Expected: exit code 0 for all three commands.

- [ ] **Step 6: Run the Impeccable detector on changed UI**

Run from the repository root:

~~~bash
node .agents/skills/impeccable/scripts/detect.mjs --json
~~~

Expected: no mechanical findings that require changes; record any remaining review-only findings for handoff.

- [ ] **Step 7: Inspect the final rendered surface**

Capture one desktop and one mobile view of \`/organizations\`, an organization workspace, and a project Overview in both theme settings if the running local stack is available. Check first viewport hierarchy, register overflow, focus visibility, dark/light contrast, and that Assignment remains the only detailed outcome surface.

- [ ] **Step 8: Commit the final design pass**

~~~bash
git add web/src/app/globals.css web/src/app/responsive-layout.test.tsx web/src/app/theme.test.ts DESIGN.md
git commit -m "polish: finish compliance workspace redesign"
~~~
