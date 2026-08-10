# Clean Editorial Casefile Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current dashboard-heavy visual treatment with a calm, white-first Clean Editorial Casefile layout that makes Counselor reading and Stakeholder input visibly distinct.

**Architecture:** Keep the existing Next.js page state, API calls, role checks, and assessment components. Add one small read-only assessment context rail, reorganize the existing project view into a reading column with supporting rails, and use a compact CSS token system for the 60/30/10 palette and responsive layout.

**Tech Stack:** Next.js 16, React 19, TypeScript, plain CSS, Vitest, Testing Library.

## Global Constraints

- Preserve existing API, workflow, role permissions, calculations, and content terminology.
- Counselor reads/interprets/reviews; Stakeholder enters responses and evidence.
- Stakeholders must not gain counselor-only editing controls or unselected outcomes.
- Use white surfaces, light neutral structure, and restrained teal in an approximate 60/30/10 balance.
- Do not add a UI framework, state library, icon package, or backend endpoint.
- Keep the existing system font stack and support long reading sessions, keyboard focus, reduced motion, and mobile reflow.

---

## Task 1: Establish the root contract and visual tokens

**Files:**
- Modify: `web/src/app/layout.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Produces the document-level Clean Editorial Casefile contract, CSS tokens, surfaces, controls, and responsive primitives used by all later tasks.

- [ ] **Step 1: Add the root contract marker before the page content**

Keep the contract as the first source-level child of `<body>` and retain an emitted, invisible metadata marker so the production HTML can be audited for the seed key. Preserve `lang="en"` and the existing metadata.

- [ ] **Step 2: Replace the old visual tokens and layout rules**

Update `globals.css` so white is the primary canvas/surface, neutral lines and panels structure the page, teal is reserved for active/action/status, shadows are minimal, reading measure is constrained, and the existing class names continue to work while later markup changes land.

- [ ] **Step 3: Verify the token-only change**

Run from `web`:

```powershell
npm run test -- --runInBand
```

Expected: the existing suite still passes; no component behavior changes are introduced by CSS.

## Task 2: Add the contextual assessment rail with TDD

**Files:**
- Create: `web/src/components/AssessmentRail.tsx`
- Create: `web/src/components/AssessmentRail.test.tsx`
- Modify: `web/src/app/page.tsx`

**Interfaces:**
- `AssessmentRail({ role, project, summary, selectedCode, outcomeCount })` renders a read-only `<aside aria-label="Assessment context">`.
- Counselor copy emphasizes reading/review; stakeholder copy emphasizes completing responses/evidence.

- [ ] **Step 1: Write the failing role hierarchy tests**

Add tests that render the rail for `counselor` and `assessor`, assert the landmark exists, assert the counselor text contains `Read and review`, assert the stakeholder text contains `Complete assigned inputs`, and assert coverage plus current/target context are visible.

- [ ] **Step 2: Run the focused test and confirm RED**

```powershell
npm run test -- AssessmentRail.test.tsx
```

Expected: fail because `AssessmentRail` does not exist.

- [ ] **Step 3: Implement the smallest read-only rail**

Use existing `Role`, `Project`, and `Summary` types. Render project status, selected Function, outcome count, coverage, included/pending/rejected counts, and one role-specific next-action sentence. Do not add state or API calls.

- [ ] **Step 4: Wire the rail into the project view**

In `page.tsx`, keep `FunctionSidebar` and `ProfileEditor` behavior unchanged, but wrap the project content in a `project-layout` with a `reading-column` and the new `AssessmentRail`.

- [ ] **Step 5: Run the focused test and full component tests**

```powershell
npm run test -- AssessmentRail.test.tsx
npm run test
```

Expected: both commands pass.

## Task 3: Make the assessment surface read-first and input-clear

**Files:**
- Modify: `web/src/app/page.tsx`
- Modify: `web/src/components/FunctionSidebar.tsx`
- Modify: `web/src/components/SummaryCards.tsx`
- Modify: `web/src/components/AssessmentCard.tsx`
- Modify: `web/src/components/StakeholderResponsePanel.tsx`
- Modify: `web/src/components/ProfileEditor.tsx`
- Modify: existing related tests under `web/src/components` and `web/src/app`

**Interfaces:**
- Existing callbacks and role checks remain unchanged.
- Existing accessible names for tests and users remain stable unless the new label is clearer.

- [ ] **Step 1: Add failing assertions for the new semantic regions**

Extend tests to assert the project view exposes named `CSF functions`, `Assessment context`, and a named `Outcome assessments` region. Add a Stakeholder assertion that the response panel remains editable only for the existing `org_admin`/`assessor` roles and a Viewer assertion that it remains read-only.

- [ ] **Step 2: Run the focused tests and confirm RED for the new regions**

```powershell
npm run test -- FunctionSidebar.test.tsx ProfileEditor.test.tsx StakeholderResponsePanel.test.tsx page.test.tsx
```

Expected: only the newly added semantic assertions fail before the markup changes.

- [ ] **Step 3: Implement the editorial hierarchy**

Give the project page one clear heading, keep the Function index as the left/orientation rail, keep summary values in a compact status band, and label the assessment list as the central reading region. Keep current and target profile sections distinct, with response/evidence below the related outcome rather than mixed into counselor fields.

- [ ] **Step 4: Run the focused tests and full suite**

```powershell
npm run test -- FunctionSidebar.test.tsx ProfileEditor.test.tsx StakeholderResponsePanel.test.tsx page.test.tsx
npm run test
```

Expected: all tests pass and role behavior is unchanged.

## Task 4: Apply the Clean Editorial composition to organization, project, and auth surfaces

**Files:**
- Modify: `web/src/components/OrganizationDashboard.tsx`
- Modify: `web/src/components/OrganizationWorkspace.tsx`
- Modify: `web/src/components/ProjectDashboard.tsx`
- Modify: `web/src/components/LoginForm.tsx`
- Modify: `web/src/components/AcceptInvitationForm.tsx`
- Modify: `web/src/app/globals.css`

**Interfaces:**
- Keep all existing form labels, callback signatures, delete confirmation, invitation controls, and loading/error states.

- [ ] **Step 1: Add stable region labels where the composition changes**

Use named `header`, `section`, `main`, and `aside` landmarks for organization context, project index, access list, sign-in, and invitation activation. Do not change user-facing product facts or role visibility.

- [ ] **Step 2: Replace card-grid emphasis with editorial lists and workpaper panels**

Use stronger section starts, a readable list measure, quiet borders, numbered index markers, and teal only for active/primary states. Keep destructive actions visually secondary but clear. Keep create/invite forms close to the section they affect.

- [ ] **Step 3: Add the responsive rules**

At desktop, use the Function rail / reading column / context rail. At smaller widths, turn the Function rail into an ordered horizontal index and stack content in reading order. Confirm inputs and actions are full-width where needed and no page introduces horizontal overflow.

- [ ] **Step 4: Run component tests**

```powershell
npm run test
```

Expected: all existing behavior and accessibility tests pass.

## Task 5: Run the bounded design QA and document the built world

**Files:**
- Modify: `web/src/app/layout.tsx`, `web/src/app/globals.css`, and any component files changed above as required by QA.
- Create: `DESIGN.md`

- [ ] **Step 1: Run the impeccable detector once**

```powershell
node .agents/skills/impeccable/scripts/detect.mjs --json --scope layout web/src/app/layout.tsx web/src/app/globals.css web/src/app/page.tsx web/src/components
```

Fix mechanical findings from this run only; do not start a second detector loop.

- [ ] **Step 2: Capture one desktop/mobile screenshot round**

Review login, organization workspace, and project assessment surfaces at desktop and mobile sizes. Check reading order, 60/30/10 balance, focus visibility, status legibility, input affordances, and horizontal overflow. Batch material fixes once and recapture at most one final round.

- [ ] **Step 3: Verify the implementation**

```powershell
npm run test
npx tsc --noEmit --incremental false
npm run build
```

Run the build only after stopping any development server. Confirm the built output retains the design contract seed key.

- [ ] **Step 4: Write DESIGN.md from the implementation**

Document the actual tokens, type scale, layout rails, responsive behavior, role hierarchy, status treatment, and component patterns that shipped. Record any intentional deviation from the approved spec.

- [ ] **Step 5: Commit the implementation**

```powershell
git add web/src docs/superpowers/plans/2026-08-10-clean-editorial-casefile.md DESIGN.md
git commit -m "style: redesign workspace as clean editorial casefile"
git push origin main
```
