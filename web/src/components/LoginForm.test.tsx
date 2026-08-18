import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { LoginForm } from "./LoginForm";

test("submits normalized login credentials", () => {
  const onSubmit = vi.fn();
  render(<LoginForm loading={false} error="" onSubmit={onSubmit} />);

  fireEvent.change(screen.getByLabelText(/email/i), { target: { value: " Admin@example.com " } });
  fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "secret-password" } });
  fireEvent.click(screen.getByRole("button", { name: /sign in/i }));

  expect(onSubmit).toHaveBeenCalledWith({ email: "admin@example.com", password: "secret-password" });
});

test("provides a password recovery link", () => {
  render(<LoginForm loading={false} error="" onSubmit={vi.fn()} />);
  expect(screen.getByRole("link", { name: /forgot password/i }).getAttribute("href")).toBe("/forgot-password");
});
