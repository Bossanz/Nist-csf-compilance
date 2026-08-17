"use client";

import { useMemo, useState } from "react";
import type { EvidencePreview, FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../lib/types";
import { FunctionSidebar, type FunctionProgress, type WorkspaceMode } from "./FunctionSidebar";
import { SummaryCards } from "./SummaryCards";
import { AssignmentProgress } from "./AssignmentProgress";
import { ProfileEditor } from "./ProfileEditor";

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
};

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
}: Props) {
  const isCounselor = user.userType === "counselor";
  const canEditScope = isCounselor;
  const canEditProfile = user.role === "org_admin" || user.role === "assessor";
  const workspaceMode: WorkspaceMode = isCounselor
    ? "Scope & Assignment"
    : user.role === "reviewer"
      ? "Review Queue"
      : user.role === "viewer"
        ? "Read-only"
        : "My Work";
  const selectedFunction = functions.find((fn) => fn.code === selectedCode);
  const taskHint = isCounselor
    ? "Set project scope and read stakeholder progress."
    : user.role === "reviewer"
      ? "Review submitted stakeholder responses and record a final decision."
      : user.role === "viewer"
        ? "Read included outcomes, responses, and evidence."
        : "Complete your assigned outcomes and attach supporting evidence.";
  const [bulkState, setBulkState] = useState<"idle" | "saving" | "error">("idle");
  const [bulkError, setBulkError] = useState("");
  const functionRows = useMemo(() => profile.filter((row) => row.functionCode === selectedCode), [profile, selectedCode]);
  const responseBySubcategoryID = useMemo(
    () => new Map(responses.map((response) => [response.subcategoryID, response] as const)),
    [responses],
  );
  const functionProgress = useMemo<Record<string, FunctionProgress>>(
    () => Object.fromEntries(functions.map((fn) => {
      const includedRows = profile.filter((row) => row.functionCode === fn.code && row.included);
      const attention = includedRows.filter((row) => {
        const response = responseBySubcategoryID.get(row.subcategoryID);
        if (isCounselor) return row.assignedUserID === null;
        if (user.role === "reviewer") return response?.status === "submitted";
        if (user.role === "viewer") return false;
        return row.assignedUserID === user.id && (!response || response.status === "draft" || response.status === "needs_more_info");
      }).length;
      const attentionLabel = isCounselor ? "unassigned" : user.role === "reviewer" ? "to review" : user.role === "viewer" ? undefined : "open";
      return [fn.code, { value: includedRows.length, label: "included", attention, attentionLabel }];
    })),
    [functions, isCounselor, profile, responseBySubcategoryID, user.role],
  );
  const allIncluded = functionRows.length > 0 && functionRows.every((row) => row.included);
  const assignmentProgress = useMemo(() => {
    const includedRows = profile.filter((row) => row.included);
    const assigned = includedRows.filter((row) => row.assignedUserID !== null).length;
    return { included: includedRows.length, assigned, unassigned: includedRows.length - assigned };
  }, [profile]);
  const visibleProfile = useMemo(
    () => profile.filter((row) => {
      if (row.functionCode !== selectedCode) return false;
      if (isCounselor) return true;
      if (!row.included) return false;
      if (user.role === "reviewer" || user.role === "viewer") return true;
      return canEditProfile && row.assignedUserID === user.id;
    }),
    [canEditProfile, isCounselor, profile, selectedCode, user.id, user.role],
  );
  const emptyQueueMessage = visibleProfile.length > 0
    ? ""
    : isCounselor
      ? "No outcomes found in this Function."
      : user.role === "reviewer"
        ? functionRows.some((row) => row.included)
          ? "No submitted review work is available in this Function."
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

  return (
    <div className="shell">
      <FunctionSidebar functions={functions} selectedCode={selectedCode} onSelect={onSelectFunction} progressByFunction={functionProgress} mode={workspaceMode} />
      <main className="main project-main">
        <header className="project-header">
          <button className="text-button back-button" onClick={onBack}>Back to organization</button>
          <div className="project-context">
            <span>{organization.name}</span>
            <span aria-hidden="true">/</span>
            <span>{project.status.replaceAll("_", " ")}</span>
          </div>
          <h1>{project.name}</h1>
          <p className="project-subtitle">{taskHint}</p>
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
