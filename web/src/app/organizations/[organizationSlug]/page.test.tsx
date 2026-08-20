import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import OrganizationPage from "./page";
import { api } from "../../../lib/api";
import type { Organization, Project, User } from "../../../lib/types";

const router = vi.hoisted(() => ({ push: vi.fn(), replace: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => router, useParams: () => ({ organizationSlug: "acme-corporation" }) }));
vi.mock("../../../lib/api", () => ({
  APIError: class APIError extends Error {
    constructor(message: string, public status: number) { super(message); }
  },
  api: {
    me: vi.fn(), getOrganizationBySlug: vi.fn(), getOrganizationProjects: vi.fn(), getOrganizationUsers: vi.fn(), getOrganizationInvitations: vi.fn(),
    createOrganizationProject: vi.fn(), deleteProject: vi.fn(), createInvitation: vi.fn(), resendInvitation: vi.fn(), cancelInvitation: vi.fn(), updateOrganizationUser: vi.fn(),
  },
}));

const user: User = { id: "user-1", organizationID: null, name: "Consultant", email: "c@example.com", userType: "counselor", role: "counselor", status: "active" };
const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: organization.name, name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-06T03:00:00Z" };

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  const promise = new Promise<T>((nextResolve) => { resolve = nextResolve; });
  return { promise, resolve };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockResolvedValue(user);
  vi.mocked(api.getOrganizationBySlug).mockResolvedValue(organization);
  vi.mocked(api.getOrganizationProjects).mockResolvedValue([project]);
  vi.mocked(api.getOrganizationUsers).mockResolvedValue([]);
  vi.mocked(api.getOrganizationInvitations).mockResolvedValue([]);
  vi.mocked(api.updateOrganizationUser).mockResolvedValue({} as User);
});

test("opens a project using organization and project slugs", async () => {
  render(<OrganizationPage />);
  fireEvent.click(await screen.findByRole("button", { name: /open readiness/i }));
  await waitFor(() => expect(router.push).toHaveBeenCalledWith("/organizations/acme-corporation/projects/readiness"));
});

test("starts authentication and organization lookup together", async () => {
  const auth = deferred<User>();
  const organizationLookup = deferred<Organization>();
  vi.mocked(api.me).mockReturnValue(auth.promise);
  vi.mocked(api.getOrganizationBySlug).mockReturnValue(organizationLookup.promise);
  vi.mocked(api.getOrganizationProjects).mockResolvedValue([]);
  vi.mocked(api.getOrganizationUsers).mockResolvedValue([]);

  render(<OrganizationPage />);
  await waitFor(() => expect(api.me).toHaveBeenCalled());
  expect(api.getOrganizationBySlug).toHaveBeenCalledWith("acme-corporation");

  auth.resolve(user);
  organizationLookup.resolve(organization);
  expect(await screen.findByRole("heading", { name: "Acme Corporation" })).toBeTruthy();
});

test("updates an organization user access", async () => {
  const member: User = { id: "member-1", organizationID: "org-1", name: "Member", email: "member@acme.test", userType: "stakeholder", role: "assessor", status: "active" };
  vi.mocked(api.getOrganizationUsers).mockResolvedValue([member]);
  vi.mocked(api.updateOrganizationUser).mockResolvedValue({ ...member, role: "reviewer", status: "disabled" });
  render(<OrganizationPage />);
  await screen.findByRole("heading", { name: "Acme Corporation" });
  fireEvent.change(screen.getByLabelText(/role for member@acme.test/i), { target: { value: "reviewer" } });
  fireEvent.change(screen.getByLabelText(/status for member@acme.test/i), { target: { value: "disabled" } });
  fireEvent.click(screen.getByRole("button", { name: /save access for member@acme.test/i }));
  await waitFor(() => expect(api.updateOrganizationUser).toHaveBeenCalledWith("org-1", "member-1", { role: "reviewer", status: "disabled" }));
});

test("shows a retryable error when organization data cannot be loaded", async () => {
  vi.mocked(api.getOrganizationProjects).mockRejectedValueOnce(new Error("Workspace API is unavailable"));
  render(<OrganizationPage />);

  expect(await screen.findByRole("heading", { name: /could not load organization/i })).toBeTruthy();
  expect(screen.getByText("Workspace API is unavailable")).toBeTruthy();

  vi.mocked(api.getOrganizationProjects).mockResolvedValueOnce([project]);
  fireEvent.click(screen.getByRole("button", { name: /try again/i }));

  expect(await screen.findByRole("heading", { name: "Acme Corporation" })).toBeTruthy();
});
