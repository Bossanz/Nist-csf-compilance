import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ActionPlan } from "./ActionPlan";
import type { ProfileRow, RemediationAction, StakeholderResponse, User } from "../lib/types";

const counselor: User = { id: "counselor-1", organizationID: null, name: "Counselor", email: "counselor@example.com", userType: "counselor", role: "counselor", status: "active" };
const assessor: User = { id: "assessor-1", organizationID: "org-1", name: "Assigned Assessor", email: "assessor@example.com", userType: "stakeholder", role: "assessor", status: "active" };
const reviewer: User = { ...assessor, id: "reviewer-1", name: "Reviewer", role: "reviewer" };

const profile: ProfileRow = {
  id: "profile-1", projectID: "project-1", subcategoryID: "subcategory-1", functionCode: "GV", categoryCode: "GV.OC", subcategoryCode: "GV.OC-01", description: "The organizational mission is understood.", included: true, rationale: "Required for governance alignment.", currentPriority: "medium", currentCoverageLevel: "partial", currentStatusText: "", currentPoliciesText: "", currentTier: "", targetPriority: "high", targetCoverageLevel: "full", targetApproachText: "", targetTier: "", notes: "", considerations: "", reviewStatus: "draft", assignedUserID: assessor.id,
};
const approvedResponse: StakeholderResponse = { id: "response-1", projectID: "project-1", subcategoryID: profile.subcategoryID, responseText: "Approved response", status: "reviewed", respondedBy: assessor.id, submittedAt: "2026-08-10T00:00:00Z", reviewComment: "Approved", reviewedBy: reviewer.id, reviewedAt: "2026-08-11T00:00:00Z", createdAt: "2026-08-09T00:00:00Z", updatedAt: "2026-08-11T00:00:00Z", documents: [] };
const action: RemediationAction = {
  id: "action-1", projectID: "project-1", subcategoryID: profile.subcategoryID, outcomeCode: profile.subcategoryCode, outcomeDescription: profile.description, currentCoverageLevel: "partial", targetCoverageLevel: "full", title: "Centralize security logs", description: "Forward application and API logs.", desiredResult: "Security events are searchable and retained.", priority: "high", ownerUserID: assessor.id, ownerName: assessor.name, ownerEmail: assessor.email, dueDate: "2020-09-30T00:00:00Z", status: "open", progressNote: "", reviewComment: "", createdBy: counselor.id, submittedAt: null, closedBy: null, closedAt: null, createdAt: "2026-08-12T00:00:00Z", updatedAt: "2026-08-12T00:00:00Z", evidence: [],
};
const callbacks = {
  onCreate: vi.fn().mockResolvedValue(undefined),
  onUpdate: vi.fn().mockResolvedValue(undefined),
  onSaveProgress: vi.fn().mockResolvedValue(undefined),
  onSubmit: vi.fn().mockResolvedValue(undefined),
  onReview: vi.fn().mockResolvedValue(undefined),
  onUploadEvidence: vi.fn().mockResolvedValue(undefined),
  onDeleteEvidence: vi.fn().mockResolvedValue(undefined),
  onDownloadEvidence: vi.fn().mockResolvedValue(undefined),
};

function renderPlan(user: User, actions: RemediationAction[] = [action]) {
  Object.values(callbacks).forEach((callback) => callback.mockClear());
  return render(<ActionPlan user={user} profile={[profile]} responses={[approvedResponse]} actions={actions} assigneeOptions={[assessor]} {...callbacks} />);
}

test("Counselor creates an action only from an approved coverage gap", async () => {
  renderPlan(counselor, []);
  expect(screen.getByText("1 approved gap ready for planning")).toBeTruthy();
  fireEvent.change(screen.getByLabelText("Outcome"), { target: { value: profile.subcategoryID } });
  fireEvent.change(screen.getByLabelText("Action title"), { target: { value: "Centralize security logs" } });
  fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Forward application and API logs." } });
  fireEvent.change(screen.getByLabelText("Desired result"), { target: { value: "Security events are searchable." } });
  fireEvent.change(screen.getByLabelText("Priority"), { target: { value: "high" } });
  fireEvent.change(screen.getByLabelText("Owner"), { target: { value: assessor.id } });
  fireEvent.change(screen.getByLabelText("Due date"), { target: { value: "2026-09-30" } });
  fireEvent.click(screen.getByRole("button", { name: "Create action" }));
  await waitFor(() => expect(callbacks.onCreate).toHaveBeenCalledWith(expect.objectContaining({ subcategoryID: profile.subcategoryID, title: "Centralize security logs", ownerUserID: assessor.id, priority: "high" })));
});

test("assigned Assessor saves progress before sending for review", async () => {
  renderPlan(assessor);
  expect(screen.getAllByText("Overdue").length).toBe(2);
  const submit = screen.getByRole("button", { name: "Send for review" });
  expect((submit as HTMLButtonElement).disabled).toBe(true);
  fireEvent.change(screen.getByLabelText("Progress update"), { target: { value: "SIEM forwarding is enabled in staging." } });
  fireEvent.click(screen.getByRole("button", { name: "Save progress" }));
  await waitFor(() => expect(callbacks.onSaveProgress).toHaveBeenCalledWith(action.id, "SIEM forwarding is enabled in staging."));
  expect((submit as HTMLButtonElement).disabled).toBe(false);
  fireEvent.click(submit);
  await waitFor(() => expect(callbacks.onSubmit).toHaveBeenCalledWith(action.id));
});

test("Reviewer has a read-only action view", () => {
  renderPlan(reviewer);
  expect(screen.getByText(action.title)).toBeTruthy();
  expect(screen.queryByRole("button", { name: "Save progress" })).toBeNull();
  expect(screen.queryByRole("button", { name: "Close action" })).toBeNull();
});

test("Counselor returns or closes an action awaiting review", async () => {
  renderPlan(counselor, [{ ...action, status: "awaiting_review", progressNote: "Production logging enabled." }]);
  const returnButton = screen.getByRole("button", { name: "Return for more work" });
  expect((returnButton as HTMLButtonElement).disabled).toBe(true);
  fireEvent.change(screen.getByLabelText("Counselor review"), { target: { value: "Attach the production deployment record." } });
  fireEvent.click(returnButton);
  await waitFor(() => expect(callbacks.onReview).toHaveBeenCalledWith(action.id, "return", "Attach the production deployment record."));
  fireEvent.click(screen.getByRole("button", { name: "Close action" }));
  await waitFor(() => expect(callbacks.onReview).toHaveBeenCalledWith(action.id, "close", "Attach the production deployment record."));
});
