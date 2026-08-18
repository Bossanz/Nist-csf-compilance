# Role-first Evidence Workspace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Project Assessment screen as a role-first evidence workspace so Counselor, Stakeholder, Reviewer, and Viewer can immediately see the work relevant to them without changing the existing workflow or data contracts.

**Architecture:** Keep the existing ProjectAssessmentWorkspace composition and handlers as the integration boundary. Derive role-specific queue labels, attention counts, visibility, and action permissions from the existing project/profile/response props, then present them through a shared project context, Function index, outcome queue, and role-specific expanded content. Keep the existing ProfileEditor and StakeholderResponsePanel responsibilities unless a small extraction is required to keep a role body testable and readable.

**Tech Stack:** Next.js 16, React 19, TypeScript, CSS in web/src/app/globals.css, Vitest, and React Testing Library.

## Global Constraints

- Preserve the existing API endpoints, database model, calculations, authentication, role permissions, slug routes, invitation behavior, and handler signatures.
- Do not add a workflow engine, dashboard subsystem, state-management library, database migration, or new persistence field.
- Use optional Project metadata already present in web/src/lib/types.ts when available: objective, assessmentPeriod, targetCompletionDate, scopeBoundary, and complianceDriver. Do not invent values when a field is absent.
- Keep Current Profile and Target Profile separate and explicitly labeled in both editable and read-only presentations.
- Counselor owns scope, rationale, and responsible stakeholder assignment. Counselor must not edit stakeholder Current, Target, response, or evidence content.
- Stakeholder roles keep the existing permission model: assigned editable work can edit Current, Target, response, and evidence through the existing handlers.
- Reviewer can read submitted work and make a final decision only through the existing review handler; reviewer decision controls appear only for submitted responses.
- Viewer receives readable values and evidence actions only; do not show disabled mutation inputs or disabled mutation buttons.
- Evidence Preview stays inline and Download stays available where the existing handler supports it.
- Follow the approved white-first visual system: white reading surfaces, cool neutral canvas/rules, deep ink, teal action/focus, restrained blue Current, restrained sand Target, and text labels for state.
- Preserve accessible names, keyboard expansion behavior, visible focus, adequate contrast, and reduced-motion behavior.
- Use test-first development for every behavior change: write or extend a focused test, run it and observe the failure, implement the smallest change, rerun the focused test, then run the relevant broader suite.
- Do not overwrite or discard the existing uncommitted read-only response-panel changes in web/src/components/StakeholderResponsePanel.tsx, web/src/components/StakeholderResponsePanel.test.tsx, and web/src/app/globals.css; treat their passing tests as the starting baseline.

---

## Task 1: Lock the role matrix and workspace queue contract

**Files:**
- Modify web/src/components/ProjectAssessmentWorkspace.test.tsx
- Modify web/src/components/FunctionSidebar.test.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.tsx
- Modify web/src/components/FunctionSidebar.tsx

- [ ] Add failing tests covering the role contract at the workspace boundary:
  - Counselor sees included and excluded rows for scope configuration and sees unassigned included work as attention.
  - Stakeholder sees only included rows assigned to the current user and sees open work as attention.
  - Reviewer sees included rows and counts submitted responses as to review.
  - Viewer sees included rows, no mutation affordance, and no attention count that implies an action.
  - The active role mode is expressed in accessible text as Scope & Assignment, My Work, Review Queue, or Read-only.
- [ ] Run the focused RED test from web:
  
      npm.cmd test -- src/components/ProjectAssessmentWorkspace.test.tsx src/components/FunctionSidebar.test.tsx

- [ ] Implement the smallest derived role/view model in ProjectAssessmentWorkspace.tsx and FunctionSidebar.tsx. Reuse existing user, profile, response, and permission props; do not add an API call or a second source of truth.
- [ ] Add an explanatory empty queue state in ProjectAssessmentWorkspace.tsx for a role with no visible outcomes, with text that explains whether the user has no assignment, no submitted review work, or no included work.
- [ ] Rerun the focused GREEN test and then the existing workspace component tests.
- [ ] Commit this task as: feat: clarify role-first assessment queues

---

## Task 2: Add the compact shared project context

**Files:**
- Modify web/src/components/ProjectAssessmentWorkspace.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.test.tsx
- Modify web/src/app/globals.css

- [ ] Add failing tests for the first viewport:
  - Organization and project name are visible.
  - Optional objective, assessment period, target completion date, scope boundary, and compliance driver render only when present.
  - Project status, overall coverage/progress, active Function, and included outcome count are visible.
  - The role task hint is specific to the active mode and does not use a generic dashboard label.
- [ ] Run:

      npm.cmd test -- src/components/ProjectAssessmentWorkspace.test.tsx

- [ ] Implement the context block using existing Project and Summary fields. Keep it compact and text-led; do not create a new dashboard card system or duplicate SummaryCards.
- [ ] Add accessible labels and a stable structure for the project metadata so long values wrap without pushing controls off screen.
- [ ] Style the context as a bordered reading header with the approved white-first palette and a modest sticky treatment only if it does not obscure the queue on small screens.
- [ ] Rerun the focused test and the current page/component tests.
- [ ] Commit this task as: feat: add project context to assessment workspace

---

## Task 3: Make the Function index a role-aware work queue

**Files:**
- Modify web/src/components/FunctionSidebar.tsx
- Modify web/src/components/FunctionSidebar.test.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.test.tsx
- Modify web/src/app/globals.css

- [ ] Add failing tests for Function navigation:
  - Every Function shows its included count.
  - Counselor attention is labeled unassigned.
  - Stakeholder attention is labeled open.
  - Reviewer attention is labeled to review.
  - Viewer does not receive a misleading action count.
  - The selected Function remains keyboard-accessible and exposes aria-current.
- [ ] Run:

      npm.cmd test -- src/components/FunctionSidebar.test.tsx src/components/ProjectAssessmentWorkspace.test.tsx

- [ ] Implement the smallest role-aware Function progress contract. Keep counts derived from the profile and responses already loaded by the page.
- [ ] Ensure the queue header mirrors the selected Function and visible outcome count, including a clear distinction between all counselor rows and included stakeholder/reviewer/viewer rows.
- [ ] Keep the Counselor Include every outcome control available only to Counselor and preserve its existing callback.
- [ ] Rerun focused tests and verify no existing bulk-scope behavior regresses.
- [ ] Commit this task as: feat: surface role attention in function index

---

## Task 4: Improve outcome summaries for scanning

**Files:**
- Modify web/src/components/AssessmentCard.tsx
- Modify web/src/components/AssessmentCard.test.tsx
- Modify web/src/components/ProjectAssessmentWorkspace.test.tsx
- Modify web/src/app/globals.css

- [ ] Add failing tests for the collapsed outcome summary:
  - Outcome code, description, included/out-of-scope state, assignment, response status, and evidence count are visible without opening the row when evidence data exists.
  - Current and Target coverage are explicitly labeled; an unlabeled arrow or raw values alone is not sufficient.
  - The expansion button keeps aria-expanded and aria-controls behavior.
  - Read-only roles do not receive disabled form controls or disabled mutation buttons in the collapsed or expanded summary.
- [ ] Run:

      npm.cmd test -- src/components/AssessmentCard.test.tsx src/components/ProjectAssessmentWorkspace.test.tsx

- [ ] Implement the summary as a workpaper row: strong code column, readable description, compact status metadata, explicit Current/Target coverage labels, assignment, and evidence count.
- [ ] Keep color as secondary reinforcement only; status text and accessible labels must carry meaning.
- [ ] Keep evidence count derived from response.documents.length and avoid adding a new response field.
- [ ] Rerun the focused tests and existing response-panel tests.
- [ ] Commit this task as: feat: make assessment rows scannable

---

## Task 5: Separate role-specific expanded content without duplicating workflow logic

**Files:**
- Modify web/src/components/AssessmentCard.tsx
- Modify web/src/components/AssessmentCard.test.tsx
- Modify web/src/components/StakeholderResponsePanel.tsx
- Modify web/src/components/StakeholderResponsePanel.test.tsx
- Modify web/src/components/ProfileEditor.tsx only if an existing prop boundary must be adjusted

- [ ] Add failing tests for expanded content:
  - Counselor can edit only include/rationale/responsible stakeholder and sees profile/response/evidence as read-only reference.
  - Stakeholder can edit Current, Target, response, and evidence for the visible assigned outcome using the existing save, submit, upload, delete, preview, and download callbacks.
  - Reviewer sees read-only Current, Target, response, and evidence plus decision controls only when the response status is submitted.
  - Viewer sees readable Current, Target, response, and evidence with Preview/Download where supported, without disabled textarea, upload, delete, save, or submit controls.
  - Review comments and sent-back/reviewed status remain visible as context after a decision.
- [ ] Run:

      npm.cmd test -- src/components/AssessmentCard.test.tsx src/components/StakeholderResponsePanel.test.tsx

- [ ] Implement the smallest conditional presentation change. Preserve the current passing read-only response-panel behavior and existing callback contracts.
- [ ] Use the existing ProfileReference/read-only value pattern for Counselor, Reviewer, and Viewer. Do not render a disabled copy of stakeholder forms.
- [ ] Keep Reviewer action controls adjacent to, but visually separated from, the evidence reading region.
- [ ] Keep inline evidence preview loading, error, close, and download behavior inside the same outcome.
- [ ] Rerun focused tests, then all component tests that cover profile, response, and invitation-independent assessment behavior.
- [ ] Commit this task as: feat: align assessment detail with role ownership

---

## Task 6: Apply the editorial workpaper layout responsively

**Files:**
- Modify web/src/app/globals.css
- Modify web/src/app/responsive-layout.test.ts
- Modify web/src/components/FunctionSidebar.test.tsx if responsive semantics need a component assertion
- Modify web/src/components/AssessmentCard.test.tsx if responsive action ordering needs an accessible assertion

- [ ] Add failing CSS assertions for the approved layout:
  - Desktop presents Function index and reading queue as two regions.
  - Medium widths make the Function index horizontally usable without forcing the queue into a narrow column.
  - Mobile order is project context, Function index, queue, detail, evidence, and actions.
  - Current and Target stack while retaining their labels.
  - Stakeholder and Reviewer actions become full-width touch targets.
  - Long evidence names wrap or truncate without pushing actions off screen.
  - Nonessential transitions are disabled under prefers-reduced-motion.
- [ ] Run:

      npm.cmd test -- src/app/responsive-layout.test.ts

- [ ] Update CSS using the existing class system and media-query strategy. Prefer border-led grouping and whitespace over new decorative containers.
- [ ] Check that the new context and queue do not introduce horizontal overflow at narrow widths.
- [ ] Rerun the responsive test and the full frontend test suite.
- [ ] Commit this task as: style: refine role-first workpaper layout

---

## Task 7: Run the full verification gate

**Files:**
- No production file changes are expected unless verification exposes a regression.
- If a regression is found, modify only the smallest affected file and add the corresponding regression test before fixing it.

- [ ] Run the full test suite:

      npm.cmd test

- [ ] Run TypeScript without incremental output:

      .\node_modules\.bin\tsc.cmd --noEmit --incremental false

- [ ] Run the production build:

      npm.cmd run build

- [ ] Run the Impeccable detector:

      node ..\.agents\skills\impeccable\scripts\detect.mjs --json web/src

- [ ] Check whitespace and staged/uncommitted diff quality:

      git -c safe.directory=C:/Acuitmesh/NIST-CSF-Compliance diff --check

- [ ] Manually verify the four role modes in the running app using representative data:
  - Counselor: scope and assignment are obvious; stakeholder fields are read-only.
  - Stakeholder: only assigned included outcomes appear; Current, Target, response, and evidence are actionable.
  - Reviewer: submitted work is easy to find; decision is available only for submitted responses.
  - Viewer: all visible content is readable; no disabled mutation controls appear.
- [ ] Confirm routes, auth/invitation flows, API calls, project calculations, and evidence preview/download behavior are unchanged.
- [ ] Summarize the verification output and any remaining uncommitted changes before offering the final commit and push.

## Completion Criteria

The implementation is ready for review when the approved Role-first Evidence Workspace spec is covered, all four role modes have explicit and test-backed behavior, the responsive layout remains readable, all existing tests/typecheck/build/detector checks pass, and no API, database, route, authentication, or permission contract has changed.
