"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../../../../lib/api";
import type { AuditPackageData, Organization, Project } from "../../../../../../lib/types";
import { projectPath } from "../../../../../../lib/routes";
import { AuditPackageView } from "../../../../../../components/AuditPackageView";

export default function AuditPackagePage() {
  const router = useRouter();
  const params = useParams<{ organizationSlug: string; projectSlug: string }>();
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [auditPackage, setAuditPackage] = useState<AuditPackageData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void load(active);
    return () => { active = false; };
  }, [params.organizationSlug, params.projectSlug]);

  async function load(active: boolean) {
    setLoading(true);
    setError("");
    try {
      await api.me();
      const nextOrganization = await api.getOrganizationBySlug(params.organizationSlug);
      const nextProject = await api.getOrganizationProjectBySlug(nextOrganization.id, params.projectSlug);
      const nextPackage = await api.getAuditPackage(nextProject.id);
      if (!active) return;
      setOrganization(nextOrganization);
      setProject(nextProject);
      setAuditPackage(nextPackage);
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 401) {
        router.replace("/login");
      } else if (active) {
        setError(cause instanceof Error ? cause.message : "Could not load audit package");
      }
    } finally {
      if (active) setLoading(false);
    }
  }

  async function downloadCSV() {
    if (!project) return;
    const blob = await api.downloadAuditPackageCSV(project.id);
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${project.slug}-audit-package.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  if (loading) return <main className="screen-center" role="status">Loading audit package…</main>;
  if (!organization || !project || !auditPackage) return <main className="screen-center"><section className="empty-state" role="alert"><h1>Could not load audit package</h1><p>{error || "This audit package is unavailable."}</p><button className="primary" type="button" onClick={() => void load(true)}>Try again</button></section></main>;

  return <AuditPackageView auditPackage={auditPackage} onBack={() => router.push(projectPath(organization, project))} onDownloadCSV={() => void downloadCSV()} />;
}
