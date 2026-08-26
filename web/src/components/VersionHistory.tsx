"use client";

import { useState } from "react";
import type { Project } from "../lib/types";
import { useDialogFocus } from "../lib/useDialogFocus";

type Props = {
  currentProject: Project;
  versions: Project[];
  canCreate: boolean;
  loading?: boolean;
  error?: string;
  onCreateVersion: () => Promise<void>;
  onOpenVersion: (project: Project) => void;
  onRetry?: () => void;
};

function versionNumber(project: Project) {
  return project.versionNumber ?? 1;
}

function statusLabel(status: string) {
  if (status === "closed") return "Finalized";
  if (status === "in_review") return "In review";
  return status.replaceAll("_", " ");
}

function versionDate(project: Project) {
  if (!project.createdAt) return "";
  const date = new Date(project.createdAt);
  if (Number.isNaN(date.getTime())) return project.createdAt;
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

export function VersionHistory({ currentProject, versions, canCreate, loading = false, error = "", onCreateVersion, onOpenVersion, onRetry }: Props) {
  const [state, setState] = useState<"idle" | "creating" | "error">("idle");
  const [createError, setCreateError] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);
  const currentVersion = versionNumber(currentProject);
  const dialogFocus = useDialogFocus(confirmOpen, () => setConfirmOpen(false));

  async function createVersion() {
    setConfirmOpen(false);
    setState("creating");
    setCreateError("");
    try {
      await onCreateVersion();
      setState("idle");
    } catch (cause) {
      setState("error");
      setCreateError(cause instanceof Error ? cause.message : "Could not start a new assessment version");
    }
  }

  return (
    <section className="version-history-panel panel" aria-labelledby="version-history-heading">
      <div className="version-history-heading">
        <div>
          <span className="eyebrow">Assessment lifecycle</span>
          <h2 id="version-history-heading">Version history</h2>
        </div>
        {canCreate && (
          <button className="primary" type="button" disabled={state === "creating"} onClick={(event) => { dialogFocus.triggerRef.current = event.currentTarget; setConfirmOpen(true); }}>
            {state === "creating" ? "Starting…" : "Start new assessment"}
          </button>
        )}
      </div>
      <p className="version-history-description">Each assessment version keeps its own responses, evidence, review decisions, reports, and Action Plan.</p>
      {loading ? (
        <p className="version-history-state" role="status" aria-busy="true">Loading version history…</p>
      ) : error ? (
        <div className="version-history-state" role="alert">
          <p>Version history could not be loaded.</p>
          {onRetry && <button className="secondary" type="button" onClick={onRetry}>Try again</button>}
        </div>
      ) : (
        <div className="version-history-list">
          {versions.map((project) => {
            const number = versionNumber(project);
            const current = project.id === currentProject.id;
            return (
              <article className={`version-history-item ${current ? "is-current" : ""}`} key={project.id}>
                <div className="version-history-item-copy">
                  <span className="version-history-number">Assessment v{number}</span>
                  <strong>{project.name}</strong>
                  <small>{statusLabel(project.status)}{versionDate(project) ? ` · ${versionDate(project)}` : ""}</small>
                </div>
                {current ? (
                  <span className="status-chip">Current version</span>
                ) : (
                  <button className="secondary" type="button" aria-label={`Open assessment v${number}`} onClick={() => onOpenVersion(project)}>Open version</button>
                )}
              </article>
            );
          })}
        </div>
      )}
      {createError && <p className="error version-history-error" role="alert">{createError}</p>}
      {confirmOpen && (
        <section ref={dialogFocus.dialogRef} className="delete-confirmation version-history-confirmation" role="dialog" aria-modal="true" aria-labelledby="version-confirmation-title">
          <div>
            <h2 id="version-confirmation-title">Start a new assessment?</h2>
            <p>Start after Assessment v{currentVersion}. Scope assignments will be copied; responses, evidence, reviews, and Action Plan items will start empty in the new version.</p>
          </div>
          <div className="confirmation-actions">
            <button className="secondary" type="button" onClick={() => setConfirmOpen(false)}>Cancel</button>
            <button className="primary" type="button" onClick={() => void createVersion()}>Confirm start</button>
          </div>
        </section>
      )}
    </section>
  );
}
