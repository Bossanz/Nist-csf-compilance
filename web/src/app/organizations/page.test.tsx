import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import OrganizationsPage from "./page";
import { api } from "../../lib/api";
import type { Organization, User } from "../../lib/types";

const router = vi.hoisted(() => ({ push: vi.fn(), replace: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => router }));
vi.mock("../../lib/api", () => ({
  APIError: class APIError extends Error {
    constructor(message: string, public status: number) { super(message); }
  },
  api: { me: vi.fn(), getOrganizations: vi.fn(), getCounselors: vi.fn(), createOrganization: vi.fn(), deleteOrganization: vi.fn(), updateCounselor: vi.fn(), createCounselorInvitation: vi.fn(), logout: vi.fn() },
}));

const user: User = { id: "user-1", organizationID: null, name: "Consultant", email: "c@example.com", userType: "counselor", role: "counselor", status: "active" };
const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockResolvedValue(user);
  vi.mocked(api.getOrganizations).mockResolvedValue([organization]);
  vi.mocked(api.getCounselors).mockResolvedValue([]);
});

test("opens an organization using its slug route", async () => {
  render(<OrganizationsPage />);
  fireEvent.click(await screen.findByRole("button", { name: /open acme corporation/i }));
  await waitFor(() => expect(router.push).toHaveBeenCalledWith("/organizations/acme-corporation"));
});

test("loads counselors for a counselor admin", async () => {
  vi.mocked(api.me).mockResolvedValue({ ...user, role: "counselor_admin" });
  render(<OrganizationsPage />);
  await screen.findByRole("heading", { name: "Organizations" });
  await waitFor(() => expect(api.getCounselors).toHaveBeenCalledTimes(1));
});

test("shows a retryable error when organizations cannot be loaded", async () => {
  vi.mocked(api.getOrganizations).mockRejectedValueOnce(new Error("Organizations API is unavailable"));
  render(<OrganizationsPage />);

  expect(await screen.findByRole("heading", { name: /could not load organizations/i })).toBeTruthy();
  expect(screen.getByText("Organizations API is unavailable")).toBeTruthy();

  vi.mocked(api.getOrganizations).mockResolvedValueOnce([organization]);
  fireEvent.click(screen.getByRole("button", { name: /try again/i }));

  expect(await screen.findByRole("button", { name: /open acme corporation/i })).toBeTruthy();
});
