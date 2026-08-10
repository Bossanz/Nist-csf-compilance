"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../../../lib/api";
import type { FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../../../../../lib/types";
import { organizationPath } from "../../../../../lib/routes";
import { ProjectAssessmentWorkspace } from "../../../../../components/ProjectAssessmentWorkspace";

const emptySummary: Summary = { coveragePct: 0, includedCount: 0, pendingCount: 0, rejectedCount: 0, functions: [] };

export default function ProjectPage() {
  const router = useRouter();
  const params = useParams<{ organizationSlug: string; projectSlug: string }>();
  const organizationSlug = params.organizationSlug;
  const projectSlug = params.projectSlug;
  const [user, setUser] = useState<User | null>(null);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [functions, setFunctions] = useState<FunctionNode[]>([]);
  const [profile, setProfile] = useState<ProfileRow[]>([]);
  const [responses, setResponses] = useState<StakeholderResponse[]>([]);
  const [summary, setSummary] = useState<Summary>(emptySummary);
  const [selectedCode, setSelectedCode] = useState("");
  const [authChecked, setAuthChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [organizationSlug, projectSlug]);

  async function initialize(active: boolean) {
    setLoading(true);
    setError("");
    setNotFound(false);
    try {
      const currentUser = await api.me();
      const nextOrganization = await api.getOrganizationBySlug(organizationSlug);
      const nextProject = await api.getOrganizationProjectBySlug(nextOrganization.id, projectSlug);
      const [functionRows, nextProfile, nextSummary, nextResponses] = await Promise.all([
        api.getFunctions(),
        api.getProfile(nextProject.id),
        api.getSummary(nextProject.id),
        api.getResponses(nextProject.id),
      ]);
      if (!active) return;
      setUser(currentUser);
      setOrganization(nextOrganization);
      setProject(nextProject);
      setFunctions(functionRows);
      setSelectedCode(functionRows[0]?.code ?? "");
      setProfile(nextProfile);
      setSummary(nextSummary);
      setResponses(nextResponses);
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 401) {
        router.replace("/login");
      } else if (active) {
        setNotFound(cause instanceof APIError && cause.status === 404);
        setError(messageOf(cause));
      }
    } finally {
      if (active) {
        setLoading(false);
        setAuthChecked(true);
      }
    }
  }

  async function saveProfile(subcategoryID: string, patch: ProfilePatch) {
    if (!project) return;
    const updated = await api.updateProfile(project.id, subcategoryID, patch);
    setProfile((rows) => rows.map((row) => row.subcategoryID === subcategoryID ? updated : row));
    setSummary(await api.getSummary(project.id));
  }

  function replaceResponse(next: StakeholderResponse) {
    setResponses((rows) => {
      const found = rows.some((row) => row.subcategoryID === next.subcategoryID);
      return found ? rows.map((row) => row.subcategoryID === next.subcategoryID ? next : row) : [...rows, next];
    });
  }

  async function saveResponse(subcategoryID: string, responseText: string) {
    if (!project) return;
    replaceResponse(await api.saveResponse(project.id, subcategoryID, responseText));
  }

  async function submitResponse(subcategoryID: string) {
    if (!project) return;
    replaceResponse(await api.submitResponse(project.id, subcategoryID));
  }

  async function reviewResponse(subcategoryID: string, status: "reviewed" | "needs_more_info", comment: string) {
    if (!project) return;
    replaceResponse(await api.reviewResponse(project.id, subcategoryID, { status, comment }));
  }

  async function uploadEvidence(subcategoryID: string, file: File) {
    if (!project) return;
    let current = responses.find((row) => row.subcategoryID === subcategoryID);
    if (!current?.id) {
      current = await api.saveResponse(project.id, subcategoryID, "");
      replaceResponse(current);
    }
    const document = await api.uploadResponseDocument(project.id, subcategoryID, file);
    replaceResponse({ ...current, documents: [...current.documents, document] });
  }

  async function deleteEvidence(subcategoryID: string, documentID: string) {
    if (!project) return;
    await api.deleteResponseDocument(project.id, subcategoryID, documentID);
    setResponses((rows) => rows.map((row) => row.subcategoryID === subcategoryID ? { ...row, documents: row.documents.filter((document) => document.id !== documentID) } : row));
  }

  async function downloadEvidence(subcategoryID: string, evidenceDocument: ResponseDocument) {
    if (!project) return;
    const blob = await api.downloadResponseDocument(project.id, subcategoryID, evidenceDocument.id);
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = evidenceDocument.originalName;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  if (!authChecked || loading && !project) return <main className="screen-center">Loading...</main>;
  if (notFound || !user || !organization || !project) {
    return (
      <main className="main dashboard">
        <button className="text-button back-button" type="button" onClick={() => router.push("/organizations")}>Organizations</button>
        <section className="empty-state" aria-labelledby="project-not-found">
          <h1 id="project-not-found">Project not found</h1>
          <p>{error || "This assessment is unavailable."}</p>
        </section>
      </main>
    );
  }

  return (
    <ProjectAssessmentWorkspace
      user={user}
      organization={organization}
      project={project}
      functions={functions}
      profile={profile}
      responses={responses}
      summary={summary}
      selectedCode={selectedCode}
      error={error}
      onBack={() => router.push(organizationPath(organization))}
      onSelectFunction={setSelectedCode}
      onSaveProfile={saveProfile}
      onSaveResponse={saveResponse}
      onSubmitResponse={submitResponse}
      onReviewResponse={reviewResponse}
      onUploadEvidence={uploadEvidence}
      onDeleteEvidence={deleteEvidence}
      onDownloadEvidence={downloadEvidence}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
