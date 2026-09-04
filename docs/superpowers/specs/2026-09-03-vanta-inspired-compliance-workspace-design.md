# Vanta-inspired Compliance Workspace Design

**Status:** Approved direction for implementation

## Goal

Make the existing V3 application feel like a focused compliance operating workspace: a counselor should see portfolio posture and the next review action immediately, while a stakeholder should see a short, assigned work queue instead of a decorative dashboard.

## Direction

Use an evidence-workbench metaphor. The interface behaves like a shared review desk with three persistent questions:

1. What is the current posture?
2. What needs attention next?
3. Who owns the next action?

The reference is Vanta's compliance-software operating pattern—status-first work queues, owners, evidence and review state—not its brand, assets, or proprietary UI. Existing QiD/Versotis tokens, 60/30/10 color balance, and the V3 workflow remain the product's visual and functional authority.

## Scope

### Organization portfolio (/organizations)

- Replace the card-first organization list with a portfolio header, compact posture metrics, and a readable client workspace register.
- Keep organization create, delete, counselor access management, password link, sign-out, and error/retry behavior unchanged.
- Make the primary action for each row opening the client workspace; keep destructive deletion secondary and protected by the existing confirmation dialog.
- Surface the number of client workspaces, active workspaces, finalized workspaces, and total people from data already available in the component. These are derived UI summaries, not new API fields.

### Organization workspace (/organizations/[organizationSlug])

- Present organization identity and project portfolio as the primary surface.
- Turn project cards into a work register showing project status, version, created date, and open action.
- Keep project creation, deletion, stakeholder invitations, invitation lifecycle controls, and access management behavior unchanged.
- Give project creation a clearly separated “new workspace” panel so it does not compete with active work.

### Project workspace (/organizations/[organizationSlug]/projects/[projectSlug])

- Keep the existing left rail, but make the surface hierarchy explicit: Overview, Assignment, Action Plan, and Log.
- Keep Overview as the initial local surface and ensure the rail highlights only the active surface/function.
- Recompose Project overview around:
  - project identity and status;
  - overall coverage and included/reviewing/returned counts;
  - a role-aware “Next up” work queue;
  - a Function register with coverage percentage, included count, and attention count;
  - existing Version History, Remediation status, Assignment progress, Stakeholder overview, and Finalization gate in their current role/state conditions.
- Keep Assignment as the detailed outcome surface. Its existing scope, profile, response, evidence, preview, reviewer and read-only behavior must not change.
- Use labels Reviewing and Approved in rendered UI while preserving existing API status values.

## Role behavior

- Counselor/Counselor Admin: portfolio posture, unassigned scope items, review/finalization readiness, and action-plan state.
- Assessor/Org Admin: assigned work, draft/returned responses, evidence completion, and a direct route to Assignment.
- Reviewer: responses awaiting review, returned items, and approved count.
- Viewer/Auditor: read-only project posture and evidence/review state; no mutation controls are introduced.

The redesign may change where an existing action is surfaced, but it must not grant a role a control that the current component does not grant.

## Visual system

- Mode: Operate. The screen is a long-session work surface, not a marketing page.
- Material: quiet paper/workbench surfaces, thin rules, compact registers, and one clear accent rail for focus.
- Color strategy: restrained. White/light neutral or graphite canvas dominates; secondary surfaces create structure; magenta/purple identifies the current location and primary action. Current/Target and semantic states retain their text labels and accessible contrast.
- Typography: preserve Space Grotesk headings, Inter body copy, and JetBrains Mono for codes, labels, and measurements.
- Density: prefer one wide register over equal-sized card grids. Keep copy at a readable measure and align metadata to the top edge.
- Motion: use existing restrained transitions only; no decorative gradients, chart animations, or new loading choreography.

## Accessibility and responsive behavior

- Every register remains keyboard reachable with visible focus.
- Status, ownership, and attention are conveyed by text in addition to color.
- Organization/project registers collapse to stacked rows below the mobile breakpoint without hiding actions.
- Long project names, emails, and outcome descriptions wrap without horizontal page overflow.
- Empty, loading, error, disabled, and finalized states remain explicit.

## Non-goals

- No backend/API/schema changes.
- No automated evidence integrations, AI, scheduling, email behavior, or new roles.
- No new project status values or changes to the finalization gate.
- No replacement of QiD/Versotis tokens with Vanta brand colors.

## Acceptance criteria

- Existing functional tests continue to pass, with new assertions for the portfolio register, project register, role-aware next-up content, and active navigation state.
- Organization and project pages render correctly in light and dark themes at desktop and mobile widths.
- The first viewport makes the current status, the next action, and the owner/attention context understandable without opening a card.
- Existing create/delete/invite/open/report/audit actions remain discoverable and preserve their current callbacks.
- Typecheck, frontend tests, production build, and the Impeccable detector pass for changed UI files.

