# User and Role Management Design

## Goal

Expose the existing account-management API through the UI so administrators can maintain Counselor and Stakeholder access without using an API client.

## Scope

### Counselor Admin view: `/organizations`

- Add a `Counselors` section below the Organization list.
- List Counselor name, email, role, and status.
- Allow inviting a new `counselor` or `counselor_admin` and display the one-time invitation URL.
- Allow changing an existing Counselor role and status.
- Keep the current user's Disable action unavailable; the API remains the final guard.

### Organization view: `/organizations/{organization-slug}`

- Add role and status controls to the existing Stakeholders list.
- `counselor`, `counselor_admin`, and `org_admin` can manage stakeholder roles/status because the current API authorizes those roles.
- Stakeholder role options remain `org_admin`, `assessor`, `reviewer`, and `viewer`.
- Do not allow disabling the current user's own account.
- Keep the existing invitation flow and project controls unchanged.

## Data flow

- Reuse `api.getCounselors`, `api.updateCounselor`, and `api.createCounselorInvitation` on the Organizations page.
- Reuse `api.updateOrganizationUser` on the Organization page.
- Page components own the canonical user arrays and replace the updated row after a successful save.
- Invitation results remain one-time URLs and are displayed inline.

## Error and loading behavior

- Each account row has its own save state so one failing update does not block other rows.
- API errors are shown in an inline alert near the affected control.
- Disable controls are disabled while their row is saving.
- Existing page-level load and authorization errors remain unchanged.

## Acceptance criteria

1. Counselor Admin can see and manage Counselors from `/organizations`.
2. Counselor Admin can invite a Counselor and receive an invitation URL.
3. Counselor and Organization Admin can change Stakeholder role/status in an Organization.
4. Users cannot disable their own account from the UI; the backend restriction remains covered by existing tests.
5. Existing project creation, invitations, scope assignment, and assessment behavior remain unchanged.
6. Web tests, TypeScript, Go tests, and production build pass.
