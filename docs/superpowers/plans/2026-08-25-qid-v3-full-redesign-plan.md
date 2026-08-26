# QID v3 Full Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the NIST CSF Compliance frontend as a coherent QID v3 workspace: dark-first, high-contrast, compact, and operationally clear across authentication, organization, assessment, action-plan, report, and audit-package routes while preserving the existing API contracts, roles, workflow gates, and calculations.

**Architecture:** Keep the current Next.js App Router and component boundaries. Establish the QID v3 token contract once in `globals.css`, expose the theme at the root layout, then migrate shared shell primitives and route-specific surfaces in dependency order. Markup changes are limited to improving hierarchy, landmarks, labels, and responsive composition; event handlers, API calls, authorization decisions, and status transitions remain unchanged.

**Tech Stack:** Next.js 16, React 19, TypeScript, CSS Modules-by-convention through `web/src/app/globals.css`, Vitest, Testing Library, Go API, PostgreSQL, Docker.

**Spec:** `docs/superpowers/specs/2026-08-25-qid-v3-full-redesign-design.md`

**Reference:** [Versotis QID v3 design tokens](https://www.versotis.com/assets/design/qid/tokens/v3.html)

## Global Constraints

- Do not change API routes, request payloads, database models, role permissions, workflow states, status labels, coverage formulas, or report data contracts as part of this redesign.
- Preserve the existing public route paths and slug-based URLs.
- Preserve the current semantic mapping: `submitted` is rendered as `Reviewing`, and `reviewed` is rendered as `Approved`.
- Keep Overview as the only surface for project-level summary and finalization; Assignment and Action Plan must remain task-oriented surfaces.
- Keep `no-print` behavior and the existing print output contract for reports and audit packages.
- Keep all existing mutation callbacks and loading/error/empty states. A visual refactor must not silently remove an action or change who can see it.
- Do not add a UI library, state-management library, routing library, or icon package. Use existing components and the project’s current icon approach.
- Do not commit or push automatically. Wait for an explicit user request after verification.
- Run focused tests after each task and the full frontend test suite plus typecheck/build before declaring the redesign complete.

## Task 1: Establish the QID v3 token foundation and root theme contract

**Files:**

- Modify `web/src/app/globals.css`.
- Modify `web/src/app/layout.tsx`.
- Modify `web/src/app/theme.test.ts`.
- Modify `web/src/app/responsive-layout.test.ts` only where assertions describe replaced tokens or layout constants.
- Add or update `web/src/components/ThemeToggle.test.tsx` if the root theme contract requires a new assertion.

**Interfaces and invariants:** `layout.tsx` continues to render the same children and metadata. `ThemeToggle` continues to control the same persisted theme behavior. CSS class names used by existing components remain available until the component task that consumes them is complete.

### 1.1 Write the failing contract tests first

- [ ] Replace the old palette assertions in `theme.test.ts` with exact QID v3 assertions for Space Grotesk, Inter, JetBrains Mono, brand pink/purple, dark surfaces, light surfaces, semantic states, radius, spacing, content width, and focus treatment.
- [ ] Add assertions that the stylesheet contains both explicit theme selectors and the shared theme variables:

```ts
expect(css).toContain('[data-theme="dark"]');
expect(css).toContain('--qid-bg: #0b0914;');
expect(css).toContain('--qid-surface: #13101f;');
expect(css).toContain('--qid-accent: #eb147c;');
expect(css).toContain('--qid-secondary: #6a32de;');
expect(css).toContain('--qid-focus-ring:');
```

- [ ] Add an assertion that the layout contract names the QID workspace and loads the QID font families without retaining IBM Plex Sans as the primary body face.
- [ ] Update responsive assertions that currently hard-code the old `1320px`/gutter contract so they describe the QID `1200px` content grid, 12-column desktop grid, and mobile collapse.
- [ ] Run the focused tests and record the expected failure before changing CSS:

```powershell
cd web
npm test -- --run src/app/theme.test.ts src/app/responsive-layout.test.ts src/components/ThemeToggle.test.tsx
```

### 1.2 Implement the token and root-theme layer

- [ ] Replace the old root token layer with named QID tokens while retaining aliases for selectors that are migrated later in this plan.
- [ ] Implement explicit light and dark variable sets. Dark is the default workspace presentation; light remains a first-class user-selected mode rather than relying only on `prefers-color-scheme`.
- [ ] Use this token shape as the source of truth:

```css
:root {
  --qid-font-heading: "Space Grotesk", system-ui, sans-serif;
  --qid-font-body: "Inter", system-ui, sans-serif;
  --qid-font-mono: "JetBrains Mono", ui-monospace, monospace;
  --qid-pink: #eb147c;
  --qid-purple: #6a32de;
  --qid-radius-sm: 6px;
  --qid-radius-md: 8px;
  --qid-radius-lg: 12px;
  --qid-radius-xl: 16px;
  --qid-content-max: 1200px;
  --qid-grid-gap: 24px;
}

:root,
:root[data-theme="light"] {
  --qid-bg: #faf9fd;
  --qid-surface: #ffffff;
  --qid-surface-2: #f2f0f7;
  --qid-surface-3: #e6e3f0;
  --qid-border: #dfdce8;
  --qid-text: #1a1725;
  --qid-muted: #625d75;
}

:root[data-theme="dark"] {
  --qid-bg: #0b0914;
  --qid-surface: #13101f;
  --qid-surface-2: #1c182d;
  --qid-surface-3: #26213c;
  --qid-border: #342e4f;
  --qid-text: #f8f8fc;
  --qid-muted: #a19db5;
}
```

- [ ] Keep the global focus ring, reduced-motion rule, and semantic colors usable in both themes. Do not use pink text on white or purple text on dark surfaces when contrast fails.
- [ ] Update the body contract, font links, `color-scheme`, and initial theme attributes without changing data fetching or auth behavior.

### 1.3 Verify the foundation

- [ ] Run the focused tests again and expect them to pass.
- [ ] Run `npm run typecheck` from `web`.
- [ ] Run `git diff --check -- web/src/app/globals.css web/src/app/layout.tsx web/src/app/theme.test.ts web/src/app/responsive-layout.test.ts`.

## Task 2: Build the shared app shell and authentication surfaces

**Files:**

- Modify `web/src/app/layout.tsx`.
- Modify `web/src/app/login/page.tsx` and `web/src/components/LoginForm.tsx`.
- Modify `web/src/app/forgot-password/page.tsx` and `web/src/components/ForgotPasswordForm.tsx`.
- Modify `web/src/app/reset-password/page.tsx` and `web/src/components/ResetPasswordForm.tsx`.
- Modify `web/src/app/account/password/page.tsx` and `web/src/components/ChangePasswordForm.tsx`.
- Modify `web/src/app/invite/[token]/page.tsx` and `web/src/components/AcceptInvitationForm.tsx`.
- Modify `web/src/components/ThemeToggle.tsx` and its test when needed.
- Extend the related existing page/component tests.
- Modify `web/src/app/globals.css` for appbar, auth, form, and invitation selectors.

**Interfaces and invariants:** Form submit functions, error messages, invite token handling, redirect destinations, and password rules remain unchanged. The shell must expose a consistent brand mark, current context, theme control, and focus order without requiring authenticated data.

### 2.1 Define the shell and form behavior tests

- [ ] Add assertions for a semantic app shell (`header`, `main`, and an accessible theme button) and a two-column auth layout that collapses to one column.
- [ ] Add regression assertions that login, reset, password-change, and invitation forms still expose their existing labels, submit buttons, validation errors, and loading text.
- [ ] Add a dark-mode assertion that theme controls remain reachable and do not depend on hover-only affordances.
- [ ] Run focused tests before implementation:

```powershell
cd web
npm test -- --run src/app/login/page.test.tsx src/app/invite/[token]/page.test.tsx src/components/LoginForm.test.tsx src/components/ForgotPasswordForm.test.tsx src/components/ResetPasswordForm.test.tsx src/components/ChangePasswordForm.test.tsx src/components/AcceptInvitationForm.test.tsx src/components/ThemeToggle.test.tsx
```

### 2.2 Implement the shared shell

- [ ] Give authenticated and unauthenticated surfaces a common visual grammar: 64px app bar, compact brand lockup, context label, theme control, and a centered `1200px` content frame.
- [ ] Use QID spacing and surfaces so forms read as a focused workbench rather than a generic centered card.
- [ ] Keep mobile auth forms full-width with a minimum 44px touch target and preserve keyboard-first order.
- [ ] Implement the shell with existing markup where possible; only add wrappers when they create a landmark or a responsive grid boundary:

```tsx
<div className="qid-shell">
  <header className="qid-appbar">...</header>
  <main className="qid-page-frame">...</main>
</div>
```

- [ ] Style fields with QID surface-3 backgrounds, border-2 focus states, inline errors, and visible disabled/loading states. Keep destructive/error color separate from the pink action gradient.

### 2.3 Verify auth and shell regression safety

- [ ] Run the focused auth tests and `npm run typecheck`.
- [ ] Confirm no changed auth test requires a different API response or redirect.
- [ ] Run `git diff --check` on all Task 2 files.

## Task 3: Redesign organization and project entry surfaces

**Files:**

- Modify `web/src/app/organizations/page.tsx` and `web/src/app/organizations/[organizationSlug]/page.tsx`.
- Modify `web/src/components/OrganizationDashboard.tsx` and `web/src/components/OrganizationWorkspace.tsx`.
- Modify `web/src/components/InvitationList.tsx`.
- Modify `web/src/components/ProjectDashboard.tsx`.
- Modify the matching organization/project component and page tests.
- Modify `web/src/app/globals.css` for directory, workspace, project, invitation, and empty-state selectors.

**Interfaces and invariants:** Keep organization creation, project creation/deletion, invitation lifecycle actions, role labels, permission-driven visibility, and slug links exactly as they are. Counselor Admin and Org Admin controls must remain distinct.

### 3.1 Write the entry-surface tests

- [ ] Assert that organization cards and project cards render as responsive grid items with a visible name, status, owner/context, and primary action.
- [ ] Assert that invitation rows retain resend/cancel/expire controls only when the existing permission logic says they are available.
- [ ] Assert that empty states provide one clear next action and do not render placeholder metrics as real data.
- [ ] Assert that the create-project form keeps all current fields: project name, objective, assessment period, target completion date, scope boundary, and compliance driver.
- [ ] Run focused tests before styling/markup changes:

```powershell
cd web
npm test -- --run src/app/organizations/page.test.tsx src/app/organizations/[organizationSlug]/page.test.tsx src/components/OrganizationDashboard.test.tsx src/components/OrganizationWorkspace.test.tsx src/components/InvitationList.test.tsx src/components/ProjectDashboard.test.tsx
```

### 3.2 Implement the QID directory/workspace composition

- [ ] Use a clear directory hierarchy: page eyebrow, title, supporting sentence, compact metrics, then organization/project workspaces.
- [ ] Use surface-2 cards with one accent edge or badge; reserve the pink-purple gradient for primary actions and important completion cues.
- [ ] Align project metadata to the top of each column so long objective/scope text does not make date fields appear vertically centered or detached.
- [ ] Make the create-project form a two-column QID grid on desktop and a single-column stack below 760px.
- [ ] Keep long email addresses, filenames, and project names wrapping safely:

```css
.qid-data-grid {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: var(--qid-grid-gap);
}

@media (max-width: 760px) {
  .qid-data-grid { grid-template-columns: 1fr; }
}
```

### 3.3 Verify organization/project entry behavior

- [ ] Run the focused tests, typecheck, and the existing route tests for project links and redirect behavior.
- [ ] Confirm the invitation actions still produce the same API calls and permission errors.
- [ ] Run `git diff --check` on Task 3 files.

## Task 4: Redesign the assessment workspace without changing the workflow

**Files:**

- Modify `web/src/components/FunctionSidebar.tsx`.
- Modify `web/src/components/ProjectAssessmentWorkspace.tsx`.
- Modify `web/src/components/AssessmentCard.tsx`.
- Modify `web/src/components/ProfileEditor.tsx`.
- Modify `web/src/components/StakeholderResponsePanel.tsx`.
- Modify `web/src/components/AssignmentProgress.tsx`.
- Modify `web/src/components/AuditTimeline.tsx`.
- Modify `web/src/components/ProjectFinalizationPanel.tsx`.
- Modify matching component tests.
- Modify `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx` only if a semantic wrapper or stable test hook is needed; do not change its data/handler logic.
- Modify `web/src/app/globals.css` for sidebar, workspace, outcome card, profile, evidence, final gate, and responsive selectors.

**Interfaces and invariants:** Counselor scope/assignment, stakeholder response/evidence, reviewer `Reviewing`/`Approved` transitions, assessor assignment, finalization gate, outcome inclusion, current/target profiles, and coverage calculations all retain their existing behavior.

### 4.1 Write the workspace behavior and accessibility tests

- [ ] Assert the workspace exposes the three existing surfaces—Overview, Assignment, and Action Plan—with the active surface visibly and programmatically identifiable.
- [ ] Assert only Overview renders the project summary/finalization gate; other surfaces keep their focused content.
- [ ] Assert FunctionSidebar keeps every function’s included count and percentage where those metrics currently exist, and that the metrics are not repeated on non-Overview content unnecessarily.
- [ ] Assert AssessmentCard keeps outcome code, inclusion state, role assignment, current/target route, evidence count, expand/collapse control, and status label.
- [ ] Keep the evidence count accessible and bounded so long labels cannot overflow the card.
- [ ] Assert reviewer controls remain gated by the same status/role conditions and preserve the existing visible labels `Reviewing` and `Approved`.
- [ ] Run focused tests before implementation:

```powershell
cd web
npm test -- --run src/components/FunctionSidebar.test.tsx src/components/ProjectAssessmentWorkspace.test.tsx src/components/AssessmentCard.test.tsx src/components/ProfileEditor.test.tsx src/components/StakeholderResponsePanel.test.tsx src/components/AssignmentProgress.test.tsx src/components/AuditTimeline.test.tsx src/components/ProjectFinalizationPanel.test.tsx
```

### 4.2 Implement the QID assessment workspace

- [ ] Replace the current visual hierarchy with a persistent workspace rail and a readable main column: compact rail on desktop, horizontal/condensed navigation on mobile.
- [ ] Build outcome cards around a consistent header/body/footer rhythm. Put outcome code and status at the left, title and assignment context in the center, and current-to-target profile plus evidence count in a bounded summary area.
- [ ] Use progressive disclosure for long scope context, counselor reason, profile fields, stakeholder response, and evidence. The closed card must still communicate status and next action.
- [ ] Use distinct QID surfaces for current profile, target profile, stakeholder response, review decision, and final gate. Avoid decorative gradients on large reading surfaces.
- [ ] Keep read-only counselor context visibly read-only for stakeholder/assessor views, and keep editable controls role-gated by existing props.
- [ ] Keep dropdowns for current/target priority values and preserve the existing three-value option set.
- [ ] Keep evidence actions keyboard reachable and ensure filenames wrap or truncate with a full accessible label:

```css
.evidence-count {
  min-width: 0;
  max-width: 12ch;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.outcome-card:focus-within {
  border-color: var(--qid-pink);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--qid-pink) 22%, transparent);
}
```

- [ ] Keep finalization copy and button placement clear, but do not change the finalization API or the rule that all included outcomes must be approved first.

### 4.3 Verify the assessment workflow surface

- [ ] Run all Task 4 tests and `npm run typecheck`.
- [ ] Run the project route test and confirm no role/status test requires changed behavior.
- [ ] Run the Impeccable detector against the changed assessment components:

```powershell
node .agents/skills/impeccable/scripts/detect.mjs --json --scope layout web/src/components/FunctionSidebar.tsx web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/AssessmentCard.tsx web/src/app/globals.css
```

- [ ] Run `git diff --check` on Task 4 files.

## Task 5: Redesign Action Plan, Final Report, and Audit Package surfaces

**Files:**

- Modify `web/src/components/ActionPlan.tsx`.
- Modify `web/src/components/FinalReport.tsx`.
- Modify `web/src/components/AuditPackageView.tsx`.
- Modify `web/src/components/AuditTimeline.tsx` if report/audit timeline presentation needs the shared QID treatment.
- Modify `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/report/page.tsx`.
- Modify `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/audit/page.tsx`.
- Modify `web/src/app/globals.css` for remediation rows, report sections, audit evidence, tables, and print rules.
- Modify matching tests: `ActionPlan.test.tsx`, `FinalReport.test.tsx`, `AuditPackageView.test.tsx`, `AuditTimeline.test.tsx`.

**Interfaces and invariants:** Action Plan remains the place for remediation ownership/status/dates; Final Report remains the read-facing assessment output; Audit Package remains evidence/timeline oriented. No new auditor permissions or project versioning are introduced by this visual task.

### 5.1 Write report and audit regression tests

- [ ] Assert Action Plan retains action title, linked outcome, owner, status, due date, notes, and empty/loading/error states.
- [ ] Assert Final Report retains approved outcome data, coverage/function summaries, report metadata, and print controls.
- [ ] Assert Audit Package retains audit timeline, evidence metadata, package sections, and download/print actions.
- [ ] Assert tables are wrapped in a scroll container on narrow widths and that `.no-print` controls are hidden only during print.
- [ ] Run focused tests before implementation:

```powershell
cd web
npm test -- --run src/components/ActionPlan.test.tsx src/components/FinalReport.test.tsx src/components/AuditPackageView.test.tsx src/components/AuditTimeline.test.tsx
```

### 5.2 Implement the QID report hierarchy

- [ ] Use report-level page headers, metadata rows, section bands, and readable tables rather than adding more nested cards.
- [ ] Give remediation status and due dates strong scan hierarchy with semantic badges; reserve warning/error colors for actual risk states.
- [ ] Keep long evidence filenames and outcome text readable on mobile and in print.
- [ ] Keep print styles explicit and isolated:

```css
@media print {
  .no-print { display: none !important; }
  .qid-report-shell { max-width: none; padding: 0; }
  .qid-report-section { break-inside: avoid; }
}
```

### 5.3 Verify report and audit behavior

- [ ] Run focused report/audit tests, typecheck, and the report/audit page tests.
- [ ] Confirm no report data fields, API calls, or print selectors were removed.
- [ ] Run `git diff --check` on Task 5 files.

## Task 6: Cross-route responsive, theme, and accessibility QA

**Files:**

- Modify `web/src/app/responsive-layout.test.ts`.
- Modify `web/src/app/theme.test.ts` and `web/src/components/ThemeToggle.test.tsx` if findings require more coverage.
- Add only narrowly scoped tests beside the component under test when a discovered regression cannot be expressed in existing suites.
- Modify `web/src/app/globals.css` and the affected component only for verified QA findings.

**Interfaces and invariants:** This task is a verification pass. Any fix must be a UI-only correction and must keep route behavior, API contracts, and role gates unchanged.

### 6.1 Add deterministic responsive/theme assertions

- [ ] Assert 12-column desktop layout, 260px navigation rail, 1200px content max, 24px grid gap, and the 760px single-column breakpoint.
- [ ] Assert all interactive controls have a visible focus style and minimum 44px touch height where the product exposes them.
- [ ] Assert cards, evidence names, emails, long project metadata, and tables wrap without horizontal page overflow.
- [ ] Assert reduced-motion CSS disables workspace transitions while preserving readable state changes.
- [ ] Run the complete deterministic suite:

```powershell
cd web
npm test
npm run typecheck
```

### 6.2 Perform visual QA at representative routes

- [ ] Start the existing local app using the project’s documented Docker/dev command and inspect these routes in both themes:
  - `/login`
  - `/organizations`
  - `/organizations/{organizationSlug}`
  - `/organizations/{organizationSlug}/projects/{projectSlug}` on Overview, Assignment, and Action Plan
  - `/organizations/{organizationSlug}/projects/{projectSlug}/report`
  - `/organizations/{organizationSlug}/projects/{projectSlug}/audit`
- [ ] Check at desktop width (1440×900), tablet width (1024×900), and mobile width (390×844).
- [ ] Verify the following manually: no clipped evidence count, no metadata vertical drift, no empty right-side dead zone on desktop, no unreadable dark-theme text, no hidden primary action, and clear role/status language.
- [ ] Capture screenshots only for actual defects or review evidence; do not add screenshot assets to the repository unless explicitly requested.

### 6.3 Run the Impeccable visual checks

- [ ] Run the detector over the final changed frontend scope:

```powershell
node .agents/skills/impeccable/scripts/detect.mjs --json --scope layout web/src/app web/src/components
```

- [ ] Run the full frontend build and confirm it completes:

```powershell
cd web
npm run build
```

- [ ] Resolve every actionable detector/build/test finding before moving to documentation.

## Task 7: Update design documentation and handoff artifacts

**Files:**

- Modify `DESIGN.md` to describe the implemented QID v3 visual system, route shell, theme behavior, responsive breakpoints, accessibility rules, and component patterns.
- Modify `.impeccable/design.json` only to reflect tokens and decisions that are actually present in code.
- Modify `README.md` only if the redesign changes local run/verification instructions; do not rewrite product workflow documentation unnecessarily.

**Interfaces and invariants:** Documentation must describe the shipped UI, not aspirational features. It must not claim auditor role, versioning, production deployment, HTTPS, secrets, or backup work that is outside this redesign.

### 7.1 Document the implemented system

- [ ] Record the QID token source, light/dark values, typography, layout grid, content width, navigation rail, status semantics, focus rules, and print behavior.
- [ ] Record the route-to-surface map and the rule that Overview owns project-level summary/finalization.
- [ ] Record any compatibility aliases retained for existing CSS selectors and the intended migration boundary.

### 7.2 Final verification and handoff

- [ ] Run the repository placeholder-copy scan against `web/src`, `DESIGN.md`, and `.impeccable/design.json`; resolve any new placeholder copy before handoff.
- [ ] Run `git diff --check` across all changed files.
- [ ] Run the complete frontend test, typecheck, build, and detector commands from Task 6 one final time.
- [ ] Report changed files, verification commands/results, any known non-blocking warnings, and the fact that no commit/push was performed.

## Completion Criteria

- [ ] All routes use the QID v3 token layer and share the same shell language.
- [ ] Dark and light themes both meet contrast and focus requirements.
- [ ] Desktop, tablet, and mobile layouts preserve all primary actions without clipping or unexplained empty columns.
- [ ] Counselor, Org Admin, Stakeholder, Assessor, and Reviewer surfaces retain their existing visibility and workflow behavior.
- [ ] Overview, Assignment, Action Plan, Final Report, and Audit Package remain distinct and understandable.
- [ ] Evidence counts, filenames, long metadata, and table content remain readable.
- [ ] Existing tests pass, new visual-contract tests pass, typecheck passes, build passes, and the Impeccable detector reports no actionable layout findings.
- [ ] `DESIGN.md` and `.impeccable/design.json` match the implemented design.
