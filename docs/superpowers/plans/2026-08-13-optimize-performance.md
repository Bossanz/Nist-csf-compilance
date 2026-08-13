# Assessment Performance Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce unnecessary waiting and repeated work when loading and rendering assessment workspaces without changing role permissions or assessment behavior.

**Architecture:** Keep the existing API and component boundaries. Start independent authentication and organization requests together, replace repeated response scans with one lookup map, and let the browser skip layout/paint work for off-screen assessment cards. No new package, cache, or virtualization layer is introduced.

**Tech Stack:** Next.js 16, React 19, TypeScript, Vitest, CSS.

## Global Constraints

- Preserve Counselor/Stakeholder permissions and all existing API semantics.
- Do not add runtime dependencies.
- Keep the current reader-friendly white editorial layout.
- Verify with focused tests, the full test suite, typecheck, build, and `git diff --check`.

---

### Task 1: Remove the organization-page request waterfall

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/page.tsx`
- Test: `web/src/app/organizations/[organizationSlug]/page.test.tsx`

**Interfaces:**
- Consumes: existing `api.me()` and `api.getOrganizationBySlug()` calls.
- Produces: identical loaded organization state, with both independent requests started in the same effect tick.

- [x] **Step 1: Write the failing test**

Add a test with deferred `api.me()` and `api.getOrganizationBySlug()` promises. Assert that both mocks have been called before either deferred promise resolves; keep the existing resolved organization assertions unchanged.

- [x] **Step 2: Run the focused test to verify it fails**

Run: `npm.cmd test -- --run src/app/organizations/[organizationSlug]/page.test.tsx`

Expected: the new concurrency assertion fails because the current implementation awaits `api.me()` before calling `api.getOrganizationBySlug()`.

- [x] **Step 3: Write the minimal implementation**

Start both promises together and await them together:

```ts
const [currentUser, nextOrganization] = await Promise.all([
  api.me(),
  api.getOrganizationBySlug(organizationSlug),
]);
```

Keep the existing project/member request and error handling unchanged.

- [x] **Step 4: Run the focused test to verify it passes**

Run: `npm.cmd test -- --run src/app/organizations/[organizationSlug]/page.test.tsx`

Expected: PASS, including the new concurrency assertion.

### Task 2: Remove the project-page request waterfall

**Files:**
- Modify: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx`
- Test: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`

**Interfaces:**
- Consumes: existing `api.me()` and `api.getOrganizationBySlug()` calls.
- Produces: identical project assessment state, with authentication and organization lookup started together.

- [x] **Step 1: Write the failing test**

Add a deferred-promise test that proves `api.me()` and `api.getOrganizationBySlug()` start before either resolves. Leave project-dependent calls behind the organization result.

- [x] **Step 2: Run the focused test to verify it fails**

Run: `npm.cmd test -- --run src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`

Expected: the concurrency assertion fails against the sequential implementation.

- [x] **Step 3: Write the minimal implementation**

Replace only the independent first two awaits with `Promise.all`, then continue using `nextOrganization.id` for project-dependent requests.

- [x] **Step 4: Run the focused test to verify it passes**

Run: `npm.cmd test -- --run src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`

Expected: PASS.

### Task 3: Avoid repeated response scans and off-screen card work

**Files:**
- Modify: `web/src/components/ProfileEditor.tsx`
- Modify: `web/src/app/globals.css`
- Test: `web/src/components/ProfileEditor.test.tsx`

**Interfaces:**
- Consumes: existing `rows` and `responses` props.
- Produces: the same response passed to each `AssessmentCard`, with one response lookup map per render and browser-native containment for the long list.

- [x] **Step 1: Write the failing test**

Add a behavior test rendering multiple rows whose responses are intentionally out of order. Assert each outcome receives the matching response text, including a row without a response receiving an empty draft. This locks the lookup contract before changing the implementation.

- [x] **Step 2: Run the focused test to verify it fails**

Run: `npm.cmd test -- --run src/components/ProfileEditor.test.tsx`

Expected: the new test fails to resolve because the test file does not yet exist.

- [x] **Step 3: Write the minimal implementation**

Build `responseBySubcategoryID` with `useMemo` from `responses`, look up each row with `.get()`, and retain `emptyResponse(row)` as the fallback. Add `content-visibility: auto` and a conservative `contain-intrinsic-size` to `.assessment-card`; do not apply paint containment or change card semantics.

- [x] **Step 4: Run the focused test to verify it passes**

Run: `npm.cmd test -- --run src/components/ProfileEditor.test.tsx`

Expected: PASS with the same response text/status behavior.

### Task 4: Full verification

**Files:**
- Verify: all modified files above.

- [x] **Step 1: Run the complete frontend test suite**

Run: `npm.cmd test -- --run`

Expected: all test files pass.

- [x] **Step 2: Run typecheck and production build**

Run: `npx.cmd tsc --noEmit --incremental false` and `npm.cmd run build` from `web`.

Expected: both commands exit successfully and the build reports the existing routes.

- [x] **Step 3: Check the diff for whitespace errors**

Run: `git diff --check`

Expected: no whitespace errors.
