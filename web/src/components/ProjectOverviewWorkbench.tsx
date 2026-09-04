import type { FunctionNode, Project, Summary, User } from "../lib/types";
import { getAttentionLabel, getProjectStatusLabel } from "../lib/workspaceMetrics";
import type { FunctionProgress } from "./FunctionSidebar";

type Props = {
  user: User;
  project: Project;
  summary: Summary;
  functions: FunctionNode[];
  functionProgress: Record<string, FunctionProgress>;
  scopeSubmitted: boolean;
  isCounselor: boolean;
  onOpenAssignment: () => void;
};

function boundedPercentage(value: number | undefined) {
  const safeValue = Number.isFinite(value) ? value ?? 0 : 0;
  return Math.round(Math.min(100, Math.max(0, safeValue)));
}

function getNextAction(user: User, scopeSubmitted: boolean, isCounselor: boolean) {
  if (isCounselor && !scopeSubmitted) {
    return {
      label: "Set and submit project scope",
      description: "Select included outcomes and assign the responsible stakeholder.",
      actionLabel: "Open scope assignment",
    };
  }
  if (isCounselor) {
    return {
      label: "Review stakeholder progress",
      description: "Open Assignment to inspect responses, evidence, and outcomes that need support.",
      actionLabel: "Open scope assignment",
    };
  }
  if (user.role === "reviewer") {
    return {
      label: "Review submitted outcomes",
      description: "Open your review queue and record a decision for each submitted response.",
      actionLabel: "Open review queue",
    };
  }
  if (user.role === "viewer" || user.role === "auditor") {
    return {
      label: "Read project posture",
      description: "Review included outcomes, evidence, and the activity history available to you.",
      actionLabel: "Open read-only assignment",
    };
  }
  return {
    label: "Needs your input",
    description: "Open your assigned outcomes and complete the response and evidence fields.",
    actionLabel: "Open assigned work",
  };
}

export function ProjectOverviewWorkbench({ user, project, summary, functions, functionProgress, scopeSubmitted, isCounselor, onOpenAssignment }: Props) {
  const nextAction = getNextAction(user, scopeSubmitted, isCounselor);
  const attentionLabel = getAttentionLabel(user.role);

  return (
    <>
      <section className="overview-workbench posture-panel" aria-label="Project posture">
        <div className="workbench-section-heading">
          <div>
            <p className="section-context">Current posture</p>
            <h2>Project posture</h2>
            <p className="workbench-description">A compact view of coverage, review state, and the people responsible for moving this assessment forward.</p>
          </div>
          <span className="status-chip">{getProjectStatusLabel(project.status)}</span>
        </div>
        <div className="posture-project-line">
          <strong>{project.name}</strong>
          <span>{project.organizationName}</span>
        </div>
        <section className="workbench-summary-metrics" aria-label="Assessment workflow summary">
          <dl className="workbench-metrics">
            <div className="workbench-metric-accent">
              <dt>Overall coverage</dt>
              <dd>{boundedPercentage(summary.coveragePct)}%</dd>
              <small>Across included outcomes</small>
            </div>
            <div>
              <dt>Included</dt>
              <dd>{summary.includedCount}</dd>
              <small>In the submitted scope</small>
            </div>
            <div>
              <dt>Pending</dt>
              <dd>{summary.pendingCount}</dd>
              <small>Waiting for a response or review</small>
            </div>
            <div>
              <dt>Returned</dt>
              <dd>{summary.rejectedCount}</dd>
              <small>Need more information</small>
            </div>
          </dl>
        </section>
      </section>

      <section className="overview-workbench next-action-panel" aria-label="Next up">
        <div className="next-action-marker" aria-hidden="true">NEXT</div>
        <div className="next-action-copy">
          <p className="section-context">{attentionLabel}</p>
          <h2>{nextAction.label}</h2>
          <p>{nextAction.description}</p>
        </div>
        <button className="primary" type="button" aria-label="Open Assignment from Next up" onClick={onOpenAssignment}>{nextAction.actionLabel}</button>
      </section>

      <section className="overview-workbench function-register" aria-labelledby="function-register-heading">
        <div className="workbench-section-heading">
          <div>
            <p className="section-context">Coverage by Function</p>
            <h2 id="function-register-heading">Function register</h2>
          </div>
          <span className="register-count">{functions.length} function{functions.length === 1 ? "" : "s"}</span>
        </div>
        <div className="function-register-list">
          {functions.map((fn) => {
            const progress = functionProgress[fn.code];
            const summaryRow = summary.functions.find((item) => item.code === fn.code);
            const includedCount = progress?.includedCount ?? summaryRow?.includedCount ?? 0;
            const attention = progress?.attention ?? 0;
            const attentionText = attention > 0 ? `${attention} ${progress?.attentionLabel ?? "open"}` : "No open items";
            return (
              <div className="function-register-row" key={fn.code}>
                <div className="function-register-code">{fn.code}</div>
                <div className="function-register-copy">
                  <strong>{fn.name}</strong>
                  <small>{includedCount} included · {attentionText}</small>
                </div>
                <div className="function-register-progress" role="progressbar" aria-label={`${fn.name} coverage`} aria-valuemin={0} aria-valuemax={100} aria-valuenow={boundedPercentage(progress?.coveragePct ?? summaryRow?.coveragePct)}>
                  <div className="function-register-progress-label"><span>Coverage</span><strong>{boundedPercentage(progress?.coveragePct ?? summaryRow?.coveragePct)}%</strong></div>
                  <div className="function-register-progress-track" aria-hidden="true"><span style={{ width: `${boundedPercentage(progress?.coveragePct ?? summaryRow?.coveragePct)}%` }} /></div>
                </div>
                <div className="function-register-status"><span>Status</span><strong>{getProjectStatusLabel(project.status)}</strong></div>
              </div>
            );
          })}
        </div>
      </section>
    </>
  );
}
