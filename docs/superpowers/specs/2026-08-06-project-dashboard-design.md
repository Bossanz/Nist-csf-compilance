# Project Dashboard Design

**Date:** 2026-08-06
**Status:** Approved
**Goal:** Let users see existing projects on the home page and reopen an assessment after a refresh.

## Scope

This increment adds a lean project dashboard to the existing single-page application. It includes project listing, project selection, project creation, returning from an assessment to the list, and loading, empty, and error states.

It does not add pagination, search, deletion, authentication, client-side persistence, a state library, or a new project route.

## API and Data

Add `GET /api/projects`. It returns a JSON array ordered by `projects.created_at DESC`, with a stable secondary order by project ID. Each item contains:

- `id`
- `organizationID`
- `organizationName`
- `name`
- `status`
- `createdAt`

The store query joins `projects` to `organizations`; no schema migration is required. Existing create-project and get-project responses also include `organizationName` so the `Project` contract stays consistent.

Database errors from the list endpoint return HTTP 500 with the existing `internal_error` envelope. A successful empty result returns HTTP 200 with `[]`.

## Web UI

On initial load, the home page requests the function catalog and project list. Until projects resolve, it shows a readable loading state. If the list is empty, it explains that no projects exist yet. If projects exist, it renders one card per project with project name, organization name, status, creation date, and an `Open project` action.

The create-project form remains on the dashboard. After creation, the new project is added to the local list and opened immediately. Selecting an existing project loads its profile and summary in parallel, then opens the current assessment workspace. A `Back to projects` action returns to the already loaded dashboard without deleting or resetting project data.

Opening failures remain on the dashboard and show a readable error. The action is disabled while a project is opening to prevent duplicate requests. Assessment save behavior is unchanged.

## Components and Boundaries

- `api/internal/store/projects.go` owns the project-list query and consistent project reads.
- `api/internal/httpapi/handler.go` exposes `GET /api/projects` through the existing handler.
- `web/src/lib/api.ts` adds the typed list request.
- `web/src/components/ProjectDashboard.tsx` renders list, empty state, create form, and dashboard errors. It receives data and callbacks; it does not call the API directly.
- `web/src/app/page.tsx` owns screen selection and coordinates API calls.

This extracts the dashboard from `page.tsx` while leaving assessment components unchanged.

## Testing

Development follows red-green-refactor:

- Go handler test: `GET /api/projects` returns the project list and the empty list contract.
- Store behavior is covered by the existing Docker smoke path using real PostgreSQL, extended to verify a created project appears in the list.
- Web component test: project details render, selecting a project calls the open callback, and the empty state is readable.
- Web page behavior is kept thin; API coordination is verified through type-check, build, and Docker smoke testing.

Final verification runs Go tests, web tests, web type-check and build, Docker Compose validation, and the project-list smoke test.

## Acceptance Criteria

1. Refreshing the home page shows every database-backed project newest first.
2. A user can open an existing project and see its saved profile and summary.
3. A user can return from the assessment to the project list.
4. A user can create a project from the dashboard and start assessing it immediately.
5. Empty, loading, and API failure states are visible and do not discard stored data.
6. No deferred feature or new infrastructure is introduced.
