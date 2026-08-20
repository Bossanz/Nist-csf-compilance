import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { InvitationList } from "./InvitationList";
import type { Invitation, Project } from "../lib/types";

const project: Project = { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "RU Registration", slug: "ru-registration", status: "setup", createdAt: "2026-08-20T00:00:00Z" };
const pending: Invitation = { id: "invite-1", organizationID: "org-1", email: "auditor@example.com", userType: "stakeholder", role: "auditor", expiresAt: "2026-08-23T00:00:00Z", acceptedAt: null, cancelledAt: null, supersededAt: null, status: "pending", projectIDs: [project.id], invitationURL: "" };
const expired: Invitation = { ...pending, id: "invite-2", email: "expired@example.com", expiresAt: "2026-08-19T00:00:00Z", status: "expired", projectIDs: [] };
const accepted: Invitation = { ...pending, id: "invite-3", email: "accepted@example.com", status: "accepted", acceptedAt: "2026-08-18T00:00:00Z" };

test("renders invitation lifecycle and project scope controls", () => {
  const onResend = vi.fn();
  const onCancel = vi.fn();
  render(<InvitationList invitations={[pending, expired, accepted]} projects={[project]} busyInvitationID="" canManageLifecycle={true} onResend={onResend} onCancel={onCancel} />);

  expect(screen.getByText("auditor@example.com")).toBeTruthy();
  expect(screen.getAllByText(/RU Registration/)).toHaveLength(2);
  expect(screen.getByText("pending")).toBeTruthy();
  expect(screen.getByText("expired")).toBeTruthy();
  expect(screen.getByText("accepted")).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /resend invitation for auditor@example.com/i }));
  fireEvent.click(screen.getByRole("button", { name: /cancel invitation for auditor@example.com/i }));
  fireEvent.click(screen.getByRole("button", { name: /resend invitation for expired@example.com/i }));

  expect(onResend).toHaveBeenCalledWith(pending);
  expect(onResend).toHaveBeenCalledWith(expired);
  expect(onCancel).toHaveBeenCalledWith(pending);
});

test("renders a regular invitation when the API omits project scope", () => {
  const invitationWithoutScope = {
    ...pending,
    email: "viewer@example.com",
    role: "viewer",
    projectIDs: undefined,
  } as unknown as Invitation;

  expect(() => render(<InvitationList invitations={[invitationWithoutScope]} projects={[project]} busyInvitationID="" canManageLifecycle={false} onResend={vi.fn()} onCancel={vi.fn()} />)).not.toThrow();
  expect(screen.getByText("viewer@example.com")).toBeTruthy();
});

test("hides lifecycle controls when the user cannot manage invitations", () => {
  render(<InvitationList invitations={[pending, expired]} projects={[project]} busyInvitationID="" canManageLifecycle={false} onResend={vi.fn()} onCancel={vi.fn()} />);

  expect(screen.queryByRole("button", { name: /resend invitation/i })).toBeNull();
  expect(screen.queryByRole("button", { name: /cancel invitation/i })).toBeNull();
});
