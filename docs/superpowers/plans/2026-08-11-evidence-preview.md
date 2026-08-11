# Evidence Preview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add inline preview for PDF and image evidence while preserving authenticated downloads for every evidence file.

**Architecture:** Reuse the existing authenticated document download endpoint and return a browser object URL from the page-level API callback. Keep preview state in the project page so the response panel remains presentational. Render only MIME types the browser can display reliably; keep DOCX/XLSX download-only.

**Tech Stack:** Next.js, React, TypeScript, Vitest, Testing Library, Go HTTP API, Docker Compose.

## Global Constraints

- Do not add a document conversion service or a new client dependency.
- Do not change evidence permissions, storage, or database schema.
- Preserve the existing Download action for all file types.
- Follow TDD: write a failing test, verify RED, implement the smallest change, verify GREEN.

---

### Task 1: Add preview behavior to the response panel

**Files:**
- Modify: `web/src/components/StakeholderResponsePanel.tsx`
- Test: `web/src/components/StakeholderResponsePanel.test.tsx`

**Interfaces:**
- Consumes: `ResponseDocument.mimeType`, `onPreview(document: ResponseDocument)`, and `preview` state supplied by the project page.
- Produces: A `Preview` button for `application/pdf`, `image/png`, `image/jpeg`, and `image/jpg`; an inline preview region with close action; the existing download button remains unchanged.

- [ ] **Step 1: Write the failing tests**

  Add a supported document to the existing response fixture and assert that clicking `Preview` calls the preview callback. Add a supplied preview URL and assert that the panel renders an image/PDF preview and a `Close preview` button. Add an unsupported DOCX document and assert that it has no `Preview` button but still calls `onDownload` when its filename is clicked.

- [ ] **Step 2: Run the focused test to verify RED**

  Run from `web`:

  ```powershell
  npm.cmd test -- --run src/components/StakeholderResponsePanel.test.tsx
  ```

  Expected: the new preview assertions fail because the panel has no preview interface or rendering.

- [ ] **Step 3: Implement the minimal panel changes**

  Extend props with:

  ```ts
  onPreview: (document: ResponseDocument) => Promise<void>;
  preview?: { documentID: string; url: string; mimeType: string } | null;
  onClosePreview: () => void;
  ```

  Add a supported-MIME helper, render a `Preview` button only for supported files, and render `<img>` for image MIME types or an `<iframe title="Evidence preview">` for PDFs. Show the preview only for the selected document and keep the filename wired to `onDownload`.

- [ ] **Step 4: Run the focused test to verify GREEN**

  Run the same Vitest command and expect all response panel tests to pass.

- [ ] **Step 5: Commit the panel change**

  ```powershell
  git add web/src/components/StakeholderResponsePanel.tsx web/src/components/StakeholderResponsePanel.test.tsx
  git commit -m "feat: add evidence preview panel"
  ```

### Task 2: Fetch and clean up preview object URLs at the project page

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx`
- Modify: `web/src/components/ProjectAssessmentWorkspace.tsx`
- Modify: `web/src/components/ProfileEditor.tsx`
- Test: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`

**Interfaces:**
- Consumes: `api.downloadResponseDocument(projectID, subcategoryID, documentID)` returning a `Blob`.
- Produces: `previewEvidence(subcategoryID, document)` that creates an object URL, stores the selected preview, and revokes old URLs on replacement/unmount.

- [ ] **Step 1: Write the failing page test**

  Add a test response with a PDF document, trigger the response panel's Preview action, resolve the mocked Blob, and assert that `URL.createObjectURL` is called and the preview is rendered. Assert that closing the preview calls `URL.revokeObjectURL`.

- [ ] **Step 2: Run the focused page test to verify RED**

  Run:

  ```powershell
  npm.cmd test -- --run "src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx"
  ```

  Expected: the preview interaction fails because the page does not pass preview state or callbacks.

- [ ] **Step 3: Implement page-level preview state and prop plumbing**

  Add a nullable preview state containing `subcategoryID`, `documentID`, `url`, and `mimeType`. Fetch the blob with the existing API method, create an object URL, revoke the previous URL before replacement, and clear/revoke on close and component unmount. Pass `onPreview`, `preview`, and `onClosePreview` through `ProjectAssessmentWorkspace` to both `ProfileEditor` and `StakeholderResponsePanel` without changing permission checks.

- [ ] **Step 4: Run focused and full web tests**

  Run the focused page test, then:

  ```powershell
  npm.cmd test
  npm.cmd run typecheck -- --incremental false
  ```

  Expected: all existing and new tests pass with no TypeScript errors.

- [ ] **Step 5: Commit the page plumbing**

  ```powershell
  git add "web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx" web/src/components/ProjectAssessmentWorkspace.tsx web/src/components/ProfileEditor.tsx "web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx"
  git commit -m "feat: load evidence previews in project workspace"
  ```

### Task 3: Verify production behavior and document the feature

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Document preview support**

  Add a short note that PDF and image evidence can be previewed inline and DOCX/XLSX remain download-only.

- [ ] **Step 2: Run the final verification commands**

  ```powershell
  Set-Location api
  $env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'
  go test ./...
  Set-Location ..\web
  npm.cmd run build
  Set-Location ..
  docker compose up --build -d
  docker compose ps
  ```

  Expected: Go tests, TypeScript/build, and Docker services pass/healthy.

- [ ] **Step 3: Commit documentation**

  ```powershell
  git add README.md
  git commit -m "docs: describe evidence preview support"
  ```
