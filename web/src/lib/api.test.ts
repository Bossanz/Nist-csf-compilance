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
