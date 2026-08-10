import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import InvitationPage from "./page";
import { api } from "../../../lib/api";

vi.mock("next/navigation", () => ({ useParams: () => ({ token: "invite-token" }) }));
vi.mock("next/link", () => ({
  default: ({ href, children, ...props }: { href: string; children: React.ReactNode }) => <a href={href} {...props}>{children}</a>,
}));
vi.mock("../../../lib/api", () => ({
  api: { acceptInvitation: vi.fn(), logout: vi.fn() },
}));

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.acceptInvitation).mockResolvedValue({} as never);
  vi.mocked(api.logout).mockResolvedValue(undefined);
});

test("clears the existing session and sends an activated account to login", async () => {
  render(<InvitationPage />);

  fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "Stakeholder" } });
  fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "StrongPass!2026" } });
  fireEvent.click(screen.getByRole("button", { name: /activate account/i }));

  await waitFor(() => expect(api.acceptInvitation).toHaveBeenCalledWith("invite-token", { name: "Stakeholder", password: "StrongPass!2026" }));
  expect(api.logout).toHaveBeenCalledOnce();
  expect((await screen.findByRole("link", { name: /continue to sign in/i })).getAttribute("href")).toBe("/login");
});
