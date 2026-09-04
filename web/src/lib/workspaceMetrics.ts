import type { Organization, Role } from "./types";

export type OrganizationPortfolioMetrics = {
  total: number;
  active: number;
  finalized: number;
  totalProjects: number;
};

export function getOrganizationPortfolioMetrics(
  organizations: Organization[],
  projectCounts: Record<string, number> = {},
): OrganizationPortfolioMetrics {
  const totalProjects = organizations.reduce((total, organization) => total + Math.max(0, projectCounts[organization.id] ?? 0), 0);
  return {
    total: organizations.length,
    active: organizations.length,
    finalized: 0,
    totalProjects,
  };
}

export function getProjectStatusLabel(status: string) {
  if (status === "closed") return "Finalized";
  if (status === "in_review") return "Reviewing";
  if (status === "in_progress") return "In progress";
  if (status === "setup") return "Setup";
  return status.replaceAll("_", " ").replace(/^\w/, (letter) => letter.toUpperCase());
}

export function getAttentionLabel(role: Role) {
  if (role === "reviewer") return "To review";
  if (role === "assessor" || role === "org_admin") return "Needs your input";
  if (role === "viewer" || role === "auditor") return "Read-only";
  return "Scope & assignment";
}
