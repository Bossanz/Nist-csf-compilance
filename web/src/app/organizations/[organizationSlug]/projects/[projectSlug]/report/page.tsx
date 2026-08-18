"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../../../../lib/api";
import type { FinalReportData, Organization, Project } from "../../../../../../lib/types";
import { projectPath } from "../../../../../../lib/routes";
import { FinalReport } from "../../../../../../components/FinalReport";

export default function FinalReportPage() {
  const router = useRouter();
  const params = useParams<{ organizationSlug: string; projectSlug: string }>();
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [report, setReport] = useState<FinalReportData | null>(null);
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
      const nextReport = await api.getFinalReport(nextProject.id);
      if (!active) return;
      setOrganization(nextOrganization);
      setProject(nextProject);
      setReport(nextReport);
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 401) {
        router.replace("/login");
      } else if (active) {
        setError(cause instanceof Error ? cause.message : "Could not load final report");
      }
    } finally {
      if (active) setLoading(false);
    }
  }

  if (loading) return <main className="screen-center" role="status">Loading final report…</main>;
  if (!organization || !project || !report) return <main className="screen-center"><section className="empty-state" role="alert"><h1>Could not load final report</h1><p>{error || "This report is unavailable."}</p><button className="primary" type="button" onClick={() => void load(true)}>Try again</button></section></main>;

  const workspacePath = projectPath(organization, project);
  return <FinalReport report={report} onBack={() => router.push(workspacePath)} onOpenAudit={() => router.push(`${workspacePath}/audit`)} />;
}
