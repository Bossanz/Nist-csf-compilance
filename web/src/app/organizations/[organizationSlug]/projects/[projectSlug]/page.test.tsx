import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import ProjectPage from "./page";
import { api } from "../../../../../lib/api";
import type { FunctionNode, Organization, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../../../../../lib/types";

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
    downloadResponseDocument: vi.fn(), getOrganizationUsers: vi.fn(),
  },
}));

const user: User = { id: "user-1", organizationID: null, name: "Consultant", email: "c@example.com", userType: "counselor", role: "counselor", status: "active" };
const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: organization.name, name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-06T03:00:00Z" };
const functions: FunctionNode[] = [{ id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] }];
const summary: Summary = { coveragePct: 0, includedCount: 0, pendingCount: 0, rejectedCount: 0, functions: [] };
const profile: ProfileRow[] = [];
const previewProfile: ProfileRow = {
  id: "profile-1", projectID: "project-1", subcategoryID: "subcategory-1", functionCode: "GV", categoryCode: "GV.OC", subcategoryCode: "GV.OC-01", description: "The organizational mission is understood", included: true, rationale: "", currentPriority: "", currentCoverageLevel: "none", currentStatusText: "", currentPoliciesText: "", currentTier: "", targetPriority: "", targetCoverageLevel: "none", targetApproachText: "", targetTier: "", notes: "", considerations: "", reviewStatus: "draft", assignedUserID: "assessor-1",
};
const previewDocument: ResponseDocument = { id: "doc-1", responseID: "response-1", originalName: "evidence.pdf", mimeType: "application/pdf", sizeBytes: 12, uploadedBy: "assessor-1", createdAt: "2026-08-07T00:00:00Z" };
const previewResponse: StakeholderResponse = { id: "response-1", projectID: "project-1", subcategoryID: "subcategory-1", responseText: "Evidence attached", status: "draft", respondedBy: "assessor-1", submittedAt: null, reviewComment: "", reviewedBy: null, reviewedAt: null, createdAt: "2026-08-07T00:00:00Z", updatedAt: "2026-08-07T00:00:00Z", documents: [previewDocument] };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockResolvedValue(user);
  vi.mocked(api.getOrganizationBySlug).mockResolvedValue(organization);
  vi.mocked(api.getOrganizationProjectBySlug).mockResolvedValue(project);
  vi.mocked(api.getFunctions).mockResolvedValue(functions);
  vi.mocked(api.getProfile).mockResolvedValue(profile);
  vi.mocked(api.getSummary).mockResolvedValue(summary);
  vi.mocked(api.getResponses).mockResolvedValue([]);
  vi.mocked(api.getOrganizationUsers).mockResolvedValue([]);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

test("loads assessment data after resolving the project slug", async () => {
  render(<ProjectPage />);
  expect(await screen.findByRole("heading", { name: "Readiness" })).toBeTruthy();
  expect(api.getOrganizationProjectBySlug).toHaveBeenCalledWith("org-1", "readiness");
  expect(api.getProfile).toHaveBeenCalledWith("project-1");
  expect(screen.queryByRole("complementary", { name: /assessment context/i })).toBeNull();
});

test("loads organization users for Counselor assignment controls", async () => {
  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });

  expect(api.getOrganizationUsers).toHaveBeenCalledWith("org-1");
});

test("back navigation returns to the organization slug route", async () => {
  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("button", { name: /back to organization/i }));
  await waitFor(() => expect(router.push).toHaveBeenCalledWith("/organizations/acme-corporation"));
});

test("loads and closes a supported evidence preview", async () => {
  const createObjectURL = vi.fn().mockReturnValue("blob:evidence-preview");
  const revokeObjectURL = vi.fn();
  vi.stubGlobal("URL", { createObjectURL, revokeObjectURL });
  vi.mocked(api.getProfile).mockResolvedValue([previewProfile]);
  vi.mocked(api.getResponses).mockResolvedValue([previewResponse]);
  vi.mocked(api.downloadResponseDocument).mockResolvedValue(new Blob(["%PDF-1.7"], { type: "application/pdf" }));

  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  fireEvent.click(screen.getByRole("button", { name: /preview evidence\.pdf/i }));

  await waitFor(() => expect(api.downloadResponseDocument).toHaveBeenCalledWith("project-1", "subcategory-1", "doc-1"));
  expect(createObjectURL).toHaveBeenCalled();
  expect(screen.getByTitle("evidence.pdf preview").getAttribute("src")).toBe("blob:evidence-preview");
  fireEvent.click(screen.getByRole("button", { name: /close preview/i }));
  expect(revokeObjectURL).toHaveBeenCalledWith("blob:evidence-preview");
});
