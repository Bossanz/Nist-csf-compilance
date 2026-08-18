import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ChangePasswordForm } from "./ChangePasswordForm";

test("submits current and new passwords", () => {
  const onSubmit = vi.fn();
  render(<ChangePasswordForm loading={false} error="" onSubmit={onSubmit} />);
  fireEvent.change(screen.getByLabelText(/current password/i), { target: { value: "old-password" } });
  fireEvent.change(screen.getByLabelText(/^new password$/i), { target: { value: "new-password" } });
  fireEvent.change(screen.getByLabelText(/confirm new password/i), { target: { value: "new-password" } });
  fireEvent.click(screen.getByRole("button", { name: /change password/i }));
  expect(onSubmit).toHaveBeenCalledWith("old-password", "new-password");
});
