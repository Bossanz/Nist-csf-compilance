import type { Organization, Project } from "./types";

function segment(value: string) {
  return encodeURIComponent(value);
}

export function organizationPath(organization: Pick<Organization, "slug">) {
  return `/organizations/${segment(organization.slug)}`;
}

export function projectPath(organization: Pick<Organization, "slug">, project: Pick<Project, "slug">) {
  return `${organizationPath(organization)}/projects/${segment(project.slug)}`;
}
