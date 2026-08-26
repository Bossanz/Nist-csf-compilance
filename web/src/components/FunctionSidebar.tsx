import { useState } from "react";
import type { FunctionNode } from "../lib/types";
import { ThemeToggle } from "./ThemeToggle";

export type FunctionProgress = {
  value: number;
  label: string;
  coveragePct?: number;
  includedCount?: number;
  attention?: number;
  attentionLabel?: string;
};

export type WorkspaceMode = "Scope & Assignment" | "My Work" | "Review Queue" | "Read-only" | "Audit View";
export type WorkspaceSurface = "overview" | "assignment" | "actions" | "log";

function displayProgressLabel(label: string) {
  if (label === "submitted") return "Reviewing";
  if (label === "reviewed") return "Approved";
  return label;
}

function displayCoverage(value: number) {
  const bounded = Math.min(100, Math.max(0, Number.isFinite(value) ? value : 0));
  return `${Math.round(bounded)}%`;
}

type Props = {
  functions: FunctionNode[];
  selectedCode: string;
  onSelect: (code: string) => void;
  progressByFunction?: Record<string, FunctionProgress>;
  mode?: WorkspaceMode;
  activeSurface?: WorkspaceSurface;
  onSelectSurface?: (surface: WorkspaceSurface) => void;
  onBack?: () => void;
};

export function FunctionSidebar({
  functions,
  selectedCode,
  onSelect,
  progressByFunction,
  mode,
  activeSurface = "assignment",
  onSelectSurface,
  onBack,
}: Props) {
  const [assignmentExpanded, setAssignmentExpanded] = useState(true);
  const selectSurface = onSelectSurface ?? (() => undefined);

  return (
    <nav className="sidebar" aria-label="CSF functions and workspace navigation">
      <div className="sidebar-content">
        <div className="brand">CSF / Workspace</div>
        {mode && <p className="workspace-mode sr-only" aria-label="Active role mode">{mode}</p>}
        <div className="sidebar-nav">
          <button
            className={`sidebar-link ${activeSurface === "overview" ? "active" : ""}`}
            type="button"
            aria-current={activeSurface === "overview" ? "page" : undefined}
            onClick={() => selectSurface("overview")}
          >
            Overview
          </button>
          <button
            className={`sidebar-link sidebar-group-toggle ${activeSurface === "assignment" ? "active" : ""}`}
            type="button"
            aria-expanded={assignmentExpanded}
            aria-controls="workspace-assignment-functions"
            onClick={() => setAssignmentExpanded((expanded) => !expanded)}
          >
            <span>Assignment</span>
            <svg className={`sidebar-chevron ${assignmentExpanded ? "is-open" : ""}`} viewBox="0 0 16 16" aria-hidden="true"><path d="m4 6 4 4 4-4" /></svg>
          </button>
          {assignmentExpanded && (
            <div className="sidebar-function-list" id="workspace-assignment-functions" role="group" aria-label="Assignment functions">
              {functions.map((fn) => {
                const progress = progressByFunction?.[fn.code];
                const progressLabel = progress
                  ? progress.coveragePct !== undefined
                    ? `${displayCoverage(progress.coveragePct)} ${progress.includedCount ?? progress.value} included`
                    : `${progress.value} ${displayProgressLabel(progress.label)}`
                  : "";
                const attentionLabel = progress?.attention !== undefined && progress.attention > 0
                  ? `${progress.attention} ${progress.attentionLabel ?? "attention"}`
                  : "";
                return (
                  <button
                    className={`sidebar-function ${activeSurface === "assignment" && selectedCode === fn.code ? "active" : ""}`}
                    type="button"
                    aria-current={activeSurface === "assignment" && selectedCode === fn.code ? "page" : undefined}
                    aria-label={`${fn.code} ${fn.name}${progressLabel ? ` ${progressLabel}` : ""}${attentionLabel ? ` ${attentionLabel}` : ""}`}
                    key={fn.code}
                    title={fn.name}
                    onClick={() => {
                      selectSurface("assignment");
                      onSelect(fn.code);
                    }}
                  >
                    <span className="sidebar-function-code">{fn.code}</span>
                    {progress && (
                      <span className="sidebar-function-meta" aria-hidden="true">
                        {progress.coveragePct !== undefined ? (
                          <>
                            <strong>{displayCoverage(progress.coveragePct)}</strong>
                            <small>{progress.includedCount ?? progress.value} included</small>
                          </>
                        ) : (
                          <small>{progress.value} {displayProgressLabel(progress.label)}</small>
                        )}
                        {progress.attention !== undefined && progress.attention > 0 && (
                          <em>{progress.attention} {progress.attentionLabel}</em>
                        )}
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          )}
          <button
            className={`sidebar-link ${activeSurface === "actions" ? "active" : ""}`}
            type="button"
            aria-current={activeSurface === "actions" ? "page" : undefined}
            onClick={() => selectSurface("actions")}
          >
            Action Plan
          </button>
          <button
            className={`sidebar-link ${activeSurface === "log" ? "active" : ""}`}
            type="button"
            aria-current={activeSurface === "log" ? "page" : undefined}
            onClick={() => selectSurface("log")}
          >
            Log
          </button>
        </div>
      </div>
      <div className="sidebar-footer">
        <button className="sidebar-icon-button" type="button" aria-label="Back to organization" title="Back to organization" onClick={onBack}>
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M8 3H4.5A1.5 1.5 0 0 0 3 4.5v11A1.5 1.5 0 0 0 4.5 17H8M11 6l4 4-4 4M7 10h8" /></svg>
        </button>
        <ThemeToggle compact />
      </div>
    </nav>
  );
}
