# Delete Organization Design

## Goal

Allow a Counselor Admin to permanently delete a client organization from the organization index without adding archive or soft-delete infrastructure.

## Authorization

- Only `counselor_admin` can see and use the delete action.
- The API continues to enforce the same permission independently of the frontend.
- Other counselor and stakeholder roles cannot delete organizations.

## User experience

- Each organization card shows a `Delete` action only for a Counselor Admin.
- Selecting it opens an inline confirmation panel for that organization.
- The user must type the exact organization name before `Delete permanently` is enabled.
- The user can cancel without changing data.
- While deletion is running, repeat submission is disabled.
- Success removes the organization from the index immediately.
- API failures, including authorization and database errors, are displayed on the page rather than becoming unhandled promises.

## Data behavior

Deleting an organization permanently deletes its dependent data in one database transaction:

- Projects and their function/profile assessment rows
- Stakeholder users and their sessions
- Pending invitations
- Organization audit logs
- The organization itself

Counselor accounts remain because they are not owned by a client organization. The database store performs explicit dependent deletes so behavior does not rely on adding a broad cascade to existing foreign keys.

## API behavior

`DELETE /api/organizations/{id}` returns:

- `204 No Content` after successful deletion
- `403 Forbidden` for roles other than `counselor_admin`
- `404 Not Found` when the organization does not exist
- `500 Internal Server Error` for an unexpected transactional failure

## Testing

- Store test proves an organization with projects, profiles, stakeholders, sessions, invitations, and audit rows is deleted while counselor accounts remain.
- Handler authorization tests retain coverage for Counselor Admin versus other roles.
- Dashboard component tests prove the delete action is role-gated and requires exact-name confirmation.
- Page test proves successful deletion removes the organization and API errors appear in the UI.

## Out of scope

- Archive and restore
- Undo or recycle bin
- Bulk deletion
- Deleting counselor accounts
