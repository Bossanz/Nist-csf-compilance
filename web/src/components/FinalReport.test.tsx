import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import type { FinalReportData, ProfileRow } from "../lib/types";
import { FinalReport } from "./FinalReport";

const profile: ProfileRow = {
  id: "profile-1", projectID: "project-1", subcategoryID: "subcategory-1", functionCode: "GV", categoryCode: "GV.OC", subcategoryCode: "GV.OC-01", description: "Mission informs risk management", included: true, rationale: "Relevant to the project", currentPriority: "High", currentCoverageLevel: "partial", currentStatusText: "Quarterly review", currentPoliciesText: "Risk policy", currentTier: "", targetPriority: "High", targetCoverageLevel: "full", targetApproachText: "Formalize the review", targetTier: "", notes: "", considerations: "", reviewStatus: "approved", assignedUserID: "assessor-1", assignedUserName: "Assessor", assignedUserEmail: "assessor@example.com",
};

const report: FinalReportData = {
  project: { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "Readiness Review", slug: "readiness-review", status: "closed", createdAt: "2026-08-18T00:00:00Z", objective: "Assess readiness", assessmentPeriod: "Q3 2026", targetCompletionDate: "2026-09-30", scopeBoundary: "Production systems", complianceDriver: "Customer assurance", finalizedAt: "2026-08-18T10:00:00Z", finalizedBy: "counselor-1" },
  summary: { coveragePct: 50, includedCount: 1, approvedCount: 1, reviewingCount: 0, returnedCount: 0, pendingCount: 0, evidenceCount: 1, functions: [{ code: "GV", coveragePct: 50, includedCount: 1, approvedCount: 1, reviewingCount: 0, returnedCount: 0, pendingCount: 0, evidenceCount: 1 }] },
  outcomes: [{ profile, response: { id: "response-1", responseText: "Reviewed quarterly.", status: "reviewed", respondedBy: "assessor-1", submittedAt: "2026-08-17T00:00:00Z", reviewComment: "Evidence is sufficient.", reviewedBy: "reviewer-1", reviewedAt: "2026-08-18T00:00:00Z" }, evidence: [{ id: "evidence-1", originalName: "risk-review.pdf", mimeType: "application/pdf", sizeBytes: 1024, uploadedBy: "assessor-1", createdAt: "2026-08-17T00:00:00Z" }] }],
  remediationSummary: { openCount: 0, inProgressCount: 1, awaitingReviewCount: 0, overdueCount: 0, closedCount: 0 },
  remediationActions: [{ id: "action-1", projectID: "project-1", subcategoryID: "subcategory-1", outcomeCode: "GV.OC-01", outcomeDescription: profile.description, currentCoverageLevel: "partial", targetCoverageLevel: "full", title: "Centralize security logs", description: "Forward logs to the SIEM.", desiredResult: "Searchable events.", priority: "high", ownerUserID: "assessor-1", ownerName: "Assessor", ownerEmail: "assessor@example.com", dueDate: "2026-09-30T00:00:00Z", status: "in_progress", progressNote: "Staging configured.", reviewComment: "", createdBy: "counselor-1", submittedAt: null, closedBy: null, closedAt: null, createdAt: "2026-08-18T00:00:00Z", updatedAt: "2026-08-18T00:00:00Z", evidence: [] }],
};

afterEach(() => vi.restoreAllMocks());

test("renders finalized project results and evidence", () => {
  render(<FinalReport report={report} onBack={vi.fn()} onOpenAudit={vi.fn()} />);

  expect(screen.getByRole("heading", { name: "Readiness Review" })).toBeTruthy();
  expect(screen.getByText("Finalized")).toBeTruthy();
  expect(screen.getAllByText("GV.OC-01")).toHaveLength(2);
  expect(screen.getByText("risk-review.pdf")).toBeTruthy();
  expect(screen.getAllByText("50%", { exact: false }).length).toBeGreaterThan(0);
  expect(screen.getByRole("heading", { name: "Remediation progress" })).toBeTruthy();
  expect(screen.getByText("Centralize security logs")).toBeTruthy();
});

test("prints the final report and opens the audit package", () => {
  const print = vi.spyOn(window, "print").mockImplementation(() => undefined);
  const onOpenAudit = vi.fn();
  render(<FinalReport report={report} onBack={vi.fn()} onOpenAudit={onOpenAudit} />);

  fireEvent.click(screen.getByRole("button", { name: /print.*save as pdf/i }));
  fireEvent.click(screen.getByRole("button", { name: /open audit package/i }));

  expect(print).toHaveBeenCalled();
  expect(onOpenAudit).toHaveBeenCalled();
});
