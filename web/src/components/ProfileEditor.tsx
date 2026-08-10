import type { ProfilePatch, ProfileRow, ResponseDocument, Role, StakeholderResponse, User } from "../lib/types";
import { AssessmentCard } from "./AssessmentCard";

export function ProfileEditor({
  rows,
  onSave,
  canEditScope,
  canEditProfile,
  assigneeOptions,
  role,
  responses = [],
  onSaveResponse,
  onSubmitResponse,
  onReviewResponse,
  onUploadEvidence,
  onDeleteEvidence,
  onDownloadEvidence,
}: {
  rows: ProfileRow[];
  onSave: (id: string, patch: ProfilePatch) => Promise<void>;
  canEditScope: boolean;
  canEditProfile: boolean;
  assigneeOptions: User[];
  role?: Role;
  responses?: StakeholderResponse[];
  onSaveResponse?: (id: string, responseText: string) => Promise<void>;
  onSubmitResponse?: (id: string) => Promise<void>;
  onReviewResponse?: (id: string, status: "reviewed" | "needs_more_info", comment: string) => Promise<void>;
  onUploadEvidence?: (id: string, file: File) => Promise<void>;
  onDeleteEvidence?: (id: string, documentID: string) => Promise<void>;
  onDownloadEvidence?: (id: string, document: ResponseDocument) => Promise<void>;
}) {
  if (rows.length === 0) {
    return <div className="empty-state">No outcomes found for this Function.</div>;
  }

  return (
    <div className="assessment-list">
      {rows.map((row) => <AssessmentCard
        key={row.id}
        row={row}
        onSave={onSave}
        canEditScope={canEditScope}
        canEditProfile={canEditProfile}
        assigneeOptions={assigneeOptions}
        role={role}
        response={responses.find((item) => item.subcategoryID === row.subcategoryID) ?? emptyResponse(row)}
        onSaveResponse={onSaveResponse}
        onSubmitResponse={onSubmitResponse}
        onReviewResponse={onReviewResponse}
        onUploadEvidence={onUploadEvidence}
        onDeleteEvidence={onDeleteEvidence}
        onDownloadEvidence={onDownloadEvidence}
      />)}
    </div>
  );
}

function emptyResponse(row: ProfileRow): StakeholderResponse {
  return { id: "", projectID: row.projectID, subcategoryID: row.subcategoryID, responseText: "", status: "draft", respondedBy: null, submittedAt: null, reviewComment: "", reviewedBy: null, reviewedAt: null, createdAt: "", updatedAt: "", documents: [] };
}
