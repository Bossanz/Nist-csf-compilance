# Role-first Evidence Workspace Design

## Status

Approved direction: Role-first Evidence Workspace.

This document defines the first redesign pass for the Project Assessment surface. It changes the visual structure and role-specific presentation while preserving the current workflow, API contracts, database model, routes, and terminology.

## Goal

Make the next action obvious for every role in a compliance project:

- Counselor decides scope and assignment.
- Stakeholder completes assigned inputs and evidence.
- Reviewer reads submitted work and closes an outcome.
- Viewer reads project information without mutation controls.

The redesign should reduce the cognitive load of a shared assessment card that currently contains several different jobs.

## Product truth and constraints

- The product is a shared NIST CSF 2.0 workspace for long compliance-reading sessions.
- Counselor-owned decisions are scope, rationale, and stakeholder assignment.
- Stakeholder-owned work is Current and Target input, response text, evidence upload, save, and submit.
- Reviewer-owned work is the final decision on submitted responses.
- Viewer access is read-only.
- Current Profile and Target Profile remain distinct and visibly labeled.
- Existing email/password authentication, role permissions, API behavior, calculations, and slug routes remain unchanged.
- The implementation must stay small enough for v1 and must not introduce a new dashboard subsystem, workflow engine, or database migration.

## Audience and operating mode

This is an authenticated Operate surface with a Read-oriented assessment body.

- Counselor arrives to understand project coverage and resolve unassigned or incomplete work.
- Stakeholder arrives to complete a known set of assigned outcomes, usually during a long form-filling session.
- Reviewer arrives to process submitted outcomes, inspect evidence, and make a final decision.
- Viewer arrives to understand project progress and supporting information without changing it.

## Selected direction

The visual world is a calm, evidence-led workpaper rather than a generic SaaS dashboard.

The Project Assessment page uses a strong grid, a persistent Function index, a focused outcome queue, and role-specific content inside the same project context. The interface should feel like one shared case file with three clearly marked work modes: Scope & Assignment, My Work, and Review Queue.

The first viewport proves three things immediately:

1. Which organization and project is open.
2. What work is relevant to the current role.
3. Which outcome should be opened or acted on next.

## Surface topology

### Shared project context

The top of the assessment route contains:

- Organization and Project name.
- Objective, assessment period, target completion date, and compliance driver when available.
- Project status and overall progress.
- A short role-specific task hint.
- The active Function and included outcome count.

The context stays compact while scrolling so the reader does not lose location, but it must not become a large dashboard hero.

### Counselor mode: Scope and Assignment

The Function navigation remains the primary orientation rail. The main queue emphasizes:

- Included or excluded state.
- Counselor rationale.
- Responsible stakeholder.
- Current-to-target coverage summary.
- Response status and evidence presence as supporting review signals.

Opening an outcome shows scope controls and the read-only profile/response reference. Counselor can save scope and assignment but cannot edit Stakeholder Current, Target, response, or evidence content.

### Stakeholder mode: My Work

The queue is filtered to included outcomes assigned to the current stakeholder. Each row makes completion state, response status, and evidence count visible.

Opening an outcome presents:

- Current Profile and Target Profile input sections.
- Response input.
- Evidence upload and inline preview.
- Save and Submit actions.
- Review comment/status as read-only context when a response has already been reviewed or sent back.

Stakeholder controls are direct inputs, not hidden behind disabled copies of Counselor controls.

### Reviewer mode: Review Queue

The queue emphasizes submitted outcomes and items that need attention. Opening an outcome presents:

- Read-only Current Profile and Target Profile.
- Read-only stakeholder response content.
- Evidence list with Preview and Download actions.
- Reviewer decision panel only while a response is submitted.
- A decision comment that remains visible after the outcome is reviewed or sent back.

The reviewer action is visually separated from the evidence reading area, but remains in the same outcome context.

### Viewer mode

Viewer receives the same reading structure as Counselor/Reviewer without mutation controls:

- Read-only Profile Reference.
- Read-only response and evidence.
- Preview and Download for supported evidence.
- No disabled textarea, upload input, delete button, Save button, or Submit button.

## Outcome anatomy

Every outcome row has a consistent summary:

- Outcome code and description.
- Included or Out of scope label.
- Current and Target coverage with explicit labels.
- Assignment label.
- Response status.
- Evidence count when available.
- Expand/collapse affordance with keyboard and screen-reader state.

Expanded content follows the current role. A read-only role receives text and labeled values, not controls with disabled styling.

## Interaction and data flow

1. Project page loads project context, Functions, profile rows, responses, and organization users using the existing page/API flow.
2. The active role determines queue filtering and available actions in the workspace.
3. Opening an outcome does not change data; it only reveals its role-specific body.
4. Counselor saves scope/rationale/assignment through the existing profile patch.
5. Stakeholder saves or submits response data through the existing response handlers and uploads evidence through the existing evidence handler.
6. Reviewer saves a final decision through the existing review handler.
7. Evidence Preview remains inline and is closed without leaving the outcome.
8. After a successful mutation, the existing save state and response status are updated; failed mutations keep the user in context and show an inline error with the existing retry path.

No new API endpoint or persistence field is required for this redesign.

## State and error behavior

- Loading state keeps the current page-level loading pattern.
- Empty queues explain why no outcomes are visible for the current role.
- Read-only surfaces use a visible Read only label and a quiet surface tone.
- Saving shows Saving, then Saved; dirty fields show Unsaved changes.
- Mutation errors remain adjacent to the action area and do not replace the assessment content.
- Reviewer decision controls appear only for submitted responses.
- Evidence preview loading and errors remain inside the evidence region.
- Status meaning is always conveyed with text, not color alone.
- At narrow widths, actions stack in reading order and remain reachable without horizontal scrolling.

## Visual system

The redesign keeps the approved white-first palette while changing the composition:

- White work surfaces dominate the reading area.
- Cool neutral canvas and rules provide structure.
- Deep ink carries long-form content and labels.
- Teal marks active navigation, primary actions, focus, and meaningful progress.
- Current uses the existing restrained blue surface; Target uses the existing restrained sand surface.
- Warning and error colors are reserved for actionable states.
- System sans typography, thin rules, modest 8px/12px corners, and border-led elevation remain.
- No gradients, glass effects, decorative illustrations, dense KPI tiles, or color-only status signals.
- Motion is limited to state feedback and outcome expansion; reduced motion must disable nonessential transitions.

## Responsive behavior

- Desktop keeps the Function index and main queue in a two-region layout.
- At medium widths, the Function index becomes a horizontal scrollable row and the queue remains the primary reading column.
- At small widths, all regions stack in this order: project context, Function index, queue, outcome detail, evidence, actions.
- Current and Target sections stack without losing their labels.
- Review controls and stakeholder actions become full-width touch targets.
- Long evidence names wrap or truncate without pushing actions off screen.

## Component boundary

The first implementation should reuse the existing component surface and avoid a broad refactor.

Expected files to modify:

- web/src/components/ProjectAssessmentWorkspace.tsx for role-specific queue hints, filtering, and shell composition.
- web/src/components/FunctionSidebar.tsx for role-aware Function progress and attention labels.
- web/src/components/AssessmentCard.tsx for the shared outcome summary and role-specific expanded body.
- web/src/components/StakeholderResponsePanel.tsx for read-only, stakeholder, and reviewer response states.
- web/src/app/globals.css for the workpaper grid, queue, read-only surfaces, and responsive rules.
- The matching component tests for each behavior change.

New components should be extracted only when a role-specific body becomes independently testable and the existing component cannot remain readable. No new state-management layer is required.

## Acceptance criteria

1. A Counselor can identify unassigned outcomes and scope/assignment actions from the queue without opening every outcome.
2. A Stakeholder sees only relevant assigned work and can complete Current, Target, response, and evidence inputs.
3. A Reviewer can identify submitted work, read the response and evidence, and make a final decision in the same outcome.
4. A Viewer can read profile, response, and evidence content without seeing disabled mutation forms.
5. Current and Target remain visually and textually distinct at every viewport.
6. Evidence Preview remains inline and Download remains available where supported.
7. Existing API behavior, role permissions, routes, project calculations, and invitation/auth flows remain unchanged.
8. No primary action is hidden behind a disabled control or communicated by color alone.
9. Existing frontend tests, TypeScript checks, production build, and Impeccable detector pass.

## Out of scope

- New API endpoints, database migrations, or workflow state machines.
- Report export, notifications, activity timeline, or a separate reviewer dashboard.
- Changing the NIST data model or assessment calculations.
- Replacing authentication or invitation behavior.
- A full visual redesign of login, organization, or project-index surfaces in this first pass.
