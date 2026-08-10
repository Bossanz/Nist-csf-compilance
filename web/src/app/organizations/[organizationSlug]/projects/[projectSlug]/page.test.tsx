import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import ProjectPage from "./page";
import { api } from "../../../../../lib/api";
import type { FunctionNode, Organization, ProfileRow, Project, Summary, User } from "../../../../../lib/types";

const router = vi.hoisted(() => ({ push: vi.fn(), replace: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => router, useParams: () => ({ organizationSlug: "acme-corporation", projectSlug: "readiness" }) }));
vi.mock("../../../../../lib/api", () => ({
  APIError: class APIError extends Error {
    constructor(message: string, public status: number) { super(message); }
  },
  api: {
    me: vi.fn(), getOrganizationBySlug: vi.fn(), getOrganizationProjectBySlug: vi.fn(), getFunctions: vi.fn(),
    getProfile: vi.fn(), getSummary: vi.fn(), getResponses: vi.fn(), updateProfile: vi.fn(), saveResponse: vi.fn(),
    submitResponse: vi.fn(), reviewResponse: vi.fn(), uploadResponseDocument: vi.fn(), deleteResponseDocument: vi.fn(),
    downloadResponseDocument: vi.fn(),
  },
}));

const user: User = { id: "user-1", organizationID: null, name: "Consultant", email: "c@example.com", userType: "counselor", role: "counselor", status: "active" };
const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: organization.name, name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-06T03:00:00Z" };
const functions: FunctionNode[] = [{ id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] }];
const summary: Summary = { coveragePct: 0, includedCount: 0, pendingCount: 0, rejectedCount: 0, functions: [] };
const profile: ProfileRow[] = [];

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockResolvedValue(user);
  vi.mocked(api.getOrganizationBySlug).mockResolvedValue(organization);
  vi.mocked(api.getOrganizationProjectBySlug).mockResolvedValue(project);
  vi.mocked(api.getFunctions).mockResolvedValue(functions);
  vi.mocked(api.getProfile).mockResolvedValue(profile);
  vi.mocked(api.getSummary).mockResolvedValue(summary);
  vi.mocked(api.getResponses).mockResolvedValue([]);
});

test("loads assessment data after resolving the project slug", async () => {
  render(<ProjectPage />);
  expect(await screen.findByRole("heading", { name: "Readiness" })).toBeTruthy();
  expect(api.getOrganizationProjectBySlug).toHaveBeenCalledWith("org-1", "readiness");
  expect(api.getProfile).toHaveBeenCalledWith("project-1");
  expect(screen.queryByRole("complementary", { name: /assessment context/i })).toBeNull();
});

test("back navigation returns to the organization slug route", async () => {
  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("button", { name: /back to organization/i }));
  await waitFor(() => expect(router.push).toHaveBeenCalledWith("/organizations/acme-corporation"));
});
