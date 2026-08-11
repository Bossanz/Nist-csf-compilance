"use client";

import { useMemo } from "react";
import type { EvidencePreview, FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../lib/types";
import { FunctionSidebar } from "./FunctionSidebar";
import { SummaryCards } from "./SummaryCards";
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
}: Props) {
  const isCounselor = user.userType === "counselor";
  const canEditScope = isCounselor;
  const canEditProfile = user.role === "org_admin" || user.role === "assessor";
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

  return (
    <div className="shell">
      <FunctionSidebar functions={functions} selectedCode={selectedCode} onSelect={onSelectFunction} />
      <main className="main project-main">
        <header className="project-header">
          <button className="text-button back-button" onClick={onBack}>Back to organization</button>
          <div className="project-context">
            <span>{organization.name}</span>
            <span aria-hidden="true">/</span>
            <span>{project.status.replaceAll("_", " ")}</span>
          </div>
          <h1>{project.name}</h1>
          <p className="project-subtitle">Current and target profile assessment</p>
        </header>
        {error && <div className="error" role="alert">{error}</div>}
        <div className="project-layout">
          <div className="reading-column">
            <SummaryCards summary={summary} />
            <section className="assessment-region" aria-labelledby="outcome-assessments-heading">
              <div className="workspace-heading">
                <div>
                  <p className="section-context">Function {selectedCode}</p>
                  <h2 id="outcome-assessments-heading">Outcome assessments</h2>
                </div>
                <div className="outcome-count"><strong>{visibleProfile.length}</strong><span>outcomes</span></div>
              </div>
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
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}
