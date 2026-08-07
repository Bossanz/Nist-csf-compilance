# Design Spec: Clean Editorial Casefile

Date: 2026-08-07  
Status: Approved direction; awaiting spec review before implementation  
Scope: Existing Next.js frontend only

## Goal

Make the compliance workspace easier to read during long sessions while making the role split explicit:

- Counselor reads, interprets, reviews, and manages assessment decisions.
- Stakeholder fills in responses and supporting evidence.

The redesign changes the visual system and information hierarchy only. Existing API contracts, database behavior, calculations, routes, role permissions, and workflow must remain intact.

## Direction

### Clean Editorial Casefile

The product should feel like a calm, well-structured assessment workpaper: a white reading surface, a light neutral filing structure, and restrained teal marks for orientation and action. The visual language is editorial rather than dashboard-heavy.

The interface should communicate this mechanism:

1. A project is the case file.
2. Functions and outcomes are the reading index.
3. Counselor-owned current/target decisions are the interpretation layer.
4. Stakeholder responses and evidence are the input layer.
5. Review status makes the handoff visible.

The counselor’s reading context is primary on review and overview screens. The stakeholder’s next input is primary on response screens.

## Root layout contract

The first child of `<body>` in `web/src/app/layout.tsx` will be a short HTML comment containing this product contract. It is an implementation guardrail, not visible UI:

```html
<!--
THESIS: Readability is the product surface: help counselors scan, interpret, and decide.
OWN-WORLD: Clean Editorial Casefile — white paper, soft neutral structure, restrained teal marks.
STORY: Counselor reads and reviews; Stakeholder fills and supports with evidence.
FIRST VIEWPORT: Project context, Function index, current/target status, and one clear next action.
FORM: Clean Editorial Casefile; grounded direction 4/7; source key 313a6a7a. FINISH: document the built world.
-->
```

The comment must stay under 150 words and must not become a second requirements document.

## Layout model

### Desktop

Use a three-part composition inside the existing application shell:

- Left: a compact Function index. It anchors the reader to the CSF structure and exposes progress/status without becoming a second dashboard.
- Center: a constrained reading column, approximately 700–780px wide. Long labels, descriptions, assessment guidance, and review notes should have a comfortable measure.
- Right: a contextual status/input rail. For Counselor it shows current/target status, review state, and relevant summary. For Stakeholder it shows the response action, evidence action, and submission state.

The project name and organization context remain visible above the composition. Page titles should identify the current task in plain language. The main column should have one dominant heading and one obvious next action.

### Mobile

Reflow into one reading order:

1. Project and organization context.
2. Page title and status summary.
3. Function index as a compact horizontal scroller or collapsible index.
4. Main reading/input content.
5. Supporting status and actions.

Do not preserve desktop columns by shrinking text. Avoid horizontal overflow and keep actions reachable after the related content.

## Role-aware hierarchy

### Counselor

Counselor screens should foreground reading and interpretation:

- Show the relevant Function/outcome context before details.
- Keep current and target values visibly distinct and easy to compare.
- Surface response/review status without requiring a deep navigation step.
- Keep editing controls for counselor-owned fields grouped and quiet.
- Make stakeholder response and evidence read-only where the existing permission model requires it.

The counselor view should not look like a form unless the counselor is editing a counselor-owned profile field.

### Stakeholder

Stakeholder screens should foreground completion:

- Show only the outcomes assigned/selected for the project.
- Put the response field, evidence upload, and save/submit action near the relevant guidance.
- Present counselor-owned priority, coverage, current profile, and target profile as read-only context.
- Keep review notes/status visible but secondary to the requested input.
- Use plain copy for what is required, saved, submitted, or returned for changes.

This is a visual hierarchy decision; it must not weaken server-side authorization or add access to unselected subcategories.

## Visual system: 60/30/10

The ratio is a visual budgeting rule, not a requirement to count pixels exactly:

- 60% white: page background, reading surfaces, cards, form fields, and open space.
- 30% light neutral: borders, dividers, muted panels, navigation grouping, and inactive states.
- 10% teal: active Function marker, focus ring, primary action, progress/status emphasis, and links.

Use color with text, labels, icons, or borders so status is never conveyed by color alone. Avoid adding new gradients, saturated success/error fills, or large decorative shapes.

## Typography and spacing

- Keep the existing dependency footprint; use the existing/system font stack unless the repository already includes a suitable font.
- Use a clear display size for page titles, readable body text, and a smaller but still accessible metadata scale.
- Favor line-height and measure over dense card grids.
- Use tabular numbers for progress and comparison values where supported.
- Keep focus states visible against white and neutral surfaces.
- Prefer a consistent spacing rhythm based on the existing CSS variables; add tokens only where they clarify the new system.

## Component scope

Expected implementation touch points:

- `web/src/app/layout.tsx`: root contract comment and document-level shell semantics.
- `web/src/app/globals.css`: white-first tokens, editorial spacing, surfaces, borders, focus states, responsive rules.
- Existing shell/navigation components: Function index and project context hierarchy.
- Existing dashboard, workspace, assessment, profile, and stakeholder response components: role-aware ordering and summary/input rails.
- Existing tests: update accessible names/regions only when the DOM hierarchy changes; add coverage for counselor/stakeholder visibility where needed.

Do not introduce a new UI framework, state library, icon package, or backend endpoint for this redesign.

## Accessibility, behavior, and performance

- Preserve semantic headings and landmarks.
- Keep one logical heading hierarchy per view.
- Use labelled regions for the Function index, reading content, status, and input areas.
- Preserve keyboard navigation, visible focus, form labels, error announcements, and disabled/loading states.
- Keep expandable detail sections keyboard-operable and associated with their controls.
- Avoid motion that is required to understand state; keep transitions short and respect reduced motion.
- Keep the DOM and CSS simple enough for fast initial render on the existing stack.
- Preserve loading, empty, error, and API-failure states in the new visual system.

## Non-goals

- No changes to authentication, invitation, deletion, API, database schema, or role authorization.
- No change to the definition of Counselor, Stakeholder, Reviewer, or Viewer.
- No change to CSF calculations, priority/coverage rules, selected subcategories, or review workflow.
- No new image assets are required for this data-heavy product.

## Acceptance criteria

1. A Counselor can scan project context, Function navigation, current/target status, and review state without opening several nested cards.
2. A Stakeholder can identify the assigned input, evidence action, and submission state without seeing unselected outcomes or counselor-only editing controls.
3. Desktop uses a readable central column with clear supporting rails; mobile becomes a single readable flow with no horizontal overflow.
4. White surfaces dominate, light neutrals structure the page, and teal is reserved for orientation/action/status in an approximate 60/30/10 balance.
5. Focus, labels, headings, status text, validation, loading, empty, and error states remain understandable without relying on color.
6. Existing frontend tests, TypeScript checking, and production build pass.
7. The final visual world is documented in `DESIGN.md` from the implemented tokens, components, and layout rather than from the proposal alone.

## Verification plan

- Run the existing frontend test suite.
- Run `npx tsc --noEmit --incremental false`.
- Run the production build after stopping any development server.
- Run the impeccable layout detector once over changed UI targets and fix mechanical findings.
- Review one batched desktop/mobile screenshot round for the login, Counselor, and Stakeholder surfaces; fix material issues and review once more if needed.
- Write `DESIGN.md` from the final implementation and record any intentional deviations from this spec.
