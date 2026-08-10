import type { Summary } from "../lib/types";
export function SummaryCards({ summary }: { summary: Summary }) {
  return (
    <section className="cards" aria-label="Assessment summary">
      <div className="card"><span className="muted">Overall coverage</span><div className="metric">{summary.coveragePct.toFixed(0)}%</div></div>
      <div className="card"><span className="muted">Included</span><div className="metric">{summary.includedCount}</div></div>
      <div className="card"><span className="muted">Pending</span><div className="metric">{summary.pendingCount}</div></div>
      <div className="card"><span className="muted">Returned</span><div className="metric">{summary.rejectedCount}</div></div>
    </section>
  );
}
