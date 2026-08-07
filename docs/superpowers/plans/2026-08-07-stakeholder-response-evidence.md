# Stakeholder Response and Evidence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let Organization Admins and Assessors answer existing NIST Subcategories and attach evidence while limiting Priority/Coverage assessment to Counselors and adding Reviewer workflow.

**Architecture:** Add one response row per Project/Subcategory with a small lifecycle state machine. Keep document metadata in PostgreSQL and document bytes in a named Docker volume. Expose separate response/review APIs and render three role-aware regions inside each outcome card.

**Tech Stack:** Go 1.24 standard library, pgx v5, PostgreSQL 16, Next.js 16, React 19, TypeScript, Vitest, Testing Library, Docker Compose.

## Global Constraints

- Version 1 uses each existing NIST CSF Subcategory as the question.
- Only Counselor Admin and Counselor edit Counselor assessment fields.
- Organization Admin and Assessor edit response text and documents only while the response is `draft` or `needs_more_info`.
- Reviewer changes only `reviewed`/`needs_more_info` and the review comment.
- Viewer is read-only.
- Maximum file size is 20 MB; allowed formats are PDF, DOCX, XLSX, PNG, and JPEG.
- Do not add a custom question bank, assignments, deadlines, notifications, document versioning, antivirus infrastructure, or cloud object storage.
- Do not add a new npm or Go dependency.
- All project and organization access checks remain server-side.

---

### Task 1: Response Schema and Lifecycle Domain

**Files:** Create `db/init/004_stakeholder_responses.sql`, `api/internal/domain/response.go`, `api/internal/domain/response_test.go`; modify `api/internal/store/models.go`.

**Produces:** `domain.ResponseStatus`, `domain.CanTransitionResponse`, `stakeholder_responses`, and `response_documents`.

- [ ] Write a table-driven failing test with literal expectations: `draft → submitted`, `needs_more_info → submitted`, `submitted → reviewed`, and `submitted → needs_more_info` are valid; all other transitions are invalid.
- [ ] Run `cd api && go test ./internal/domain -run TestCanTransitionResponse -count=1`; verify RED because the types do not exist.
- [ ] Define the four statuses and one explicit transition function. Add complete `StakeholderResponse` and `ResponseDocument` store structs with JSON tags.
- [ ] Create an idempotent migration. `stakeholder_responses` has unique `(project_id, subcategory_id)`, response text, status check, actors/timestamps, and Project cascade. `response_documents` has response cascade, unique storage key, original name, MIME, byte size, uploader, and timestamp.
- [ ] Run `gofmt`, domain tests, and `docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/004_stakeholder_responses.sql`.
- [ ] Commit: `git add db/init/004_stakeholder_responses.sql api/internal/domain api/internal/store/models.go; git commit -m "feat: add stakeholder response lifecycle"`.

### Task 2: Response Persistence, Lifecycle API, and Field-Level RBAC

**Files:** Create `api/internal/store/responses.go`, `api/internal/store/responses_integration_test.go`, `api/internal/httpapi/responses_handler.go`, and `api/internal/httpapi/responses_handler_test.go`; modify `api/internal/httpapi/handler.go`, `authorization.go`, and tests.

**Produces:** `ListResponses`, `SaveResponseDraft`, `SubmitResponse`, `ReviewResponse` and:

- `GET /api/projects/{projectID}/responses`
- `PUT /api/projects/{projectID}/responses/{subcategoryID}`
- `POST /api/projects/{projectID}/responses/{subcategoryID}/submit`
- `POST /api/projects/{projectID}/responses/{subcategoryID}/review`

- [ ] Write failing HTTP tests for Organization Admin/Assessor save, Reviewer review, Viewer mutation rejection, invalid lifecycle `409`, and cross-organization isolation.
- [ ] Run `cd api && go test ./internal/httpapi -run 'Response|Assessor|Reviewer|Viewer' -count=1`; verify RED.
- [ ] Implement `ListResponses` with a left join to NIST Subcategories. Implement draft upsert with `ON CONFLICT`, allowing edits only in `draft`/`needs_more_info`.
- [ ] Implement submit/review transactions using `domain.CanTransitionResponse`, actor IDs, timestamps, and `pgx.ErrNoRows` for missing data.
- [ ] Make each handler load Project → Organization before authorizing. Return `403` forbidden, `404` missing, `409` invalid transition, and `500` storage errors.
- [ ] Change existing profile update authorization so only `counselor_admin` and `counselor` can update Priority/Coverage and other Counselor assessment fields. Add an Assessor `403` test.
- [ ] Run with `TEST_DATABASE_URL=postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable`: `cd api && go test ./...`.
- [ ] Commit: `git add api/internal/store/responses.go api/internal/store/responses_integration_test.go api/internal/httpapi; git commit -m "feat: add response review API"`.

### Task 3: Local Evidence Storage and Document API

**Files:** Create `api/internal/evidence/files.go`, `files_test.go`, `api/internal/store/documents.go`, `api/internal/httpapi/documents_handler.go`, and `documents_handler_test.go`; modify `handler.go`.

**Produces:** `evidence.FileStore{Save, Open, Delete, DeleteKeys}` and:

- `POST /api/projects/{projectID}/responses/{subcategoryID}/documents`
- `GET /api/projects/{projectID}/responses/{subcategoryID}/documents/{documentID}`
- `DELETE /api/projects/{projectID}/responses/{subcategoryID}/documents/{documentID}`

- [ ] Write failing `t.TempDir()` tests for allowed PDF/DOCX/XLSX/PNG/JPEG, unsupported types, 20 MB boundary, opaque storage key, safe display name, and cleanup after metadata failure.
- [ ] Run `cd api && go test ./internal/evidence -run TestSave -count=1`; verify RED because the package is absent.
- [ ] Implement with standard library only: `crypto/rand` 32-byte keys, `io.LimitReader`, `os.CreateTemp`, `os.Rename`, extension/MIME allow lists, and typed errors `ErrFileTooLarge`, `ErrUnsupportedType`, and `ErrInvalidName`. Never use the original filename as a path.
- [ ] Save binary first and metadata second; remove the binary on metadata failure. Verify Project/Subcategory/Organization on every request. Set safe attachment headers and `X-Content-Type-Options: nosniff` on download.
- [ ] Test Organization Admin/Assessor upload/delete only in draft or needs-more-info; Reviewer/Viewer download only; Counselor download only; reviewed responses reject mutation with `409`.
- [ ] Run `cd api && gofmt -w internal/evidence internal/store/documents.go internal/httpapi/documents_handler.go && go test ./...`.
- [ ] Commit: `git add api/internal/evidence api/internal/store/documents.go api/internal/httpapi; git commit -m "feat: add local evidence uploads"`.

### Task 4: Evidence Volume and Deletion Cleanup

**Files:** Modify `api/cmd/server/main.go`, `api/internal/store/projects.go`, `api/internal/store/organizations.go`, and `api/internal/httpapi/handler.go`; create `api/internal/store/document_cleanup_integration_test.go`; modify `docker-compose.yml`, `.env.example`, and `README.md`.

- [ ] Write a failing integration test that creates document metadata and a matching temp file, deletes Project/Organization, and asserts database metadata and file are gone; missing binary cleanup must be idempotent.
- [ ] Run with `TEST_DATABASE_URL` and `TEST_EVIDENCE_DIR`; verify RED because current delete paths do not know storage keys.
- [ ] Before deleting Project/Organization rows, collect document storage keys. Delete business rows transactionally, then remove binaries and log failed cleanup keys without restoring deleted business data.
- [ ] Create `EVIDENCE_DIR` on API startup. Mount the named volume exactly as follows:

```yaml
api:
  environment:
    EVIDENCE_DIR: /data/evidence
  volumes:
    - evidence_data:/data/evidence
volumes:
  evidence_data:
```

- [ ] Document `EVIDENCE_DIR=/data/evidence` and that `docker compose down -v` removes uploaded evidence.
- [ ] Run `cd api && go test ./...`, `docker compose config`, `docker compose up --build -d`, and `docker compose ps`; expect all services healthy.
- [ ] Commit the runtime/configuration changes with `git add api docker-compose.yml .env.example README.md; git commit -m "feat: persist evidence in Docker volume"`.

### Task 5: Role-Specific Frontend Regions

**Files:** Create `web/src/components/StakeholderResponsePanel.tsx`, `ReviewPanel.tsx`, and their tests; modify `AssessmentCard.tsx`, `ProfileEditor.tsx`, `web/src/lib/types.ts`, and `web/src/lib/api.ts`.

- [ ] Write failing component tests: Assessor edits response but sees disabled Priority/Coverage; Reviewer sees actions only for submitted; Viewer sees no mutation controls; failed upload preserves typed response text.
- [ ] Run `cd web && npm.cmd test -- src/components/StakeholderResponsePanel.test.tsx src/components/ReviewPanel.test.tsx`; verify RED.
- [ ] Add complete response/document types and client methods `getResponses`, `saveResponse`, `submitResponse`, `reviewResponse`, `uploadDocument`, `downloadDocument`, and `deleteDocument`. Omit JSON `Content-Type` when request body is `FormData`.
- [ ] Build the response region with labeled textarea, draft/submit controls, status, document list, file input, upload, and delete. Allow edits only for draft/needs-more-info and Organization Admin/Assessor.
- [ ] Build the review region with comment and Reviewed/Need more info only for Reviewer with submitted status. Show status/comments/documents to authorized readers.
- [ ] Set existing ProfileEditor read-only for every role except `counselor_admin` and `counselor`; preserve Counselor save behavior.
- [ ] Run new component tests plus `ProfileEditor.test.tsx`; commit frontend region/API changes.

### Task 6: Project Integration and End-to-End Verification

**Files:** Modify `web/src/app/page.tsx`, `web/src/app/page.test.tsx`, `web/src/app/globals.css`, and `README.md`.

- [ ] Write failing page tests for response loading, Assessor disabled Priority/Coverage, Counselor editable Priority/Coverage, Reviewer action, and Viewer read-only state.
- [ ] Run `cd web && npm.cmd test -- src/app/page.test.tsx`; verify RED because Home does not load responses.
- [ ] Load profile, summary, and responses in parallel when opening a Project. Pass the response matching each outcome and keep response/document/review errors scoped to that outcome.
- [ ] Render separate `Client response and documents`, `Review`, and `Counselor assessment` regions with labels, focus states, loading/empty/error states, and responsive layout using existing tokens.
- [ ] Document `Stakeholder answers → Reviewer checks → Counselor sets Priority/Coverage`, role permissions, migration command, file limits, and evidence volume.
- [ ] Run full verification:

```powershell
cd api
$env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
go test ./...
cd ..\web
npx.cmd tsc --noEmit --incremental false
npm.cmd test
npm.cmd run build
cd ..
docker compose up --build -d
docker compose ps
```

- [ ] Browser-test Assessor answer/upload/submit, Reviewer request-more-info/review, Assessor resubmit, Counselor-only Priority/Coverage editing, Viewer read-only access, persistence after web restart, and zero unhandled console errors.
- [ ] Commit integration: `git add web/src/app/page.tsx web/src/app/page.test.tsx web/src/app/globals.css README.md; git commit -m "feat: integrate stakeholder response workflow"`.
