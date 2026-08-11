# Counselor Assignment Progress Design

## Goal

Help a Counselor see which included outcomes still need a stakeholder assignment without opening every outcome card.

## Approved behavior

- The progress indicator is visible only to users with `userType === "counselor"`.
- `Included` counts profile rows where `included` is `true`.
- `Assigned` counts included rows where `assignedUserID` is not null.
- `Waiting for assignment` counts included rows where `assignedUserID` is null.
- Excluded outcomes do not count as waiting for assignment.
- The indicator is derived from the profile rows already loaded by the Project page; no API, database, or new endpoint is needed.
- Stakeholder, Reviewer, and Viewer views remain unchanged.

## Placement and copy

Place a compact assignment-progress strip directly below the existing assessment summary cards and before the outcome list. Use the existing editorial card styles and readable labels:

`Assignment progress` · `Included 12` · `Assigned 5` · `Waiting for assignment 7`

On narrow screens, the items wrap rather than creating horizontal overflow.

## Data flow

`ProjectPage` continues to own the canonical `profile` state. It passes `profile` and `user` to `ProjectAssessmentWorkspace`, which derives the three counts with a memoized calculation and renders the strip for Counselors.

## Acceptance criteria

1. A Counselor with included assigned and unassigned rows sees all three counts.
2. Changing scope or assigning a stakeholder updates the counts from the refreshed profile state.
3. A non-Counselor does not see the assignment-progress strip.
4. Existing scope, response, evidence, and review behavior remains unchanged.
5. Web tests, Go tests, typecheck, production build, and Docker health checks pass.
