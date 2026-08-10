"use client";

import { useMemo } from "react";
import type { FunctionNode, Organization, ProfilePatch, ProfileRow, Project, ResponseDocument, StakeholderResponse, Summary, User } from "../lib/types";
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
}: Props) {
  const visibleProfile = useMemo(
    () => profile.filter((row) => row.functionCode === selectedCode && (user.userType !== "stakeholder" || row.included)),
    [profile, selectedCode, user.userType],
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
                readOnly={!['counselor_admin', 'counselor'].includes(user.role)}
                role={user.role}
                responses={responses}
                onSaveResponse={onSaveResponse}
                onSubmitResponse={onSubmitResponse}
                onReviewResponse={onReviewResponse}
                onUploadEvidence={onUploadEvidence}
                onDeleteEvidence={onDeleteEvidence}
                onDownloadEvidence={onDownloadEvidence}
              />
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}
