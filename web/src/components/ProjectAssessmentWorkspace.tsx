"use client";

import { useMemo, useState } from "react";
import type { EvidencePreview, FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../lib/types";
import { FunctionSidebar, type FunctionProgress, type WorkspaceMode } from "./FunctionSidebar";
import { SummaryCards } from "./SummaryCards";
import { AssignmentProgress } from "./AssignmentProgress";
import { ProfileEditor } from "./ProfileEditor";
import { ProjectFinalizationPanel } from "./ProjectFinalizationPanel";

type Props = {
  user: User;
  organization: Organization;
  project: Project;
  functions: FunctionNode[];
  organizationUsers: User[];
  profile: ProfileRow[];
  responses: StakeholderResponse[];
  summary: Summary;
  selectedCode: string;
  error: string;
  onBack: () => void;
  onSelectFunction: (code: string) => void;
  onSaveProfile: (subcategoryID: string, patch: ProfilePatch) => Promise<void>;
  onSaveResponse: (subcategoryID: string, responseText: string) => Promise<void>;
  onSubmitResponse: (subcategoryID: string) => Promise<void>;
  onReviewResponse: (subcategoryID: string, status: "reviewed" | "needs_more_info", comment: string) => Promise<void>;
  onUploadEvidence: (subcategoryID: string, file: File) => Promise<void>;
  onDeleteEvidence: (subcategoryID: string, documentID: string) => Promise<void>;
  onDownloadEvidence: (subcategoryID: string, document: ResponseDocument) => Promise<void>;
  onPreviewEvidence?: (subcategoryID: string, document: ResponseDocument) => Promise<void>;
  evidencePreview?: EvidencePreview | null;
  previewTargetSubcategoryID?: string | null;
  previewLoading?: boolean;
  previewError?: string;
  onCloseEvidencePreview?: () => void;
  onSetFunctionIncluded?: (functionCode: string, included: boolean) => Promise<void>;
  onSubmitScope: () => Promise<void>;
  onFinalizeProject?: () => Promise<void>;
  onOpenFinalReport?: () => void;
  onOpenAuditPackage?: () => void;
};

function formatProjectDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

export function ProjectAssessmentWorkspace({
  user,
  organization,
  project,
  functions,
  organizationUsers,
  profile,
  responses,
  summary,
  selectedCode,
  error,
  onBack,
  onSelectFunction,
  onSaveProfile,
  onSaveResponse,
  onSubmitResponse,
  onReviewResponse,
  onUploadEvidence,
  onDeleteEvidence,
  onDownloadEvidence,
  onPreviewEvidence,
  evidencePreview,
  previewTargetSubcategoryID,
  previewLoading,
  previewError,
  onCloseEvidencePreview,
  onSetFunctionIncluded,
  onSubmitScope,
  onFinalizeProject,
  onOpenFinalReport,
  onOpenAuditPackage,
}: Props) {
  const isCounselor = user.userType === "counselor";
  const isFinalized = project.status === "closed";
  const scopeSubmitted = project.status !== "setup";
  const canEditScope = isCounselor && !isFinalized;
  const canEditProfile = !isFinalized && (user.role === "org_admin" || user.role === "assessor");
  const workspaceMode: WorkspaceMode = isFinalized
    ? "Read-only"
    : isCounselor
    ? "Scope & Assignment"
    : user.role === "reviewer"
      ? "Review Queue"
      : user.role === "viewer"
        ? "Read-only"
        : "My Work";
  const selectedFunction = functions.find((fn) => fn.code === selectedCode);
  const taskHint = isFinalized
    ? "Project finalized. Read the final report and audit package."
    : isCounselor
    ? scopeSubmitted
      ? "Read stakeholder progress and final results."
      : "Select outcomes, add rationale, and assign an Assessor before submitting the scope."
    : user.role === "reviewer"
      ? "Review outcomes in Reviewing status and record a final decision."
      : user.role === "viewer"
        ? "Read included outcomes, responses, and evidence."
        : "Complete your assigned outcomes and attach supporting evidence.";
  const [bulkState, setBulkState] = useState<"idle" | "saving" | "error">("idle");
  const [bulkError, setBulkError] = useState("");
  const [scopeSubmitState, setScopeSubmitState] = useState<"idle" | "submitting" | "error">("idle");
  const [scopeSubmitError, setScopeSubmitError] = useState("");
  const functionRows = useMemo(() => profile.filter((row) => row.functionCode === selectedCode), [profile, selectedCode]);
  const responseBySubcategoryID = useMemo(
    () => new Map(responses.map((response) => [response.subcategoryID, response] as const)),
    [responses],
  );
  const functionProgress = useMemo<Record<string, FunctionProgress>>(
    () => Object.fromEntries(functions.map((fn) => {
      const includedRows = profile.filter((row) => row.functionCode === fn.code && row.included && (isCounselor || scopeSubmitted));
      const attention = includedRows.filter((row) => {
        const response = responseBySubcategoryID.get(row.subcategoryID);
        if (isCounselor) return row.assignedUserID === null;
        if (user.role === "reviewer") return response?.status === "submitted";
        if (user.role === "viewer") return false;
        return row.assignedUserID === user.id && (!response || response.status === "draft" || response.status === "needs_more_info");
      }).length;
      const attentionLabel = isCounselor ? "unassigned" : user.role === "reviewer" ? "to review" : user.role === "viewer" ? undefined : "open";
      const functionSummary = summary.functions.find((item) => item.code === fn.code);
      return [fn.code, {
        value: includedRows.length,
        label: "included",
        coveragePct: functionSummary?.coveragePct ?? 0,
        includedCount: functionSummary?.includedCount ?? includedRows.length,
        attention,
        attentionLabel,
      }];
    })),
    [functions, isCounselor, profile, responseBySubcategoryID, scopeSubmitted, summary.functions, user.role],
  );
  const allIncluded = functionRows.length > 0 && functionRows.every((row) => row.included);
  const assignmentProgress = useMemo(() => {
    const includedRows = profile.filter((row) => row.included);
    const assigned = includedRows.filter((row) => row.assignedUserID !== null).length;
    return { included: includedRows.length, assigned, unassigned: includedRows.length - assigned };
  }, [profile]);
  const finalizationReadiness = useMemo(() => {
    const includedRows = profile.filter((row) => row.included);
    const approvedCount = includedRows.filter((row) => responseBySubcategoryID.get(row.subcategoryID)?.status === "reviewed").length;
    const remaining = includedRows.filter((row) => responseBySubcategoryID.get(row.subcategoryID)?.status !== "reviewed").map((row) => {
      const status = responseBySubcategoryID.get(row.subcategoryID)?.status;
      const reason = status === "submitted" ? "Reviewing" : status === "needs_more_info" ? "Returned for changes" : "Waiting for Assessor response";
      return { code: row.subcategoryCode, reason };
    });
    return { includedCount: includedRows.length, approvedCount, remaining };
  }, [profile, responseBySubcategoryID]);
  const visibleProfile = useMemo(
    () => profile.filter((row) => {
      if (row.functionCode !== selectedCode) return false;
      if (isCounselor) return true;
      if (!scopeSubmitted) return false;
      if (!row.included) return false;
      if (user.role === "reviewer" || user.role === "viewer") return true;
      return canEditProfile && row.assignedUserID === user.id;
    }),
    [canEditProfile, isCounselor, profile, scopeSubmitted, selectedCode, user.id, user.role],
  );
  const activeFunctionIncludedCount = isCounselor ? functionRows.filter((row) => row.included).length : visibleProfile.filter((row) => row.included).length;
  const activeFunctionLabel = selectedFunction ? `${selectedCode} — ${selectedFunction.name}` : selectedCode;
  const projectMetadata = [
    { label: "Objective", value: project.objective },
    { label: "Assessment period", value: project.assessmentPeriod },
    { label: "Target completion", value: project.targetCompletionDate ? formatProjectDate(project.targetCompletionDate) : undefined },
    { label: "Scope boundary", value: project.scopeBoundary },
    { label: "Compliance driver", value: project.complianceDriver },
  ].filter((item): item is { label: string; value: string } => Boolean(item.value?.trim()));
  const includedOutcomeLabel = `${activeFunctionIncludedCount} included ${activeFunctionIncludedCount === 1 ? "outcome" : "outcomes"}`;
  const emptyQueueMessage = visibleProfile.length > 0
    ? ""
    : isCounselor
      ? "No outcomes found in this Function."
      : !scopeSubmitted
        ? "The Counselor has not submitted the Project scope yet."
        : user.role === "reviewer"
        ? functionRows.some((row) => row.included)
          ? "No outcomes awaiting review are available in this Function."
          : "No included outcomes are available in this Function."
        : user.role === "viewer"
          ? "No included outcomes are available in this Function."
          : "No included outcomes are assigned to you in this Function.";

  async function setFunctionIncluded(included: boolean) {
    if (!onSetFunctionIncluded) return;
    setBulkState("saving");
    setBulkError("");
    try {
      await onSetFunctionIncluded(selectedCode, included);
      setBulkState("idle");
    } catch (cause) {
      setBulkState("error");
      setBulkError(cause instanceof Error ? cause.message : "Could not update outcomes");
    }
  }
  async function submitScope() {
    setScopeSubmitState("submitting");
    setScopeSubmitError("");
    try {
      await onSubmitScope();
      setScopeSubmitState("idle");
    } catch (cause) {
      setScopeSubmitState("error");
      setScopeSubmitError(cause instanceof Error ? cause.message : "Could not submit project scope");
    }
  }

  return (
    <div className="shell">
      <FunctionSidebar functions={functions} selectedCode={selectedCode} onSelect={onSelectFunction} progressByFunction={functionProgress} mode={workspaceMode} />
      <main className="main project-main">
        <header className="project-header">
          <button className="text-button back-button" onClick={onBack}>Back to organization</button>
          <div className="project-context">
            <span>{organization.name}</span>
            <span aria-hidden="true">/</span>
            <span>{displayProjectStatus(project.status)}</span>
          </div>
          <h1>{project.name}</h1>
          <p className="project-subtitle">{taskHint}</p>
          <section className="project-context-panel" aria-label="Project context">
            <div className="project-context-overview">
              <div>
                <span className="context-label">Project status</span>
            <strong>{displayProjectStatus(project.status)}</strong>
              </div>
              <div>
                <span className="context-label">Workspace mode</span>
                <strong>{workspaceMode}</strong>
              </div>
              <div>
                <span className="context-label">Active Function</span>
                <strong>{activeFunctionLabel}</strong>
              </div>
              <div>
                <span className="context-label">Included outcomes</span>
                <strong>{includedOutcomeLabel}</strong>
              </div>
            </div>
            {projectMetadata.length > 0 && (
              <dl className="project-metadata">
                {projectMetadata.map((item) => (
                  <div key={item.label}>
                    <dt>{item.label}</dt>
                    <dd>{item.value}</dd>
                  </div>
                ))}
              </dl>
            )}
          </section>
          {isCounselor && !scopeSubmitted && (
            <section className="scope-submit-panel" aria-label="Scope submission">
              <div className="scope-submit-copy">
                <strong>Scope draft</strong>
                <p>Stakeholder response fields stay hidden until you submit the selected scope.</p>
              </div>
              <button className="primary" type="button" disabled={scopeSubmitState === "submitting"} onClick={() => void submitScope()}>
                {scopeSubmitState === "submitting" ? "Submitting…" : "Submit scope"}
              </button>
              {scopeSubmitError && <p className="error scope-submit-error" role="alert">{scopeSubmitError}</p>}
            </section>
          )}
          {isCounselor && (project.status === "in_review" || isFinalized) && onFinalizeProject && onOpenFinalReport && onOpenAuditPackage && (
            <ProjectFinalizationPanel
              status={project.status}
              includedCount={finalizationReadiness.includedCount}
              approvedCount={finalizationReadiness.approvedCount}
              remaining={finalizationReadiness.remaining}
              onFinalize={onFinalizeProject}
              onOpenReport={onOpenFinalReport}
              onOpenAudit={onOpenAuditPackage}
            />
          )}
        </header>
        {error && <div className="error" role="alert">{error}</div>}
        <div className="project-layout">
          <div className="reading-column">
            <SummaryCards summary={summary} />
            {isCounselor && <AssignmentProgress {...assignmentProgress} />}
            <section className="assessment-region" aria-labelledby="outcome-assessments-heading">
              <div className="workspace-heading">
                <div>
                    <p className="section-context">Function: {selectedCode}{selectedFunction ? ` — ${selectedFunction.name}` : ""}</p>
                  <h2 id="outcome-assessments-heading">Outcomes in this Function</h2>
                </div>
                <div className="workspace-tools">
                  {isCounselor && onSetFunctionIncluded && functionRows.length > 0 && (
                    <label className="check-field bulk-scope-toggle">
                      <input type="checkbox" aria-label="Include all outcomes in this Function" checked={allIncluded} disabled={bulkState === "saving"} onChange={(event) => void setFunctionIncluded(event.target.checked)} />
                      <span>{bulkState === "saving" ? "Applying scope…" : `Include every outcome in ${selectedFunction?.name ?? selectedCode}`}</span>
                    </label>
                  )}
                  <div className="outcome-count"><strong>{visibleProfile.length}</strong><span>{isCounselor ? "outcomes in this Function" : "included outcomes"}</span></div>
                  {bulkError && <span className="error bulk-scope-error" role="alert">{bulkError}</span>}
                </div>
              </div>
              {emptyQueueMessage ? (
                <div className="empty-state" role="status">{emptyQueueMessage}</div>
              ) : (
                <ProfileEditor
                  rows={visibleProfile}
                  onSave={onSaveProfile}
                  canEditScope={canEditScope}
                  canEditProfile={canEditProfile}
                  scopeSubmitted={scopeSubmitted}
                  assigneeOptions={organizationUsers}
                  role={user.role}
                  responses={responses}
                  onSaveResponse={onSaveResponse}
                  onSubmitResponse={onSubmitResponse}
                  onReviewResponse={onReviewResponse}
                  onUploadEvidence={onUploadEvidence}
                  onDeleteEvidence={onDeleteEvidence}
                  onDownloadEvidence={onDownloadEvidence}
                  onPreviewEvidence={onPreviewEvidence}
                  evidencePreview={evidencePreview}
                  previewTargetSubcategoryID={previewTargetSubcategoryID}
                  previewLoading={previewLoading}
                  previewError={previewError}
                  onCloseEvidencePreview={onCloseEvidencePreview}
                />
              )}
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}

function displayProjectStatus(status: string) {
  if (status === "closed") return "Finalized";
  return status.replaceAll("_", " ");
}
