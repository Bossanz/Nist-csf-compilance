"use client";

import { useState } from "react";
import type { CoverageLevel, ProfilePatch, ProfileRow, ResponseDocument, Role, StakeholderResponse } from "../lib/types";
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

function createDraft(row: ProfileRow): Draft {
  return {
    included: row.included,
    rationale: row.rationale,
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

export function AssessmentCard({
  row,
  onSave,
  readOnly = false,
  role,
  response,
  onSaveResponse,
  onSubmitResponse,
  onReviewResponse,
  onUploadEvidence,
  onDeleteEvidence,
  onDownloadEvidence,
}: {
  row: ProfileRow;
  onSave: (id: string, patch: ProfilePatch) => Promise<void>;
  readOnly?: boolean;
  role?: Role;
  response?: StakeholderResponse;
  onSaveResponse?: (id: string, responseText: string) => Promise<void>;
  onSubmitResponse?: (id: string) => Promise<void>;
  onReviewResponse?: (id: string, status: "reviewed" | "needs_more_info", comment: string) => Promise<void>;
  onUploadEvidence?: (id: string, file: File) => Promise<void>;
  onDeleteEvidence?: (id: string, documentID: string) => Promise<void>;
  onDownloadEvidence?: (id: string, document: ResponseDocument) => Promise<void>;
}) {
  const [expanded, setExpanded] = useState(false);
  const [draft, setDraft] = useState<Draft>(() => createDraft(row));
  const [state, setState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [error, setError] = useState("");
  const detailID = `assessment-body-${row.id}`;

  function update<K extends keyof Draft>(field: K, value: Draft[K]) {
    setDraft((current) => ({ ...current, [field]: value }));
    setState("idle");
  }

  async function save() {
    setError("");
    setState("saving");
    try {
      await onSave(row.subcategoryID, draft);
      setState("saved");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Save failed");
      setState("error");
    }
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
        </span>
        <span className="coverage-route" aria-label={`Coverage ${draft.currentCoverageLevel} to ${draft.targetCoverageLevel}`}>
          <span>{draft.currentCoverageLevel}</span>
          <svg className="coverage-arrow" viewBox="0 0 16 16" aria-hidden="true"><path d="M2 8h11M9 4l4 4-4 4" /></svg>
          <span>{draft.targetCoverageLevel}</span>
        </span>
        <span className="expand-mark" aria-hidden="true"><svg viewBox="0 0 16 16"><path d={expanded ? "M4 8h8" : "M8 4v8M4 8h8"} /></svg></span>
      </button>

      {expanded ? (
        <>
        <fieldset id={detailID} className="assessment-body" disabled={readOnly}>
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
          </div>

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

          <div className="assessment-actions">
            <span className={`save-state ${state}`} role="status">
              {state === "saved" ? "Saved" : state === "error" ? error : "Changes are saved per outcome"}
            </span>
            {!readOnly && <button className="primary" type="button" disabled={state === "saving"} onClick={save}>
              {state === "saving" ? "Saving…" : "Save assessment"}
            </button>}
          </div>
        </fieldset>
        {role && response && onSaveResponse && onSubmitResponse && onReviewResponse && onUploadEvidence && onDeleteEvidence && onDownloadEvidence && (
          <StakeholderResponsePanel
            role={role}
            response={response}
            onSave={(responseText) => onSaveResponse(row.subcategoryID, responseText)}
            onSubmit={() => onSubmitResponse(row.subcategoryID)}
            onReview={(status, comment) => onReviewResponse(row.subcategoryID, status, comment)}
            onUpload={(file) => onUploadEvidence(row.subcategoryID, file)}
            onDelete={(documentID) => onDeleteEvidence(row.subcategoryID, documentID)}
            onDownload={(document) => onDownloadEvidence(row.subcategoryID, document)}
          />
        )}
        </>
      ) : null}
    </article>
  );
}
