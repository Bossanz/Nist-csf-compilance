import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import Home from "./page";
import { api } from "../lib/api";
import type { FunctionNode, ProfileRow, Project, Summary } from "../lib/types";

vi.mock("../lib/api", () => ({
  api: {
    getFunctions: vi.fn(),
    getProjects: vi.fn(),
    getProfile: vi.fn(),
    getSummary: vi.fn(),
    createProject: vi.fn(),
    deleteProject: vi.fn(),
    updateProfile: vi.fn(),
  },
}));

const project: Project = { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "Readiness Review", status: "setup", createdAt: "2026-08-06T03:00:00Z" };
const otherProject: Project = { id: "project-2", organizationID: "org-2", organizationName: "Beta", name: "Second Review", status: "setup", createdAt: "2026-08-05T03:00:00Z" };
const functions: FunctionNode[] = [{ id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] }];
const profile: ProfileRow[] = [{ id: "profile-1", projectID: "project-1", subcategoryID: "subcategory-1", functionCode: "GV", categoryCode: "GV.OC", subcategoryCode: "GV.OC-01", description: "The organizational mission is understood", included: false, rationale: "", currentPriority: "", currentCoverageLevel: "none", currentStatusText: "", currentPoliciesText: "", currentTier: "", targetPriority: "", targetCoverageLevel: "none", targetApproachText: "", targetTier: "", notes: "", considerations: "", reviewStatus: "draft" }];
const summary: Summary = { coveragePct: 0, includedCount: 0, pendingCount: 0, rejectedCount: 0, functions: [] };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.getFunctions).mockResolvedValue(functions);
  vi.mocked(api.getProjects).mockResolvedValue([project]);
  vi.mocked(api.getProfile).mockImplementation(async id => {
    if (id !== project.id) throw new Error("wrong project requested");
    return profile;
  });
  vi.mocked(api.getSummary).mockImplementation(async id => {
    if (id !== project.id) throw new Error("wrong project requested");
    return summary;
  });
  vi.mocked(api.deleteProject).mockImplementation(async id => {
    if (id !== project.id) throw new Error("wrong project deleted");
  });
});

afterEach(() => vi.restoreAllMocks());

test("opens a persisted project and returns to the dashboard", async () => {
  render(<Home />);

  fireEvent.click(await screen.findByRole("button", { name: /open readiness review/i }));
  expect(await screen.findByRole("heading", { name: "Readiness Review" })).toBeTruthy();
  expect(screen.getByText("The organizational mission is understood")).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /back to projects/i }));
  expect(screen.getByText("Acme")).toBeTruthy();
});

test("removes only the project deleted successfully", async () => {
  vi.mocked(api.getProjects).mockResolvedValue([project, otherProject]);
  vi.spyOn(window, "confirm").mockReturnValue(true);
  render(<Home />);

  fireEvent.click(await screen.findByRole("button", { name: /delete readiness review/i }));

  await waitFor(() => expect(screen.queryByText("Readiness Review")).toBeNull());
  expect(screen.getByText("Second Review")).toBeTruthy();
});
