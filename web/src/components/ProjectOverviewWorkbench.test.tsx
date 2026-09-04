import { fireEvent, render, screen, within } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ProjectOverviewWorkbench } from "./ProjectOverviewWorkbench";
import type { FunctionNode, Organization, Project, Summary, User } from "../lib/types";
import type { FunctionProgress } from "./FunctionSidebar";

const organization: Organization = { id: "org-1", name: "Acme", slug: "acme", type: "client" };
const project: Project = { id: "project-1", organizationID: organization.id, organizationName: organization.name, name: "Readiness", slug: "readiness", status: "in_review", createdAt: "2026-08-10T00:00:00Z" };
const functions: FunctionNode[] = [
  { id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] },
  { id: "function-2", code: "PR", name: "Protect", description: "Protection", categories: [] },
];
const summary: Summary = { coveragePct: 42, includedCount: 4, pendingCount: 1, rejectedCount: 0, functions: [{ code: "GV", coveragePct: 60, includedCount: 2 }, { code: "PR", coveragePct: 25, includedCount: 2 }] };
const counselor: User = { id: "counselor-1", organizationID: null, name: "Counselor", email: "counselor@example.com", userType: "counselor", role: "counselor", status: "active" };
const assessor: User = { id: "assessor-1", organizationID: organization.id, name: "Assessor", email: "assessor@example.com", userType: "stakeholder", role: "assessor", status: "active" };
const progressByFunction: Record<string, FunctionProgress> = {
  GV: { value: 2, label: "included", coveragePct: 60, includedCount: 2, attention: 1, attentionLabel: "unassigned" },
  PR: { value: 2, label: "included", coveragePct: 25, includedCount: 2, attention: 0, attentionLabel: "unassigned" },
};

test("organizes project posture, next action, and Function progress", () => {
  const onOpenAssignment = vi.fn();

  render(<ProjectOverviewWorkbench user={counselor} project={project} summary={summary} functions={functions} functionProgress={progressByFunction} scopeSubmitted={false} isCounselor onOpenAssignment={onOpenAssignment} />);

  expect(within(screen.getByRole("region", { name: "Project posture" })).getByText("42%")).toBeTruthy();
  expect(screen.getByRole("region", { name: "Next up" }).textContent).toContain("Set and submit project scope");
  expect(screen.getByRole("region", { name: /function register/i }).textContent).toContain("Govern");
  expect(screen.getByRole("button", { name: /open assignment from next up/i })).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /open assignment from next up/i }));
  expect(onOpenAssignment).toHaveBeenCalledOnce();
});

test("gives an Assessor a role-aware next action", () => {
  render(<ProjectOverviewWorkbench user={assessor} project={project} summary={summary} functions={functions} functionProgress={progressByFunction} scopeSubmitted isCounselor={false} onOpenAssignment={vi.fn()} />);

  const nextUp = screen.getByRole("region", { name: "Next up" });
  expect(nextUp.textContent).toContain("Needs your input");
  expect(nextUp.textContent).toContain("Open your assigned outcomes and complete the response and evidence fields.");
});
