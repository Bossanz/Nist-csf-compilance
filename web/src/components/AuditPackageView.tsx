"use client";

import type { AuditPackageData } from "../lib/types";

type Props = {
  auditPackage: AuditPackageData;
  onBack: () => void;
  onDownloadCSV: () => void;
};

function formatDate(value?: string | null) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function statusLabel(status?: string) {
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

export function AuditPackageView({ auditPackage, onBack, onDownloadCSV }: Props) {
  const { project, summary } = auditPackage;
  return (
    <main className="report-page audit-page">
      <div className="report-toolbar no-print">
        <button className="text-button" type="button" onClick={onBack}>Back to project</button>
        <div className="report-toolbar-actions">
          <button className="secondary" type="button" onClick={onDownloadCSV}>Download CSV</button>
          <button className="primary" type="button" onClick={() => window.print()}>Print / Save as PDF</button>
        </div>
      </div>
      <header className="report-header panel">
        <div><span className="eyebrow">{project.name}</span><h1>Audit package</h1><p>{project.organizationName} · Traceability register for NIST CSF 2.0 assessment</p></div>
        <div className="report-final-status"><span className="status-chip">{project.status === "closed" ? "Finalized" : "In progress"}</span><small>{summary.includedCount} included outcomes</small></div>
      </header>
      <section className="report-summary-grid" aria-label="Audit package summary">
        <div className="report-metric report-metric-accent"><span>Coverage</span><strong>{Math.round(summary.coveragePct)}%</strong></div>
        <div className="report-metric"><span>Included</span><strong>{summary.includedCount}</strong></div>
        <div className="report-metric"><span>Approved</span><strong>{summary.approvedCount}</strong></div>
        <div className="report-metric"><span>Audit events</span><strong>{auditPackage.auditTrail.length}</strong></div>
      </section>
      <section className="report-section">
        <div className="report-section-heading"><div><span className="eyebrow">01</span><h2>Scope &amp; assignment</h2></div><span className="muted">{auditPackage.scope.length} catalog outcomes</span></div>
        <div className="report-table-wrap"><table className="report-table"><thead><tr><th>Outcome</th><th>Included</th><th>Rationale</th><th>Assigned assessor</th></tr></thead><tbody>
          {auditPackage.scope.map(({ profile }) => <tr key={profile.subcategoryID}><th scope="row">{profile.subcategoryCode}<small>{profile.description}</small></th><td>{profile.included ? "Yes" : "No"}</td><td>{profile.rationale || "—"}</td><td>{profile.assignedUserName || profile.assignedUserEmail || "—"}</td></tr>)}
        </tbody></table></div>
      </section>
      <section className="report-section">
        <div className="report-section-heading"><div><span className="eyebrow">02</span><h2>Evidence register</h2></div><span className="muted">{summary.evidenceCount} files</span></div>
        <div className="report-table-wrap"><table className="report-table"><thead><tr><th>Outcome</th><th>File</th><th>Type</th><th>Uploaded</th><th>Size</th></tr></thead><tbody>
          {auditPackage.outcomes.flatMap((outcome) => outcome.evidence.map((evidence) => <tr key={evidence.id}><th scope="row">{outcome.profile.subcategoryCode}</th><td>{evidence.originalName}</td><td>{evidence.mimeType}</td><td>{formatDate(evidence.createdAt)}</td><td>{Math.ceil(evidence.sizeBytes / 1024)} KB</td></tr>))}
          {summary.evidenceCount === 0 && <tr><td colSpan={5}>No evidence files were attached.</td></tr>}
        </tbody></table></div>
      </section>
      <section className="report-section">
        <div className="report-section-heading"><div><span className="eyebrow">03</span><h2>Review history</h2></div><span className="muted">Response decisions</span></div>
        <div className="report-table-wrap"><table className="report-table"><thead><tr><th>Outcome</th><th>Status</th><th>Reviewer comment</th><th>Reviewed at</th></tr></thead><tbody>
          {auditPackage.outcomes.map((outcome) => <tr key={outcome.profile.subcategoryID}><th scope="row">{outcome.profile.subcategoryCode}</th><td>{statusLabel(outcome.response?.status)}</td><td>{outcome.response?.reviewComment || "—"}</td><td>{formatDate(outcome.response?.reviewedAt)}</td></tr>)}
        </tbody></table></div>
      </section>
      <section className="report-section">
        <div className="report-section-heading"><div><span className="eyebrow">05</span><h2>Audit trail</h2></div><span className="muted">Chronological activity</span></div>
        <ol className="audit-timeline">
          {auditPackage.auditTrail.map((event) => <li key={event.id}><div><strong>{event.action}</strong><span>{event.entityType}{event.entityID ? ` · ${event.entityID}` : ""}</span></div><div><span>{event.actorName || event.actorEmail || "System"}</span><time dateTime={event.createdAt}>{formatDate(event.createdAt)}</time></div></li>)}
          {auditPackage.auditTrail.length === 0 && <li className="muted">No audit events have been recorded.</li>}
        </ol>
      </section>
      <section className="report-section">
        <div className="report-section-heading"><div><span className="eyebrow">04</span><h2>Remediation register</h2></div><span className="muted">{auditPackage.remediationActions.length} actions · {auditPackage.remediationSummary.overdueCount} overdue</span></div>
        <div className="report-table-wrap"><table className="report-table"><thead><tr><th>Outcome / action</th><th>Owner</th><th>Priority</th><th>Due</th><th>Status</th><th>Evidence</th></tr></thead><tbody>
          {auditPackage.remediationActions.map((action) => <tr key={action.id}><th scope="row">{action.outcomeCode}<small>{action.title}</small></th><td>{action.ownerName}</td><td>{action.priority}</td><td>{formatDate(action.dueDate)}</td><td>{remediationStatusLabel(action.status)}</td><td>{action.evidence.map((evidence) => <span key={evidence.id} className="report-evidence-name">{evidence.originalName}</span>)}</td></tr>)}
          {auditPackage.remediationActions.length === 0 && <tr><td colSpan={6}>No remediation actions recorded.</td></tr>}
        </tbody></table></div>
      </section>
      <footer className="report-footer"><span>Traceability chain: Scope → Assignment → Response → Evidence → Review → Finalization → Remediation</span><span>Project {project.id}</span></footer>
    </main>
  );
}
