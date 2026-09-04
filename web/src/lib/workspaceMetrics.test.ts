import { expect, test } from "vitest";
import { getAttentionLabel, getOrganizationPortfolioMetrics, getProjectStatusLabel } from "./workspaceMetrics";

test("derives portfolio workspace counts from available project totals", () => {
  expect(getOrganizationPortfolioMetrics(
    [
      { id: "org-1", name: "Acme", slug: "acme", type: "client" },
      { id: "org-2", name: "Beta", slug: "beta", type: "client" },
    ],
    { "org-1": 2, "org-2": 1 },
  )).toEqual({ total: 2, active: 2, finalized: 0, totalProjects: 3 });
});

test("formats workflow statuses for the register", () => {
  expect(getProjectStatusLabel("in_review")).toBe("Reviewing");
  expect(getProjectStatusLabel("closed")).toBe("Finalized");
});

test("names the role-aware next action", () => {
  expect(getAttentionLabel("reviewer")).toBe("To review");
  expect(getAttentionLabel("assessor")).toBe("Needs your input");
});
