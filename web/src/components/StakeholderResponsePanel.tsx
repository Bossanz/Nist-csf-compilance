"use client";

import { useEffect, useState } from "react";
import type { EvidencePreview, ResponseDocument, ResponseStatus, Role, StakeholderResponse } from "../lib/types";

const statusLabels: Record<ResponseStatus, string> = {
  draft: "Draft",
  submitted: "Submitted",
  reviewed: "Reviewed",
  needs_more_info: "Needs more information",
};

type SaveState = "saved" | "dirty" | "saving" | "error";

type Props = {
  role: Role;
  response: StakeholderResponse;
  onSave: (responseText: string) => Promise<void>;
  onSubmit: () => Promise<void>;
  onReview: (status: "reviewed" | "needs_more_info", comment: string) => Promise<void>;
  onUpload: (file: File) => Promise<void>;
  onDelete: (documentID: string) => Promise<void>;
  onDownload: (document: ResponseDocument) => Promise<void>;
  onPreview?: (document: ResponseDocument) => Promise<void>;
  preview?: EvidencePreview | null;
  onClosePreview?: () => void;
  previewLoading?: boolean;
  previewError?: string;
};

export function StakeholderResponsePanel({ role, response, onSave, onSubmit, onReview, onUpload, onDelete, onDownload, onPreview, preview, onClosePreview, previewLoading, previewError }: Props) {
  const [responseText, setResponseText] = useState(response.responseText);
  const [reviewStatus, setReviewStatus] = useState<"reviewed" | "needs_more_info">(response.status === "needs_more_info" ? "needs_more_info" : "reviewed");
  const [reviewComment, setReviewComment] = useState(response.reviewComment);
  const [state, setState] = useState<SaveState>("saved");
  const [error, setError] = useState("");
  const canRespond = role === "org_admin" || role === "assessor";
  const canEditResponse = canRespond && (response.status === "draft" || response.status === "needs_more_info");
  const canReview = role === "reviewer" && response.status === "submitted";

  useEffect(() => {
    setResponseText(response.responseText);
    setReviewStatus(response.status === "needs_more_info" ? "needs_more_info" : "reviewed");
    setReviewComment(response.reviewComment);
    setState("saved");
  }, [response.responseText, response.status, response.reviewComment]);

  async function run(action: () => Promise<void>) {
    setState("saving");
    setError("");
    try {
      await action();
      setState("saved");
    } catch (cause) {
      setState("error");
      setError(cause instanceof Error ? cause.message : "Could not save response");
    }
  }

  async function upload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    await run(() => onUpload(file));
    event.target.value = "";
  }

  return (
    <section className="response-panel" aria-labelledby={`response-${response.subcategoryID}`}>
      <div className="response-heading">
        <div>
          <h3 id={`response-${response.subcategoryID}`}>Stakeholder response</h3>
          <p className="response-note">Response and evidence for this outcome.</p>
        </div>
        <span className={`response-status status-${response.status}`} role="status" aria-live="polite">{statusLabels[response.status]}</span>
      </div>

      <label className="field">
        <span>Client response</span>
        <textarea
          rows={4}
          value={responseText}
          disabled={!canEditResponse}
          onChange={(event) => { setResponseText(event.target.value); setState("dirty"); }}
          placeholder="Describe how this outcome is handled in your organization"
        />
      </label>

      {canEditResponse && (
        <div className="response-actions">
          <label className="upload-control">
            <span>Upload evidence</span>
            <input type="file" accept=".pdf,.docx,.xlsx,.png,.jpg,.jpeg" onChange={upload} disabled={state === "saving"} />
            <small>PDF, DOCX, XLSX, PNG, or JPEG · max 20 MB</small>
          </label>
          <div className="response-buttons">
            <button className="secondary" type="button" disabled={state === "saving"} onClick={() => void run(() => onSave(responseText))}>Save response</button>
            <button className="primary" type="button" disabled={state === "saving" || !response.id} onClick={() => void run(onSubmit)}>Submit response</button>
          </div>
        </div>
      )}

      {canReview && (
        <div className="review-panel">
          <h4>Reviewer final decision</h4>
          <div className="field-grid">
            <label className="field">
              <span>Review status</span>
              <select aria-label={"Review status for " + response.subcategoryID} value={reviewStatus} onChange={(event) => { setReviewStatus(event.target.value as "reviewed" | "needs_more_info"); setState("dirty"); }}>
                <option value="reviewed">Reviewed</option>
                <option value="needs_more_info">Needs more information</option>
              </select>
            </label>
            <label className="field">
              <span>Review comment</span>
              <textarea aria-label={"Review comment for " + response.subcategoryID} rows={2} value={reviewComment} onChange={(event) => { setReviewComment(event.target.value); setState("dirty"); }} placeholder="What should be accepted or clarified?" />
            </label>
          </div>
          <button className="primary" type="button" disabled={state === "saving"} onClick={() => void run(() => onReview(reviewStatus, reviewComment))}>Save review</button>
        </div>
      )}

      {response.documents.length > 0 && (
        <div className="evidence-list">
          <span className="field-label">Supporting evidence</span>
          <ul>
            {response.documents.map((document) => <li key={document.id}>
              <button className="button-link" type="button" onClick={() => void onDownload(document)}>{document.originalName}</button>
              <small>{formatBytes(document.sizeBytes)}</small>
              {canPreview(document.mimeType) && onPreview && <button className="secondary" type="button" aria-label={`Preview ${document.originalName}`} onClick={() => void onPreview(document)}>Preview</button>}
              {canEditResponse && <button className="danger" type="button" onClick={() => void run(() => onDelete(document.id))}>Delete</button>}
              {preview?.documentID === document.id && (
                <div className="evidence-preview" aria-label={`${document.originalName} preview`}>
                  <div className="evidence-preview-heading">
                    <strong>{document.originalName}</strong>
                    {onClosePreview && <button className="text-button" type="button" aria-label="Close preview" onClick={onClosePreview}>Close preview</button>}
                  </div>
                  {preview.mimeType.startsWith("image/") ? (
                    <img src={preview.url} alt={`${document.originalName} preview`} />
                  ) : (
                    <iframe title={`${document.originalName} preview`} src={preview.url} />
                  )}
                </div>
              )}
            </li>)}
          </ul>
          {previewLoading && <span className="save-state" role="status">Loading preview...</span>}
          {previewError && <div className="error" role="alert">{previewError}</div>}
        </div>
      )}

      <div className="response-footer">
        <span className={"save-state " + state} role="status">
          {state === "saved" ? "Saved" : state === "dirty" ? "Unsaved changes" : state === "saving" ? "Saving…" : error}
        </span>
        {response.status === "draft" && <span className="response-note">Not submitted yet</span>}
        {response.status !== "draft" && response.reviewComment && <span className="response-note">{response.reviewComment}</span>}
      </div>
    </section>
  );
}

function formatBytes(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function canPreview(mimeType: string) {
  return mimeType === "application/pdf" || mimeType === "image/png" || mimeType === "image/jpeg" || mimeType === "image/jpg";
}
