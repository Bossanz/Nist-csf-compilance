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
    getRemediationActions: vi.fn(), createRemediationAction: vi.fn(), updateRemediationAction: vi.fn(),
    updateRemediationProgress: vi.fn(), submitRemediationAction: vi.fn(), reviewRemediationAction: vi.fn(),
    uploadRemediationEvidence: vi.fn(), downloadRemediationEvidence: vi.fn(), deleteRemediationEvidence: vi.fn(),
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
const secondBulkProfile: ProfileRow = { ...previewProfile, id: "profile-2", subcategoryID: "subcategory-2", subcategoryCode: "GV.OC-02", included: false };
const previewDocument: ResponseDocument = { id: "doc-1", responseID: "response-1", originalName: "evidence.pdf", mimeType: "application/pdf", sizeBytes: 12, uploadedBy: "assessor-1", createdAt: "2026-08-07T00:00:00Z" };
const secondPreviewDocument: ResponseDocument = { ...previewDocument, id: "doc-2", originalName: "second-evidence.pdf" };
const previewResponse: StakeholderResponse = { id: "response-1", projectID: "project-1", subcategoryID: "subcategory-1", responseText: "Evidence attached", status: "draft", respondedBy: "assessor-1", submittedAt: null, reviewComment: "", reviewedBy: null, reviewedAt: null, createdAt: "2026-08-07T00:00:00Z", updatedAt: "2026-08-07T00:00:00Z", documents: [previewDocument] };

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  const promise = new Promise<T>((nextResolve) => { resolve = nextResolve; });
  return { promise, resolve };
}

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
  vi.mocked(api.getRemediationActions).mockResolvedValue([]);
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

test("starts authentication and organization lookup together", async () => {
  const auth = deferred<User>();
  const organizationLookup = deferred<Organization>();
  vi.mocked(api.me).mockReturnValue(auth.promise);
  vi.mocked(api.getOrganizationBySlug).mockReturnValue(organizationLookup.promise);

  render(<ProjectPage />);
  await waitFor(() => expect(api.me).toHaveBeenCalled());
  expect(api.getOrganizationBySlug).toHaveBeenCalledWith("acme-corporation");

  auth.resolve(user);
  organizationLookup.resolve(organization);
  expect(await screen.findByRole("heading", { name: "Readiness" })).toBeTruthy();
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
  vi.mocked(api.getOrganizationProjectBySlug).mockResolvedValue({ ...project, status: "in_review" });
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

  await waitFor(() => expect(api.downloadResponseDocument).toHaveBeenCalledWith("project-1", "subcategory-1", "doc-1", expect.any(AbortSignal)));
  expect(createObjectURL).toHaveBeenCalled();
  expect(screen.getByTitle("evidence.pdf preview").getAttribute("src")).toBe("blob:evidence-preview");
  fireEvent.click(screen.getByRole("button", { name: /close preview/i }));
  expect(revokeObjectURL).toHaveBeenCalledWith("blob:evidence-preview");
});

test("loads remediation actions and opens the Action Plan workspace", async () => {
  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  expect(api.getRemediationActions).toHaveBeenCalledWith("project-1");
  fireEvent.click(screen.getByRole("button", { name: "Action Plan" }));
  expect(screen.getByRole("heading", { name: "Action Plan" })).toBeTruthy();
});

test("cancels a stale evidence preview request before starting another", async () => {
  vi.mocked(api.getOrganizationProjectBySlug).mockResolvedValue({ ...project, status: "in_review" });
  vi.mocked(api.getProfile).mockResolvedValue([previewProfile]);
  vi.mocked(api.getResponses).mockResolvedValue([{ ...previewResponse, documents: [previewDocument, secondPreviewDocument] }]);
  vi.stubGlobal("URL", { createObjectURL: vi.fn().mockReturnValue("blob:second-evidence-preview"), revokeObjectURL: vi.fn() });
  const firstDownload = deferred<Blob>();
  const secondDownload = deferred<Blob>();
  vi.mocked(api.downloadResponseDocument)
    .mockReturnValueOnce(firstDownload.promise)
    .mockReturnValueOnce(secondDownload.promise);

  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  fireEvent.click(screen.getByRole("button", { name: /preview evidence\.pdf/i }));
  await waitFor(() => expect(api.downloadResponseDocument).toHaveBeenCalledTimes(1));

  fireEvent.click(screen.getByRole("button", { name: /preview second-evidence\.pdf/i }));
  await waitFor(() => expect(api.downloadResponseDocument).toHaveBeenCalledTimes(2));

  const firstCall = vi.mocked(api.downloadResponseDocument).mock.calls[0] as unknown as [string, string, string, AbortSignal];
  expect(firstCall[3]).toBeInstanceOf(AbortSignal);
  expect(firstCall[3].aborted).toBe(true);

  secondDownload.resolve(new Blob(["%PDF-1.7"], { type: "application/pdf" }));
  await waitFor(() => expect(screen.getByTitle("second-evidence.pdf preview")).toBeTruthy());
  fireEvent.click(screen.getByRole("button", { name: /close preview/i }));
});

test("includes every outcome in the selected Function", async () => {
  vi.mocked(api.getProfile).mockResolvedValue([previewProfile, secondBulkProfile]);
  vi.mocked(api.getResponses).mockResolvedValue([]);

  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("checkbox", { name: /include all outcomes in this function/i }));

  await waitFor(() => {
    expect(api.updateProfile).toHaveBeenNthCalledWith(1, "project-1", "subcategory-1", { included: true });
    expect(api.updateProfile).toHaveBeenNthCalledWith(2, "project-1", "subcategory-2", { included: true });
  });
});

test("shows a retryable error when assessment data cannot be loaded", async () => {
  vi.mocked(api.getProfile).mockRejectedValueOnce(new Error("Assessment API is unavailable"));
  render(<ProjectPage />);

  expect(await screen.findByRole("heading", { name: /could not load project/i })).toBeTruthy();
  expect(screen.getByText("Assessment API is unavailable")).toBeTruthy();

  vi.mocked(api.getProfile).mockResolvedValueOnce(profile);
  fireEvent.click(screen.getByRole("button", { name: /try again/i }));

  expect(await screen.findByRole("heading", { name: "Readiness" })).toBeTruthy();
});

test("serializes bulk outcome updates and keeps the first update before a later failure", async () => {
  vi.mocked(api.getProfile).mockResolvedValue([previewProfile, secondBulkProfile]);
  let resolveFirst: (value: ProfileRow) => void = () => undefined;
  const firstUpdate = new Promise<ProfileRow>((resolve) => { resolveFirst = resolve; });
  vi.mocked(api.updateProfile)
    .mockReturnValueOnce(firstUpdate)
    .mockRejectedValueOnce(new Error("Second outcome failed"));

  render(<ProjectPage />);
  await screen.findByRole("heading", { name: "Readiness" });
  fireEvent.click(screen.getByRole("checkbox", { name: /include all outcomes in this function/i }));

  await waitFor(() => expect(api.updateProfile).toHaveBeenCalledTimes(1));
  expect(api.updateProfile).toHaveBeenNthCalledWith(1, "project-1", "subcategory-1", { included: true });
  resolveFirst({ ...previewProfile, included: true });
  await waitFor(() => expect(api.updateProfile).toHaveBeenCalledTimes(2));
  expect(api.updateProfile).toHaveBeenNthCalledWith(2, "project-1", "subcategory-2", { included: true });
  expect(await screen.findByText("Second outcome failed")).toBeTruthy();
});
