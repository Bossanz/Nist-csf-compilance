# Project Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show persisted projects on the home page and let users reopen an assessment after refresh.

**Architecture:** Add one read-only project-list method and REST route to the existing Go API. Keep screen state in `page.tsx`, and move dashboard rendering into one presentational React component. Reuse the existing PostgreSQL schema, API client, and assessment workspace.

**Tech Stack:** Go standard-library HTTP, pgx/PostgreSQL, Next.js 16, React 19, TypeScript, Vitest, Testing Library, Docker Compose.

## Global Constraints

- Do not add pagination, search, deletion, authentication, browser persistence, a state library, a schema migration, or a new route.
- `GET /api/projects` returns newest projects first and an empty JSON array when none exist.
- Project responses consistently include `organizationName`.
- Use red-green-refactor for every behavior change.

---

### Task 1: Project List API

**Files:**
- Modify: `api/internal/store/models.go`
- Modify: `api/internal/store/projects.go`
- Modify: `api/internal/httpapi/handler.go`
- Modify: `api/internal/httpapi/handler_test.go`

**Interfaces:**
- Produces: `Store.ListProjects(context.Context) ([]Project, error)`
- Produces: `GET /api/projects -> store.Project[]`
- Changes: `Project` gains `OrganizationName string` serialized as `organizationName`

- [ ] **Step 1: Introduce a test seam and write the failing route tests**

Change `Handler.Store` to this package-local interface:

```go
type dataStore interface {
    ListFunctions(context.Context) ([]store.Function, error)
    CreateProject(context.Context, string, string) (store.Project, error)
    ListProjects(context.Context) ([]store.Project, error)
    GetProject(context.Context, string) (store.Project, error)
    ListProfile(context.Context, string) ([]store.ProfileRow, error)
    UpdateProfile(context.Context, string, string, store.ProfilePatch) (store.ProfileRow, error)
}
```

Add this minimal fake to `handler_test.go`:

```go
type fakeStore struct {
    projects []store.Project
    listErr  error
}

func (f fakeStore) ListProjects(context.Context) ([]store.Project, error) { return f.projects, f.listErr }
func (f fakeStore) ListFunctions(context.Context) ([]store.Function, error) { return nil, nil }
func (f fakeStore) CreateProject(context.Context, string, string) (store.Project, error) { return store.Project{}, nil }
func (f fakeStore) GetProject(context.Context, string) (store.Project, error) { return store.Project{}, nil }
func (f fakeStore) ListProfile(context.Context, string) ([]store.ProfileRow, error) { return nil, nil }
func (f fakeStore) UpdateProfile(context.Context, string, string, store.ProfilePatch) (store.ProfileRow, error) {
    return store.ProfileRow{}, nil
}
```

Add these tests:

```go
func TestListProjects(t *testing.T) {
    expected := []store.Project{{ID: "project-1", Name: "Readiness", OrganizationName: "Acme", Status: "setup"}}
    r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
    w := httptest.NewRecorder()

    (&Handler{Store: fakeStore{projects: expected}}).ServeHTTP(w, r)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
    var got []store.Project
    if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
        t.Fatal(err)
    }
    if !reflect.DeepEqual(got, expected) {
        t.Fatalf("expected %#v, got %#v", expected, got)
    }
}

func TestListProjectsReturnsEmptyArray(t *testing.T) {
    r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
    w := httptest.NewRecorder()

    (&Handler{Store: fakeStore{projects: []store.Project{}}}).ServeHTTP(w, r)

    if w.Code != http.StatusOK || w.Body.String() != "[]\n" {
        t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
    }
}

func TestListProjectsHandlesStoreFailure(t *testing.T) {
    r := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
    w := httptest.NewRecorder()

    (&Handler{Store: fakeStore{listErr: errors.New("database unavailable")}}).ServeHTTP(w, r)

    if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"internal_error"`) {
        t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
    }
}
```

- [ ] **Step 2: Run the API tests and observe the expected failure**

Run: `cd api; go test ./internal/httpapi -run ListProjects -v`

Expected: compilation fails because `Project.OrganizationName`, `fakeStore`, and the `GET /api/projects` behavior do not exist.

- [ ] **Step 3: Implement the minimal store and route behavior**

Add `OrganizationName` to `store.Project`. Implement a shared row scanner query shape for create/get/list so all project responses include it. `ListProjects` executes:

```sql
SELECT p.id, p.organization_id, o.name, p.name, p.status, p.created_at::text
FROM projects p
JOIN organizations o ON o.id = p.organization_id
ORDER BY p.created_at DESC, p.id DESC
```

Add this route before the existing `POST /api/projects` route:

```go
if len(parts) == 2 && parts[0] == "api" && parts[1] == "projects" && r.Method == http.MethodGet {
    h.projects(w, r)
    return
}
```

Add the handler:

```go
func (h *Handler) projects(w http.ResponseWriter, r *http.Request) {
    projects, err := h.Store.ListProjects(r.Context())
    if err != nil {
        writeError(w, http.StatusInternalServerError, "internal_error", "could not load projects")
        return
    }
    writeJSON(w, http.StatusOK, projects)
}
```

- [ ] **Step 4: Run all Go tests**

Run: `cd api; go test ./...`

Expected: all packages pass.

- [ ] **Step 5: Commit the backend increment**

```powershell
git add api/internal/store/models.go api/internal/store/projects.go api/internal/httpapi/handler.go api/internal/httpapi/handler_test.go
git commit -m "feat: list compliance projects"
```

---

### Task 2: Project Dashboard Component

**Files:**
- Create: `web/src/components/ProjectDashboard.tsx`
- Create: `web/src/components/ProjectDashboard.test.tsx`
- Modify: `web/src/lib/types.ts`

**Interfaces:**
- Consumes: `Project[]`
- Produces: `ProjectDashboard({ projects, loading, openingID, error, onOpen, onCreate })`
- `onCreate` accepts `{ name: string; organizationName: string }`

- [ ] **Step 1: Write failing component tests**

Cover three independent behaviors:

```tsx
test("renders persisted projects and opens the selected project", () => {
  const onOpen = vi.fn();
  render(<ProjectDashboard projects={[project]} loading={false} openingID="" error="" onOpen={onOpen} onCreate={vi.fn()} />);
  expect(screen.getByText("Readiness Review")).toBeInTheDocument();
  expect(screen.getByText("Acme")).toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: /open readiness review/i }));
  expect(onOpen).toHaveBeenCalledWith(project);
});

test("renders a readable empty state", () => {
  render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={vi.fn()} />);
  expect(screen.getByText(/no projects yet/i)).toBeInTheDocument();
});

test("submits trimmed project details", () => {
  const onCreate = vi.fn();
  render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={onCreate} />);
  fireEvent.change(screen.getByLabelText(/project name/i), { target: { value: "  Readiness Review  " } });
  fireEvent.change(screen.getByLabelText(/organization name/i), { target: { value: "  Acme  " } });
  fireEvent.click(screen.getByRole("button", { name: /create project/i }));
  expect(onCreate).toHaveBeenCalledWith({ name: "Readiness Review", organizationName: "Acme" });
});
```

- [ ] **Step 2: Run the component test and observe the expected failure**

Run: `cd web; npm test -- src/components/ProjectDashboard.test.tsx`

Expected: test suite fails because `ProjectDashboard` does not exist and `Project` lacks `organizationName`.

- [ ] **Step 3: Implement the presentational dashboard**

Create a client component that:

- displays `Loading projects…` when `loading` is true;
- displays `No projects yet. Create one to begin an assessment.` for an empty list;
- maps projects to semantic `<article>` cards;
- formats `createdAt` with `new Intl.DateTimeFormat("en-GB", { dateStyle: "medium" })`;
- labels each action `Open ${project.name}` and disables actions while `openingID` is non-empty;
- owns only the two controlled form fields and trims values before calling `onCreate`;
- renders the supplied `error` with the existing `.error` class.

Add `organizationName: string` to the TypeScript `Project` type.

- [ ] **Step 4: Run the component and existing web tests**

Run: `cd web; npm test`

Expected: all tests pass.

- [ ] **Step 5: Commit the component increment**

```powershell
git add web/src/components/ProjectDashboard.tsx web/src/components/ProjectDashboard.test.tsx web/src/lib/types.ts
git commit -m "feat: add project dashboard"
```

---

### Task 3: Home-Screen Coordination and End-to-End Verification

**Files:**
- Create: `web/src/app/page.test.tsx`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/app/page.tsx`
- Modify: `web/src/app/globals.css`
- Modify: `scripts/smoke-test.ps1`

**Interfaces:**
- Produces: `api.getProjects(): Promise<Project[]>`
- Consumes: existing `api.getProfile` and `api.getSummary`

- [ ] **Step 1: Write failing page-coordination tests**

Mock the exported `api` object and add one test that resolves `getFunctions` and `getProjects`, clicks `Open Readiness Review`, and verifies `getProfile("project-1")`, `getSummary("project-1")`, and the assessment heading. Add a second test that clicks `Back to projects` and verifies the project card is visible again.

```tsx
vi.mock("../lib/api", () => ({
  api: {
    getFunctions: vi.fn(),
    getProjects: vi.fn(),
    getProfile: vi.fn(),
    getSummary: vi.fn(),
    createProject: vi.fn(),
    updateProfile: vi.fn(),
  },
}));

test("opens a persisted project and returns to the dashboard", async () => {
  vi.mocked(api.getFunctions).mockResolvedValue(functions);
  vi.mocked(api.getProjects).mockResolvedValue([project]);
  vi.mocked(api.getProfile).mockResolvedValue(profile);
  vi.mocked(api.getSummary).mockResolvedValue(summary);

  render(<Home />);
  fireEvent.click(await screen.findByRole("button", { name: /open readiness review/i }));

  await screen.findByRole("heading", { name: "Readiness Review" });
  expect(api.getProfile).toHaveBeenCalledWith("project-1");
  expect(api.getSummary).toHaveBeenCalledWith("project-1");

  fireEvent.click(screen.getByRole("button", { name: /back to projects/i }));
  expect(screen.getByText("Acme")).toBeInTheDocument();
});
```

- [ ] **Step 2: Run the page test and observe the expected failure**

Run: `cd web; npm test -- src/app/page.test.tsx`

Expected: the test fails because `api.getProjects`, persisted project selection, and `Back to projects` do not exist.

- [ ] **Step 3: Implement the minimal page coordination**

Add to `api`:

```ts
getProjects: () => request<Project[]>("/api/projects"),
```

Refactor `Home` to keep these states: `functions`, `projects`, `project`, `profile`, `summary`, `selected`, `loadingProjects`, `openingID`, and `error`.

On mount, load `getFunctions()` and `getProjects()` together. Render `ProjectDashboard` while `project === null`. Implement `openProject(selectedProject)` by loading profile and summary in parallel and setting the project only after both succeed. Implement creation by calling `createProject`, prepending the result to `projects`, then opening it. Add a `Back to projects` button that sets `project` to `null`, clears the dashboard error, and leaves persisted assessment data untouched.

Add only dashboard layout styles: `.dashboard-header`, `.project-grid`, `.project-card`, `.project-meta`, `.secondary`, and responsive single-column behavior below 640px. Reuse existing color, border, radius, panel, form, error, and button tokens.

- [ ] **Step 4: Extend the real PostgreSQL smoke test**

Immediately after project creation in `scripts/smoke-test.ps1`, add:

```powershell
$projects = Invoke-RestMethod http://localhost:8080/api/projects
$listed = $projects | Where-Object { $_.id -eq $project.id }
if (-not $listed) { throw 'created project was not returned by project list' }
if ($listed.organizationName -ne 'Smoke Org') { throw 'project list omitted organization name' }
```

- [ ] **Step 5: Run focused and full verification**

Run:

```powershell
Set-Location api
go test ./...
Set-Location ..\web
npm test
npm run typecheck
npm run build
Set-Location ..
docker compose config
docker compose up --build -d
powershell -ExecutionPolicy Bypass -File scripts/smoke-test.ps1
```

Expected: Go tests pass, all Vitest tests pass, type-check and Next.js build exit 0, Compose config is valid, and smoke output is `smoke test passed`.

- [ ] **Step 6: Manually verify the user workflow**

Open `http://localhost:3000`, verify the smoke project appears, open it, confirm 106 assessment outcomes load, use `Back to projects`, and confirm the project list remains visible.

- [ ] **Step 7: Commit the integrated feature**

```powershell
git add web/src/app/page.test.tsx web/src/lib/api.ts web/src/app/page.tsx web/src/app/globals.css scripts/smoke-test.ps1
git commit -m "feat: reopen existing assessments"
```

- [ ] **Step 8: Stop the verification environment**

Run: `docker compose down`

Expected: web, API, and PostgreSQL containers stop; the PostgreSQL named volume remains intact.
