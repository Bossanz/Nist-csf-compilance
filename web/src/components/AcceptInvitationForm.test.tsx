import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { AcceptInvitationForm } from "./AcceptInvitationForm";

test("accepts an invitation with name and password", async () => {
  const onAccept = vi.fn().mockResolvedValue(undefined);
  render(<AcceptInvitationForm loading={false} error="" onAccept={onAccept} />);

  fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "Jane Stakeholder" } });
  fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: "StrongPass!2026" } });
  fireEvent.click(screen.getByRole("button", { name: /activate account/i }));

  await waitFor(() => expect(onAccept).toHaveBeenCalledWith({ name: "Jane Stakeholder", password: "StrongPass!2026" }));
});
