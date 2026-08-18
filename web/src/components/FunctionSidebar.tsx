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

export type WorkspaceMode = "Scope & Assignment" | "My Work" | "Review Queue" | "Read-only";

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
};

export function FunctionSidebar({ functions, selectedCode, onSelect, progressByFunction, mode }: Props) {
  return (
    <nav className="sidebar" aria-label="CSF functions">
      <div className="brand">CSF / Workspace</div>
      <div className="sidebar-theme-control"><ThemeToggle /></div>
      {mode && <p className="workspace-mode" aria-label="Active role mode">{mode}</p>}
      <div>
        {functions.map((fn) => (
          <button className={"nav-item " + (selectedCode === fn.code ? "active" : "")} type="button" aria-current={selectedCode === fn.code ? "page" : undefined} key={fn.code} onClick={() => onSelect(fn.code)}>
            <span className="nav-label">{fn.code} <span className="muted">{fn.name}</span></span>
            {progressByFunction?.[fn.code] && (
              <span
                className="nav-meta"
              >
                <strong>
                  {progressByFunction[fn.code]!.coveragePct !== undefined
                    ? displayCoverage(progressByFunction[fn.code]!.coveragePct ?? progressByFunction[fn.code]!.value)
                    : progressByFunction[fn.code]!.value}
                </strong>
                <small>
                  {progressByFunction[fn.code]!.coveragePct !== undefined
                    ? `${progressByFunction[fn.code]!.includedCount ?? progressByFunction[fn.code]!.value} included`
                    : displayProgressLabel(progressByFunction[fn.code]!.label)}
                </small>
                {progressByFunction[fn.code]!.attention !== undefined && progressByFunction[fn.code]!.attention! > 0 && (
                  <em>{progressByFunction[fn.code]!.attention} {progressByFunction[fn.code]!.attentionLabel}</em>
                )}
              </span>
            )}
          </button>
        ))}
      </div>
    </nav>
  );
}
