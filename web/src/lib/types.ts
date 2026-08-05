export type CoverageLevel = "none" | "partial" | "substantial" | "full";
export type FunctionNode = { id: string; code: string; name: string; description: string; categories: CategoryNode[] };
export type CategoryNode = { id: string; code: string; name: string; description: string; subcategories: SubcategoryNode[] };
export type SubcategoryNode = { id: string; code: string; description: string };
export type Project = { id: string; organizationID: string; name: string; status: string; createdAt: string };
export type ProfileRow = { id: string; projectID: string; subcategoryID: string; functionCode: string; categoryCode: string; subcategoryCode: string; description: string; included: boolean; rationale: string; currentPriority: string; currentCoverageLevel: CoverageLevel; currentStatusText: string; currentPoliciesText: string; currentTier: string; targetPriority: string; targetCoverageLevel: CoverageLevel; targetApproachText: string; targetTier: string; notes: string; considerations: string; reviewStatus: string };
export type ProfilePatch = Partial<Pick<ProfileRow, "included" | "rationale" | "currentPriority" | "currentCoverageLevel" | "currentStatusText" | "currentPoliciesText" | "targetPriority" | "targetCoverageLevel" | "targetApproachText" | "notes" | "considerations">>;
export type Summary = { coveragePct: number; includedCount: number; pendingCount: number; rejectedCount: number; functions: { code: string; coveragePct: number; includedCount: number }[] };
