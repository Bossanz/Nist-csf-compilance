import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ForgotPasswordForm } from "./ForgotPasswordForm";

test("submits a normalized email for password recovery", () => {
  const onSubmit = vi.fn();
  render(<ForgotPasswordForm loading={false} error="" submitted={false} onSubmit={onSubmit} />);
  fireEvent.change(screen.getByLabelText(/email/i), { target: { value: " Person@example.com " } });
  fireEvent.click(screen.getByRole("button", { name: /send reset link/i }));
  expect(onSubmit).toHaveBeenCalledWith("person@example.com");
});
