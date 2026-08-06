# Delete Project Design

**Date:** 2026-08-06
**Status:** Approved
**Goal:** Permanently delete a project from the dashboard with an explicit confirmation and predictable cleanup.

## Scope

This increment adds permanent project deletion. It does not add archive, restore, bulk deletion, a recycle bin, or deletion from the assessment workspace.

## API and Data

Add `DELETE /api/projects/:id`. A successful deletion returns HTTP 204 with no body. An unknown or malformed project ID returns the existing HTTP 404 `not_found` error envelope. Database failures return HTTP 500 `internal_error`.

Deletion runs in one transaction:

1. Delete the selected row from `projects` and return its `organization_id`.
2. PostgreSQL automatically deletes related `project_functions` and `project_subcategory_profiles` through existing `ON DELETE CASCADE` constraints.
3. Delete the organization only when no remaining project or user references it.
4. Commit the transaction.

The NIST catalog and unrelated projects are never deleted. No schema migration is required.

## Dashboard Behavior

Each project card gets a visually secondary `Delete` button. Clicking it opens the browser confirmation dialog containing the project name and stating that its assessment data will be permanently deleted.

Cancellation performs no API request. Confirmation disables project actions while deletion is in progress. On HTTP 204, the card is removed from local state. On failure, the card remains and the existing dashboard error area shows a readable message.

## Components and Boundaries

- `api/internal/store/projects.go` owns transactional project deletion and conditional organization cleanup.
- `api/internal/httpapi/handler.go` exposes the DELETE route and maps not-found versus internal errors.
- `web/src/lib/api.ts` sends the DELETE request and supports empty successful responses.
- `ProjectDashboard` owns confirmation and emits the selected project to `onDelete`.
- `page.tsx` calls the API and removes the project from state only after success.

## Testing

Use red-green-refactor:

- Go handler tests cover HTTP 204, not found, and database failure.
- Dashboard tests cover confirmation, cancellation, and disabled actions during deletion.
- Home workflow test verifies that a successful deletion removes only the selected card.
- Docker smoke test creates a project, deletes it, verifies it is absent from the list, and verifies its profile endpoint no longer succeeds.

## Acceptance Criteria

1. A user can permanently delete a project from its dashboard card.
2. The project name and irreversible effect are shown before deletion.
3. Cancelling leaves the project unchanged.
4. Success removes the project and its dependent assessment rows.
5. Failure leaves the project visible and shows an error.
6. Unrelated projects, organizations still in use, users, and catalog rows remain intact.
