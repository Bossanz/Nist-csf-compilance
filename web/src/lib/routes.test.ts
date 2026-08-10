import { expect, test } from "vitest";
import { organizationPath, projectPath } from "./routes";
import type { Organization, Project } from "./types";

const organization: Organization = { id: "org-1", name: "Acme Corporation", slug: "acme-corporation", type: "client" };
const project: Project = { id: "project-1", organizationID: "org-1", organizationName: organization.name, name: "Readiness", slug: "readiness", status: "setup", createdAt: "2026-08-06T03:00:00Z" };

test("builds organization and project paths from slugs instead of IDs", () => {
  expect(organizationPath(organization)).toBe("/organizations/acme-corporation");
  expect(projectPath(organization, project)).toBe("/organizations/acme-corporation/projects/readiness");
  expect(organizationPath(organization)).not.toContain(organization.id);
  expect(projectPath(organization, project)).not.toContain(project.id);
});

test("encodes Unicode slugs as URL path segments", () => {
  const thaiOrganization = { ...organization, slug: "บริษัท-เอ-บี-ซี" };
  expect(organizationPath(thaiOrganization)).toBe(`/organizations/${encodeURIComponent(thaiOrganization.slug)}`);
});
