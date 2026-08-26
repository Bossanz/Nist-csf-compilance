import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import type { AuditPackageData, ProfileRow } from "../lib/types";
import { AuditPackageView } from "./AuditPackageView";

const profile: ProfileRow = {
  id: "profile-1", projectID: "project-1", subcategoryID: "subcategory-1", functionCode: "GV", categoryCode: "GV.OC", subcategoryCode: "GV.OC-01", description: "Mission informs risk management", included: true, rationale: "Relevant", currentPriority: "High", currentCoverageLevel: "partial", currentStatusText: "", currentPoliciesText: "", currentTier: "", targetPriority: "High", targetCoverageLevel: "full", targetApproachText: "Formalize", targetTier: "", notes: "", considerations: "", reviewStatus: "approved", assignedUserID: "assessor-1", assignedUserName: "Assessor", assignedUserEmail: "assessor@example.com",
};

const auditPackage: AuditPackageData = {
  project: { id: "project-1", organizationID: "org-1", organizationName: "Acme", name: "Readiness Review", slug: "readiness-review", status: "closed", createdAt: "2026-08-18T00:00:00Z", versionNumber: 2, previousVersionID: "project-previous", isLatest: true },
  summary: { coveragePct: 50, includedCount: 1, approvedCount: 1, reviewingCount: 0, returnedCount: 0, pendingCount: 0, evidenceCount: 1, functions: [] },
  scope: [{ profile }],
  outcomes: [{ profile, response: { id: "response-1", responseText: "Reviewed quarterly.", status: "reviewed", respondedBy: "assessor-1", submittedAt: "2026-08-17T00:00:00Z", reviewComment: "Approved.", reviewedBy: "reviewer-1", reviewedAt: "2026-08-18T00:00:00Z" }, evidence: [{ id: "evidence-1", originalName: "risk-review.pdf", mimeType: "application/pdf", sizeBytes: 1024, uploadedBy: "assessor-1", createdAt: "2026-08-17T00:00:00Z" }] }],
  auditTrail: [{ id: "event-1", actorUserID: "reviewer-1", actorName: "Reviewer", actorEmail: "reviewer@example.com", action: "response.reviewed", entityType: "response", entityID: "response-1", metadata: {}, createdAt: "2026-08-18T00:00:00Z" }],
  remediationSummary: { openCount: 0, inProgressCount: 1, awaitingReviewCount: 0, overdueCount: 0, closedCount: 0 },
  remediationActions: [{ id: "action-1", projectID: "project-1", subcategoryID: "subcategory-1", outcomeCode: "GV.OC-01", outcomeDescription: profile.description, currentCoverageLevel: "partial", targetCoverageLevel: "full", title: "Centralize security logs", description: "Forward logs to the SIEM.", desiredResult: "Searchable events.", priority: "high", ownerUserID: "assessor-1", ownerName: "Assessor", ownerEmail: "assessor@example.com", dueDate: "2026-09-30T00:00:00Z", status: "in_progress", progressNote: "Staging configured.", reviewComment: "", createdBy: "counselor-1", submittedAt: null, closedBy: null, closedAt: null, createdAt: "2026-08-18T00:00:00Z", updatedAt: "2026-08-18T00:00:00Z", evidence: [{ id: "rem-evidence-1", actionID: "action-1", originalName: "deployment.pdf", mimeType: "application/pdf", sizeBytes: 2048, uploadedBy: "assessor-1", createdAt: "2026-08-18T00:00:00Z" }] }],
};

test("renders audit scope, evidence, and activity trail", () => {
  render(<AuditPackageView auditPackage={auditPackage} onBack={vi.fn()} onDownloadCSV={vi.fn()} />);

  expect(screen.getByRole("heading", { name: /audit package/i })).toBeTruthy();
  expect(screen.getByText(/Assessment v2/i)).toBeTruthy();
  expect(screen.getByText(/Previous v1/i)).toBeTruthy();
  expect(screen.getByText("Scope & assignment")).toBeTruthy();
  expect(screen.getByText("risk-review.pdf")).toBeTruthy();
  expect(screen.getByText("response.reviewed")).toBeTruthy();
  expect(screen.getByRole("heading", { name: "Remediation register" })).toBeTruthy();
  expect(screen.getByText("deployment.pdf")).toBeTruthy();
  expect(screen.getByRole("region", { name: "Scope and assignment report table" }).getAttribute("tabindex")).toBe("0");
  expect(screen.getByRole("region", { name: "Scope & assignment" })).toBeTruthy();
});

test("downloads the CSV register", () => {
  const onDownloadCSV = vi.fn();
  render(<AuditPackageView auditPackage={auditPackage} onBack={vi.fn()} onDownloadCSV={onDownloadCSV} />);

  fireEvent.click(screen.getByRole("button", { name: /download csv/i }));

  expect(onDownloadCSV).toHaveBeenCalled();
});
