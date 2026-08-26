import type { RemediationOverviewStatus, RemediationStatusSummary } from "../lib/remediationStatus";
import { remediationOverviewStatusLabel } from "../lib/remediationStatus";

type Props = {
  assessmentStatus: string;
  summary: RemediationStatusSummary;
  loading?: boolean;
  error?: string;
  onOpenActionPlan: () => void;
};

function coverageGapLabel(count: number) {
  return `${count} coverage ${count === 1 ? "gap" : "gaps"}`;
}

function actionLabel(count: number) {
  return `${count} action ${count === 1 ? "item" : "items"}`;
}

function statusDescription(status: RemediationOverviewStatus, assessmentStatus: string, summary: RemediationStatusSummary) {
  if (status === "not_required") return "No current-to-target coverage gaps are recorded.";
  if (status === "not_started") {
    return assessmentStatus === "closed"
      ? "Assessment finalized. Coverage gaps remain without an Action Plan."
      : "No remediation actions have been opened yet. This work is separate from the assessment Finalize gate.";
  }
  if (status === "complete") return "Assessment and remediation are tracked separately. All recorded actions are closed.";
  return `${assessmentStatus === "closed" ? "Assessment finalized." : "Assessment is still in review."} ${summary.openActionCount} ${summary.openActionCount === 1 ? "action remains" : "actions remain"} open in the Action Plan.`;
}

export function RemediationStatusPanel({ assessmentStatus, summary, loading = false, error = "", onOpenActionPlan }: Props) {
  return (
    <section className="remediation-status-panel" aria-labelledby="remediation-status-heading">
      <div className="remediation-status-copy">
        <span className="eyebrow">Remediation</span>
        <h2 id="remediation-status-heading">Remediation status</h2>
        <p>{statusDescription(summary.status, assessmentStatus, summary)}</p>
        {loading && <small className="remediation-status-loading" role="status">Refreshing Action Plan details…</small>}
        {error && <small className="error remediation-status-error" role="alert">Action Plan details could not be loaded yet.</small>}
      </div>
      <div className="remediation-status-content">
        <div className={`remediation-state remediation-state-${summary.status}`}>
          <span>Current status</span>
          <strong>{remediationOverviewStatusLabel(summary.status)}</strong>
        </div>
        <dl className="remediation-status-stats">
          <div><dt>Coverage gaps</dt><dd>{summary.gapCount}</dd></div>
          <div><dt>Action items</dt><dd>{summary.actionCount}</dd></div>
          <div><dt>Open actions</dt><dd>{summary.openActionCount}</dd></div>
          <div><dt>Closed actions</dt><dd>{summary.closedActionCount}</dd></div>
        </dl>
        <button className="secondary remediation-status-action" type="button" onClick={onOpenActionPlan}>
          Open Action Plan
        </button>
      </div>
      <span className="sr-only">{coverageGapLabel(summary.gapCount)} and {actionLabel(summary.actionCount)}.</span>
    </section>
  );
}
