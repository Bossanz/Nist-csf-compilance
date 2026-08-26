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

test("uses controlled dropdowns for Current and Target priority", () => {
  render(<AssessmentCard row={row} onSave={vi.fn()} canEditScope={false} canEditProfile assigneeOptions={[]} />);

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  for (const label of ["Current priority", "Target priority"]) {
    const select = screen.getByRole("combobox", { name: label }) as HTMLSelectElement;
    expect(Array.from(select.options).map((option) => option.textContent)).toEqual([
      "Select priority",
      "Low",
      "Medium",
      "High",
    ]);
  }
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

  expect(screen.getByText("Reviewing")).toBeTruthy();
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
  expect(screen.getByRole("group", { name: /coverage: partial to substantial/i })).toBeTruthy();
  expect(screen.getByText("1 evidence file")).toBeTruthy();
  expect(summary.getAttribute("aria-expanded")).toBe("false");
});

test("keeps the full evidence count available when the summary must condense", () => {
  render(
    <AssessmentCard
      row={row}
      onSave={vi.fn()}
      canEditScope={false}
      canEditProfile={false}
      assigneeOptions={[]}
      response={{ ...submittedResponse, documents: [evidenceDocument] }}
    />
  );

  const evidenceCount = screen.getByText("1 evidence file");
  expect(evidenceCount.getAttribute("aria-label")).toBe("1 evidence file");
});

test("names the outcome article and its grouped form sections for assistive technology", () => {
  render(<AssessmentCard row={row} onSave={vi.fn()} canEditScope={true} canEditProfile={false} assigneeOptions={[]} />);

  const article = screen.getByRole("article", { name: /GV\.OC-01: The organizational mission is understood/i });
  expect(article).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));
  expect(screen.getByRole("group", { name: /scope and assignment/i })).toBeTruthy();
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

test("locks Current and Target profile fields while an outcome is under review", () => {
  render(
    <AssessmentCard
      row={{ ...row, assignedUserID: "assessor-1" }}
      onSave={vi.fn()}
      canEditScope={false}
      canEditProfile
      assigneeOptions={[]}
      role="assessor"
      response={submittedResponse}
    />,
  );

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  expect(screen.getByRole("heading", { name: /assessment profile/i })).toBeTruthy();
  expect(screen.queryByLabelText("Current priority")).toBeNull();
  expect(screen.queryByRole("button", { name: /save assessment/i })).toBeNull();
});

test("shows Counselor rationale to an Assessor as read-only context", () => {
  const rationale = "This outcome is in scope because the application handles regulated registration data.";

  render(
    <AssessmentCard
      row={{ ...row, rationale, assignedUserID: "assessor-1" }}
      onSave={vi.fn()}
      canEditScope={false}
      canEditProfile
      assigneeOptions={[]}
      role="assessor"
      scopeSubmitted
    />,
  );

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  const rationaleRegion = screen.getByRole("region", { name: /counselor rationale/i });
  expect(rationaleRegion.textContent).toContain(rationale);
  expect(screen.queryByRole("textbox", { name: /counselor rationale/i })).toBeNull();
});

test("Counselor draft hides profile and response work while scope is not submitted", () => {
  render(
    <AssessmentCard
      row={row}
      onSave={vi.fn()}
      canEditScope
      canEditProfile={false}
      assigneeOptions={[]}
      role="counselor"
      scopeSubmitted={false}
      response={submittedResponse}
      onSaveResponse={vi.fn().mockResolvedValue(undefined)}
      onSubmitResponse={vi.fn().mockResolvedValue(undefined)}
      onReviewResponse={vi.fn().mockResolvedValue(undefined)}
      onUploadEvidence={vi.fn().mockResolvedValue(undefined)}
      onDeleteEvidence={vi.fn().mockResolvedValue(undefined)}
      onDownloadEvidence={vi.fn().mockResolvedValue(undefined)}
    />,
  );

  fireEvent.click(screen.getByRole("button", { name: /GV\.OC-01/i }));

  expect(screen.getByText("Rationale")).toBeTruthy();
  expect(screen.getByLabelText("Responsible stakeholder for GV.OC-01")).toBeTruthy();
  expect(screen.queryByRole("heading", { name: /assessment profile/i })).toBeNull();
  expect(screen.queryByRole("heading", { name: /stakeholder response/i })).toBeNull();
});
