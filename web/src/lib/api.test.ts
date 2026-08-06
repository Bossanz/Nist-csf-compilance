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
