"use client";

import type { FinalReportData, ReportOutcome } from "../lib/types";

type Props = {
  report: FinalReportData;
  onBack: () => void;
  onOpenAudit: () => void;
};

function formatDate(value?: string | null) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function percent(value: number) {
  return `${Math.round(Math.max(0, Math.min(100, value)))}%`;
}

function statusLabel(status: string | undefined) {
  if (status === "reviewed") return "Approved";
  if (status === "submitted") return "Reviewing";
  if (status === "needs_more_info") return "Returned";
  if (status === "draft") return "Pending";
  return status || "Pending";
}

function remediationStatusLabel(status: string) {
  if (status === "in_progress") return "In progress";
  if (status === "awaiting_review") return "Awaiting review";
  return status.charAt(0).toUpperCase() + status.slice(1);
}

function OutcomeResult({ outcome }: { outcome: ReportOutcome }) {
  const profile = outcome.profile;
  return (
    <article className="report-outcome">
      <header className="report-outcome-heading">
        <div>
          <span className="eyebrow">{profile.subcategoryCode}</span>
          <h3>{profile.description}</h3>
          <p>{profile.functionCode} / {profile.categoryCode}</p>
        </div>
        <span className="status-chip">{statusLabel(outcome.response?.status)}</span>
      </header>
      <div className="report-outcome-grid">
        <section>
          <h4>Current profile</h4>
          <dl className="report-facts">
            <div><dt>Priority</dt><dd>{profile.currentPriority || "—"}</dd></div>
            <div><dt>Coverage</dt><dd>{profile.currentCoverageLevel}</dd></div>
            <div><dt>Activities</dt><dd>{profile.currentStatusText || "—"}</dd></div>
            <div><dt>Policies</dt><dd>{profile.currentPoliciesText || "—"}</dd></div>
          </dl>
        </section>
        <section>
          <h4>Target profile</h4>
          <dl className="report-facts">
            <div><dt>Priority</dt><dd>{profile.targetPriority || "—"}</dd></div>
            <div><dt>Coverage</dt><dd>{profile.targetCoverageLevel}</dd></div>
            <div><dt>Approach</dt><dd>{profile.targetApproachText || "—"}</dd></div>
          </dl>
        </section>
      </div>
      <section className="report-response-block">
        <h4>Stakeholder response</h4>
        <p>{outcome.response?.responseText || "No response recorded."}</p>
        {outcome.response?.reviewComment && <p className="report-review-comment"><strong>Reviewer comment:</strong> {outcome.response.reviewComment}</p>}
      </section>
      <section className="report-evidence-block">
        <h4>Evidence ({outcome.evidence.length})</h4>
        {outcome.evidence.length > 0 ? (
          <ul>
            {outcome.evidence.map((evidence) => <li key={evidence.id}><span>{evidence.originalName}</span><small>{evidence.mimeType} · {Math.ceil(evidence.sizeBytes / 1024)} KB</small></li>)}
          </ul>
        ) : <p className="muted">No evidence attached.</p>}
      </section>
    </article>
  );
}

export function FinalReport({ report, onBack, onOpenAudit }: Props) {
  const { project, summary } = report;
  const assessmentVersion = project.versionNumber ?? 1;
  const previousVersion = assessmentVersion > 1 ? assessmentVersion - 1 : null;
  return (
    <main className="report-page">
      <div className="report-toolbar no-print">
        <button className="text-button" type="button" onClick={onBack}>Back to project</button>
        <div className="report-toolbar-actions">
          <button className="secondary" type="button" onClick={onOpenAudit}>Open Audit Package</button>
          <button className="primary" type="button" onClick={() => window.print()}>Print / Save as PDF</button>
        </div>
      </div>
      <header className="report-header panel">
        <div>
          <span className="eyebrow">Final assessment report · Assessment v{assessmentVersion}</span>
          <h1>{project.name}</h1>
          <p>{project.organizationName} · {project.assessmentPeriod || "Assessment period not specified"}</p>
        </div>
        <div className="report-final-status">
          <span className="status-chip">{project.status === "closed" ? "Finalized" : "Draft report"}</span>
          <small>Finalized {formatDate(project.finalizedAt)}</small>
        </div>
      </header>
      <section className="report-meta panel" aria-label="Project metadata">
        <div><span>Objective</span><strong>{project.objective || "—"}</strong></div>
        <div><span>Scope boundary</span><strong>{project.scopeBoundary || "—"}</strong></div>
        <div><span>Compliance driver</span><strong>{project.complianceDriver || "—"}</strong></div>
        <div><span>Assessment version</span><strong>v{assessmentVersion}</strong></div>
        {previousVersion && <div><span>Previous version</span><strong>v{previousVersion}</strong></div>}
        <div><span>Finalized by</span><strong>{project.finalizedBy || "—"}</strong></div>
      </section>
      <section className="report-summary-grid" aria-label="Final assessment summary">
        <div className="report-metric report-metric-accent"><span>Overall coverage</span><strong>{percent(summary.coveragePct)}</strong></div>
        <div className="report-metric"><span>Included outcomes</span><strong>{summary.includedCount}</strong></div>
        <div className="report-metric"><span>Approved</span><strong>{summary.approvedCount}</strong></div>
        <div className="report-metric"><span>Evidence files</span><strong>{summary.evidenceCount}</strong></div>
      </section>
      <section className="report-section" aria-labelledby="coverage-by-function-title">
        <div className="report-section-heading"><div><span className="eyebrow">01</span><h2 id="coverage-by-function-title">Coverage by Function</h2></div><span className="muted">Current profile coverage</span></div>
        <div className="report-table-wrap" role="region" tabIndex={0} aria-label="Coverage by Function report table">
          <table className="report-table"><thead><tr><th>Function</th><th>Coverage</th><th>Included</th><th>Approved</th><th>Reviewing</th><th>Returned</th></tr></thead><tbody>
            {summary.functions.map((item) => <tr key={item.code}><th scope="row">{item.code}</th><td><strong>{percent(item.coveragePct)}</strong></td><td>{item.includedCount}</td><td>{item.approvedCount}</td><td>{item.reviewingCount}</td><td>{item.returnedCount}</td></tr>)}
          </tbody></table>
          <span className="report-table-hint" aria-hidden="true">Scroll horizontally to view all columns</span>
        </div>
      </section>
      <section className="report-section" aria-labelledby="included-outcomes-title">
        <div className="report-section-heading"><div><span className="eyebrow">02</span><h2 id="included-outcomes-title">Included outcomes</h2></div><span className="muted">{report.outcomes.length} outcome{report.outcomes.length === 1 ? "" : "s"}</span></div>
        <div className="report-outcome-list">
          {report.outcomes.length > 0 ? report.outcomes.map((outcome) => <OutcomeResult key={outcome.profile.subcategoryID} outcome={outcome} />) : <div className="empty-state">No included outcomes are available in this report.</div>}
        </div>
      </section>
      <section className="report-section" aria-labelledby="remediation-progress-title">
        <div className="report-section-heading"><div><h2 id="remediation-progress-title">Remediation progress</h2></div><span className="muted">Live action status; finalized assessment results remain unchanged</span></div>
        <div className="report-summary-grid remediation-report-summary" aria-label="Remediation summary">
          <div className="report-metric"><span>Open</span><strong>{report.remediationSummary.openCount}</strong></div>
          <div className="report-metric"><span>In progress</span><strong>{report.remediationSummary.inProgressCount}</strong></div>
          <div className="report-metric"><span>Awaiting review</span><strong>{report.remediationSummary.awaitingReviewCount}</strong></div>
          <div className="report-metric"><span>Overdue</span><strong>{report.remediationSummary.overdueCount}</strong></div>
          <div className="report-metric"><span>Closed</span><strong>{report.remediationSummary.closedCount}</strong></div>
        </div>
        <div className="report-table-wrap" role="region" tabIndex={0} aria-label="Remediation progress report table"><table className="report-table"><thead><tr><th>Outcome / action</th><th>Owner</th><th>Priority</th><th>Due</th><th>Status</th><th>Closed</th></tr></thead><tbody>
          {report.remediationActions.map((action) => <tr key={action.id}><th scope="row">{action.outcomeCode}<small>{action.title}</small></th><td>{action.ownerName}</td><td>{action.priority}</td><td>{formatDate(action.dueDate)}</td><td>{remediationStatusLabel(action.status)}</td><td>{formatDate(action.closedAt)}</td></tr>)}
          {report.remediationActions.length === 0 && <tr><td colSpan={6}>No remediation actions recorded.</td></tr>}
        </tbody></table><span className="report-table-hint" aria-hidden="true">Scroll horizontally to view all columns</span></div>
      </section>
      <footer className="report-footer"><span>Generated from the finalized assessment workspace.</span><span>{formatDate(new Date().toISOString())}</span></footer>
    </main>
  );
}
