import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vitest";
import { ProfileEditor } from "./ProfileEditor";
import type { ProfileRow, StakeholderResponse, User } from "../lib/types";

const row: ProfileRow = {
  id: "profile-1",
  projectID: "project-1",
  subcategoryID: "subcategory-1",
  functionCode: "GV",
  categoryCode: "GV.OC",
  subcategoryCode: "GV.OC-01",
  description: "The organizational mission is understood",
  included: false,
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
  assignedUserID: null,
};
const assessor: User = { id: "assessor-1", organizationID: "org-1", name: "Assessment Owner", email: "owner@example.com", userType: "stakeholder", role: "assessor", status: "active" };
const assignedRow: ProfileRow = { ...row, included: true, assignedUserID: assessor.id };
const response: StakeholderResponse = { id: "response-1", projectID: "project-1", subcategoryID: "subcategory-1", responseText: "", status: "draft", respondedBy: null, submittedAt: null, reviewComment: "", reviewedBy: null, reviewedAt: null, createdAt: "", updatedAt: "", documents: [] };

describe("ProfileEditor", () => {
  test("Counselor edits scope and assignee but cannot edit Current or Target", () => {
    render(<ProfileEditor rows={[row]} onSave={vi.fn()} role="counselor" canEditScope canEditProfile={false} assigneeOptions={[assessor]} />);

    fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

    expect((screen.getByRole("checkbox", { name: /include in profile/i }) as HTMLInputElement).disabled).toBe(false);
    expect((screen.getByLabelText(/responsible stakeholder/i) as HTMLSelectElement).disabled).toBe(false);
    expect(screen.queryByLabelText(/current priority/i)).toBeNull();
    expect(screen.getByRole("region", { name: /profile reference/i })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: /stakeholder response/i })).toBeNull();
  });

  test("Stakeholder sees profile input without counselor scope controls", () => {
    render(<ProfileEditor rows={[assignedRow]} onSave={vi.fn()} role="assessor" canEditScope={false} canEditProfile assigneeOptions={[]} />);

    fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

    expect(screen.getByLabelText(/current priority/i)).toBeTruthy();
    expect(screen.queryByRole("checkbox", { name: /include in profile/i })).toBeNull();
  });

  test("assigned Assessor edits Current and Target without sending scope fields", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<ProfileEditor rows={[assignedRow]} onSave={onSave} role="assessor" canEditScope={false} canEditProfile assigneeOptions={[]} />);

    fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
    fireEvent.change(screen.getByLabelText(/current priority/i), { target: { value: "High" } });
    fireEvent.click(screen.getByRole("button", { name: /save assessment/i }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("subcategory-1", expect.objectContaining({ currentPriority: "High" })));
    expect(onSave.mock.calls[0][1]).not.toHaveProperty("included");
    expect(onSave.mock.calls[0][1]).not.toHaveProperty("assignedUserID");
  });

  test("keeps assessments viewable but not editable in read-only mode", () => {
    render(<ProfileEditor rows={[row]} onSave={vi.fn()} role="viewer" canEditScope={false} canEditProfile={false} assigneeOptions={[]} />);

    const summary = screen.getByRole("button", { name: /GV\.OC-01/i });
    expect(summary.getAttribute("aria-expanded")).toBe("false");
    expect(summary.getAttribute("aria-controls")).toBe("assessment-body-profile-1");
    fireEvent.click(summary);

    expect(screen.getByRole("region", { name: /profile reference/i })).toBeTruthy();
    expect(screen.queryByRole("checkbox", { name: /include in profile/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /save assessment/i })).toBeNull();
  });

  test("explains when an Assessor has no assigned outcomes", () => {
    render(<ProfileEditor rows={[]} onSave={vi.fn()} role="assessor" canEditScope={false} canEditProfile assigneeOptions={[]} />);

    expect(screen.getByText("No included outcomes are assigned to you in this Function.")).toBeTruthy();
  });

  test("saves the editable profile fields instead of mutating the source row", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<ProfileEditor rows={[assignedRow]} onSave={onSave} role="assessor" canEditScope={false} canEditProfile assigneeOptions={[]} />);

    fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
    fireEvent.change(screen.getByLabelText(/current priority/i), { target: { value: "High" } });
    fireEvent.change(screen.getByLabelText(/current activities/i), { target: { value: "Quarterly review" } });
    fireEvent.change(screen.getByLabelText(/policies and procedures/i), { target: { value: "Risk policy" } });
    fireEvent.change(screen.getByLabelText(/current coverage/i), { target: { value: "partial" } });
    fireEvent.change(screen.getByLabelText(/target priority/i), { target: { value: "High" } });
    fireEvent.change(screen.getByLabelText(/target coverage/i), { target: { value: "full" } });
    fireEvent.change(screen.getByLabelText(/target approach/i), { target: { value: "Formalize governance review" } });
    fireEvent.change(screen.getByLabelText(/^notes$/i), { target: { value: "Owner to confirm" } });
    fireEvent.change(screen.getByLabelText(/considerations/i), { target: { value: "Align with ERM" } });
    fireEvent.click(screen.getByRole("button", { name: /save assessment/i }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("subcategory-1", {
      currentPriority: "High",
      currentCoverageLevel: "partial",
      currentStatusText: "Quarterly review",
      currentPoliciesText: "Risk policy",
      targetPriority: "High",
      targetCoverageLevel: "full",
      targetApproachText: "Formalize governance review",
      notes: "Owner to confirm",
      considerations: "Align with ERM",
    }));

    expect(assignedRow.included).toBe(true);
    expect(assignedRow.currentCoverageLevel).toBe("none");
  });

  test("does not scan the response list once per outcome", () => {
    const rows = [row, { ...row, id: "profile-2", subcategoryID: "subcategory-2", subcategoryCode: "GV.OC-02" }];
    const responses = [response];
    const responseFind = vi.spyOn(responses, "find");

    render(<ProfileEditor rows={rows} onSave={vi.fn()} canEditScope={false} canEditProfile={false} assigneeOptions={[]} responses={responses} />);

    try {
      expect(responseFind).not.toHaveBeenCalled();
    } finally {
      responseFind.mockRestore();
    }
  });
});
