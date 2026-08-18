---
target: rganizations-organizationslug-projects-projectslug
total_score: 22
max_score: 40
p0_count: 0
p1_count: 3
p2_count: 2
p3_count: 0
na_heuristics: 
timestamp: 2026-08-13T08-53-00Z
slug: rganizations-organizationslug-projects-projectslug
---
# Impeccable Critique — Project Assessment Workspace

## Method

Dual-agent critique:

- Assessment A: source/design review, isolated from detector findings.
- Assessment B: detector and browser evidence, isolated from Assessment A.

Target: web/src/app/organizations/[organizationSlug]/projects/[projectSlug]
Resolved slug: rganizations-organizationslug-projects-projectslug
Live URL attempted: http://localhost:3000/organizations/pea/projects/tmls
Ignore file: none

## Design Specificity Verdict

Moderately specific: the workflow is clearly NIST-CSF-specific, while the shell remains category-familiar. Specific evidence includes CSF Function navigation, Current/Target profiles, stakeholder assignment, response, evidence, and reviewer decisions. The white-first teal-accented visual system matches the Clean Editorial Casefile direction. The 220px rail, four KPI cards, and accordion-card composition could still appear in a generic audit or risk product. The page is missing stronger case-file context such as objective, assessment period, target date, category grouping, and a reviewer queue.

## Design Health Score

| Heuristic | Score |
|---|---:|
| Visibility of system status | 2/4 |
| Match between system and real world | 3/4 |
| User control and freedom | 2/4 |
| Consistency and standards | 3/4 |
| Error prevention | 2/4 |
| Recognition rather than recall | 2/4 |
| Flexibility and efficiency | 2/4 |
| Aesthetic and minimalist design | 3/4 |
| Error recognition and recovery | 2/4 |
| Help and documentation | 1/4 |
| Total | 22/40 — Acceptable |

The visual foundation is solid; workflow clarity and state communication need improvement before long-session use.

## Cognitive Load and Emotional Journey

Cognitive load is high for stakeholder and reviewer work. A single expanded card combines Current, Target, notes, response, evidence, save, submit, and review. Grouping and initial progressive disclosure are good, but visual hierarchy, single-focus work, working-memory support, and chunking are weak.

Arrival feels calm and credible. Orientation is adequate but does not clearly tell stakeholders which outcome to do next. Expanding the first card sharply increases effort. The phrase “Changes are saved per outcome” creates the largest trust problem when edits may not yet be saved. Reviewers must open cards to find statuses, so the end state is only moderately reassuring.

## Strengths

- Disciplined white surfaces, thin rules, restrained teal, and Current/Target tonal separation.
- Role boundaries are structurally respected for Counselor, Stakeholder, Reviewer, and Viewer.
- Expandable controls, live response status, evidence preview labels, and role-aware empty states are thoughtful accessibility foundations.

## Priority Issues

### P1 — “Saved” appears before save confirmation

AssessmentCard.tsx:250 shows an idle message that can imply persistence while a visible Save assessment action still exists. Distinguish Unsaved changes, Saving, and server-confirmed Saved; give response edits the same dirty-state vocabulary. Suggested command: $impeccable clarify.

### P1 — Review triage is hidden in accordions

AssessmentCard.tsx:130 exposes outcome information but not response/review status, assignee, or last activity. FunctionSidebar.tsx has no progress indicators. Add status, assignment, and attention counts to collapsed rows and the Function rail; unify Returned with Needs more information. Suggested command: $impeccable layout.

### P1 — One card carries four different jobs

AssessmentCard.tsx:150 combines scope, assignment, profiles, stakeholder response, evidence, and review. Make the expanded experience role-specific: Stakeholder starts at response/evidence, Counselor at scope/assignment, Reviewer at submitted evidence and decision context. Suggested command: $impeccable distill.

### P2 — Read-only roles look like disabled forms

Disabled fieldsets make Viewer/Reviewer reading feel unavailable. Use deliberate read-only content with an explicit permission label and separate review controls. Suggested command: $impeccable audit.

### P2 — Bulk scope changes lack partial-failure recovery

The Include all action can update outcomes serially. Show affected count, mixed state, partial-result summary, retry, or undo. Suggested command: $impeccable harden.

## Persona Red Flags

- Alex/Counselor: assignment remains one-card-at-a-time; no Function-level attention queue or returned/unassigned filters; misleading save state risks false confidence.
- Sam/Stakeholder: profile fields, response, evidence, and submission are combined; Submit can be disabled without explaining that response must be saved first; no assigned-outcome queue.
- Casey/Reviewer/Viewer: review state is hidden in expanded cards; Viewer sees disabled inputs instead of a confident reading surface; Current/Target chips lack explicit labels.

## Minor Observations

- The intended 820px editorial column may be left-anchored on wide screens.
- Project header omits objective, assessment period, target date, and compliance driver even though project data supports them.
- Category hierarchy is not surfaced; outcomes show only subcategory code and description.
- Coverage should explicitly label Current and Target rather than relying on chip order.
- Lowercase project status text is produced by replaceAll.
- Needs more information should require a review comment.
- A reviewed response without a comment can leave the response footer visually empty.

## Questions to Consider

- Is the primary task configuring scope, completing outcomes, reviewing submissions, or reading evidence? Why do all four share one accordion?
- Could the Function rail become an attention queue?
- What exactly does “saved” mean: locally changed, server-confirmed, submitted, or reviewed?
- If Casey is read-only, why does the interface look like a disabled form?

## Evidence Packet

Detector command:

    node .agents/skills/impeccable/scripts/detect.mjs --json web/src

Exit code: 0; JSON output: []; findings: 0; rule IDs and locations: none. No ignore file was present.

Browser evidence: a fresh in-app browser tab attempted the project URL but received net::ERR_CONNECTION_REFUSED. DOM snapshot, screenshot, responsive checks, overflow, focus, target size, loading/error states, and console evidence were unavailable. Visualization overlay injection was skipped because the document never loaded. The live server was not left running; temporary logs were cleaned. No repository source files were edited.
