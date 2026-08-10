---
name: CSF Compliance
description: A calm case-file workspace for reading, completing, and reviewing NIST CSF assessments.
colors:
  primary: "#0d7567"
  primary-deep: "#0b4d44"
  accent-soft: "#e7f3ef"
  neutral-bg: "#eef3f1"
  surface: "#ffffff"
  surface-subtle: "#f7faf9"
  ink: "#17231f"
  muted: "#61716b"
  muted-strong: "#42534d"
  line: "#dbe5e1"
  line-strong: "#b4c7c0"
  current: "#edf4fb"
  target: "#f8f1e4"
  error: "#a9342b"
  success: "#207052"
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
  sm: "8px"
  md: "12px"
  pill: "999px"
spacing:
  sm: "8px"
  md: "16px"
  lg: "22px"
  xl: "40px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "#ffffff"
    rounded: "{rounded.sm}"
    padding: "10px 16px"
    height: "42px"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.primary-deep}"
    rounded: "{rounded.sm}"
    padding: "10px 16px"
    height: "42px"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.sm}"
    padding: "11px 12px"
  status-chip:
    backgroundColor: "{colors.accent-soft}"
    textColor: "{colors.primary-deep}"
    rounded: "{rounded.pill}"
    padding: "5px 9px"

# Design System: CSF Compliance

## Overview

**Creative North Star: "Clean Editorial Casefile"**

The interface treats every assessment as a calm working case file. White reading surfaces hold the content; cool neutral structure separates pages, rails, and states; restrained teal marks the active Function, the next action, and the few statuses that need attention. The result is deliberately quiet because the real product value is the relationship between counselor interpretation and stakeholder evidence.

The system is built for long sessions rather than quick dashboard glances. A Counselor gets a reading column with surrounding context; a Stakeholder gets the same evidence in a clearer input path. The visual language avoids gradients, decorative chrome, dense metric tiles, and heavy shadows that compete with compliance text.

**Key Characteristics:**

- White-first work surfaces with a cool neutral page canvas.
- Teal used as orientation and action, not as decoration.
- Editorial reading measure with explicit current/target comparison.
- Thin borders, modest corners, and one soft ambient shadow for raised panels.
- Role-aware context that explains whether the next action is reading, reviewing, or completing input.

## Colors

The palette follows the approved 60/30/10 rule: white surfaces dominate, light neutrals provide structure, and teal is reserved for orientation, action, and status.

### Primary

- **Casefile Teal** (`{colors.primary}`): Primary buttons, active Function navigation, focus borders, and the assessment rail rule.
- **Deep Teal Ink** (`{colors.primary-deep}`): High-emphasis numbers and readable text on teal-tinted surfaces.
- **Teal Wash** (`{colors.accent-soft}`): Active navigation, the leading summary value, and low-intensity status context.

### Neutral

- **Cool Desk** (`{colors.neutral-bg}`): The page canvas around white work surfaces.
- **Paper White** (`{colors.surface}`): Reading columns, forms, cards, rails, and authentication panels.
- **Quiet Paper** (`{colors.surface-subtle}`): Expanded assessment bodies and low-emphasis containers.
- **Ink** (`{colors.ink}`): Main headings, field values, and assessment copy.
- **Muted Graphite** (`{colors.muted}`): Supporting copy, metadata, and secondary labels.
- **Rule Gray** (`{colors.line}`): Dividers and default borders.
- **Strong Rule** (`{colors.line-strong}`): Input borders, focus-adjacent structure, and prominent dividers.
- **Current Blue** (`{colors.current}`) and **Target Sand** (`{colors.target}`): Small comparison surfaces for Current Profile and Target Profile; they are always paired with text labels.

### Named Rules

**The Teal Rarity Rule.** Teal should explain where the reader is or what action is available; it should not fill whole sections without a functional reason.

## Typography

**Display Font:** System UI stack (`system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`)
**Body Font:** The same system UI stack
**Label/Mono Font:** None; technical tone comes from hierarchy and data, not costume typography.

**Character:** Neutral, highly legible, and compact enough for forms while giving project titles enough weight to orient a long reading session.

### Hierarchy

- **Display** (800, `clamp(2rem, 4vw, 3.25rem)`, 1.18): Project, organization, and primary auth headings.
- **Headline** (700, `clamp(1.35rem, 2vw, 1.85rem)`, 1.18): Section headings and assessment titles.
- **Title** (700, `1.04rem`, 1.18): Outcome and form group titles.
- **Body** (400, `15px`, 1.65): Long-form descriptions and field content, kept to a readable measure.
- **Label** (750, `0.78rem`, 1.3): Field labels, metadata, and action context.

## Layout

The application shell uses a 220px Function navigation rail and a fluid content area. Project assessment screens use three logical zones: the Function index, a central reading column capped around 820px, and a 235–280px contextual rail. The main page container uses generous horizontal padding and a readable content measure rather than stretching text across the viewport.

Organization and project indexes use two-column work lists at larger widths. Create and invite forms remain adjacent to the list they affect. At 1100px the Function navigation becomes a horizontal index and the project layout stacks the context rail before the reading column. At 760px all lists and profile comparisons become one column, actions become full width, and response controls follow the content they act on.

The spacing rhythm is built from 8px, 16px, 22px, and 40px relationships. Headings receive more space above than below so sections read as editorial passages instead of stacked cards.

## Elevation & Depth

The system is border-led and flat by default. White surfaces sit on the cool desk canvas through tonal contrast and thin rules. Only panels that need separation from surrounding content use the soft ambient shadow `0 10px 26px rgba(22, 49, 42, 0.07)`. Focus uses a teal ring as an interaction state, not as ambient decoration.

### Shadow Vocabulary

- **Ambient panel** (`0 10px 26px rgba(22, 49, 42, 0.07)`): Organization forms, deletion confirmation, and auth panels that need a quiet lift.
- **No rest shadow:** Assessment cards, project rows, profile columns, and rails rely on borders and surface tone.

### Named Rules

**The Flat Workpaper Rule.** If a border and surface change can establish hierarchy, do not add a shadow.

## Shapes

Surfaces use gently curved 12px corners; controls use 8px corners; compact status chips and current/target value markers use pill shapes only when they communicate a bounded state. Borders are 1px by default, with a 2px top rule reserved for primary context or destructive state. No side stripes or hard offset shadows define the system.

## Components

### Buttons

- **Shape:** Small 8px corners with a 42px minimum height.
- **Anchor primary:** The invitation completion link uses the same teal action treatment as a primary button.
- **Primary:** Casefile Teal fill, white text, and 10px 16px padding. Hover deepens to Deep Teal Ink.
- **Secondary:** White fill with a Strong Rule border and Deep Teal Ink text; hover receives a very light Teal Wash.
- **Danger:** Text-first red action for destructive operations; the confirmation surface carries the structural warning rule.
- **Focus:** A visible 3px teal outline with 3px offset.

### Chips

- **Style:** Small pill with a light tinted background and readable text. Used for roles, project status, and response status.
- **State:** Current, target, submitted, reviewed, and needs-more-information states pair color with text labels.

### Cards / Containers

- **Corner Style:** 12px for surfaces, 8px for controls.
- **Background:** Paper White for primary content; Quiet Paper for expanded or supporting regions.
- **Shadow Strategy:** Border-led; Ambient panel shadow only where separation is needed.
- **Border:** Thin neutral rule with a restrained top rule for active context.
- **Internal Padding:** 16px for compact rows, 22px for primary panels, and 13–16px on mobile.

### Inputs / Fields

- **Style:** White field, 1px Strong Rule border, 8px corner, 11px 12px padding, and 1.65 line-height for textareas.
- **Focus:** Teal border plus a soft, offset focus ring.
- **Error / Disabled:** Error uses red text and a lightly tinted surface; read-only fields use Quiet Paper and muted text without hiding their values.

### Navigation

The Function index is a labeled navigation landmark. Active items use Teal Wash, a thin neutral boundary, Deep Teal Ink text, and `aria-current="page"`. At smaller widths it becomes a horizontally scrollable row without changing the reading order.

### Assessment Rail

The read-only context rail names the role-specific job, project, Function, outcome count, project status, coverage, pending count, and returned count. It stays beside the Counselor's reading column and moves before it when the layout stacks.

### Role boundary

Counselor users can read the complete profile, including out-of-scope outcomes, so they can maintain the assessment. Stakeholder users receive only included outcomes and use the response/evidence controls assigned to their role; the same boundary is enforced by the profile, response, and evidence APIs. Organization and project deletion are guarded by typed confirmation and report API failures in the confirmation surface.

## Do's and Don'ts

### Do:

- **Do** keep compliance copy on white surfaces with a readable measure.
- **Do** make current and target profiles visually distinct and textually labeled.
- **Do** reserve teal for active location, primary action, focus, and meaningful status.
- **Do** keep Counselor interpretation separate from Stakeholder response/evidence input.
- **Do** preserve visible keyboard focus and role/status text at every breakpoint.

### Don't:

- **Don't** reintroduce gradients, glass effects, decorative illustrations, or dense dashboard chrome.
- **Don't** use color alone to communicate response, coverage, or permission state.
- **Don't** put a heavy colored stripe on the side of a card or list item.
- **Don't** use a monospace or novelty display face as a technical costume.
- **Don't** make a Stakeholder's input controls look like Counselor-owned profile editing controls.
