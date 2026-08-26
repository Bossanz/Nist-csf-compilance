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

test("requests a password reset without exposing the email flow to the caller", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 202, json: async () => ({ message: "If an active account exists" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.requestPasswordReset("person@example.com");

  expect(fetchMock).toHaveBeenCalledWith(
    "/api/auth/password-reset/request",
    expect.objectContaining({ method: "POST", body: JSON.stringify({ email: "person@example.com" }) }),
  );
});

test("confirms a password reset and changes an authenticated password", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 204 });
  vi.stubGlobal("fetch", fetchMock);

  await api.confirmPasswordReset("reset-token", "new-password");
  await api.changePassword("old-password", "new-password");

  expect(fetchMock).toHaveBeenNthCalledWith(
    1,
    "/api/auth/password-reset/confirm",
    expect.objectContaining({ method: "POST", body: JSON.stringify({ token: "reset-token", password: "new-password" }) }),
  );
  expect(fetchMock).toHaveBeenNthCalledWith(
    2,
    "/api/auth/password",
    expect.objectContaining({ method: "PUT", body: JSON.stringify({ currentPassword: "old-password", newPassword: "new-password" }) }),
  );
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

test("loads project version history and starts a new version", async () => {
  const fetchMock = vi.fn()
    .mockResolvedValueOnce({ ok: true, status: 200, json: async () => [{ id: "project-2", versionNumber: 2 }] })
    .mockResolvedValueOnce({ ok: true, status: 201, json: async () => ({ id: "project-2", versionNumber: 2 }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.getProjectVersions("project-1");
  await api.createProjectVersion("project-1");

  expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/projects/project-1/versions", expect.any(Object));
  expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/projects/project-1/versions", expect.objectContaining({ method: "POST", body: "{}" }));
});

test("updates a whole Function scope in one request", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => [] });
  vi.stubGlobal("fetch", fetchMock);

  await api.updateFunctionScope("project-1", "GV", true);

  expect(fetchMock).toHaveBeenCalledWith(
    "/api/projects/project-1/functions/GV/scope",
    expect.objectContaining({ method: "PUT", body: JSON.stringify({ included: true }) }),
  );
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

test("serializes the remediation action lifecycle", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "action-1" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.createRemediationAction("project-1", {
    subcategoryID: "outcome-1",
    title: "Centralize logs",
    description: "Forward application logs.",
    desiredResult: "Security events are searchable.",
    priority: "high",
    ownerUserID: "assessor-1",
    dueDate: "2026-09-30",
  });
  await api.updateRemediationProgress("project-1", "action-1", "SIEM forwarding is enabled.");
  await api.submitRemediationAction("project-1", "action-1");
  await api.reviewRemediationAction("project-1", "action-1", { decision: "return", comment: "Attach the deployment record." });

  expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/projects/project-1/remediation-actions", expect.objectContaining({ method: "POST" }));
  expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/projects/project-1/remediation-actions/action-1/progress", expect.objectContaining({ method: "PATCH", body: JSON.stringify({ progressNote: "SIEM forwarding is enabled." }) }));
  expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/projects/project-1/remediation-actions/action-1/submit", expect.objectContaining({ method: "POST" }));
  expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/projects/project-1/remediation-actions/action-1/review", expect.objectContaining({ method: "POST", body: JSON.stringify({ decision: "return", comment: "Attach the deployment record." }) }));
});
