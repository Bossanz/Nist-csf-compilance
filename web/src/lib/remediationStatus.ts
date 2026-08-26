import type { CoverageLevel, ProfileRow, RemediationAction } from "./types";

export type RemediationOverviewStatus = "not_required" | "not_started" | "in_progress" | "complete";

export type RemediationStatusSummary = {
  status: RemediationOverviewStatus;
  gapCount: number;
  actionCount: number;
  openActionCount: number;
  closedActionCount: number;
};

const coverageRank: Record<CoverageLevel, number> = {
  none: 0,
  partial: 1,
  substantial: 2,
  full: 3,
};

export function getRemediationStatusSummary(
  profile: Array<Pick<ProfileRow, "included" | "currentCoverageLevel" | "targetCoverageLevel">>,
  actions: Array<Pick<RemediationAction, "status">>,
): RemediationStatusSummary {
  const gapCount = profile.filter((row) => row.included && coverageRank[row.currentCoverageLevel] < coverageRank[row.targetCoverageLevel]).length;
  const closedActionCount = actions.filter((action) => action.status === "closed").length;
  const openActionCount = actions.length - closedActionCount;
  const status: RemediationOverviewStatus = actions.length === 0
    ? gapCount === 0 ? "not_required" : "not_started"
    : openActionCount === 0 ? "complete" : "in_progress";

  return {
    status,
    gapCount,
    actionCount: actions.length,
    openActionCount,
    closedActionCount,
  };
}

export function remediationOverviewStatusLabel(status: RemediationOverviewStatus) {
  if (status === "not_required") return "Not required";
  if (status === "not_started") return "Not started";
  if (status === "in_progress") return "In progress";
  return "Complete";
}
