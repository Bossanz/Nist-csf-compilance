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
    me: vi.fn(), getOrganizationBySlug: vi.fn(), getOrganizationProjects: vi.fn(), getOrganizationUsers: vi.fn(),
    createOrganizationProject: vi.fn(), deleteProject: vi.fn(), createInvitation: vi.fn(),
  },
}));

const user: User = { id: "user-1", organizationID: null, name: "Consultant", email: "c@example.com", userType: "counselor", role: "counselor", status: "active" };
const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: organization.name, name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-06T03:00:00Z" };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockResolvedValue(user);
  vi.mocked(api.getOrganizationBySlug).mockResolvedValue(organization);
  vi.mocked(api.getOrganizationProjects).mockResolvedValue([project]);
  vi.mocked(api.getOrganizationUsers).mockResolvedValue([]);
});

test("opens a project using organization and project slugs", async () => {
  render(<OrganizationPage />);
  fireEvent.click(await screen.findByRole("button", { name: /open readiness/i }));
  await waitFor(() => expect(router.push).toHaveBeenCalledWith("/organizations/acme-corporation/projects/readiness"));
});
