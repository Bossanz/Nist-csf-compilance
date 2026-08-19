"use client";

import { useMemo, useState } from "react";
import type { ProfileRow, RemediationAction, RemediationCreateInput, RemediationPatchInput, RemediationPriority, StakeholderResponse, User } from "../lib/types";

type Props = {
  user: User;
  profile: ProfileRow[];
  responses: StakeholderResponse[];
  actions: RemediationAction[];
  assigneeOptions: User[];
  onCreate: (input: RemediationCreateInput) => Promise<void>;
  onUpdate: (actionID: string, patch: RemediationPatchInput) => Promise<void>;
  onSaveProgress: (actionID: string, progressNote: string) => Promise<void>;
  onSubmit: (actionID: string) => Promise<void>;
  onReview: (actionID: string, decision: "close" | "return", comment: string) => Promise<void>;
  onUploadEvidence: (actionID: string, file: File) => Promise<void>;
  onDeleteEvidence: (actionID: string, evidenceID: string) => Promise<void>;
  onDownloadEvidence: (actionID: string, evidenceID: string, originalName: string) => Promise<void>;
};

const coverageRank = { none: 0, partial: 1, substantial: 2, full: 3 } as const;
const statusLabel = { open: "Open", in_progress: "In progress", awaiting_review: "Awaiting review", closed: "Closed" } as const;
const priorities: RemediationPriority[] = ["low", "medium", "high", "critical"];

function dateInput(value: string) {
  return value.slice(0, 10);
}

function displayDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat("en-GB", { dateStyle: "medium" }).format(date);
}

function isOverdue(action: RemediationAction) {
  if (action.status === "closed") return false;
  const due = new Date(dateInput(action.dueDate) + "T23:59:59");
  return !Number.isNaN(due.getTime()) && due.getTime() < Date.now();
}

export function ActionPlan({ user, profile, responses, actions, assigneeOptions, onCreate, onUpdate, onSaveProgress, onSubmit, onReview, onUploadEvidence, onDeleteEvidence, onDownloadEvidence }: Props) {
  const isCounselor = user.userType === "counselor";
  const isOwnerRole = user.role === "org_admin" || user.role === "assessor";
  const approvedByOutcome = useMemo(() => new Map(responses.map((response) => [response.subcategoryID, response.status === "reviewed"])), [responses]);
  const eligibleOutcomes = useMemo(() => profile.filter((row) => row.included && approvedByOutcome.get(row.subcategoryID) && coverageRank[row.currentCoverageLevel] < coverageRank[row.targetCoverageLevel]), [approvedByOutcome, profile]);
  const owners = assigneeOptions.filter((option) => option.status === "active" && (option.role === "org_admin" || option.role === "assessor"));
  const visibleActions = isOwnerRole ? actions.filter((action) => action.ownerUserID === user.id) : actions;
  const counts = {
    open: visibleActions.filter((action) => action.status === "open").length,
    inProgress: visibleActions.filter((action) => action.status === "in_progress").length,
    awaiting: visibleActions.filter((action) => action.status === "awaiting_review").length,
    overdue: visibleActions.filter(isOverdue).length,
    closed: visibleActions.filter((action) => action.status === "closed").length,
  };
  const [form, setForm] = useState<RemediationCreateInput>({ subcategoryID: "", title: "", description: "", desiredResult: "", priority: "medium", ownerUserID: "", dueDate: "" });
  const [createState, setCreateState] = useState<"idle" | "saving" | "error">("idle");
  const [createError, setCreateError] = useState("");

  async function createAction(event: React.FormEvent) {
    event.preventDefault();
    setCreateState("saving");
    setCreateError("");
    try {
      await onCreate(form);
      setForm({ subcategoryID: "", title: "", description: "", desiredResult: "", priority: "medium", ownerUserID: "", dueDate: "" });
      setCreateState("idle");
    } catch (cause) {
      setCreateState("error");
      setCreateError(cause instanceof Error ? cause.message : "Could not create action");
    }
  }

  return (
    <section className="action-plan" aria-labelledby="action-plan-heading">
      <header className="action-plan-heading">
        <div>
          <h2 id="action-plan-heading">Action Plan</h2>
          <p>Turn approved coverage gaps into assigned remediation work without changing the finalized assessment.</p>
        </div>
        {isCounselor && <strong>{eligibleOutcomes.length} approved {eligibleOutcomes.length === 1 ? "gap" : "gaps"} ready for planning</strong>}
      </header>

      <dl className="action-summary" aria-label="Remediation status summary">
        <div><dt>Open</dt><dd>{counts.open}</dd></div>
        <div><dt>In progress</dt><dd>{counts.inProgress}</dd></div>
        <div><dt>Awaiting review</dt><dd>{counts.awaiting}</dd></div>
        <div className={counts.overdue > 0 ? "action-summary-warning" : ""}><dt>Overdue</dt><dd>{counts.overdue}</dd></div>
        <div><dt>Closed</dt><dd>{counts.closed}</dd></div>
      </dl>

      {isCounselor && (
        <form className="action-create panel" onSubmit={(event) => void createAction(event)}>
          <div className="action-create-intro"><h3>Create remediation action</h3><p>Choose an approved outcome with a measurable Current-to-Target gap.</p></div>
          <label>Outcome<select aria-label="Outcome" required value={form.subcategoryID} onChange={(event) => setForm({ ...form, subcategoryID: event.target.value })}><option value="">Select approved gap</option>{eligibleOutcomes.map((row) => <option key={row.subcategoryID} value={row.subcategoryID}>{row.subcategoryCode} — {row.description}</option>)}</select></label>
          <label>Action title<input aria-label="Action title" required value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} /></label>
          <label className="action-field-wide">Description<textarea aria-label="Description" rows={3} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
          <label className="action-field-wide">Desired result<textarea aria-label="Desired result" rows={3} value={form.desiredResult} onChange={(event) => setForm({ ...form, desiredResult: event.target.value })} /></label>
          <label>Priority<select aria-label="Priority" value={form.priority} onChange={(event) => setForm({ ...form, priority: event.target.value as RemediationPriority })}>{priorities.map((priority) => <option key={priority} value={priority}>{priority[0].toUpperCase() + priority.slice(1)}</option>)}</select></label>
          <label>Owner<select aria-label="Owner" required value={form.ownerUserID} onChange={(event) => setForm({ ...form, ownerUserID: event.target.value })}><option value="">Select owner</option>{owners.map((owner) => <option key={owner.id} value={owner.id}>{owner.name} — {owner.role.replaceAll("_", " ")}</option>)}</select></label>
          <label>Due date<input aria-label="Due date" type="date" required value={form.dueDate} onChange={(event) => setForm({ ...form, dueDate: event.target.value })} /></label>
          <div className="action-create-footer">{createError && <span className="error" role="alert">{createError}</span>}<button className="primary" type="submit" disabled={createState === "saving" || eligibleOutcomes.length === 0}>{createState === "saving" ? "Creating…" : "Create action"}</button></div>
        </form>
      )}

      <div className="action-list">
        {visibleActions.length === 0 ? <div className="empty-state" role="status">{isOwnerRole ? "No remediation actions are assigned to you." : "No remediation actions have been created yet."}</div> : visibleActions.map((action) => (
          <ActionRow key={action.id} action={action} user={user} owners={owners} onUpdate={onUpdate} onSaveProgress={onSaveProgress} onSubmit={onSubmit} onReview={onReview} onUploadEvidence={onUploadEvidence} onDeleteEvidence={onDeleteEvidence} onDownloadEvidence={onDownloadEvidence} />
        ))}
      </div>
    </section>
  );
}

type RowProps = Pick<Props, "user" | "onUpdate" | "onSaveProgress" | "onSubmit" | "onReview" | "onUploadEvidence" | "onDeleteEvidence" | "onDownloadEvidence"> & { action: RemediationAction; owners: User[] };

function ActionRow({ action, user, owners, onUpdate, onSaveProgress, onSubmit, onReview, onUploadEvidence, onDeleteEvidence, onDownloadEvidence }: RowProps) {
  const isCounselor = user.userType === "counselor";
  const isOwner = action.ownerUserID === user.id && (user.role === "org_admin" || user.role === "assessor");
  const canWork = isOwner && (action.status === "open" || action.status === "in_progress");
  const [progress, setProgress] = useState(action.progressNote);
  const [progressSaved, setProgressSaved] = useState(Boolean(action.progressNote.trim()));
  const [reviewComment, setReviewComment] = useState(action.reviewComment);
  const [state, setState] = useState<"idle" | "saving" | "error">("idle");
  const [error, setError] = useState("");
  const [edit, setEdit] = useState<RemediationPatchInput>({ title: action.title, description: action.description, desiredResult: action.desiredResult, priority: action.priority, ownerUserID: action.ownerUserID, dueDate: dateInput(action.dueDate) });

  async function run(work: () => Promise<void>) {
    setState("saving"); setError("");
    try { await work(); setState("idle"); } catch (cause) { setState("error"); setError(cause instanceof Error ? cause.message : "Could not update action"); }
  }

  return (
    <article className="action-row">
      <header className="action-row-header">
        <div className="action-outcome"><strong>{action.outcomeCode}</strong><span>{action.currentCoverageLevel} → {action.targetCoverageLevel}</span></div>
        <div className="action-title"><h3>{action.title}</h3><p>{action.outcomeDescription}</p></div>
        <div className="action-state"><span className={`status-chip remediation-${action.status}`}>{statusLabel[action.status]}</span>{isOverdue(action) && <span className="status-chip remediation-overdue">Overdue</span>}</div>
      </header>
      <dl className="action-facts">
        <div><dt>Owner</dt><dd>{action.ownerName}</dd></div><div><dt>Priority</dt><dd>{action.priority}</dd></div><div><dt>Due</dt><dd>{displayDate(action.dueDate)}</dd></div><div><dt>Evidence</dt><dd>{action.evidence.length}</dd></div>
      </dl>
      <div className="action-copy"><div><h4>Work required</h4><p>{action.description || "No additional description."}</p></div><div><h4>Desired result</h4><p>{action.desiredResult || "No desired result recorded."}</p></div></div>

      {isCounselor && action.status !== "closed" && (
        <details className="action-edit"><summary>Edit assignment</summary><div className="action-edit-grid">
          <label>Action title<input value={edit.title ?? ""} onChange={(event) => setEdit({ ...edit, title: event.target.value })} /></label>
          <label>Priority<select value={edit.priority} onChange={(event) => setEdit({ ...edit, priority: event.target.value as RemediationPriority })}>{priorities.map((priority) => <option key={priority}>{priority}</option>)}</select></label>
          <label>Owner<select value={edit.ownerUserID} onChange={(event) => setEdit({ ...edit, ownerUserID: event.target.value })}>{owners.map((owner) => <option key={owner.id} value={owner.id}>{owner.name}</option>)}</select></label>
          <label>Due date<input type="date" value={edit.dueDate ?? ""} onChange={(event) => setEdit({ ...edit, dueDate: event.target.value })} /></label>
          <label className="action-field-wide">Description<textarea rows={2} value={edit.description ?? ""} onChange={(event) => setEdit({ ...edit, description: event.target.value })} /></label>
          <label className="action-field-wide">Desired result<textarea rows={2} value={edit.desiredResult ?? ""} onChange={(event) => setEdit({ ...edit, desiredResult: event.target.value })} /></label>
          <button className="secondary" type="button" disabled={state === "saving"} onClick={() => void run(() => onUpdate(action.id, edit))}>Save assignment</button>
        </div></details>
      )}

      {(canWork || action.progressNote) && <section className="action-progress"><h4>Progress</h4>{canWork ? <><label>Progress update<textarea aria-label="Progress update" rows={4} value={progress} onChange={(event) => { setProgress(event.target.value); setProgressSaved(false); }} /></label><div className="action-buttons"><button className="secondary" type="button" disabled={state === "saving" || !progress.trim()} onClick={() => void run(async () => { await onSaveProgress(action.id, progress); setProgressSaved(true); })}>Save progress</button><button className="primary" type="button" disabled={state === "saving" || !progressSaved || !progress.trim()} onClick={() => void run(() => onSubmit(action.id))}>Send for review</button></div></> : <p>{action.progressNote}</p>}</section>}

      <section className="action-evidence"><h4>Remediation evidence</h4>{action.evidence.length > 0 && <ul>{action.evidence.map((evidence) => <li key={evidence.id}><button className="button-link" type="button" onClick={() => void onDownloadEvidence(action.id, evidence.id, evidence.originalName)}>{evidence.originalName}</button><small>{Math.ceil(evidence.sizeBytes / 1024)} KB</small>{canWork && <button className="danger" type="button" onClick={() => void run(() => onDeleteEvidence(action.id, evidence.id))}>Delete</button>}</li>)}</ul>}{action.evidence.length === 0 && <p className="muted">No evidence attached.</p>}{canWork && <label className="upload-control">Upload evidence<input aria-label={`Upload evidence for ${action.title}`} type="file" accept=".pdf,.docx,.xlsx,.png,.jpg,.jpeg" onChange={(event) => { const file = event.target.files?.[0]; if (file) void run(() => onUploadEvidence(action.id, file)); }} /></label>}</section>

      {isCounselor && action.status === "awaiting_review" && <section className="action-review"><h4>Counselor review</h4><label>Counselor review<textarea aria-label="Counselor review" rows={3} value={reviewComment} onChange={(event) => setReviewComment(event.target.value)} /></label><div className="action-buttons"><button className="secondary" type="button" disabled={state === "saving" || !reviewComment.trim()} onClick={() => void run(() => onReview(action.id, "return", reviewComment))}>Return for more work</button><button className="primary" type="button" disabled={state === "saving"} onClick={() => void run(() => onReview(action.id, "close", reviewComment))}>Close action</button></div></section>}
      {action.reviewComment && action.status !== "awaiting_review" && <p className="action-review-note"><strong>Counselor comment:</strong> {action.reviewComment}</p>}
      {error && <p className="error action-error" role="alert">{error}</p>}
    </article>
  );
}
