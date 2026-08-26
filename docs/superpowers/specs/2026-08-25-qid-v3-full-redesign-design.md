# QID v3 Full Redesign Design

**Status:** Draft for review
**Date:** 2026-08-25
**Reference:** [Versotis QiD Design System v3](https://www.versotis.com/assets/design/qid/tokens/v3.html)

## Goal

Replace the current incremental Versotis-inspired styling with one coherent QID v3 visual system across every frontend route while preserving the existing NIST CSF workflow, role permissions, API contracts, calculations, content terminology, and long-session readability.

## Scope

This redesign covers the visual system and frontend composition for:

- Authentication: login, forgot password, reset password, account password, and invitation activation.
- Organization surfaces: organization directory, organization workspace, project list, project creation, people, and invitation management.
- Assessment surfaces: project Overview, Function Assignment, outcome cards, Current/Target profile editing, stakeholder response/evidence, Reviewer review, and activity timeline.
- Handoff surfaces: Action Plan, Final Report, and Audit Package.
- Shared states: loading, empty, error, confirmation, read-only, disabled, success, focus, dark/light theme, and responsive reflow.

This redesign does not change:

- Go handlers, PostgreSQL schema, API payloads, session behavior, or authorization decisions.
- Counselor, Organization Admin, Stakeholder/Assessor, Reviewer, Viewer, and Auditor capabilities.
- The stored `submitted` and `reviewed` values; the UI continues to show them as `Reviewing` and `Approved`.
- Coverage calculations, finalization gates, remediation eligibility, evidence preview behavior, or audit traceability.

## Design direction

### Creative thesis

Make the compliance workspace feel like a precise operational instrument: the system should expose where the user is, what needs attention, and what is safe to change without decorative dashboard noise. The redesign refuses the current page-by-page styling drift and replaces it with one token-driven shell and a consistent hierarchy of work surfaces.

### Theme and physical scene

The primary scene is a long-session desktop workspace used by Counselors, Stakeholders, Reviewers, and Auditors. Dark mode is the default because it matches the QID v3 reference and creates a focused control-room surface for dense assessment work. Light mode remains available for bright offices, printing, and users who prefer a paper-like reading surface. Both themes use the same semantic roles, spacing, states, and hierarchy.

### Own-world vocabulary

- Deep violet-black canvas with layered plum surfaces.
- Pink-to-purple action gradient reserved for primary actions and active orientation.
- Fine borders and restrained elevation instead of floating cards everywhere.
- Space Grotesk headings, Inter body copy, and JetBrains Mono for IDs, labels, and compact data.
- Compact 6–32px radius scale with 8px-based rhythm and 4px micro-steps.
- Status is always text plus color; no state may depend on color alone.

## Token contract

The following tokens become the semantic source of truth in `web/src/app/globals.css`. Component CSS must consume semantic variables rather than hard-coded page-specific colors.

### Typography

```css
--font-heading: "Space Grotesk", system-ui, sans-serif;
--font-body: "Inter", "IBM Plex Sans Thai Looped", system-ui, sans-serif;
--font-mono: "JetBrains Mono", ui-monospace, monospace;
```

Heading hierarchy:

- Display: `clamp(40px, 6vw, 72px)`, weight 700, line-height 1.05–1.2.
- Page title: `40px` desktop, fluid down to `30px` mobile, weight 700.
- Section title: `28–32px`, weight 600–700.
- Component title: `16–18px`, weight 600–700.
- Body: `15px`, line-height `1.6`.
- Compact metadata: `11–13px` JetBrains Mono.

### Color primitives

Dark theme:

```css
--bg: #0b0914;
--chrome: #0b0914;
--surface: #13101f;
--surface-2: #1c182d;
--surface-3: #26213c;
--border: #26213c;
--border-2: #342e4f;
--text: #f8f8fc;
--muted: #a19db5;
--faint: #736f8a;
```

Light theme:

```css
--bg: #faf9fd;
--chrome: rgba(250, 249, 253, 0.85);
--surface: #ffffff;
--surface-2: #f2f0f7;
--surface-3: #e6e3f0;
--hover: #ebe8f4;
--border: #dfdce8;
--border-2: #d2cede;
--text: #1a1725;
--muted: #625d75;
--faint: #89859c;
```

Brand and state roles:

```css
--brand-pink-500: #eb147c;
--brand-purple-500: #6a32de;
--brand-magenta-500: #e01476;
--accent: var(--brand-pink-500);
--accent-ink: #f24e9b; /* dark */
--secondary: var(--brand-purple-500);
--ok: #10b981;
--warning: #f59e0b;
--error: #ef4444;
--info: #3b82f6;
```

### Layout, shape, depth, and motion

```css
--content-max: 1200px;
--nav-h: 64px;
--grid-cols: 12;
--grid-gap: 24px;
--radius-xs: 6px;
--radius-sm: 8px;
--radius-md: 12px;
--radius-lg: 16px;
--radius-xl: 24px;
--radius-2xl: 32px;
--radius-pill: 9999px;
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-5: 20px;
--space-6: 24px;
--space-8: 32px;
--space-10: 40px;
--space-12: 48px;
--space-16: 64px;
--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.4);
--shadow: 0 6px 20px rgba(0, 0, 0, 0.5);
--shadow-lg: 0 20px 40px rgba(0, 0, 0, 0.6);
--ease-out: cubic-bezier(0.16, 1, 0.3, 1);
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
--ease-spring: cubic-bezier(0.175, 0.885, 0.32, 1.275);
--duration-fast: 150ms;
--duration-base: 200ms;
--duration-slow: 300ms;
```

Gradients:

- `grad-brand`: pink-to-purple vertical action gradient.
- `grad-cta`: dark purple to deep magenta for primary CTA contrast.
- `grad-neutral`: layered surface gradient only where a surface needs depth.
- No gradient text, no decorative gradient on every card, and no glow unless it communicates focus or active orientation.

## Shared application architecture

The redesign keeps the current component boundaries and introduces a small set of shared visual primitives through existing components and semantic CSS. It does not add a component library or a new state-management layer.

### App bar

All authenticated routes use a 64px top app bar containing:

- A compact CSF/QiD brand mark and current organization/project context.
- The current user identity and role badge.
- Theme toggle.
- A compact account/password or sign-out affordance where the current route supports it.

The app bar is sticky, keyboard reachable, and visually separated by a single border. It must not hide the current page title or replace the route's primary action.

### Page frame

Authenticated pages use a centered frame capped at 1200px with 24px desktop gutters and 16px mobile gutters. The frame supports a two-level composition:

- Contextual navigation rail for the assessment workspace.
- Fluid reading/work column for the current task.

The organization and report pages may use the full 12-column frame without the assessment rail.

### Component rules

- Buttons: 42px minimum height, 8px radius, 10px/16px padding, explicit action text, visible focus ring.
- Inputs/selects/textareas: surface-3 background, border-2 outline, 8px radius, 10px/14px padding, accent focus ring.
- Badges: pill shape only for compact state/role metadata; state text is always visible.
- Panels: surface or surface-2 with one border; use shadow only for auth panels, overlays, and deliberate high-value surfaces.
- Tables: border-led rows, mono metadata, horizontally scrollable on narrow screens, never clipped.
- Icons: authored SVG or existing icon components only; no Unicode symbols used as interface icons.
- Motion: 150ms interaction feedback and one purposeful 300ms reveal for expandable assessment content. Respect reduced-motion preferences.

## Route-by-route design

### Authentication and invitation

Use the QID two-column auth composition on desktop:

- Left: product context, short explanation of the task, role/workspace cue, and restrained brand accent.
- Right: focused form panel with clear title, explicit labels, field-level errors, and one primary action.
- On mobile: context first, form second, full-width actions.
- Password recovery copy remains generic and must not expose account existence.
- Invitation activation must clearly say which organization and role the account will join.

### Organization directory

Use a page header with organization count and one primary “Create organization” action. Organization rows become compact, border-led workspace entries with:

- Organization name and type.
- Project and people counts.
- Last activity/status metadata.
- Open and delete actions separated by visual priority.

The empty state should lead with the create action and explain what an Organization contains.

### Organization workspace

Use a context header for organization identity, then a two-column content grid:

- Primary column: Project list with status, progress, date, and open action.
- Secondary column: People and invitation management.

Project creation becomes a focused QID section with a clear context heading and a two-column form that stacks at tablet width. Required and optional fields must be visibly distinguished without changing the current API.

Invitation management uses a compact table/list with role, project scope, lifecycle status, created/expiry metadata, and contextual Resend/Cancel actions. Destructive actions use text-first danger styling and confirmation only when required.

### Project assessment workspace

Retain the existing three workspace surfaces but express them through the new shell:

- Overview: project context, Overall coverage, Included/Pending/Returned, assignment progress, scope submission, and Final Gate.
- Assignment: contextual Function rail and outcome register only.
- Action Plan: remediation register and action detail only.

The Function rail shows code, percentage, included count, and attention count. It becomes a horizontal scroll index at narrow widths. The main column uses a 72ch reading measure for long copy while allowing evidence, tables, and profile comparisons to expand when needed.

Outcome cards use a quiet summary row with code, title, status, assignment, Current→Target coverage, evidence count, and an authored expand control. Expanded content uses surface-2 with a divider, not a floating card. Counselor scope/rationale, Stakeholder response/evidence, Reviewer decision, and Auditor read-only states remain visually distinct.

### Action Plan

Use a section header with eligible-gap count and one create action. Each action row shows outcome, title, priority, owner, due date, progress, and lifecycle status. Expanded details group:

- Remediation intent.
- Assignee progress.
- Evidence.
- Counselor review.

Progress and review controls must remain role-aware. Finalized assessment data remains read-only while remediation remains editable according to existing permissions.

### Final Report and Audit Package

Reports use the same QID tokens but a calmer reading composition:

- Report header and project context.
- Summary metrics.
- Per-Function coverage.
- Outcome register with Current/Target and review state.
- Response/evidence and remediation sections.
- Audit trail in chronological order.

The Final Report remains print-friendly and switches to light print tokens. The Audit Package remains dense, traceable, read-only, and horizontally safe for long IDs and evidence names.

## Responsive behavior

- At `1100px`: assessment rail becomes a horizontally scrollable navigation index; multi-column forms begin stacking.
- At `760px`: all creator forms, profile comparisons, report grids, tables, and auth columns become one column; actions stretch; evidence names wrap; coverage routes move below outcome titles.
- At all widths: no horizontal page overflow; long organization/project/outcome/evidence text uses `min-width: 0`, wrapping or ellipsis with a full accessible label.
- DOM order and keyboard order must remain aligned with the visual reading order.

## Accessibility and state requirements

- Every action has a visible label or accessible name.
- All focusable controls use a QID accent focus ring with sufficient contrast.
- Status, role, permission, and response state use text plus color.
- Loading, empty, error, disabled, read-only, Reviewing, Approved, Needs more information, and Finalized states have explicit copy.
- Reduced-motion mode disables non-essential transitions.
- Form errors are associated with fields and announced through an alert/status region.

## Implementation boundaries

The implementation should primarily modify:

- `web/src/app/globals.css` for token layers, theme values, shared layout, and responsive rules.
- `web/src/app/layout.tsx` for the root design contract and authenticated shell hooks.
- `web/src/app/**/page.tsx` for route-level composition only where the current structure cannot express the new layout.
- Existing components under `web/src/components` for local anatomy and role-aware presentation.
- Existing frontend tests plus new visual/semantic assertions for route states and responsive-safe structures.
- `DESIGN.md` and `.impeccable/design.json` after the built world is verified.

Do not modify API or database code for this redesign. Do not add a CSS framework, icon package, or state library unless an existing requirement cannot be met with the current stack.

## Acceptance criteria

1. Every listed frontend route uses the same QID v3 semantic tokens and typography roles.
2. Dark theme is the default authenticated presentation, with a working light alternative.
3. Login, organization, project, assessment, action, report, invitation, and password surfaces are visually coherent without relying on route-specific legacy colors.
4. Existing role visibility and mutation behavior remain unchanged.
5. Existing workflow labels and business terminology remain unchanged, including `Reviewing` and `Approved` in the UI.
6. No page-level horizontal overflow exists at desktop, tablet, or mobile widths.
7. Long outcome titles, project names, organization names, evidence names, and audit IDs remain readable or expose a full accessible label when visually condensed.
8. Keyboard focus, reduced motion, form errors, loading, empty, and read-only states are present and legible.
9. Full frontend tests, typecheck, production build, and the Impeccable detector pass.
10. `DESIGN.md` describes the shipped QID v3 world rather than the pre-redesign system.

## Verification plan

The implementation will use short TDD cycles per route group:

1. Add semantic and responsive regression assertions for the route group.
2. Run the focused test and confirm the new assertion fails for the old presentation.
3. Implement the smallest token/layout/component change.
4. Run the focused test, then the full frontend suite and typecheck.
5. Capture desktop and mobile screenshots for Login, Organization workspace, Project Overview, Assignment, Action Plan, Final Report, and Audit Package.
6. Run the Impeccable detector once over changed UI targets, fix mechanical findings, and run the production build.
7. Review the final visual result against this spec and update `DESIGN.md` from the built system.

## Open implementation decisions

- Keep the current `FunctionSidebar` as the contextual rail and adapt it to the QID app shell, or extract only a small `AppBar` wrapper around it. The implementation should choose the smaller change that preserves the current route behavior.
- Whether the brand mark is text-only or uses an existing local asset. No new logo asset should be invented without a supplied source asset.
- Whether the user preference should override the dark-first default on first visit or only after an explicit toggle. The recommended behavior is dark-first until a saved preference exists, then persist the user's choice.
