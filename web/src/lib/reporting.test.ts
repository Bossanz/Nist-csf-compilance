import { afterEach, expect, test, vi } from "vitest";
import { api } from "./api";

afterEach(() => vi.unstubAllGlobals());

test("finalizes a project through the project endpoint", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ status: "closed" }) });
  vi.stubGlobal("fetch", fetchMock);

  await api.finalizeProject("project-1");

  expect(fetchMock).toHaveBeenCalledWith("/api/projects/project-1/finalize", expect.objectContaining({ method: "POST" }));
});

test("loads the final report and audit package", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
  vi.stubGlobal("fetch", fetchMock);

  await api.getFinalReport("project-1");
  await api.getAuditPackage("project-1");

  expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/projects/project-1/final-report", expect.any(Object));
  expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/projects/project-1/audit-package", expect.any(Object));
});

test("downloads the audit package CSV", async () => {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, blob: async () => new Blob(["csv"]) });
  vi.stubGlobal("fetch", fetchMock);

  await api.downloadAuditPackageCSV("project-1");

  expect(fetchMock).toHaveBeenCalledWith("/api/projects/project-1/audit-package.csv", { signal: undefined });
});
