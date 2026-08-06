# Same-origin API Proxy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Route browser API traffic through the Next.js origin so the compliance UI no longer depends on direct cross-origin requests to port 8080.

**Architecture:** The API client sends relative `/api/...` requests. A Next.js rewrite proxies those requests server-side to a configurable internal API URL, set to the Compose service name in Docker and localhost during standalone development.

**Tech Stack:** Next.js 16, TypeScript, Vitest, Docker Compose, Go API

## Global Constraints

- Keep the Go API and its routes unchanged.
- Do not add retries, dependencies, authentication changes, or unrelated refactors.
- Keep port 8080 published for local diagnostics while normal browser traffic uses port 3000.

---

### Task 1: Route API requests through Next.js

**Files:**
- Create: `web/src/lib/api.test.ts`
- Create: `web/next.config.test.ts`
- Modify: `web/src/lib/api.ts`
- Modify: `web/next.config.ts`
- Modify: `docker-compose.yml`

**Interfaces:**
- Consumes: existing `api.getProjects(): Promise<Project[]>` and Next.js `rewrites()` configuration interface.
- Produces: browser requests to relative `/api/...` URLs and a rewrite from `/api/:path*` to `${API_INTERNAL_URL}/api/:path*`.

- [ ] **Step 1: Write failing tests for the browser URL and proxy rewrite**

Create `web/src/lib/api.test.ts`:

```ts
import { afterEach, expect, test, vi } from "vitest";
import { api } from "./api";

afterEach(() => vi.unstubAllGlobals());

test("requests projects through the current web origin", async () => {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => [],
  });
  vi.stubGlobal("fetch", fetchMock);

  await api.getProjects();

  expect(fetchMock).toHaveBeenCalledWith(
    "/api/projects",
    expect.objectContaining({ headers: { "Content-Type": "application/json" } }),
  );
});
```

Create `web/next.config.test.ts`:

```ts
import { expect, test } from "vitest";
import nextConfig from "./next.config";

test("proxies API routes to the internal Go service", async () => {
  const rewrites = await nextConfig.rewrites!();

  expect(rewrites).toContainEqual({
    source: "/api/:path*",
    destination: "http://localhost:8080/api/:path*",
  });
});
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `npm test -- src/lib/api.test.ts next.config.test.ts`

Expected: both tests fail because the client currently prepends `http://localhost:8080` and no `rewrites` function exists.

- [ ] **Step 3: Implement the minimal same-origin client and rewrite**

Change `web/src/lib/api.ts` so `base` defaults to an empty string:

```ts
const base = process.env.NEXT_PUBLIC_API_BASE_URL || "";
```

Change `web/next.config.ts` to:

```ts
import type { NextConfig } from "next";

const apiInternalURL = process.env.API_INTERNAL_URL || "http://localhost:8080";
const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiInternalURL}/api/:path*` }];
  },
};

export default nextConfig;
```

Change the `web.environment` section in `docker-compose.yml` to:

```yaml
environment:
  API_INTERNAL_URL: http://api:8080
```

- [ ] **Step 4: Run focused and complete frontend verification**

Run from `web`:

```powershell
npm test -- src/lib/api.test.ts next.config.test.ts
npm test
npm run typecheck
npm run build
```

Expected: every command exits 0 with all tests passing.

- [ ] **Step 5: Rebuild Docker and verify the running workflow**

Run from the repository root:

```powershell
docker compose up --build -d
docker compose ps
Invoke-WebRequest -UseBasicParsing http://localhost:3000/api/projects
```

Expected: PostgreSQL and API are healthy, Web is running, and the proxied projects endpoint returns HTTP 200 through port 3000.

In the in-app browser, reload `http://localhost:3000`, open the existing project, and confirm there is no `Failed to fetch` alert or browser console error.

- [ ] **Step 6: Commit the fix**

```powershell
git add web/src/lib/api.test.ts web/next.config.test.ts web/src/lib/api.ts web/next.config.ts docker-compose.yml
git commit -m "fix: proxy browser api requests through nextjs"
```
