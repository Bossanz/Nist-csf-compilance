# NIST CSF Compliance Web App Implementation Plan

> For agentic workers: use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

Goal: Build a Docker Compose vertical slice with Next.js, Go REST API, PostgreSQL schema/seed, project profile editing, and coverage summaries.

Architecture: A lean monorepo contains web/, api/, and db/. Next.js calls Go over REST; Go owns validation, persistence, and calculations; PostgreSQL stores the seeded CSF catalog and project data.

Tech Stack: Next.js, TypeScript, Go standard library HTTP server, PostgreSQL, Docker Compose, pgx driver, Vitest/React Testing Library, Go testing.

## Global Constraints

- Run with docker compose up --build.
- Seed exactly 6 Functions, 22 Categories, and 106 Subcategories from the supplied workbook.
- Coverage levels are none, partial, substantial, full with scores 0, 1, 2, 3.
- Excluded rows do not affect denominators or pending counts.
- Use TDD: failing test first, observe failure, minimal implementation, then green test.
- Do not add authentication, email, object storage, PDF, or full action-plan features.
- Do not commit secrets or rely on a git commit because this workspace has no root git repository.

### Task 1: Root project and Docker orchestration

Files:
- Create docker-compose.yml, .env.example, .gitignore, README.md
- Create api/Dockerfile and web/Dockerfile
- Create db/init/001_schema.sql and db/init/002_seed.sql
- Create scripts/verify-compose.ps1

Interfaces:
- web on port 3000, api on port 8080, postgres on port 5432
- DATABASE_URL for API and NEXT_PUBLIC_API_BASE_URL for web
- PostgreSQL init scripts run in lexical order

Steps:
- [ ] Write verify-compose.ps1 to assert web/api/postgres services. Run it before files exist; expect failure because docker-compose.yml is missing.
- [ ] Add compose services, healthchecks, API dependency on PostgreSQL health, web dependency on API health, and non-secret example defaults.
- [ ] Run pwsh -File scripts/verify-compose.ps1; expect pass.
- [ ] Run docker compose config; expect exit 0.
- [ ] Document startup, URLs, and shutdown in README.md.

### Task 2: PostgreSQL schema and CSF seed

Files:
- Modify db/init/001_schema.sql and db/init/002_seed.sql
- Create scripts/verify-seed.ps1
- Reference the supplied workbook as the source hierarchy

Interfaces:
- Tables: functions, categories, subcategories, organizations, users, projects, project_functions, project_subcategory_profiles.
- Unique codes and project joins.
- Seed counts: 6, 22, 106.

Steps:
- [ ] Write verify-seed.ps1 to query counts before seed exists; run and observe expected failure.
- [ ] Add UUID keys, foreign keys, unique constraints, indexes, timestamps, and check constraints for status and coverage enums.
- [ ] Extract deterministic UTF-8 INSERT data from the workbook hierarchy.
- [ ] Start PostgreSQL and run the verification; expect 6 Functions, 22 Categories, 106 Subcategories.
- [ ] Verify unique constraints for project/function and project/subcategory.

### Task 3: Go calculation domain with TDD

Files:
- Create api/go.mod
- Create api/internal/domain/profile.go and calculation.go
- Test api/internal/domain/calculation_test.go

Interfaces:
- CoverageLevel, Score(level) (int, error)
- ProfileScore with Included, Current, Target
- Summary with CoveragePct, IncludedCount, PendingCount, RejectedCount
- CalculateSummary(rows []ProfileScore) Summary

Steps:
- [ ] Write failing tests for none=0, partial=1, substantial=2, full=3 and invalid enum.
- [ ] Run cd api; go test ./internal/domain -v; expect failure because implementation is missing.
- [ ] Implement the minimal mapping and typed invalid-level error.
- [ ] Write failing summary tests for included full/partial, excluded rows, and zero included rows.
- [ ] Implement average current score divided by 3, excluding rows where Included is false.
- [ ] Run cd api; go test ./internal/domain -v; expect pass.

### Task 4: PostgreSQL repository and project creation

Files:
- Create api/internal/store/postgres.go, catalog.go, projects.go
- Test api/internal/store/store_test.go
- Modify api/go.mod

Interfaces:
- Store with DB pool
- ListFunctions(ctx)
- CreateProject(ctx, name, organizationName)
- GetProject(ctx, id)
- ListProfile(ctx, projectID)
- UpdateProfile(ctx, projectID, subcategoryID, patch)

Steps:
- [ ] Write integration tests using TEST_DATABASE_URL; assert project creation yields one organization, one project, 6 project-function rows, and 106 profile rows.
- [ ] Run the tests before methods exist; expect failure. Skip only when TEST_DATABASE_URL is absent.
- [ ] Implement bounded connection ping, close, catalog query ordering, and transactional project creation.
- [ ] Implement hierarchy-aware profile reads and partial updates with enum validation.
- [ ] Run TEST_DATABASE_URL=... go test ./internal/store -v; expect pass.

### Task 5: Go HTTP API and smoke tests

Files:
- Create api/cmd/server/main.go
- Create api/internal/httpapi/handler.go, json.go, errors.go
- Test api/internal/httpapi/handler_test.go

Interfaces:
- GET /healthz
- GET /api/functions
- POST /api/projects
- GET /api/projects/{id}
- GET /api/projects/{id}/profile
- PUT /api/projects/{id}/profile/{subcategoryId}
- GET /api/projects/{id}/summary

Steps:
- [ ] Write failing handler tests for health 200 JSON and empty project name 400 validation_error.
- [ ] Run cd api; go test ./internal/httpapi -v; expect failure.
- [ ] Implement JSON helpers, one-body decoding, safe path parsing, and documented error mapping.
- [ ] Implement handlers with request contexts and 201 for project creation.
- [ ] Run cd api; go test ./...; expect pass.
- [ ] Build the API container and curl /healthz; expect HTTP 200 with status ok.

### Task 6: Next.js web shell and API client

Files:
- Create web/package.json, tsconfig.json, next.config.ts
- Create web/src/app/layout.tsx, page.tsx, globals.css
- Create web/src/lib/api.ts, types.ts
- Test web/src/lib/api.test.ts

Interfaces:
- api.getFunctions()
- api.createProject(input)
- api.getProject(id)
- api.getProfile(id)
- api.updateProfile(projectId, subcategoryId, patch)
- api.getSummary(id)

Steps:
- [ ] Write failing client tests for POST JSON and non-2xx error parsing.
- [ ] Run npm test for the client; expect failure because the client is missing.
- [ ] Implement typed fetch functions using NEXT_PUBLIC_API_BASE_URL.
- [ ] Add a project creation screen and loading/error/empty states.
- [ ] Run npm test and npm run typecheck; expect pass.

### Task 7: Profile editor and coverage dashboard

Files:
- Create web/src/components/FunctionSidebar.tsx
- Create web/src/components/SummaryCards.tsx
- Create web/src/components/ProfileEditor.tsx
- Test SummaryCards.test.tsx and ProfileEditor.test.tsx
- Modify web/src/app/page.tsx

Interfaces:
- FunctionSidebar({ functions, selectedCode, onSelect })
- SummaryCards({ summary })
- ProfileEditor({ rows, onSave })

Steps:
- [ ] Write failing component tests for coverage/count cards and current/target controls.
- [ ] Run targeted component tests; expect failure because components are missing.
- [ ] Implement native controls for included, current level, target level, notes, and save.
- [ ] Preserve local values during saving/errors.
- [ ] Wire project/profile/summary loading, Function filtering, row update, and summary refresh.
- [ ] Run npm test and npm run build; expect pass.

### Task 8: Full Docker verification and handoff

Files:
- Create scripts/smoke-test.ps1
- Modify README.md
- Modify docker-compose.yml only if verification reveals a configuration issue

Interfaces:
- smoke test verifies health, catalog, project creation, 106 profile rows, one profile update, and summary JSON.

Steps:
- [ ] Write smoke-test.ps1 before the stack is running; run and observe expected failure.
- [ ] Run docker compose down -v followed by docker compose up --build -d.
- [ ] Run smoke-test.ps1; expect all endpoint checks to pass.
- [ ] Run cd api; go test ./..., web npm test, npm run typecheck, npm run build, and docker compose config.
- [ ] Update README.md with only the verified workflow and explicit v1 non-goals.

## Plan Self-Review

Spec coverage: architecture, Docker, catalog seed, schema, API contract, calculations, UI, error handling, tests, and future boundaries are covered by Tasks 1–8.

Placeholder scan: no TBD, TODO, or unspecified “add appropriate” steps remain.

Type consistency: domain types feed store types; store methods feed HTTP handlers; API client types match endpoint responses; UI components consume the profile and summary shapes.

Verification: final completion claims require fresh Go, web, Docker Compose, and smoke-test output.

