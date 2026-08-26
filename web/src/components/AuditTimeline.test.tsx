import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { AuditTimeline } from "./AuditTimeline";
import type { AuditTrailEntry } from "../lib/types";

const event: AuditTrailEntry = {
  id: "event-1",
  actorUserID: "reviewer-1",
  actorName: "Reviewer",
  actorEmail: "reviewer@example.com",
  actorRole: "reviewer",
  result: "success",
  requestID: "request-123",
  ipAddress: "192.0.2.10",
  userAgent: "audit-test/1.0",
  projectID: "project-1",
  action: "response.reviewed",
  entityType: "response",
  entityID: "response-1",
  metadata: {},
  createdAt: "2026-08-18T00:00:00Z",
};

test("renders readable activity details and trace metadata", () => {
  render(<AuditTimeline events={[event]} />);

  expect(screen.getByRole("heading", { name: "Activity trail" })).toBeTruthy();
  expect(screen.getByText("Response approved")).toBeTruthy();
  expect(screen.getByText(/Reviewer · reviewer · success/i)).toBeTruthy();
  expect(screen.getByText(/Request request-123/i)).toBeTruthy();
});

test("renders an empty state when there are no audit events", () => {
  render(<AuditTimeline events={[]} />);

  expect(screen.getByText("No audit activity has been recorded yet.")).toBeTruthy();
});

test("renders a non-blocking loading state for audit activity", () => {
  render(<AuditTimeline events={[]} loading />);

  expect(screen.getByText("Loading activity trail…").getAttribute("role")).toBe("status");
});

test("renders an audit activity error without hiding the workspace", () => {
  render(<AuditTimeline events={[]} error="Activity trail unavailable" />);

  expect(screen.getByRole("alert").textContent).toContain("Activity trail unavailable");
});
