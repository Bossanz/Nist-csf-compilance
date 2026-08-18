"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../../../lib/api";
import type { EvidencePreview, FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../../../../../lib/types";
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
  const [organizationUsers, setOrganizationUsers] = useState<User[]>([]);
  const [profile, setProfile] = useState<ProfileRow[]>([]);
  const [responses, setResponses] = useState<StakeholderResponse[]>([]);
  const [summary, setSummary] = useState<Summary>(emptySummary);
  const [selectedCode, setSelectedCode] = useState("");
  const [authChecked, setAuthChecked] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [notFound, setNotFound] = useState(false);
  const [evidencePreview, setEvidencePreview] = useState<EvidencePreview | null>(null);
  const [previewTargetSubcategoryID, setPreviewTargetSubcategoryID] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState("");
  const previewURL = useRef<string | null>(null);

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [organizationSlug, projectSlug, retryCount]);

  useEffect(() => () => {
    if (previewURL.current) URL.revokeObjectURL(previewURL.current);
  }, []);

  async function initialize(active: boolean) {
    setLoading(true);
    setError("");
    setNotFound(false);
    try {
      const [currentUser, nextOrganization] = await Promise.all([
        api.me(),
        api.getOrganizationBySlug(organizationSlug),
      ]);
      const nextProject = await api.getOrganizationProjectBySlug(nextOrganization.id, projectSlug);
      const usersPromise = currentUser.userType === "counselor" ? api.getOrganizationUsers(nextOrganization.id) : Promise.resolve<User[]>([]);
      const [functionRows, nextProfile, nextSummary, nextResponses] = await Promise.all([
        api.getFunctions(),
        api.getProfile(nextProject.id),
        api.getSummary(nextProject.id),
        api.getResponses(nextProject.id),
      ]);
      const nextOrganizationUsers = await usersPromise;
      if (!active) return;
      setUser(currentUser);
      setOrganization(nextOrganization);
      setProject(nextProject);
      setFunctions(functionRows);
      setOrganizationUsers(nextOrganizationUsers);
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

  async function setFunctionIncluded(functionCode: string, included: boolean) {
    if (!project) return;
    const functionRows = profile.filter((row) => row.functionCode === functionCode);
    for (const row of functionRows) {
      const updated = await api.updateProfile(project.id, row.subcategoryID, { included });
      const next = updated ?? { ...row, included };
      setProfile((rows) => rows.map((current) => current.subcategoryID === row.subcategoryID ? next : current));
    }
    setSummary(await api.getSummary(project.id));
  }
  async function submitScope() {
    if (!project) return;
    setProject(await api.submitProjectScope(project.id));
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

  async function previewEvidence(subcategoryID: string, evidenceDocument: ResponseDocument) {
    if (!project) return;
    setPreviewTargetSubcategoryID(subcategoryID);
    setPreviewLoading(true);
    setPreviewError("");
    if (previewURL.current) {
      URL.revokeObjectURL(previewURL.current);
      previewURL.current = null;
    }
    setEvidencePreview(null);
    try {
      const blob = await api.downloadResponseDocument(project.id, subcategoryID, evidenceDocument.id);
      const url = URL.createObjectURL(blob);
      previewURL.current = url;
      setEvidencePreview({ subcategoryID, documentID: evidenceDocument.id, url, mimeType: evidenceDocument.mimeType });
    } catch (cause) {
      setPreviewError(messageOf(cause));
    } finally {
      setPreviewLoading(false);
    }
  }

  function closeEvidencePreview() {
    if (previewURL.current) {
      URL.revokeObjectURL(previewURL.current);
      previewURL.current = null;
    }
    setEvidencePreview(null);
    setPreviewTargetSubcategoryID(null);
    setPreviewError("");
  }

  function retryLoad() {
    setUser(null);
    setOrganization(null);
    setProject(null);
    setFunctions([]);
    setOrganizationUsers([]);
    setProfile([]);
    setResponses([]);
    setSummary(emptySummary);
    setError("");
    setNotFound(false);
    setAuthChecked(false);
    setRetryCount((value) => value + 1);
  }

  if (!authChecked || loading && !project) return <main className="screen-center" role="status" aria-live="polite" aria-busy="true">Loading assessment…</main>;
  if (!user && error && !notFound) {
    return (
      <main className="screen-center" aria-labelledby="project-load-error">
        <section className="empty-state" role="alert">
          <h1 id="project-load-error">Could not load project</h1>
          <p>{error}</p>
          <button className="primary" type="button" onClick={retryLoad}>Try again</button>
        </section>
      </main>
    );
  }
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
      organizationUsers={organizationUsers}
      profile={profile}
      responses={responses}
      summary={summary}
      selectedCode={selectedCode}
      error={error}
      onBack={() => router.push(organizationPath(organization))}
      onSelectFunction={setSelectedCode}
      onSaveProfile={saveProfile}
      onSubmitScope={submitScope}
      onSaveResponse={saveResponse}
      onSubmitResponse={submitResponse}
      onReviewResponse={reviewResponse}
      onUploadEvidence={uploadEvidence}
      onDeleteEvidence={deleteEvidence}
      onDownloadEvidence={downloadEvidence}
      onSetFunctionIncluded={setFunctionIncluded}
      onPreviewEvidence={previewEvidence}
      evidencePreview={evidencePreview}
      previewTargetSubcategoryID={previewTargetSubcategoryID}
      previewLoading={previewLoading}
      previewError={previewError}
      onCloseEvidencePreview={closeEvidencePreview}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
