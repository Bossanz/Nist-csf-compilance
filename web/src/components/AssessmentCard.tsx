"use client";

import { useState } from "react";
import type { CoverageLevel, EvidencePreview, ProfilePatch, ProfileRow, ResponseDocument, Role, StakeholderResponse, User } from "../lib/types";
import { StakeholderResponsePanel } from "./StakeholderResponsePanel";

const coverageLevels: Array<{ value: CoverageLevel; label: string }> = [
  { value: "none", label: "None" },
  { value: "partial", label: "Partial" },
  { value: "substantial", label: "Substantial" },
  { value: "full", label: "Full" },
];

type Draft = {
  included: boolean;
  rationale: string;
  assignedUserID: string | null;
  currentPriority: string;
  currentCoverageLevel: CoverageLevel;
  currentStatusText: string;
  currentPoliciesText: string;
  targetPriority: string;
  targetCoverageLevel: CoverageLevel;
  targetApproachText: string;
  notes: string;
  considerations: string;
};

type SaveState = "saved" | "dirty" | "saving" | "error";

function createDraft(row: ProfileRow): Draft {
  return {
    included: row.included,
    rationale: row.rationale,
    assignedUserID: row.assignedUserID,
    currentPriority: row.currentPriority,
    currentCoverageLevel: row.currentCoverageLevel,
    currentStatusText: row.currentStatusText,
    currentPoliciesText: row.currentPoliciesText,
    targetPriority: row.targetPriority,
    targetCoverageLevel: row.targetCoverageLevel,
    targetApproachText: row.targetApproachText,
    notes: row.notes,
    considerations: row.considerations,
  };
}

const responseStatusLabels: Record<StakeholderResponse["status"], string> = {
  draft: "Draft",
  submitted: "Submitted",
  reviewed: "Reviewed",
  needs_more_info: "Needs more information",
};

function getResponseSummary(response?: StakeholderResponse) {
  const hasActivity = Boolean(response && (
    response.id
    || response.responseText.trim()
    || response.documents.length > 0
    || response.status !== "draft"
    || response.reviewComment.trim()
  ));

  return hasActivity
    ? { label: responseStatusLabels[response!.status], statusClass: response!.status, hasActivity: true }
    : { label: "Not started", statusClass: "not_started", hasActivity: false };
}

function displayProfileValue(value: string) {
  return value.trim() || "Not provided";
}

function coverageLabel(value: CoverageLevel) {
  return coverageLevels.find((level) => level.value === value)?.label ?? value;
}

function ProfileReference({ draft, id }: { draft: Draft; id: string }) {
  return (
    <section className="profile-reference" aria-label="Profile reference">
      <div className="profile-reference-heading">
        <div>
          <p className="section-context">Reference</p>
          <h3>Assessment profile</h3>
        </div>
        <span className="status-chip">Read only</span>
      </div>
      <div className="profile-reference-columns">
        <section className="profile-reference-column current-reference" aria-labelledby={"current-profile-reference-" + id}>
          <h4 id={"current-profile-reference-" + id}>Current profile</h4>
          <dl className="profile-reference-list">
            <div><dt>Priority</dt><dd>{displayProfileValue(draft.currentPriority)}</dd></div>
            <div><dt>Coverage</dt><dd>{coverageLabel(draft.currentCoverageLevel)}</dd></div>
            <div><dt>Activities</dt><dd>{displayProfileValue(draft.currentStatusText)}</dd></div>
            <div><dt>Policies and procedures</dt><dd>{displayProfileValue(draft.currentPoliciesText)}</dd></div>
          </dl>
        </section>
        <section className="profile-reference-column target-reference" aria-labelledby={"target-profile-reference-" + id}>
          <h4 id={"target-profile-reference-" + id}>Target profile</h4>
          <dl className="profile-reference-list">
            <div><dt>Priority</dt><dd>{displayProfileValue(draft.targetPriority)}</dd></div>
            <div><dt>Coverage</dt><dd>{coverageLabel(draft.targetCoverageLevel)}</dd></div>
            <div><dt>Approach</dt><dd>{displayProfileValue(draft.targetApproachText)}</dd></div>
          </dl>
        </section>
      </div>
      <dl className="profile-reference-notes">
        <div><dt>Notes</dt><dd>{displayProfileValue(draft.notes)}</dd></div>
        <div><dt>Considerations</dt><dd>{displayProfileValue(draft.considerations)}</dd></div>
      </dl>
    </section>
  );
}

export function AssessmentCard({
  row,
  onSave,
  canEditScope,
  canEditProfile,
  assigneeOptions,
  role,
  response,
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
  row: ProfileRow;
  onSave: (id: string, patch: ProfilePatch) => Promise<void>;
  canEditScope: boolean;
  canEditProfile: boolean;
  assigneeOptions: User[];
  role?: Role;
  response?: StakeholderResponse;
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
  const [expanded, setExpanded] = useState(false);
  const [draft, setDraft] = useState<Draft>(() => createDraft(row));
  const [state, setState] = useState<SaveState>("saved");
  const [error, setError] = useState("");
  const eligibleAssignees = assigneeOptions.filter((user) => user.status === "active" && (user.role === "org_admin" || user.role === "assessor"));
  const detailID = `assessment-body-${row.id}`;
  const responseSummary = getResponseSummary(response);
  const assignmentLabel = row.assignedUserID
    ? row.assignedUserName
      ? "Assigned to " + row.assignedUserName
      : "Assigned stakeholder"
    : "Unassigned";
  const isReadOnlyWorkspaceRole = role === "counselor" || role === "counselor_admin" || role === "reviewer" || role === "viewer";
  const showResponsePanel = Boolean(response && (canEditProfile || isReadOnlyWorkspaceRole || responseSummary.hasActivity));
  const currentCoverageLabel = coverageLabel(draft.currentCoverageLevel);
  const targetCoverageLabel = coverageLabel(draft.targetCoverageLevel);
  const evidenceCount = response?.documents.length ?? 0;

  function update<K extends keyof Draft>(field: K, value: Draft[K]) {
    setDraft((current) => ({ ...current, [field]: value }));
    setState("dirty");
  }

  async function save() {
    setError("");
    setState("saving");
    try {
      await onSave(row.subcategoryID, buildPatch());
      setState("saved");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Save failed");
      setState("error");
    }
  }

  function buildPatch(): ProfilePatch {
    if (canEditScope) {
      return { included: draft.included, rationale: draft.rationale, assignedUserID: draft.assignedUserID };
    }
    return {
      currentPriority: draft.currentPriority,
      currentCoverageLevel: draft.currentCoverageLevel,
      currentStatusText: draft.currentStatusText,
      currentPoliciesText: draft.currentPoliciesText,
      targetPriority: draft.targetPriority,
      targetCoverageLevel: draft.targetCoverageLevel,
      targetApproachText: draft.targetApproachText,
      notes: draft.notes,
      considerations: draft.considerations,
    };
  }

  return (
      <article className={`assessment-card ${draft.included ? "is-included" : "is-excluded"}`}>
      <button
        className="assessment-summary"
        type="button"
        aria-expanded={expanded}
        aria-controls={detailID}
        onClick={() => setExpanded((value) => !value)}
      >
        <span className="outcome-code">{row.subcategoryCode}</span>
        <span className="outcome-copy">
          <strong>{row.description}</strong>
          <small>{draft.included ? "Included in profile" : "Out of scope"}</small>
          <span className="outcome-meta">
            <span className={"status-chip outcome-status status-" + responseSummary.statusClass} aria-label={"Response status: " + responseSummary.label}>{responseSummary.label}</span>
            <span className="assignment-label">{assignmentLabel}</span>
          </span>
        </span>
        <span className="coverage-route" role="group" aria-label={`Coverage: ${currentCoverageLabel} to ${targetCoverageLabel}`}>
          <span className="coverage-value current-coverage">
            <small>Current</small>
            <strong>{currentCoverageLabel}</strong>
          </span>
          <svg className="coverage-arrow" viewBox="0 0 16 16" aria-hidden="true"><path d="M2 8h11M9 4l4 4-4 4" /></svg>
          <span className="coverage-value target-coverage">
            <small>Target</small>
            <strong>{targetCoverageLabel}</strong>
          </span>
        </span>
        {response && (
          <span className="evidence-count">
            {evidenceCount} evidence {evidenceCount === 1 ? "file" : "files"}
          </span>
        )}
        <span className="expand-mark" aria-hidden="true"><svg viewBox="0 0 16 16"><path d={expanded ? "M4 8h8" : "M8 4v8M4 8h8"} /></svg></span>
      </button>

      {expanded ? (
        <>
        <div id={detailID} className="assessment-body">
          {canEditScope && <fieldset className="scope-fields">
            <div className="scope-band">
            <label className="check-field">
              <input
                type="checkbox"
                checked={draft.included}
                onChange={(event) => update("included", event.target.checked)}
              />
              <span>Include in profile</span>
            </label>
            <label className="field grow">
              <span>Rationale</span>
              <input
                value={draft.rationale}
                onChange={(event) => update("rationale", event.target.value)}
                placeholder="Why this outcome is in or out of scope"
              />
            </label>
            <label className="field grow">
              <span>Responsible stakeholder</span>
              <select
                aria-label={`Responsible stakeholder for ${row.subcategoryCode}`}
                value={draft.assignedUserID ?? ""}
                onChange={(event) => update("assignedUserID", event.target.value || null)}
              >
                <option value="">Select stakeholder</option>
                {eligibleAssignees.map((user) => <option key={user.id} value={user.id}>{user.name} — {user.email}</option>)}
              </select>
            </label>
            </div>
          </fieldset>}

          {canEditProfile ? <fieldset className="profile-fields">
            <div className="profile-columns">
            <section className="profile-column current-column" aria-labelledby={`current-${row.id}`}>
              <div className="column-heading">
                <span>01</span>
                <h3 id={`current-${row.id}`}>Current profile</h3>
              </div>
              <div className="field-grid">
                <label className="field">
                  <span>Current priority</span>
                  <input value={draft.currentPriority} onChange={(event) => update("currentPriority", event.target.value)} placeholder="Low, Medium, High" />
                </label>
                <label className="field">
                  <span>Current coverage</span>
                  <select value={draft.currentCoverageLevel} onChange={(event) => update("currentCoverageLevel", event.target.value as CoverageLevel)}>
                    {coverageLevels.map((level) => <option key={level.value} value={level.value}>{level.label}</option>)}
                  </select>
                </label>
                <label className="field full-width">
                  <span>Current activities</span>
                  <textarea rows={3} value={draft.currentStatusText} onChange={(event) => update("currentStatusText", event.target.value)} placeholder="What is currently practiced?" />
                </label>
                <label className="field full-width">
                  <span>Policies and procedures</span>
                  <textarea rows={3} value={draft.currentPoliciesText} onChange={(event) => update("currentPoliciesText", event.target.value)} placeholder="Related policies, procedures, and guidance" />
                </label>
              </div>
            </section>

            <section className="profile-column target-column" aria-labelledby={`target-${row.id}`}>
              <div className="column-heading">
                <span>02</span>
                <h3 id={`target-${row.id}`}>Target profile</h3>
              </div>
              <div className="field-grid">
                <label className="field">
                  <span>Target priority</span>
                  <input value={draft.targetPriority} onChange={(event) => update("targetPriority", event.target.value)} placeholder="Low, Medium, High" />
                </label>
                <label className="field">
                  <span>Target coverage</span>
                  <select value={draft.targetCoverageLevel} onChange={(event) => update("targetCoverageLevel", event.target.value as CoverageLevel)}>
                    {coverageLevels.map((level) => <option key={level.value} value={level.value}>{level.label}</option>)}
                  </select>
                </label>
                <label className="field full-width">
                  <span>Target approach</span>
                  <textarea rows={7} value={draft.targetApproachText} onChange={(event) => update("targetApproachText", event.target.value)} placeholder="What needs to change to reach the target?" />
                </label>
              </div>
            </section>
            </div>

            <div className="supporting-grid">
            <label className="field">
              <span>Notes</span>
              <textarea rows={2} value={draft.notes} onChange={(event) => update("notes", event.target.value)} />
            </label>
            <label className="field">
              <span>Considerations</span>
              <textarea rows={2} value={draft.considerations} onChange={(event) => update("considerations", event.target.value)} />
            </label>
            </div>
          </fieldset> : <ProfileReference draft={draft} id={row.id} />}

          <div className="assessment-actions">
            <span className={"save-state " + state} role="status">
              {state === "saved" ? "Saved" : state === "dirty" ? "Unsaved changes" : state === "saving" ? "Saving…" : error}
            </span>
            {(canEditScope || canEditProfile) && <button className="primary" type="button" disabled={state === "saving"} onClick={save}>
              {state === "saving" ? "Saving…" : "Save assessment"}
            </button>}
          </div>
        </div>
        {showResponsePanel && role && response && onSaveResponse && onSubmitResponse && onReviewResponse && onUploadEvidence && onDeleteEvidence && onDownloadEvidence && (
          <StakeholderResponsePanel
            role={role}
            response={response}
            onSave={(responseText) => onSaveResponse(row.subcategoryID, responseText)}
            onSubmit={() => onSubmitResponse(row.subcategoryID)}
            onReview={(status, comment) => onReviewResponse(row.subcategoryID, status, comment)}
            onUpload={(file) => onUploadEvidence(row.subcategoryID, file)}
            onDelete={(documentID) => onDeleteEvidence(row.subcategoryID, documentID)}
            onDownload={(document) => onDownloadEvidence(row.subcategoryID, document)}
            onPreview={onPreviewEvidence ? (document) => onPreviewEvidence(row.subcategoryID, document) : undefined}
            preview={evidencePreview?.subcategoryID === row.subcategoryID ? evidencePreview : null}
            onClosePreview={onCloseEvidencePreview}
            previewLoading={previewTargetSubcategoryID === row.subcategoryID && previewLoading}
            previewError={previewTargetSubcategoryID === row.subcategoryID ? previewError : ""}
          />
        )}
        </>
      ) : null}
    </article>
  );
}
