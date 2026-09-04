---
name: CSF Compliance
description: A QID v3 dark-first workspace for reading, completing, reviewing, and handing off NIST CSF assessments.
colors:
  pink: "#eb147c"
  purple: "#6a32de"
  deep: "#160a2e"
  dark-bg: "#0b0914"
  dark-surface: "#13101f"
  dark-surface-2: "#1c182d"
  dark-surface-3: "#26213c"
  dark-border: "#342e4f"
  dark-text: "#f8f8fc"
  dark-muted: "#a19db5"
  light-bg: "#faf9fd"
  light-surface: "#ffffff"
  light-surface-2: "#f2f0f7"
  light-border: "#dfdce8"
  light-text: "#1a1725"
  light-muted: "#625d75"
  success: "#10b981"
  warning: "#f59e0b"
  error: "#ef4444"
  info: "#3b82f6"
  current: "#3b82f6"
  target: "#f59e0b"
  current-text: "#a9c7ff"
  target-text: "#ffd58a"
  approved: "#6ee7b7"
  link: "#d3b8ff"
typography:
  display:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "clamp(2.2rem, 4vw, 4rem)"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.04em"
  headline:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "clamp(1.5rem, 2.3vw, 2.25rem)"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.03em"
  title:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "1.04rem"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.02em"
  body:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.55
  label:
    fontFamily: "JetBrains Mono, ui-monospace, monospace"
    fontSize: "0.78rem"
    fontWeight: 750
    lineHeight: 1.3
rounded:
  xs: "6px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  pill: "999px"
  circle: "50%"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "48px"
components:
  button-primary:
    backgroundColor: "linear-gradient(135deg, {colors.pink}, {colors.purple})"
    textColor: "#ffffff"
    rounded: "{rounded.sm}"
    padding: "10px 16px"
    height: "44px"
  button-secondary:
    backgroundColor: "{colors.dark-surface-2}"
    textColor: "{colors.dark-text}"
    rounded: "{rounded.md}"
    padding: "10px 16px"
    height: "44px"
  button-danger:
    backgroundColor: "transparent"
    textColor: "{colors.error}"
    rounded: "{rounded.xs}"
    padding: "10px 8px"
    height: "44px"
  input:
    backgroundColor: "{colors.dark-surface-2}"
    textColor: "{colors.dark-text}"
    rounded: "{rounded.md}"
    padding: "11px 12px"
  status-chip:
    backgroundColor: "{colors.dark-surface-3}"
    textColor: "{colors.dark-text}"
    rounded: "{rounded.pill}"
    padding: "5px 9px"
  nav-item:
    backgroundColor: "transparent"
    textColor: "{colors.dark-muted}"
    rounded: "{rounded.md}"
    padding: "10px 12px"
    height: "44px"
  assessment-card:
    backgroundColor: "{colors.dark-surface}"
    textColor: "{colors.dark-text}"
    rounded: "{rounded.md}"
    padding: "0"
  response-panel:
    backgroundColor: "{colors.dark-surface}"
    textColor: "{colors.dark-text}"
    rounded: "{rounded.md}"
    padding: "22px"
---

# Design System: CSF Compliance — QID v3 Evidence Workbench

## Overview

**Creative North Star: "Evidence Workbench"**

The interface turns a complex assessment into a clear, accountable workspace. Its product pattern is similar to mature compliance tools such as Vanta—status-first registers, evidence ownership, and review queues—without copying another product's brand or assets. Dark graphite surfaces are the default reading environment; a light QID surface set remains available for bright environments. Magenta-to-purple brand cues mark the active Function, the next action, and the few statuses that need attention. The result follows the [Versotis QID v3 token reference](https://www.versotis.com/assets/design/qid/tokens/v3.html) without sacrificing long compliance-reading sessions.

The system keeps Counselor interpretation, Stakeholder response/evidence input, and Reviewer feedback visually legible as different kinds of work. Organization and project pages use status-first registers and compact forms; assessment pages use an editorial reading column with expandable outcome cards. Project Overview answers three questions before the reader opens an outcome: what is the posture, what needs attention next, and who owns the next action. The visual language uses the brand gradient only for actions and orientation, avoiding decorative chrome and heavy shadows that compete with compliance text. Existing component class contracts remain in place so the redesign does not alter API, role, or workflow behavior.

The public account-recovery screens (`/forgot-password` and `/reset-password`) and authenticated password screen (`/account/password`) reuse the same two-column auth layout, explicit field labels, readable error states, and visible return links. Recovery success copy remains generic so the interface does not reveal whether an email belongs to an account.

**Key Characteristics:**

- Dark-first work surfaces: `#0b0914` canvas, `#13101f` primary surface, and `#1c182d` supporting surface; light mode uses `#faf9fd` and white reading surfaces.
- Magenta-to-purple used as orientation and action, not as decoration.
- Space Grotesk for headings, Inter for long-form body copy, and JetBrains Mono for wayfinding/data labels.
- A 12-column desktop grid with a 1200px content frame, a 260px workspace rail, and explicit Current / Target comparison.
- Thin borders, 6–16px corners, and restrained QID elevation for raised panels.
- Role-aware surfaces that distinguish reading, reviewing, and completing input.
- Status-first client and project registers that expose ownership, version, status, and the next route into work.
- An Overview workbench with posture metrics, a role-aware `Next up` block, and progress by Function.
- Project context metadata aligned from a shared top edge so uneven copy lengths do not make short values appear vertically suspended.
- Contrast-safe muted text and semantic state tokens remain consistent across dark and light themes.
- Dense report tables stay keyboard-scrollable on small screens, pin the first column, and expose an explicit scroll hint.

## Colors

The palette follows the 60/30/10 rule: neutral workspace surfaces take most of the page, the secondary surface/border system provides structure, and magenta-purple is reserved for orientation, action, focus, and meaningful status.

### Primary

- **QID Pink** (`{colors.pink}`) and **QID Purple** (`{colors.purple}`): Primary actions, active Function navigation, focus borders, links, and assessment orientation cues.
- **Brand Gradient** (`linear-gradient(135deg, {colors.pink}, {colors.purple})`): Primary buttons and high-value action anchors only.
- **Accent washes:** Color-mixed pink/purple surfaces support selected navigation, role/status chips, progress context, and low-intensity regions.

### Neutral

- **Dark canvas/surfaces** (`{colors.dark-bg}`, `{colors.dark-surface}`, `{colors.dark-surface-2}`, `{colors.dark-surface-3}`): The default reading environment.
- **Light canvas/surfaces** (`{colors.light-bg}`, `{colors.light-surface}`, `{colors.light-surface-2}`): The alternate reading environment.
- **Text hierarchy** (`{colors.dark-text}`, `{colors.dark-muted}`, `{colors.light-text}`, `{colors.light-muted}`): Headings, metadata, labels, and long-form copy.
- **Borders** (`{colors.dark-border}`, `{colors.light-border}`): Dividers, card boundaries, input strokes, and structural separation.
- **Current Blue** (`{colors.current}` / `{colors.current-text}`) and **Target Amber** (`{colors.target}` / `{colors.target-text}`): Paired comparison surfaces that always include a text label.
- **Success**, **Warning**, **Error**, and **Info** states are always paired with text and never used as the only signal.
- **Approved** (`{colors.approved}`) and **Link** (`{colors.link}`) are semantic roles; their text and border treatments are theme-aware rather than hard-coded per component.

### Named Rules

**The Brand Cue Rule.** Magenta and purple should explain where the reader is or what action is available; the brand gradient belongs to actions and orientation, not whole reading sections.

**The Labeled State Rule.** Current, target, Reviewing, Approved, error, and permission states use color alongside a visible text label.

### Dark Theme

Dark mode is the default root theme and is explicitly controlled with `data-theme="dark"`. The palette uses QID graphite surfaces and soft-white text; light mode swaps the same semantic roles to the light token set. Theme selection persists through the existing `csf-theme` local preference and is available from both the authenticated rail and auth surfaces.

## Typography

**Display Font:** Space Grotesk (`Space Grotesk`, system-ui fallback)
**Body Font:** Inter (`Inter`, system-ui fallback)
**Label/Mono Font:** JetBrains Mono for outcome codes, context markers, and compact data labels.

**Character:** Strategic, technical, and crisp without becoming a marketing landing page; headings carry the brand while body copy stays comfortable for a long reading session.

### Hierarchy

- **Display** (700, `clamp(2.2rem, 4vw, 4rem)`, 1.18, `-0.04em`): Organization, project, and authentication headings.
- **Headline** (700, `clamp(1.5rem, 2.3vw, 2.25rem)`, 1.18, `-0.03em`): Section headings and major workspace titles.
- **Title** (700, `1.04rem`, 1.18, `-0.02em`): Outcome, card, and form-group titles.
- **Body** (400, `14px`, 1.55): Long-form descriptions, assessment copy, and field content; paragraph measure stays near 68ch.
- **Label** (650, `0.71rem`, 1.3): Field labels and supporting metadata. Eyebrows and context lines use compact mono uppercase tracking for orientation.

### Named Rules

**The Reading Measure Rule.** Let hierarchy and a readable line length do the organizing; do not compensate for long compliance text with oversized decorative type.

## Layout

The application shell uses a 260px sticky Function navigation rail and a fluid main area. Main content is capped at 1200px with a responsive horizontal gutter and a 12-column desktop grid. Desktop organization and project indexes use status-first work registers rather than equal-sized card grids, while the project assessment keeps a full-width workspace with prose capped by the 68ch reading measure. Orientation is carried by the project header, Function navigation, Overview workbench, assignment-progress banner, and outcome count.

The project context panel uses a five-column metadata grid on wide screens, two columns at the rail breakpoint, and one column on small screens. Each metadata cell uses top-aligned content, a compact 3px label/value gap, and a thin rule above the row so long Scope or Compliance driver copy does not vertically center shorter dates and periods. Long account, invitation, and evidence names wrap within their rows at the single-column breakpoint so metadata cannot force horizontal overflow.

At 1100px the Function rail becomes a horizontal, scrollable index and multi-column creator forms begin to stack. At 760px lists, profile comparisons, supporting grids, and authentication surfaces become one column; actions stretch to the available width; assessment summaries move the coverage route below the outcome title; and response/evidence controls follow the content they act on. Report tables remain intact as workpaper tables inside focusable horizontal-scroll regions; the first column stays pinned on small screens and an explicit hint tells keyboard and touch users how to view the remaining columns. The spacing rhythm follows 4px, 8px, 16px, 24px, and 48px QID steps.

## Elevation & Depth

The system is border-led and flat by default. Dark or light surfaces sit on their matching canvas through tonal contrast and thin rules. QID base and large shadows are reserved for forms, deletion confirmation, authentication panels, and branded primary actions that need quiet separation. Focus uses a magenta ring as an interaction state, not as ambient decoration. Expanded assessment content uses a tonal shift and a divider rather than a floating layer.

### Shadow Vocabulary

- **QID base** (`var(--qid-shadow-base)` / `0 6px 20px rgba(0, 0, 0, 0.22)`): Organization forms, deletion confirmation, and auth panels that need a quiet lift.
- **QID focus ring** (`var(--qid-focus-ring)`): Keyboard and field focus; it is not a card shadow.
- **No rest shadow:** Assessment cards, project rows, profile columns, and navigation rely on borders and surface tone.

### Named Rules

**The Flat Workpaper Rule.** If a border and surface change can establish hierarchy, do not add a shadow.

## Shapes

Surfaces use 12–16px corners; controls use 8px corners; compact danger actions use a 6px corner; status chips and coverage markers use pill shapes; expand controls use circular outlines. Borders are 1px by default. A 2px top rule marks primary auth context, destructive confirmation, or assignment progress. Avoid hard offset shadows, clipping, and decorative geometry.

## Components

### Buttons

- **Shape:** 8px corners, 44px minimum height, 10px 16px padding, and medium-bold labels.
- **Primary:** Magenta-to-purple gradient with white text; hover brightens the magenta edge; active state shifts down by 1px.
- **Secondary:** QID supporting-surface fill with a strong border and readable text; hover receives a lifted surface and accent border.
- **Danger:** Text-first red action with transparent background and a 44px minimum touch target; hover adds the soft error tint and underline.
- **Focus:** All actionable controls use a visible magenta outline with a 3px offset.

### Chips

- **Style:** Small pill with a tinted background and readable text, used for roles, project status, and response status.
- **State:** Reviewing, Approved, needs-more-information, Current, and Target states pair color with text labels.

### Cards / Containers

- **Corner Style:** 12–16px for primary surfaces and assessment cards; 8px for compact rows and controls.
- **Background:** QID primary surface for content; QID supporting surface for expanded or supporting regions.
- **Shadow Strategy:** Border-led; Ambient panel shadow only where separation is needed.
- **Border:** Thin neutral rule with a restrained top rule for active context or destructive state.
- **Internal Padding:** 16px for compact rows, 18–24px for primary panels, and 14–18px on mobile.

### Inputs / Fields

- **Style:** Explicit labels above QID supporting-surface fields, 1px strong border, 8px corner, 11px 12px padding, and 1.65 line-height for textareas.
- **Focus:** Accent border plus the soft, offset focus ring.
- **Error / Disabled:** Error uses red text and a lightly tinted surface; disabled controls reduce opacity; read-only fields use Quiet Paper and muted text without hiding their values.

### Navigation

The Function index is a labeled navigation landmark. It uses a dark or light QID surface and a thin border boundary. Active items use a pink wash, a subtle magenta boundary, readable text, and `aria-current="page"`. At smaller widths it becomes a horizontally scrollable row without changing the reading order.

### Report Tables

Report tables use a labeled `role="region"` wrapper with `tabIndex="0"` so keyboard users can reach and scroll dense registers without losing context. The wrapper exposes a small-screen scroll hint, preserves table semantics, and pins the first identifying column below the single-column breakpoint. Tables stay read-only and keep status labels in text alongside their semantic colors.

### Project Context Metadata

The project context is a quiet reading surface rather than a dashboard widget. Labels use compact mono typography; values use the body face and remain top-aligned across the grid. The structure reflows from five columns to two and then one without changing the metadata order.

### Assessment Cards

Outcome cards use a quiet summary row with the outcome code, title, current-to-target coverage route, and circular expand affordance. Expanding reveals a Quiet Paper body containing the Counselor scope/profile controls or the Stakeholder response/evidence panel. Current and Target profile columns use their paired tinted surfaces and explicit headings.

### Response & Evidence Panel

The Stakeholder response panel keeps the response textarea, evidence upload, save/submit controls, review fields, and evidence list in one bordered QID surface. Previewable evidence opens inline inside a supporting-surface preview region; status and save feedback remain text-labeled.

### Authentication Panels

Login and invitation activation use a two-column QID surface composition on desktop: a short product introduction with a magenta top rule and a focused form panel with quiet ambient lift. At mobile widths the intro and form stack, preserving the same reading order and full-width primary action.

### Coverage Summary

Coverage summaries use restrained metric blocks for Overall, Included, Pending, and Returned counts. The Project Overview workbench repeats these high-value posture signals in a role-aware context, while the Function navigation and Overview register carry the per-Function percentage, included count, and attention count. This keeps progress visible without turning every detailed outcome page into a dashboard.

### Evidence Workbench Registers

Organization and project indexes use a register anatomy: a numbered wayfinding marker, primary identity, human-readable state, supporting context, and the smallest useful action set. Rows use thin rules and tonal surfaces instead of a wall of floating cards. On mobile, metadata and actions stack below the identity so the action remains reachable without horizontal scrolling.

The Project Overview workbench uses a posture panel, one `Next up` panel, and a Function register. The next-up copy is role-aware: Counselors see scope or review readiness, Assessors and Organization Admins see their input, Reviewers see responses awaiting a decision, and read-only roles see project posture. It is a presentation layer over existing derived data; it does not create permissions or a second workflow.

### Finalization & Audit Handoff

Project finalization is a deliberate reading state, not a decorative success screen. The Counselor sees readiness counts and the remaining outcome codes; the primary action stays disabled until every included outcome is Approved by the Reviewer. Once finalized, the workspace uses a clear **Finalized** label, presents the Final Report and Audit Package links, and removes mutation controls while preserving evidence preview and download.

The Final Report is a calm, print-first reading surface: project context and finalization metadata come first, followed by overall and per-Function coverage, then the included outcome register. Each outcome keeps Current / Target, stakeholder response, reviewer decision, and evidence metadata together so an auditor can read the assessment without jumping between screens. The browser print stylesheet removes navigation and actions, keeps table headers with their rows, and preserves the Current / Target tonal distinction in print where supported.

The Audit Package uses a denser register layout appropriate for handoff. Scope and assignment appear before response/evidence records, reviewer history follows the outcome register, and the chronological audit trail closes the package. The CSV action is explicit and secondary to reading; its columns mirror the traceability chain while omitting private storage keys. Audit pages remain read-only and reuse the QID surfaces, thin rules, compact mono labels, and the same magenta/purple orientation cues as the workspace.

## Do's and Don'ts

### Do:

- **Do** keep compliance copy on a high-contrast QID surface with a readable measure.
- **Do** use the cool desk canvas and thin rules to group content before reaching for shadow.
- **Do** reserve magenta/purple for active location, primary action, focus, and meaningful status.
- **Do** distinguish Counselor interpretation from Stakeholder response/evidence input through layout and labels.
- **Do** make Current and Target profiles visually distinct and textually labeled.
- **Do** preserve visible keyboard focus, role/status text, and responsive reading order at every breakpoint.
- **Do** use semantic theme tokens for muted text, Current / Target, Approved, Error, and links so both themes keep readable contrast.

### Don't:

- **Don't** use the brand gradient outside primary actions or purposeful orientation cues; avoid glass effects, decorative illustrations, and dense dashboard chrome.
- **Don't** use color alone to communicate response, coverage, or permission state.
- **Don't** turn every list row into a raised card; use borders and tonal layers first.
- **Don't** use JetBrains Mono as a display face; reserve it for codes, labels, and measurement.
- **Don't** make Stakeholder input controls look like Counselor-owned profile editing controls.
- **Don't** stretch assessment copy across the full viewport when a readable measure is available.
- **Don't** hide dense report columns on small screens without a focusable scroll region and a visible usage hint.
