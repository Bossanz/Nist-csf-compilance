import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ProjectAssessmentWorkspace } from "./ProjectAssessmentWorkspace";
import type { FunctionNode, Organization, ProfileRow, Project, Summary, User } from "../lib/types";

const organization: Organization = { id: "org-1", name: "Acme", slug: "acme", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-10T00:00:00Z" };
const functions: FunctionNode[] = [{ id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] }];
const summary: Summary = { coveragePct: 0, includedCount: 2, pendingCount: 2, rejectedCount: 0, functions: [] };
const assessor: User = { id: "assessor-1", organizationID: "org-1", name: "Assigned Assessor", email: "assessor@example.com", userType: "stakeholder", role: "assessor", status: "active" };
const otherAssessor: User = { ...assessor, id: "assessor-2", name: "Other Assessor" };
const counselor: User = { id: "counselor-1", organizationID: null, name: "Counselor", email: "counselor@example.com", userType: "counselor", role: "counselor", status: "active" };
const reviewer: User = { ...assessor, id: "reviewer-1", name: "Reviewer", role: "reviewer" };

function row(id: string, included: boolean, assignedUserID: string | null): ProfileRow {
  return {
    id,
    projectID: "project-1",
    subcategoryID: id,
    functionCode: "GV",
    categoryCode: "GV.OC",
    subcategoryCode: id,
    description: id,
    included,
    rationale: "",
    currentPriority: "",
    currentCoverageLevel: "none",
    currentStatusText: "",
    currentPoliciesText: "",
    currentTier: "",
    targetPriority: "",
    targetCoverageLevel: "none",
    targetApproachText: "",
    targetTier: "",
    notes: "",
    considerations: "",
    reviewStatus: "draft",
    assignedUserID,
  };
}

const profile: ProfileRow[] = [
  row("GV.OC-01", true, assessor.id),
  row("GV.OC-02", true, otherAssessor.id),
  row("GV.OC-03", false, null),
];
const noop = vi.fn().mockResolvedValue(undefined);

function renderWorkspace(user: User, onSetFunctionIncluded = noop) {
  return render(
    <ProjectAssessmentWorkspace
      user={user}
      organization={organization}
      project={project}
      functions={functions}
      organizationUsers={[assessor, otherAssessor]}
      profile={profile}
      responses={[]}
      summary={summary}
      selectedCode="GV"
      error=""
      onBack={noop}
      onSelectFunction={noop}
      onSaveProfile={noop}
      onSaveResponse={noop}
      onSubmitResponse={noop}
      onReviewResponse={noop}
      onUploadEvidence={noop}
      onDeleteEvidence={noop}
      onDownloadEvidence={noop}
      onSetFunctionIncluded={onSetFunctionIncluded}
    />,
  );
}

function visibleOutcomeCount() {
  const label = screen.getByText("outcomes");
  return label.parentElement?.querySelector("strong")?.textContent;
}

test("Counselor sees every outcome for scope configuration", () => {
  renderWorkspace(counselor);
  expect(visibleOutcomeCount()).toBe("3");
});

test("assigned Assessor sees only assigned included outcomes", () => {
  renderWorkspace(assessor);
  expect(visibleOutcomeCount()).toBe("1");
});

test("Reviewer sees every included outcome", () => {
  renderWorkspace(reviewer);
  expect(visibleOutcomeCount()).toBe("2");
});

test("Counselor can include all outcomes in the selected Function", async () => {
  const onSetFunctionIncluded = vi.fn().mockResolvedValue(undefined);
  renderWorkspace(counselor, onSetFunctionIncluded);

  const toggle = screen.getByRole("checkbox", { name: /include all outcomes in this function/i }) as HTMLInputElement;
  expect(toggle.checked).toBe(false);
  fireEvent.click(toggle);

  await waitFor(() => expect(onSetFunctionIncluded).toHaveBeenCalledWith("GV", true));
});

test("stakeholder roles do not see the bulk include control", () => {
  renderWorkspace(assessor);
  expect(screen.queryByRole("checkbox", { name: /include all outcomes in this function/i })).toBeNull();
});
