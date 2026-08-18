---
name: CSF Compliance
description: A branded trust workspace for reading, completing, and reviewing NIST CSF assessments.
colors:
  magenta: "#e6007a"
  magenta-soft: "#ff2d9a"
  purple: "#7b2cbf"
  purple-soft: "#9b5de5"
  navy: "#0b1d3a"
  navy-deep: "#060e1d"
  gold: "#f5df4d"
  ink: "#0b1d3a"
  muted: "#5a6580"
  muted-strong: "#33415e"
  line: "rgba(11, 29, 58, 0.1)"
  line-strong: "rgba(11, 29, 58, 0.16)"
  accent: "#e6007a"
  accent-dark: "#7b2cbf"
  accent-soft: "#ffd1ec"
  accent-faint: "#fff1f8"
  surface: "#ffffff"
  surface-subtle: "#fafbfd"
  canvas: "#f1f4fa"
  current: "#edf2ff"
  current-ink: "#295a8a"
  target: "#fff3dc"
  target-ink: "#785b1a"
  error: "#b42318"
  error-soft: "#fff2f0"
  success: "#147d55"
  warning: "#9a5f00"
typography:
  display:
    fontFamily: "Space Grotesk, IBM Plex Sans Thai Looped, -apple-system, BlinkMacSystemFont, sans-serif"
    fontSize: "clamp(2rem, 4vw, 3.25rem)"
    fontWeight: 800
    lineHeight: 1.18
    letterSpacing: "-0.04em"
  headline:
    fontFamily: "Space Grotesk, IBM Plex Sans Thai Looped, -apple-system, BlinkMacSystemFont, sans-serif"
    fontSize: "clamp(1.35rem, 2vw, 1.85rem)"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.03em"
  title:
    fontFamily: "Space Grotesk, IBM Plex Sans Thai Looped, -apple-system, BlinkMacSystemFont, sans-serif"
    fontSize: "1.04rem"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.02em"
  body:
    fontFamily: "Space Grotesk, IBM Plex Sans Thai Looped, -apple-system, BlinkMacSystemFont, sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "JetBrains Mono, ui-monospace, monospace"
    fontSize: "0.78rem"
    fontWeight: 750
    lineHeight: 1.3
rounded:
  xs: "6px"
  sm: "10px"
  md: "14px"
  lg: "20px"
  pill: "999px"
  circle: "50%"
spacing:
  xs: "6px"
  sm: "8px"
  md: "16px"
  lg: "22px"
  xl: "40px"
components:
  button-primary:
    backgroundColor: "linear-gradient(135deg, {colors.magenta}, {colors.purple})"
    textColor: "#ffffff"
    rounded: "{rounded.sm}"
    padding: "10px 16px"
    height: "42px"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.accent-dark}"
    rounded: "{rounded.sm}"
    padding: "10px 16px"
    height: "42px"
  button-danger:
    backgroundColor: "transparent"
    textColor: "{colors.error}"
    rounded: "{rounded.xs}"
    padding: "10px 8px"
    height: "44px"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.sm}"
    padding: "11px 12px"
  status-chip:
    backgroundColor: "{colors.accent-soft}"
    textColor: "{colors.accent-dark}"
    rounded: "{rounded.pill}"
    padding: "5px 9px"
  nav-item:
    backgroundColor: "transparent"
    textColor: "{colors.muted}"
    rounded: "{rounded.sm}"
    padding: "10px 12px"
    height: "44px"
  assessment-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "0"
  response-panel:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "22px"
---

# Design System: CSF Compliance

## Overview

**Creative North Star: "Versotis Trust Workspace"**

The interface turns a complex assessment into a clear, accountable trust workspace. Paper-white reading surfaces hold the content in light mode; deep navy surfaces create the same calm reading hierarchy in dark mode; magenta-to-purple brand cues mark the active Function, the next action, and the few statuses that need attention. A visible theme toggle can select light or dark mode, while the system preference remains the default when no explicit choice is stored. The result borrows Versotis' strategic energy without sacrificing long compliance-reading sessions in either ambient light.

The system keeps Counselor interpretation, Stakeholder response/evidence input, and Reviewer feedback visually legible as different kinds of work. Organization and project pages use structured lists and compact forms; assessment pages use an editorial reading column with expandable outcome cards. The visual language uses the brand gradient only for actions and orientation, avoiding decorative chrome and heavy shadows that compete with compliance text.

**Key Characteristics:**

- Paper-white work surfaces with a cool neutral page canvas in light mode, and deep navy work surfaces with blue-black elevation in dark mode.
- Magenta-to-purple used as orientation and action, not as decoration.
- Space Grotesk for the product voice, IBM Plex Sans Thai Looped for Thai fallback, and JetBrains Mono for wayfinding/data labels.
- Editorial reading measure with explicit Current / Target comparison.
- Thin borders, modest corners, and one soft ambient shadow for raised panels.
- Role-aware surfaces that distinguish reading, reviewing, and completing input.
- Project context metadata aligned from a shared top edge so uneven copy lengths do not make short values appear vertically suspended.

## Colors

The palette follows the 60/30/10 rule: white surfaces dominate reading areas, light navy neutrals provide page structure and dividers, and the magenta-to-purple accent is reserved for orientation, action, focus, and meaningful status.

### Primary

- **Versotis Magenta** (`{colors.accent}`) and **Purple** (`{colors.accent-dark}`): Primary actions, active Function navigation, focus borders, links, and assessment orientation cues.
- **Brand Gradient** (`linear-gradient(135deg, {colors.magenta}, {colors.purple})`): Primary buttons and high-value action anchors only.
- **Pink Wash** (`{colors.accent-soft}`) and **Faint Pink** (`{colors.accent-faint}`): Selected navigation, role/status chips, progress context, and low-intensity supporting regions.

### Neutral

- **Cool Desk** (`{colors.canvas}`): The page canvas around white work surfaces.
- **Paper White** (`{colors.surface}`): Reading columns, forms, cards, authentication panels, and response surfaces.
- **Quiet Paper** (`{colors.surface-subtle}`): Expanded assessment bodies, read-only fields, previews, and supporting regions.
- **Ink** (`{colors.ink}`), **Muted Graphite** (`{colors.muted}`), and **Strong Graphite** (`{colors.muted-strong}`): Headings, metadata, labels, and long-form copy.
- **Rule Gray** (`{colors.line}`) and **Strong Rule** (`{colors.line-strong}`): Dividers, card boundaries, input strokes, and structural separation.
- **Current Blue** (`{colors.current}` / `{colors.current-ink}`) and **Target Sand** (`{colors.target}` / `{colors.target-ink}`): Paired comparison surfaces that always include a text label.
- **Error Red** (`{colors.error}` / `{colors.error-soft}`), **Success Green** (`{colors.success}`), and **Warning Ochre** (`{colors.warning}`): Feedback states that are paired with text and never used as the only signal.

### Named Rules

**The Brand Cue Rule.** Magenta and purple should explain where the reader is or what action is available; the brand gradient belongs to actions and orientation, not whole reading sections.

**The Labeled State Rule.** Current, target, submitted, reviewed, error, and permission states use color alongside a visible text label.

### Dark Theme

Dark mode follows the system `prefers-color-scheme` setting. The dark palette uses Versotis navy (`{colors.navy-deep}`) as the canvas, Versotis navy (`{colors.navy}`) for work surfaces, a lifted blue-navy for supporting regions, and the same magenta-to-purple action gradient. Body text moves to a soft white and secondary text to a cool blue-gray so long compliance copy remains readable without turning the interface into pure black and white.

## Typography

**Display Font:** Space Grotesk (`Space Grotesk`, with IBM Plex Sans Thai Looped fallback)
**Body Font:** Space Grotesk with IBM Plex Sans Thai Looped for Thai glyphs
**Label/Mono Font:** JetBrains Mono for outcome codes, context markers, and compact data labels.

**Character:** Strategic, technical, and crisp without becoming a marketing landing page; headings carry the brand while body copy stays comfortable for a long reading session.

### Hierarchy

- **Display** (800, `clamp(2rem, 4vw, 3.25rem)`, 1.18, `-0.04em`): Organization, project, and authentication headings.
- **Headline** (700, `clamp(1.35rem, 2vw, 1.85rem)`, 1.18, `-0.03em`): Section headings and major workspace titles.
- **Title** (700, `1.04rem`, 1.18, `-0.02em`): Outcome, card, and form-group titles.
- **Body** (400, `15px`, 1.65): Long-form descriptions, assessment copy, and field content; paragraph measure stays near 72ch.
- **Label** (750, `0.78rem`, 1.3): Field labels and supporting metadata. Eyebrows and context lines use compact uppercase tracking for orientation.

### Named Rules

**The Reading Measure Rule.** Let hierarchy and a readable line length do the organizing; do not compensate for long compliance text with oversized decorative type.

## Layout

The application shell uses a 220px sticky Function navigation rail and a fluid main area. Main content is capped at 1540px with a responsive horizontal gutter; organization dashboards and project workspaces share a centered 1320px content frame. Desktop organization and project indexes use two-column work lists, while the project assessment keeps its content in a full-width workspace with prose capped by the 72ch reading measure. The assessment page currently does not rely on a separate contextual rail; orientation is carried by the project header, Function navigation, assignment-progress banner, and outcome count.

The project context panel uses a five-column metadata grid on wide screens, two columns at the rail breakpoint, and one column on small screens. Each metadata cell uses top-aligned content, a compact 3px label/value gap, and a thin rule above the row so long Scope or Compliance driver copy does not vertically center shorter dates and periods.

At 1100px the Function rail becomes a horizontal, scrollable index and multi-column creator forms begin to stack. At 760px lists, profile comparisons, supporting grids, and authentication surfaces become one column; actions stretch to the available width; assessment summaries move the coverage route below the outcome title; and response/evidence controls follow the content they act on. The spacing rhythm is based on 8px, 16px, 22px, and 40px relationships, with compact 6px and 14–18px gaps where controls need tighter grouping.

## Elevation & Depth

The system is border-led and flat by default. White surfaces sit on the light navy-gray canvas through tonal contrast and thin rules. The soft ambient shadow is reserved for forms, deletion confirmation, authentication panels, and branded primary actions that need quiet separation. Focus uses a magenta ring as an interaction state, not as ambient decoration. Expanded assessment content uses a tonal shift and a divider rather than a floating layer.

### Shadow Vocabulary

- **Ambient panel** (`var(--shadow-soft)` / `0 24px 60px -30px rgba(11, 29, 58, 0.24)`): Organization forms, deletion confirmation, and auth panels that need a quiet lift.
- **Focus ring** (`var(--shadow-focus)` / `0 0 0 4px rgba(230, 0, 122, 0.18)`): Keyboard and field focus; it is not a card shadow.
- **No rest shadow:** Assessment cards, project rows, profile columns, and navigation rely on borders and surface tone.

### Named Rules

**The Flat Workpaper Rule.** If a border and surface change can establish hierarchy, do not add a shadow.

## Shapes

Surfaces use gently curved 14px corners; controls use 10px corners; compact danger actions use a 6px corner; status chips and coverage markers use pill shapes; expand controls use circular outlines. Borders are 1px by default. A 2px top rule marks primary auth context, destructive confirmation, or assignment progress. Avoid hard offset shadows, clipping, and decorative geometry.

## Components

### Buttons

- **Shape:** 10px corners, 42px minimum height, 10px 16px padding, and medium-bold labels.
- **Primary:** Magenta-to-purple gradient with white text; hover brightens the magenta edge; active state shifts down by 1px.
- **Secondary:** Paper White fill with a Strong Rule border and Purple text; hover receives a Faint Pink wash and accent border.
- **Danger:** Text-first red action with transparent background and a 44px minimum touch target; hover adds the soft error tint and underline.
- **Focus:** All actionable controls use a visible magenta outline with a 3px offset.

### Chips

- **Style:** Small pill with a tinted background and readable text, used for roles, project status, and response status.
- **State:** Submitted, reviewed, needs-more-information, Current, and Target states pair color with text labels.

### Cards / Containers

- **Corner Style:** 14px for primary surfaces and assessment cards; 10px for compact rows and controls.
- **Background:** Paper White for primary content; Quiet Paper for expanded or supporting regions.
- **Shadow Strategy:** Border-led; Ambient panel shadow only where separation is needed.
- **Border:** Thin neutral rule with a restrained top rule for active context or destructive state.
- **Internal Padding:** 16px for compact rows, 18–22px for primary panels, and 13–16px on mobile.

### Inputs / Fields

- **Style:** Explicit labels above white fields, 1px Strong Rule border, 10px corner, 11px 12px padding, and 1.65 line-height for textareas.
- **Focus:** Accent border plus the soft, offset focus ring.
- **Error / Disabled:** Error uses red text and a lightly tinted surface; disabled controls reduce opacity; read-only fields use Quiet Paper and muted text without hiding their values.

### Navigation

The Function index is a labeled navigation landmark. It uses a white surface and a thin Rule Gray boundary. Active items use Pink Wash, a subtle magenta boundary, Purple text, and `aria-current="page"`. At smaller widths it becomes a horizontally scrollable row without changing the reading order.

### Project Context Metadata

The project context is a quiet reading surface rather than a dashboard widget. Labels use compact muted typography; values use the body face and remain top-aligned across the grid. The structure reflows from five columns to two and then one without changing the metadata order.

### Assessment Cards

Outcome cards use a quiet summary row with the outcome code, title, current-to-target coverage route, and circular expand affordance. Expanding reveals a Quiet Paper body containing the Counselor scope/profile controls or the Stakeholder response/evidence panel. Current and Target profile columns use their paired tinted surfaces and explicit headings.

### Response & Evidence Panel

The Stakeholder response panel keeps the response textarea, evidence upload, save/submit controls, review fields, and evidence list in one bordered white surface. Previewable evidence opens inline inside a Quiet Paper preview region; status and save feedback remain text-labeled.

### Authentication Panels

Login and invitation activation use a two-column white-panel composition on desktop: a short product introduction with a magenta top rule and a focused form panel with quiet ambient lift. At mobile widths the intro and form stack, preserving the same reading order and full-width primary action.

### Coverage Summary

Coverage summaries use restrained metric blocks for Overall, Included, Pending, and Returned counts. The Function navigation carries the per-Function percentage, included count, and role-specific attention count so progress is visible without repeating a large dashboard banner.

## Do's and Don'ts

### Do:

- **Do** keep compliance copy on white surfaces with a readable measure.
- **Do** use the cool desk canvas and thin rules to group content before reaching for shadow.
- **Do** reserve magenta/purple for active location, primary action, focus, and meaningful status.
- **Do** distinguish Counselor interpretation from Stakeholder response/evidence input through layout and labels.
- **Do** make Current and Target profiles visually distinct and textually labeled.
- **Do** preserve visible keyboard focus, role/status text, and responsive reading order at every breakpoint.

### Don't:

- **Don't** use the brand gradient outside primary actions or purposeful orientation cues; avoid glass effects, decorative illustrations, and dense dashboard chrome.
- **Don't** use color alone to communicate response, coverage, or permission state.
- **Don't** turn every list row into a raised card; use borders and tonal layers first.
- **Don't** use JetBrains Mono as a display face; reserve it for codes, labels, and measurement.
- **Don't** make Stakeholder input controls look like Counselor-owned profile editing controls.
- **Don't** stretch assessment copy across the full viewport when a readable measure is available.
