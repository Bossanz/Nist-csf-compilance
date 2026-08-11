# Bulk Include Outcomes Design

**Status:** Approved for implementation

## Goal

Allow a Counselor to include or exclude every outcome in the currently selected CSF Function with one checkbox.

## Design

The assessment header shows `Include all outcomes in this Function` only for Counselor users. Its checked state is true only when every outcome in the selected Function is included. Toggling it updates only the `included` field for those outcomes; rationale and stakeholder assignment remain unchanged. Individual outcome checkboxes remain available for follow-up adjustments.

The existing profile update API is called once per outcome in the selected Function. No new endpoint or database field is added in v1. Stakeholder, Reviewer, and Viewer roles do not see the bulk control.

## Verification

Component tests cover Counselor visibility, checked state, and callback payload. Page tests cover updating every selected-Function outcome with `{ included }`. Existing full web and Go tests, typecheck, production build, and Docker health checks must remain green.
