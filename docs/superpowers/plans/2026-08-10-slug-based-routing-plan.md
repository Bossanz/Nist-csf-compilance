# Slug-Based Routing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** แยกหน้า Next.js เป็น route จริง และใช้ slug ของ Organization/Project ใน browser URL โดยคง UUID และ API assessment เดิมไว้ภายในระบบ

**Architecture:** เพิ่ม slug ที่ backend เป็นข้อมูลถาวรของ Organization และ Project พร้อม unique scope และตัวสร้าง slug ที่รองรับ Unicode จากนั้นเพิ่ม lookup API ที่ resolve slug เป็น UUID ก่อนใช้ endpoint เดิม หน้า Next.js แต่ละ route จะโหลดข้อมูลของตัวเองและส่ง UUID ที่ได้ไปยัง API เดิมสำหรับ assessment

**Tech Stack:** Go, net/http, PostgreSQL, Next.js App Router 16, React 19, TypeScript, Vitest, Testing Library

## Global Constraints

- ห้ามใช้ UUID เป็น segment ใน browser URL
- UUID ยังคงเป็น primary key และ internal identifier
- Slug ต้องรองรับภาษาไทยและภาษาอังกฤษ
- Organization slug ไม่ซ้ำทั้งระบบ
- Project slug ไม่ซ้ำภายใน Organization เดียวกัน
- Slug สร้างครั้งเดียวและไม่เปลี่ยนตามการแก้ชื่อในอนาคต
- คง role authorization ฝั่ง Go และ API assessment เดิม
- ไม่เพิ่ม router หรือ state-management library ใหม่
- ห้าม reset, checkout หรือเขียนทับไฟล์ที่มีการแก้ไขค้างอยู่ก่อนเริ่มงาน

---

### Task 1: Add and test the slug generator

**Files:**
- Create: `api/internal/store/slugs.go`
- Create: `api/internal/store/slugs_test.go`

**Interfaces:**
- Produces `store.Slugify(input string) string`

- [ ] **Step 1: Write the failing tests**

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        name, input, want string
    }{
        {"english words", "Acme Corporation", "acme-corporation"},
        {"punctuation", " NIST / CSF 2.0 ", "nist-csf-2-0"},
        {"thai text", "บริษัท เอ บี ซี", "บริษัท-เอ-บี-ซี"},
        {"empty punctuation", "---", "item"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if got := Slugify(test.input); got != test.want {
                t.Fatalf("Slugify(%q) = %q, want %q", test.input, got, test.want)
            }
        })
    }
}
```

- [ ] **Step 2: Run the test and verify the expected failure**

Run: `go test ./internal/store -run TestSlugify -count=1`

Expected: FAIL because `Slugify` does not exist yet.

- [ ] **Step 3: Implement the minimal generator**

Implement `Slugify` with `unicode.IsLetter`, `unicode.IsDigit`, `unicode.ToLower`, separator collapsing, trimming, and `item` fallback. Keep Unicode letters so Thai names remain readable.

- [ ] **Step 4: Run the focused test**

Run: `go test ./internal/store -run TestSlugify -count=1`

Expected: PASS.

### Task 2: Persist slugs and backfill existing data

**Files:**
- Create: `db/init/005_slug_routing.sql`
- Modify: `api/internal/store/models.go`
- Modify: `api/internal/store/organizations.go`
- Modify: `api/internal/store/projects.go`
- Modify: `api/cmd/server/main.go`
- Modify: `api/internal/store/slugs.go`
- Create: `api/internal/store/slugs_integration_test.go`

**Interfaces:**
- Produces `Store.EnsureSlugs(context.Context) error`
- Produces `Store.GetOrganizationBySlug(context.Context, string) (Organization, error)`
- Produces `Store.GetProjectBySlug(context.Context, string, string) (Project, error)`
- Extends `store.Organization` and `store.Project` with JSON field `slug`

- [ ] **Step 1: Write the integration regression test**

Add a `TEST_DATABASE_URL`-guarded test that creates an organization and two same-base projects, calls the store create methods, and asserts slugs `acme-corporation`, `readiness`, and `readiness-2`. Add a second assertion that `GetOrganizationBySlug` and `GetProjectBySlug` return the created rows.

- [ ] **Step 2: Run the integration test before implementation**

Run: `$env:TEST_DATABASE_URL='postgres://compliance:compliance@localhost:5432/compliance?sslmode=disable'; go test ./internal/store -run 'TestSlugPersistence|TestSlugLookup' -count=1`

Expected: FAIL because the schema and store methods do not exist yet. If PostgreSQL is unavailable, record the environment failure and continue with unit/API tests; do not delete the database volume.

- [ ] **Step 3: Add the idempotent schema migration**

Create `005_slug_routing.sql` with nullable `slug` columns and partial unique indexes so it can be applied to a database that already contains rows:

```sql
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS slug text;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS slug text;
CREATE UNIQUE INDEX IF NOT EXISTS organizations_slug_unique ON organizations(slug) WHERE slug IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS projects_organization_slug_unique ON projects(organization_id, slug) WHERE slug IS NOT NULL;
```

Also make `EnsureSlugs` run the same idempotent DDL at API startup so an existing Docker volume receives the columns even though Docker init scripts are only run for a new volume.

- [ ] **Step 4: Implement persistence and collision handling**

Add `slug` to all organization/project SELECT and `RETURNING` statements. Generate a base with `Slugify`, then query the relevant scope and append `-2`, `-3`, and so on until the candidate is free. Create organization/project rows with the chosen slug in the same transaction as the insert.

- [ ] **Step 5: Implement startup backfill**

`EnsureSlugs` must fill null slugs for existing organizations and projects in deterministic ID order, using the same collision suffix rules. Call it after `store.New` and before bootstrap/server startup. Return startup errors instead of serving a partially migrated API.

- [ ] **Step 6: Run store tests**

Run: `go test ./internal/store -count=1`

Expected: unit tests pass; integration tests pass when `TEST_DATABASE_URL` is available and otherwise skip only those guarded tests.

### Task 3: Add authenticated slug lookup endpoints

**Files:**
- Modify: `api/internal/httpapi/organizations_handler.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/organizations_handler_test.go`
- Modify: `api/internal/httpapi/handler_test.go`

**Interfaces:**
- Adds `GET /api/organizations/by-slug/:organizationSlug`
- Adds `GET /api/organizations/:organizationID/projects/by-slug/:projectSlug`

- [ ] **Step 1: Write failing handler tests**

Test that an authenticated counselor can resolve an organization slug and a scoped project slug, that the response includes `id`, `name`, and `slug`, and that a stakeholder cannot resolve an organization outside its organization. Use the existing `fakeStore` pattern.

- [ ] **Step 2: Run focused handler tests**

Run: `go test ./internal/httpapi -run 'Test.*Slug' -count=1`

Expected: FAIL because the fake-store interfaces and routes do not exist yet.

- [ ] **Step 3: Implement the handlers and routes**

Extend `organizationDataStore`, add authorization through the existing `authorizeOrganization`, unescape the path segment, and return the same not-found response for missing or unauthorized slugs. Register lookup routes before the existing ID routes to avoid treating `by-slug` as an ID.

- [ ] **Step 4: Run handler tests and all Go tests**

Run: `go test ./internal/httpapi ./internal/store -count=1`

Expected: PASS with no authorization regressions.

### Task 4: Extend frontend API types and route helpers

**Files:**
- Modify: `web/src/lib/types.ts`
- Modify: `web/src/lib/api.ts`
- Create: `web/src/lib/routes.ts`
- Modify: `web/src/lib/api.test.ts`
- Create: `web/src/lib/routes.test.ts`

**Interfaces:**
- Adds `Organization.slug` and `Project.slug`
- Adds `api.getOrganizationBySlug(slug)`
- Adds `api.getOrganizationProjectBySlug(organizationID, slug)`
- Produces `organizationPath(organization)` and `projectPath(organization, project)`

- [ ] **Step 1: Write failing route helper tests**

Assert that English and Thai slugs are URI-encoded as path segments and that the helper returns `/organizations/<slug>` and `/organizations/<slug>/projects/<slug>` without any UUID.

- [ ] **Step 2: Run focused frontend tests**

Run: `npm test -- --run src/lib/routes.test.ts`

Expected: FAIL because the helper does not exist.

- [ ] **Step 3: Implement API methods and helpers**

Use `encodeURIComponent` only for URL path segments. Keep all existing UUID-based API methods unchanged. Add API request tests for the two lookup paths using the existing fetch mocking pattern.

- [ ] **Step 4: Run frontend library tests and typecheck**

Run: `npm test -- --run src/lib`; `npm run typecheck`

Expected: PASS after updating fixtures with `slug` values.

### Task 5: Split the Next.js UI into route pages

**Files:**
- Modify: `web/src/app/page.tsx`
- Create: `web/src/app/login/page.tsx`
- Create: `web/src/app/organizations/page.tsx`
- Create: `web/src/app/organizations/[organizationSlug]/page.tsx`
- Create: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx`
- Create: `web/src/components/ProjectAssessmentWorkspace.tsx`
- Modify: `web/src/app/page.test.tsx`
- Create: `web/src/app/login/page.test.tsx`
- Create: `web/src/app/organizations/page.test.tsx`
- Create: `web/src/app/organizations/[organizationSlug]/page.test.tsx`
- Create: `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.test.tsx`
- Modify: `web/src/components/OrganizationDashboard.test.tsx`
- Modify: `web/src/components/OrganizationWorkspace.test.tsx`

**Interfaces:**
- `/` redirects to `/organizations`
- Each page owns its own session check, loading state, API loading, and router navigation
- Existing presentational components remain reusable and keep their role behavior

- [ ] **Step 1: Write failing page tests**

Cover these behaviors:

```text
login success -> router.replace("/organizations")
organization list -> router.push("/organizations/acme-corporation")
organization workspace -> router.push("/organizations/acme-corporation/projects/readiness")
project route -> loads profile/summary/responses after slug lookup
401 -> router.replace("/login")
```

- [ ] **Step 2: Run the new page tests**

Run: `npm test -- --run src/app`

Expected: FAIL because the route pages and navigation do not exist.

- [ ] **Step 3: Implement the root and login routes**

Make `web/src/app/page.tsx` a redirect-only page. Move login rendering and login redirect into `/login`, preserving current error/loading copy.

- [ ] **Step 4: Implement the organization list route**

Move organization loading/create/delete/logout behavior into `/organizations`. Use `organizationPath` for open navigation and preserve Counselor Admin delete behavior.

- [ ] **Step 5: Implement the organization workspace route**

Resolve the organization by slug, load projects/users by resolved UUID, and use `projectPath` on open. Preserve project creation/deletion and invitation behavior. Use a route link back to `/organizations`.

- [ ] **Step 6: Extract and implement the project assessment route**

Move the current project branch into `ProjectAssessmentWorkspace`. Resolve organization and project by slug, load functions/profile/summary/responses, and preserve all Counselor/Stakeholder read/write/review/evidence behavior. The back action must navigate to the organization slug route.

- [ ] **Step 7: Update fixtures and run page tests**

Add slugs to all Organization/Project fixtures and run: `npm test -- --run src/app src/components`

Expected: PASS.

### Task 6: Full verification and handoff

**Files:**
- Modify: `README.md` only if the route table needs documenting

- [ ] **Step 1: Run the complete frontend verification**

Run: `npm test`; `npm run typecheck`; `npm run build`

Expected: all tests pass, TypeScript exits 0, and Next production build exits 0.

- [ ] **Step 2: Run the complete Go verification**

Run: `go test ./...`

Expected: all Go tests pass; integration tests skip only when `TEST_DATABASE_URL` is not configured.

- [ ] **Step 3: Inspect the final diff**

Run: `git diff --check`; `git status --short`

Confirm no unrelated user changes were staged or overwritten, and confirm route files contain no UUID segments in generated browser paths.

- [ ] **Step 4: Report verification evidence and changed files**

Report exact test/build results and the route URLs. Do not claim completion without fresh command output.
