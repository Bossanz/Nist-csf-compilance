import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ProjectAssessmentWorkspace } from "./ProjectAssessmentWorkspace";
import type { FunctionNode, Organization, ProfileRow, Project, StakeholderResponse, Summary, User } from "../lib/types";

const organization: Organization = { id: "org-1", name: "Acme", slug: "acme", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "Readiness", slug: "readiness", status: "in_progress", createdAt: "2026-08-10T00:00:00Z", objective: "Prepare the organization for registration.", assessmentPeriod: "Q3 2026", targetCompletionDate: "2026-09-30", scopeBoundary: "Thailand operations", complianceDriver: "Customer assurance" };
const functions: FunctionNode[] = [{ id: "function-1", code: "GV", name: "Govern", description: "Governance", categories: [] }];
const summary: Summary = { coveragePct: 0, includedCount: 2, pendingCount: 2, rejectedCount: 0, functions: [] };
const assessor: User = { id: "assessor-1", organizationID: "org-1", name: "Assigned Assessor", email: "assessor@example.com", userType: "stakeholder", role: "assessor", status: "active" };
const otherAssessor: User = { ...assessor, id: "assessor-2", name: "Other Assessor" };
const counselor: User = { id: "counselor-1", organizationID: null, name: "Counselor", email: "counselor@example.com", userType: "counselor", role: "counselor", status: "active" };
const reviewer: User = { ...assessor, id: "reviewer-1", name: "Reviewer", role: "reviewer" };
const viewer: User = { ...assessor, id: "viewer-1", name: "Viewer", role: "viewer" };
const submittedResponse: StakeholderResponse = {
  id: "response-1",
  projectID: "project-1",
  subcategoryID: "GV.OC-01",
  responseText: "Evidence attached.",
  status: "submitted",
  respondedBy: assessor.id,
  submittedAt: "2026-08-10T00:00:00Z",
  reviewComment: "",
  reviewedBy: null,
  reviewedAt: null,
  createdAt: "2026-08-09T00:00:00Z",
  updatedAt: "2026-08-10T00:00:00Z",
  documents: [],
};

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
  row("GV.OC-02", true, null),
  row("GV.OC-03", false, null),
];
const noop = vi.fn().mockResolvedValue(undefined);

function renderWorkspace(user: User, onSetFunctionIncluded = noop, profileRows = profile, responseRows: StakeholderResponse[] = [], projectOverride: Project = project, onSubmitScope = noop) {
  return render(
    <ProjectAssessmentWorkspace
      user={user}
      organization={organization}
      project={projectOverride}
      functions={functions}
      organizationUsers={[assessor, otherAssessor]}
      profile={profileRows}
      responses={responseRows}
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
      onSubmitScope={onSubmitScope}
    />,
  );
}

function visibleOutcomeCount() {
  return document.querySelector(".outcome-count strong")?.textContent;
}

test("Counselor sees every outcome for scope configuration", () => {
  renderWorkspace(counselor);
  expect(visibleOutcomeCount()).toBe("3");
  expect(screen.getByRole("paragraph", { name: "Active role mode" }).textContent).toContain("Scope & Assignment");
  expect(screen.getAllByText("Included in profile").length).toBe(2);
  expect(screen.getByText("Out of scope")).toBeTruthy();
  expect(screen.getByRole("button", { name: /gv govern.*1 unassigned/i })).toBeTruthy();
});

test("assigned Assessor sees only assigned included outcomes", () => {
  renderWorkspace(assessor);
  expect(visibleOutcomeCount()).toBe("1");
  expect(screen.getByRole("paragraph", { name: "Active role mode" }).textContent).toContain("My Work");
  expect(screen.getByRole("button", { name: /gv govern.*1 open/i })).toBeTruthy();
  expect(screen.getByRole("button", { name: /GV\.OC-01/i })).toBeTruthy();
  expect(screen.queryByRole("button", { name: /GV\.OC-02/i })).toBeNull();
});

test("explains the assigned Assessor next step and selected Function", () => {
  renderWorkspace(assessor);
  expect(screen.getByText("Complete your assigned outcomes and attach supporting evidence.")).toBeTruthy();
  expect(screen.getByText("Function: GV — Govern")).toBeTruthy();
});

test("shows the shared project context and active Function progress", () => {
  renderWorkspace(assessor);

  const context = screen.getByRole("region", { name: "Project context" });
  expect(context.textContent).toContain("Prepare the organization for registration.");
  expect(context.textContent).toContain("Q3 2026");
  expect(context.textContent).toContain("2026");
  expect(context.textContent).toContain("Thailand operations");
  expect(context.textContent).toContain("Customer assurance");
  expect(context.textContent).toContain("in progress");
  expect(context.textContent).not.toContain("Overall coverage");
  expect(context.textContent).toContain("GV — Govern");
  expect(context.textContent).toContain("1 included outcome");
  expect(screen.getByRole("button", { name: /gv govern.*0% 2 included/i })).toBeTruthy();
  const summaryRegion = screen.getByRole("region", { name: /assessment workflow summary/i });
  expect(summaryRegion.textContent).toContain("Overall coverage");
  expect(summaryRegion.textContent).toContain("0%");
});

test("does not render empty optional project metadata", () => {
  const projectWithoutMetadata: Project = { ...project, objective: undefined, assessmentPeriod: undefined, targetCompletionDate: undefined, scopeBoundary: undefined, complianceDriver: undefined };
  renderWorkspace(assessor, noop, profile, [], projectWithoutMetadata);

  const context = screen.getByRole("region", { name: "Project context" });
  expect(context.textContent).not.toContain("Objective");
  expect(context.textContent).not.toContain("Assessment period");
  expect(context.textContent).not.toContain("Target completion");
  expect(context.textContent).not.toContain("Scope boundary");
  expect(context.textContent).not.toContain("Compliance driver");
});

test("Reviewer sees every included outcome", () => {
  renderWorkspace(reviewer);
  expect(visibleOutcomeCount()).toBe("2");
  expect(screen.getByRole("paragraph", { name: "Active role mode" }).textContent).toContain("Review Queue");
});

test("shows submitted work in the selected Function navigation", () => {
  renderWorkspace(reviewer, noop, profile, [submittedResponse]);

  expect(screen.getByRole("button", { name: /gv govern.*2 included.*1 to review/i })).toBeTruthy();
});

test("Viewer sees included outcomes without an action count or mutation affordance", () => {
  renderWorkspace(viewer, noop, profile, [submittedResponse]);

  expect(visibleOutcomeCount()).toBe("2");
  expect(document.querySelector(".nav-meta em")).toBeNull();
  expect(screen.queryByText("to review")).toBeNull();
  expect(screen.queryByRole("checkbox", { name: /include all outcomes in this function/i })).toBeNull();

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  expect(screen.getAllByText("Read only").length).toBeGreaterThan(0);
  expect(screen.queryByText("Include in profile")).toBeNull();
  expect(screen.queryByRole("button", { name: /save assessment/i })).toBeNull();
});

test("explains when a stakeholder has no assigned queue", () => {
  renderWorkspace(assessor, noop, profile.map((item) => ({ ...item, assignedUserID: otherAssessor.id })));

  expect(screen.getByRole("status").textContent).toContain("No included outcomes are assigned to you in this Function.");
});

test("Counselor can include all outcomes in the selected Function", async () => {
  const onSetFunctionIncluded = vi.fn().mockResolvedValue(undefined);
  renderWorkspace(counselor, onSetFunctionIncluded);

  const toggle = screen.getByRole("checkbox", { name: /include all outcomes in this function/i }) as HTMLInputElement;
  expect(toggle.checked).toBe(false);
  fireEvent.click(toggle);

  await waitFor(() => expect(onSetFunctionIncluded).toHaveBeenCalledWith("GV", true));
});

test("Counselor sees assignment progress for included outcomes", () => {
  renderWorkspace(counselor, noop, [...profile, row("GV.OC-04", true, null)]);

  const progress = screen.getByRole("region", { name: /assignment progress/i });
  expect(progress.textContent).toContain("3");
  expect(progress.textContent).toContain("2");
  expect(progress.textContent).toContain("1");
});

test("stakeholder roles do not see the bulk include control", () => {
  renderWorkspace(assessor);
  expect(screen.queryByRole("checkbox", { name: /include all outcomes in this function/i })).toBeNull();
  expect(screen.queryByRole("region", { name: /assignment progress/i })).toBeNull();
});

test("names the scope assignee control for its outcome", () => {
  renderWorkspace(counselor);
  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  expect(screen.getByLabelText("Responsible stakeholder for GV.OC-01")).toBeTruthy();
});
test("Counselor draft shows scope controls but hides stakeholder work", () => {
  const onSubmitScope = vi.fn().mockResolvedValue(undefined);
  renderWorkspace(counselor, noop, profile, [submittedResponse], { ...project, status: "setup" }, onSubmitScope);

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  expect(screen.getByText("Rationale")).toBeTruthy();
  expect(screen.getByLabelText("Responsible stakeholder for GV.OC-01")).toBeTruthy();
  expect(screen.queryByRole("heading", { name: /assessment profile/i })).toBeNull();
  expect(screen.queryByRole("heading", { name: /stakeholder response/i })).toBeNull();
  expect(screen.getByRole("button", { name: /submit scope/i })).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /submit scope/i }));
  expect(onSubmitScope).toHaveBeenCalledOnce();
});

test("Assessor sees no outcomes before Counselor submits the scope", () => {
  renderWorkspace(assessor, noop, profile, [], { ...project, status: "setup" });

  expect(visibleOutcomeCount()).toBe("0");
  expect(screen.queryByRole("button", { name: /GV\.OC-01/i })).toBeNull();
});

test("Counselor can read stakeholder work after scope submission", () => {
  renderWorkspace(counselor, noop, profile, [submittedResponse], { ...project, status: "in_review" });

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  expect(screen.getByRole("heading", { name: /stakeholder response/i })).toBeTruthy();
  expect(screen.queryByRole("button", { name: /submit scope/i })).toBeNull();
});
