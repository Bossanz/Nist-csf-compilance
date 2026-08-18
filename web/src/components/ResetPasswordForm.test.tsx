import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ResetPasswordForm } from "./ResetPasswordForm";

test("submits the reset token and new password", () => {
  const onSubmit = vi.fn();
  render(<ResetPasswordForm token="reset-token" loading={false} error="" onSubmit={onSubmit} />);
  fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: "new-password" } });
  fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: "new-password" } });
  fireEvent.click(screen.getByRole("button", { name: /reset password/i }));
  expect(onSubmit).toHaveBeenCalledWith("reset-token", "new-password");
});
