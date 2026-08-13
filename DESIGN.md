---
name: CSF Compliance
description: A calm case-file workspace for reading, completing, and reviewing NIST CSF assessments.
colors:
  ink: "#17231f"
  muted: "#61716b"
  muted-strong: "#42534d"
  line: "#dbe5e1"
  line-strong: "#b4c7c0"
  accent: "#0d7567"
  accent-dark: "#0b4d44"
  accent-soft: "#e7f3ef"
  accent-faint: "#f2f8f6"
  surface: "#ffffff"
  surface-subtle: "#f7faf9"
  canvas: "#eef3f1"
  current: "#edf4fb"
  current-ink: "#295679"
  target: "#f8f1e4"
  target-ink: "#76571f"
  error: "#a9342b"
  error-soft: "#fff4f2"
  success: "#207052"
  warning: "#9a6716"
typography:
  display:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "clamp(2rem, 4vw, 3.25rem)"
    fontWeight: 800
    lineHeight: 1.18
    letterSpacing: "-0.04em"
  headline:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "clamp(1.35rem, 2vw, 1.85rem)"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.03em"
  title:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "1.04rem"
    fontWeight: 700
    lineHeight: 1.18
    letterSpacing: "-0.02em"
  body:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: "0.78rem"
    fontWeight: 750
    lineHeight: 1.3
rounded:
  xs: "6px"
  sm: "8px"
  md: "12px"
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
    backgroundColor: "{colors.accent}"
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

**Creative North Star: "Clean Editorial Casefile"**

The interface treats every assessment as a calm working case file. White reading surfaces hold the content; the cool desk canvas gives pages and groups enough separation to scan; restrained teal marks the active Function, the next action, and the few statuses that need attention. The result is deliberately quiet because the product is used for long compliance-reading sessions, not quick dashboard glances.

The system keeps Counselor interpretation, Stakeholder response/evidence input, and Reviewer feedback visually legible as different kinds of work. Organization and project pages use structured lists and compact forms; assessment pages use an editorial reading column with expandable outcome cards. The visual language avoids gradients, decorative chrome, dense metric tiles, and heavy shadows that compete with compliance text.

**Key Characteristics:**

- White-first work surfaces with a cool neutral page canvas.
- Teal used as orientation and action, not as decoration.
- Editorial reading measure with explicit Current / Target comparison.
- Thin borders, modest corners, and one soft ambient shadow for raised panels.
- Role-aware surfaces that distinguish reading, reviewing, and completing input.

## Colors

The palette follows the 60/30/10 rule: white surfaces dominate reading areas, cool neutrals provide page structure and dividers, and teal is reserved for orientation, action, focus, and meaningful status.

### Primary

- **Casefile Teal** (`{colors.accent}`): Primary actions, active Function navigation, focus borders, links, and assessment orientation cues.
- **Deep Teal Ink** (`{colors.accent-dark}`): High-emphasis text and numbers on tinted surfaces, plus primary-action hover states.
- **Teal Wash** (`{colors.accent-soft}`) and **Faint Teal** (`{colors.accent-faint}`): Selected navigation, role/status chips, progress context, and low-intensity supporting regions.

### Neutral

- **Cool Desk** (`{colors.canvas}`): The page canvas around white work surfaces.
- **Paper White** (`{colors.surface}`): Reading columns, forms, cards, authentication panels, and response surfaces.
- **Quiet Paper** (`{colors.surface-subtle}`): Expanded assessment bodies, read-only fields, previews, and supporting regions.
- **Ink** (`{colors.ink}`), **Muted Graphite** (`{colors.muted}`), and **Strong Graphite** (`{colors.muted-strong}`): Headings, metadata, labels, and long-form copy.
- **Rule Gray** (`{colors.line}`) and **Strong Rule** (`{colors.line-strong}`): Dividers, card boundaries, input strokes, and structural separation.
- **Current Blue** (`{colors.current}` / `{colors.current-ink}`) and **Target Sand** (`{colors.target}` / `{colors.target-ink}`): Paired comparison surfaces that always include a text label.
- **Error Red** (`{colors.error}` / `{colors.error-soft}`), **Success Green** (`{colors.success}`), and **Warning Ochre** (`{colors.warning}`): Feedback states that are paired with text and never used as the only signal.

### Named Rules

**The Teal Rarity Rule.** Teal should explain where the reader is or what action is available; it should not fill whole sections without a functional reason.

**The Labeled State Rule.** Current, target, submitted, reviewed, error, and permission states use color alongside a visible text label.

## Typography

**Display Font:** System UI stack (`system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`)
**Body Font:** The same system UI stack
**Label/Mono Font:** None; the technical tone comes from hierarchy and data, not a novelty typeface.

**Character:** Neutral, highly legible, and compact enough for forms while giving organization, project, and assessment headings enough weight to orient a long reading session.

### Hierarchy

- **Display** (800, `clamp(2rem, 4vw, 3.25rem)`, 1.18, `-0.04em`): Organization, project, and authentication headings.
- **Headline** (700, `clamp(1.35rem, 2vw, 1.85rem)`, 1.18, `-0.03em`): Section headings and major workspace titles.
- **Title** (700, `1.04rem`, 1.18, `-0.02em`): Outcome, card, and form-group titles.
- **Body** (400, `15px`, 1.65): Long-form descriptions, assessment copy, and field content; paragraph measure stays near 72ch.
- **Label** (750, `0.78rem`, 1.3): Field labels and supporting metadata. Eyebrows and context lines use compact uppercase tracking for orientation.

### Named Rules

**The Reading Measure Rule.** Let hierarchy and a readable line length do the organizing; do not compensate for long compliance text with oversized decorative type.

## Layout

The application shell uses a 220px sticky Function navigation rail and a fluid main area. Main content is capped at 1540px with a responsive horizontal gutter; organization dashboards are capped at 1320px. Desktop organization and project indexes use two-column work lists, while the project assessment uses one central reading column capped at 820px. The assessment page currently does not rely on a separate contextual rail; orientation is carried by the project header, Function navigation, assignment-progress banner, and outcome count.

At 1100px the Function rail becomes a horizontal, scrollable index and multi-column creator forms begin to stack. At 760px lists, profile comparisons, supporting grids, and authentication surfaces become one column; actions stretch to the available width; assessment summaries move the coverage route below the outcome title; and response/evidence controls follow the content they act on. The spacing rhythm is based on 8px, 16px, 22px, and 40px relationships, with compact 6px and 14–18px gaps where controls need tighter grouping.

## Elevation & Depth

The system is border-led and flat by default. White surfaces sit on the cool desk canvas through tonal contrast and thin rules. The soft ambient shadow is reserved for forms, deletion confirmation, and authentication panels that need quiet separation. Focus uses a teal ring as an interaction state, not as ambient decoration. Expanded assessment content uses a tonal shift and a divider rather than a floating layer.

### Shadow Vocabulary

- **Ambient panel** (`var(--shadow-soft)` / `0 10px 26px rgba(22, 49, 42, 0.07)`): Organization forms, deletion confirmation, and auth panels that need a quiet lift.
- **Focus ring** (`var(--shadow-focus)` / `0 0 0 4px rgba(13, 117, 103, 0.16)`): Keyboard and field focus; it is not a card shadow.
- **No rest shadow:** Assessment cards, project rows, profile columns, and navigation rely on borders and surface tone.

### Named Rules

**The Flat Workpaper Rule.** If a border and surface change can establish hierarchy, do not add a shadow.

## Shapes

Surfaces use gently curved 12px corners; controls use 8px corners; compact danger actions use a 6px corner; status chips and coverage markers use pill shapes; expand controls use circular outlines. Borders are 1px by default. A 2px top rule marks primary auth context, destructive confirmation, or assignment progress. Avoid hard offset shadows, clipping, and decorative geometry.

## Components

### Buttons

- **Shape:** 8px corners, 42px minimum height, 10px 16px padding, and medium-bold labels.
- **Primary:** Casefile Teal fill with white text; hover deepens to Deep Teal Ink; active state shifts down by 1px.
- **Secondary:** Paper White fill with a Strong Rule border and Deep Teal Ink text; hover receives a Faint Teal wash and accent border.
- **Danger:** Text-first red action with transparent background and a 44px minimum touch target; hover adds the soft error tint and underline.
- **Focus:** All actionable controls use a visible teal outline with a 3px offset.

### Chips

- **Style:** Small pill with a tinted background and readable text, used for roles, project status, and response status.
- **State:** Submitted, reviewed, needs-more-information, Current, and Target states pair color with text labels.

### Cards / Containers

- **Corner Style:** 12px for primary surfaces and assessment cards; 8px for compact rows and controls.
- **Background:** Paper White for primary content; Quiet Paper for expanded or supporting regions.
- **Shadow Strategy:** Border-led; Ambient panel shadow only where separation is needed.
- **Border:** Thin neutral rule with a restrained top rule for active context or destructive state.
- **Internal Padding:** 16px for compact rows, 18–22px for primary panels, and 13–16px on mobile.

### Inputs / Fields

- **Style:** Explicit labels above white fields, 1px Strong Rule border, 8px corner, 11px 12px padding, and 1.65 line-height for textareas.
- **Focus:** Accent border plus the soft, offset focus ring.
- **Error / Disabled:** Error uses red text and a lightly tinted surface; disabled controls reduce opacity; read-only fields use Quiet Paper and muted text without hiding their values.

### Navigation

The Function index is a labeled navigation landmark. It uses a white surface and a thin Rule Gray boundary. Active items use Teal Wash, a subtle boundary, Deep Teal Ink text, and `aria-current="page"`. At smaller widths it becomes a horizontally scrollable row without changing the reading order.

### Assessment Cards

Outcome cards use a quiet summary row with the outcome code, title, current-to-target coverage route, and circular expand affordance. Expanding reveals a Quiet Paper body containing the Counselor scope/profile controls or the Stakeholder response/evidence panel. Current and Target profile columns use their paired tinted surfaces and explicit headings.

### Response & Evidence Panel

The Stakeholder response panel keeps the response textarea, evidence upload, save/submit controls, review fields, and evidence list in one bordered white surface. Previewable evidence opens inline inside a Quiet Paper preview region; status and save feedback remain text-labeled.

### Authentication Panels

Login and invitation activation use a two-column white-panel composition on desktop: a short product introduction with a teal top rule and a focused form panel with quiet ambient lift. At mobile widths the intro and form stack, preserving the same reading order and full-width primary action.

## Do's and Don'ts

### Do:

- **Do** keep compliance copy on white surfaces with a readable measure.
- **Do** use the cool desk canvas and thin rules to group content before reaching for shadow.
- **Do** reserve teal for active location, primary action, focus, and meaningful status.
- **Do** distinguish Counselor interpretation from Stakeholder response/evidence input through layout and labels.
- **Do** make Current and Target profiles visually distinct and textually labeled.
- **Do** preserve visible keyboard focus, role/status text, and responsive reading order at every breakpoint.

### Don't:

- **Don't** reintroduce gradients, glass effects, decorative illustrations, or dense dashboard chrome.
- **Don't** use color alone to communicate response, coverage, or permission state.
- **Don't** turn every list row into a raised card; use borders and tonal layers first.
- **Don't** use a monospace or novelty display face as a technical costume.
- **Don't** make Stakeholder input controls look like Counselor-owned profile editing controls.
- **Don't** stretch assessment copy across the full viewport when a readable measure is available.
