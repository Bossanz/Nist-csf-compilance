import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { AssessmentCard } from "./AssessmentCard";
import type { ProfileRow, ResponseDocument, StakeholderResponse } from "../lib/types";

const row: ProfileRow = {
  id: "profile-1",
  projectID: "project-1",
  subcategoryID: "subcategory-1",
  functionCode: "GV",
  categoryCode: "GV.OC",
  subcategoryCode: "GV.OC-01",
  description: "The organizational mission is understood",
  included: true,
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

const submittedResponse: StakeholderResponse = {
  id: "response-1",
  projectID: "project-1",
  subcategoryID: "subcategory-1",
  responseText: "Evidence attached.",
  status: "submitted",
  respondedBy: "assessor-1",
  submittedAt: "2026-08-10T00:00:00Z",
  reviewComment: "",
  reviewedBy: null,
  reviewedAt: null,
  createdAt: "2026-08-09T00:00:00Z",
  updatedAt: "2026-08-10T00:00:00Z",
  documents: [],
};
const evidenceDocument: ResponseDocument = {
  id: "document-1",
  responseID: "response-1",
  originalName: "registration-evidence.pdf",
  mimeType: "application/pdf",
  sizeBytes: 2048,
  uploadedBy: "assessor-1",
  createdAt: "2026-08-10T00:00:00Z",
};

test("shows unsaved, saving, and saved assessment states", async () => {
  let resolveSave!: () => void;
  const savePromise = new Promise<void>((resolve) => { resolveSave = resolve; });
  const onSave = vi.fn().mockReturnValue(savePromise);

  render(<AssessmentCard row={row} onSave={onSave} canEditScope={false} canEditProfile assigneeOptions={[]} />);

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  expect(screen.getByText("Saved")).toBeTruthy();

  fireEvent.change(screen.getByLabelText("Current priority"), { target: { value: "High" } });
  expect(screen.getByText("Unsaved changes")).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /save assessment/i }));
  expect(screen.getByRole("status").textContent).toContain("Saving…");

  resolveSave();
  await waitFor(() => expect(screen.getByText("Saved")).toBeTruthy());
});

test("shows response status and assignee in the collapsed outcome summary", () => {
  render(
    <AssessmentCard
      row={{ ...row, assignedUserID: "assessor-1", assignedUserName: "Ari Assessor" }}
      onSave={vi.fn()}
      canEditScope
      canEditProfile={false}
      assigneeOptions={[]}
      response={submittedResponse}
    />,
  );

  expect(screen.getByText("Submitted")).toBeTruthy();
  expect(screen.getByText("Assigned to Ari Assessor")).toBeTruthy();
});

test("makes Current and Target coverage plus evidence count scannable in the collapsed summary", () => {
  render(
    <AssessmentCard
      row={{ ...row, currentCoverageLevel: "partial", targetCoverageLevel: "substantial" }}
      onSave={vi.fn()}
      canEditScope={false}
      canEditProfile={false}
      assigneeOptions={[]}
      response={{ ...submittedResponse, documents: [evidenceDocument] }}
    />
  );

  const summary = screen.getByRole("button", { name: /GV\.OC-01/i });
  expect(summary.textContent).toContain("Current");
  expect(summary.textContent).toContain("Partial");
  expect(summary.textContent).toContain("Target");
  expect(summary.textContent).toContain("Substantial");
  expect(screen.getByText("1 evidence file")).toBeTruthy();
  expect(summary.getAttribute("aria-expanded")).toBe("false");
});
test("Counselor can read an empty stakeholder response without seeing stakeholder controls", () => {
  render(
    <AssessmentCard
      row={row}
      onSave={vi.fn()}
      canEditScope
      canEditProfile={false}
      assigneeOptions={[]}
      role="counselor"
      response={{ ...submittedResponse, id: "", responseText: "", status: "draft", submittedAt: null, documents: [] }}
      onSaveResponse={vi.fn().mockResolvedValue(undefined)}
      onSubmitResponse={vi.fn().mockResolvedValue(undefined)}
      onReviewResponse={vi.fn().mockResolvedValue(undefined)}
      onUploadEvidence={vi.fn().mockResolvedValue(undefined)}
      onDeleteEvidence={vi.fn().mockResolvedValue(undefined)}
      onDownloadEvidence={vi.fn().mockResolvedValue(undefined)}
    />
  );

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  expect(screen.getByRole("heading", { name: /stakeholder response/i })).toBeTruthy();
  expect(screen.getAllByText("Read only").length).toBeGreaterThan(0);
  expect(screen.queryByRole("textbox", { name: /client response/i })).toBeNull();
  expect(screen.queryByRole("button", { name: /save response/i })).toBeNull();
  expect(screen.queryByLabelText(/upload evidence/i)).toBeNull();
});
