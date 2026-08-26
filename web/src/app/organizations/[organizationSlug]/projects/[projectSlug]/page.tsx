"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../../../lib/api";
import type { AuditTrailEntry, EvidencePreview, FunctionNode, Organization, ProfilePatch, ProfileRow, Project, RemediationAction, RemediationCreateInput, RemediationPatchInput, ResponseDocument, StakeholderResponse, Summary, User } from "../../../../../lib/types";
import { organizationPath, projectPath } from "../../../../../lib/routes";
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
  const [projectVersions, setProjectVersions] = useState<Project[]>([]);
  const [versionHistoryLoading, setVersionHistoryLoading] = useState(false);
  const [versionHistoryError, setVersionHistoryError] = useState("");
  const [functions, setFunctions] = useState<FunctionNode[]>([]);
  const [organizationUsers, setOrganizationUsers] = useState<User[]>([]);
  const [profile, setProfile] = useState<ProfileRow[]>([]);
  const [responses, setResponses] = useState<StakeholderResponse[]>([]);
  const [auditTrail, setAuditTrail] = useState<AuditTrailEntry[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditError, setAuditError] = useState("");
  const [remediationActions, setRemediationActions] = useState<RemediationAction[]>([]);
  const [remediationLoaded, setRemediationLoaded] = useState(false);
  const [remediationLoading, setRemediationLoading] = useState(false);
  const [remediationError, setRemediationError] = useState("");
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
  const previewRequest = useRef<AbortController | null>(null);

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [organizationSlug, projectSlug, retryCount]);

  useEffect(() => () => {
    previewRequest.current?.abort();
    if (previewURL.current) URL.revokeObjectURL(previewURL.current);
  }, []);

  async function initialize(active: boolean) {
    setLoading(true);
    setError("");
    setNotFound(false);
    setProjectVersions([]);
    setVersionHistoryLoading(false);
    setVersionHistoryError("");
    setAuditTrail([]);
    setAuditLoading(false);
    setAuditError("");
    setRemediationActions([]);
    setRemediationLoaded(false);
    setRemediationLoading(false);
    setRemediationError("");
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
      setVersionHistoryLoading(true);
      void api.getProjectVersions(nextProject.id)
        .then((nextVersions) => {
          if (active) setProjectVersions(nextVersions);
        })
        .catch((cause) => {
          if (active) setVersionHistoryError(messageOf(cause));
        })
        .finally(() => {
          if (active) setVersionHistoryLoading(false);
        });
      setAuditLoading(true);
      void api.getProjectAuditLogs(nextProject.id)
        .then((nextAuditTrail) => {
          if (active) setAuditTrail(nextAuditTrail);
        })
        .catch((cause) => {
          if (active) setAuditError(messageOf(cause));
        })
        .finally(() => {
          if (active) setAuditLoading(false);
        });
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
    const updatedRows = await api.updateFunctionScope(project.id, functionCode, included);
    const updatedByID = new Map(updatedRows.map((row) => [row.subcategoryID, row]));
    setProfile((rows) => rows.map((row) => updatedByID.get(row.subcategoryID) ?? row));
    setSummary(await api.getSummary(project.id));
  }

  async function loadRemediationActions() {
    if (!project || remediationLoaded || remediationLoading) return;
    setRemediationLoading(true);
    setRemediationError("");
    try {
      setRemediationActions(await api.getRemediationActions(project.id));
      setRemediationLoaded(true);
    } catch (cause) {
      setRemediationError(messageOf(cause));
    } finally {
      setRemediationLoading(false);
    }
  }
  async function submitScope() {
    if (!project) return;
    setProject(await api.submitProjectScope(project.id));
  }

  async function finalizeProject() {
    if (!project) return;
    setProject(await api.finalizeProject(project.id));
  }

  async function createProjectVersion() {
    if (!project || !organization) return;
    const nextProject = await api.createProjectVersion(project.id);
    setVersionHistoryLoading(true);
    setVersionHistoryError("");
    try {
      setProjectVersions(await api.getProjectVersions(nextProject.id));
    } catch (cause) {
      setVersionHistoryError(messageOf(cause));
    } finally {
      setVersionHistoryLoading(false);
    }
    router.push(projectPath(organization, nextProject));
  }

  function openProjectVersion(nextProject: Project) {
    if (!organization) return;
    router.push(projectPath(organization, nextProject));
  }

  function replaceRemediationAction(next: RemediationAction) {
    setRemediationActions((rows) => rows.some((row) => row.id === next.id) ? rows.map((row) => row.id === next.id ? next : row) : [...rows, next]);
  }

  async function createRemediation(input: RemediationCreateInput) {
    if (!project) return;
    replaceRemediationAction(await api.createRemediationAction(project.id, input));
  }

  async function updateRemediation(actionID: string, patch: RemediationPatchInput) {
    if (!project) return;
    replaceRemediationAction(await api.updateRemediationAction(project.id, actionID, patch));
  }

  async function saveRemediationProgress(actionID: string, progressNote: string) {
    if (!project) return;
    replaceRemediationAction(await api.updateRemediationProgress(project.id, actionID, progressNote));
  }

  async function submitRemediation(actionID: string) {
    if (!project) return;
    replaceRemediationAction(await api.submitRemediationAction(project.id, actionID));
  }

  async function reviewRemediation(actionID: string, decision: "close" | "return", comment: string) {
    if (!project) return;
    replaceRemediationAction(await api.reviewRemediationAction(project.id, actionID, { decision, comment }));
  }

  async function uploadRemediationEvidence(actionID: string, file: File) {
    if (!project) return;
    const evidence = await api.uploadRemediationEvidence(project.id, actionID, file);
    setRemediationActions((rows) => rows.map((action) => action.id === actionID ? { ...action, evidence: [...action.evidence, evidence] } : action));
  }

  async function deleteRemediationEvidence(actionID: string, evidenceID: string) {
    if (!project) return;
    await api.deleteRemediationEvidence(project.id, actionID, evidenceID);
    setRemediationActions((rows) => rows.map((action) => action.id === actionID ? { ...action, evidence: action.evidence.filter((item) => item.id !== evidenceID) } : action));
  }

  async function downloadRemediationEvidence(actionID: string, evidenceID: string, originalName: string) {
    if (!project) return;
    const blob = await api.downloadRemediationEvidence(project.id, actionID, evidenceID);
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = originalName;
    anchor.click();
    URL.revokeObjectURL(url);
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
    previewRequest.current?.abort();
    const controller = new AbortController();
    previewRequest.current = controller;
    setPreviewTargetSubcategoryID(subcategoryID);
    setPreviewLoading(true);
    setPreviewError("");
    if (previewURL.current) {
      URL.revokeObjectURL(previewURL.current);
      previewURL.current = null;
    }
    setEvidencePreview(null);
    try {
      const blob = await api.downloadResponseDocument(project.id, subcategoryID, evidenceDocument.id, controller.signal);
      if (previewRequest.current !== controller) return;
      const url = URL.createObjectURL(blob);
      previewURL.current = url;
      setEvidencePreview({ subcategoryID, documentID: evidenceDocument.id, url, mimeType: evidenceDocument.mimeType });
    } catch (cause) {
      if (cause instanceof Error && cause.name === "AbortError") return;
      if (previewRequest.current !== controller) return;
      setPreviewError(messageOf(cause));
    } finally {
      if (previewRequest.current === controller) setPreviewLoading(false);
    }
  }

  function closeEvidencePreview() {
    previewRequest.current?.abort();
    previewRequest.current = null;
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
    setRemediationActions([]);
    setProjectVersions([]);
    setAuditTrail([]);
    setAuditLoading(false);
    setAuditError("");
    setRemediationLoaded(false);
    setRemediationLoading(false);
    setRemediationError("");
    setVersionHistoryLoading(false);
    setVersionHistoryError("");
    setAuditTrail([]);
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
      versions={projectVersions}
      versionHistoryLoading={versionHistoryLoading}
      versionHistoryError={versionHistoryError}
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
      onFinalizeProject={finalizeProject}
      onCreateProjectVersion={createProjectVersion}
      onOpenProjectVersion={openProjectVersion}
      onOpenFinalReport={() => router.push(`${projectPath(organization, project)}/report`)}
      onOpenAuditPackage={() => router.push(`${projectPath(organization, project)}/audit`)}
      auditTrail={auditTrail}
      auditLoading={auditLoading}
      auditError={auditError}
      remediationActions={remediationActions}
      remediationLoaded={remediationLoaded}
      remediationLoading={remediationLoading}
      remediationError={remediationError}
      onLoadRemediationActions={loadRemediationActions}
      onCreateRemediation={createRemediation}
      onUpdateRemediation={updateRemediation}
      onSaveRemediationProgress={saveRemediationProgress}
      onSubmitRemediation={submitRemediation}
      onReviewRemediation={reviewRemediation}
      onUploadRemediationEvidence={uploadRemediationEvidence}
      onDeleteRemediationEvidence={deleteRemediationEvidence}
      onDownloadRemediationEvidence={downloadRemediationEvidence}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
