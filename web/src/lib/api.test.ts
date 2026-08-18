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

test("restores the current authenticated user", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "user-1", role: "counselor_admin" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.me();

  expect(fetchMock).toHaveBeenCalledWith("/api/auth/me", expect.any(Object));
});

test("creates a project inside an existing organization", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 201, json: async () => ({ id: "project-1" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.createOrganizationProject("org-1", { name: "Readiness" });

  expect(fetchMock).toHaveBeenCalledWith("/api/organizations/org-1/projects", expect.objectContaining({ method: "POST", body: JSON.stringify({ name: "Readiness" }) }));
});

test("resolves an organization by slug", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "org-1", slug: "acme-corporation" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.getOrganizationBySlug("acme-corporation");

  expect(fetchMock).toHaveBeenCalledWith("/api/organizations/by-slug/acme-corporation", expect.any(Object));
});

test("resolves a project by organization and project slugs", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "project-1", slug: "readiness" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.getOrganizationProjectBySlug("org-1", "readiness");

  expect(fetchMock).toHaveBeenCalledWith("/api/organizations/org-1/projects/by-slug/readiness", expect.any(Object));
});

test("forwards an abort signal when downloading evidence", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, blob: async () => new Blob(["evidence"]) });
  vi.stubGlobal("fetch", fetchMock);
  const controller = new AbortController();

  await api.downloadResponseDocument("project-1", "subcategory-1", "document-1", controller.signal);

  expect(fetchMock).toHaveBeenCalledWith(
    "/api/projects/project-1/responses/subcategory-1/documents/document-1",
    { signal: controller.signal },
  );
});
