import { expect, test } from "vitest";
import { getRemediationStatusSummary } from "./remediationStatus";

const gap = { included: true, currentCoverageLevel: "partial" as const, targetCoverageLevel: "full" as const };
const noGap = { included: true, currentCoverageLevel: "full" as const, targetCoverageLevel: "full" as const };

test("reports a coverage gap as not started when no action exists", () => {
  expect(getRemediationStatusSummary([gap], [])).toEqual({
    status: "not_started",
    gapCount: 1,
    actionCount: 0,
    openActionCount: 0,
    closedActionCount: 0,
  });
});

test("reports remediation as in progress while an action remains open", () => {
  expect(getRemediationStatusSummary([gap], [{ status: "in_progress" }])).toEqual({
    status: "in_progress",
    gapCount: 1,
    actionCount: 1,
    openActionCount: 1,
    closedActionCount: 0,
  });
});

test("reports remediation as complete when every action is closed", () => {
  expect(getRemediationStatusSummary([gap], [{ status: "closed" }])).toEqual({
    status: "complete",
    gapCount: 1,
    actionCount: 1,
    openActionCount: 0,
    closedActionCount: 1,
  });
});

test("reports not required when included outcomes have no coverage gap", () => {
  expect(getRemediationStatusSummary([noGap], [])).toEqual({
    status: "not_required",
    gapCount: 0,
    actionCount: 0,
    openActionCount: 0,
    closedActionCount: 0,
  });
});
