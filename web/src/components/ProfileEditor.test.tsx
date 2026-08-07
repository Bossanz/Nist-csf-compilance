import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vitest";
import { ProfileEditor } from "./ProfileEditor";
import type { ProfileRow } from "../lib/types";

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
};

describe("ProfileEditor", () => {
  test("keeps assessments viewable but not editable in read-only mode", () => {
    render(<ProfileEditor rows={[row]} onSave={vi.fn()} readOnly />);

    const summary = screen.getByRole("button", { name: /GV\.OC-01/i });
    expect(summary.getAttribute("aria-expanded")).toBe("false");
    expect(summary.getAttribute("aria-controls")).toBe("assessment-body-profile-1");
    fireEvent.click(summary);

    expect((screen.getByRole("checkbox", { name: /include in profile/i }).closest("fieldset") as HTMLFieldSetElement).disabled).toBe(true);
    expect(screen.queryByRole("button", { name: /save assessment/i })).toBeNull();
  });

  test("saves the complete editable assessment instead of mutating the source row", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<ProfileEditor rows={[row]} onSave={onSave} />);

    fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
    fireEvent.click(screen.getByRole("checkbox", { name: /include in profile/i }));
    fireEvent.change(screen.getByLabelText(/rationale/i), { target: { value: "Critical business outcome" } });
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
      included: true,
      rationale: "Critical business outcome",
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

    expect(row.included).toBe(false);
    expect(row.currentCoverageLevel).toBe("none");
  });
});
