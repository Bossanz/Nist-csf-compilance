import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import LoginPage from "./page";
import { api } from "../../lib/api";

const router = vi.hoisted(() => ({ replace: vi.fn() }));

vi.mock("next/navigation", () => ({ useRouter: () => router }));
vi.mock("../../lib/api", () => ({
  APIError: class APIError extends Error {
    constructor(message: string, public status: number) { super(message); }
  },
  api: { me: vi.fn(), login: vi.fn() },
}));

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(api.me).mockRejectedValue(new Error("Authentication required"));
  vi.mocked(api.login).mockResolvedValue({} as never);
});

test("redirects to organizations after a successful login", async () => {
  render(<LoginPage />);
  await screen.findByRole("heading", { name: /sign in/i });
  fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "admin@example.com" } });
  fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "secret-password" } });
  fireEvent.click(screen.getByRole("button", { name: /sign in/i }));

  await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/organizations"));
});

test("uses the shared ellipsis while checking authentication", () => {
  vi.mocked(api.me).mockReturnValue(new Promise<never>(() => {}));

  render(<LoginPage />);

  expect(screen.getByRole("status").textContent).toContain("Loading…");
});
