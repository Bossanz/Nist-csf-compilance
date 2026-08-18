import { useMemo } from "react";
import type { EvidencePreview, ProfilePatch, ProfileRow, ResponseDocument, Role, StakeholderResponse, User } from "../lib/types";
import { AssessmentCard } from "./AssessmentCard";

export function ProfileEditor({
  rows,
  onSave,
  canEditScope,
  canEditProfile,
  scopeSubmitted = true,
  assigneeOptions,
  role,
  responses = [],
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
}: {
  rows: ProfileRow[];
  onSave: (id: string, patch: ProfilePatch) => Promise<void>;
  canEditScope: boolean;
  canEditProfile: boolean;
  assigneeOptions: User[];
  scopeSubmitted?: boolean;
  role?: Role;
  responses?: StakeholderResponse[];
  onSaveResponse?: (id: string, responseText: string) => Promise<void>;
  onSubmitResponse?: (id: string) => Promise<void>;
  onReviewResponse?: (id: string, status: "reviewed" | "needs_more_info", comment: string) => Promise<void>;
  onUploadEvidence?: (id: string, file: File) => Promise<void>;
  onDeleteEvidence?: (id: string, documentID: string) => Promise<void>;
  onDownloadEvidence?: (id: string, document: ResponseDocument) => Promise<void>;
  onPreviewEvidence?: (id: string, document: ResponseDocument) => Promise<void>;
  evidencePreview?: EvidencePreview | null;
  previewTargetSubcategoryID?: string | null;
  previewLoading?: boolean;
  previewError?: string;
  onCloseEvidencePreview?: () => void;
}) {
  const responseBySubcategoryID = useMemo(
    () => new Map(responses.map((response) => [response.subcategoryID, response] as const)),
    [responses],
  );

  if (rows.length === 0) {
    const emptyMessage = role === "counselor"
      ? "No outcomes found in this Function."
      : role === "reviewer" || role === "viewer"
        ? "No included outcomes are available in this Function."
        : "No included outcomes are assigned to you in this Function.";
    return <div className="empty-state">{emptyMessage}</div>;
  }

  return (
    <div className="assessment-list">
      {rows.map((row) => <AssessmentCard
        key={row.id}
        row={row}
        onSave={onSave}
        canEditScope={canEditScope}
        canEditProfile={canEditProfile}
        scopeSubmitted={scopeSubmitted}
        assigneeOptions={assigneeOptions}
        role={role}
        response={responseBySubcategoryID.get(row.subcategoryID) ?? emptyResponse(row)}
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
      />)}
    </div>
  );
}

function emptyResponse(row: ProfileRow): StakeholderResponse {
  return { id: "", projectID: row.projectID, subcategoryID: row.subcategoryID, responseText: "", status: "draft", respondedBy: null, submittedAt: null, reviewComment: "", reviewedBy: null, reviewedAt: null, createdAt: "", updatedAt: "", documents: [] };
}
