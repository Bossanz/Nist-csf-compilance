"use client";

import { useState } from "react";

type RemainingOutcome = { code: string; reason: string };

type Props = {
  status: string;
  includedCount: number;
  approvedCount: number;
  remaining: RemainingOutcome[];
  onFinalize: () => Promise<void>;
  onOpenReport: () => void;
  onOpenAudit: () => void;
};

export function ProjectFinalizationPanel({ status, includedCount, approvedCount, remaining, onFinalize, onOpenReport, onOpenAudit }: Props) {
  const [state, setState] = useState<"idle" | "saving" | "error">("idle");
  const [error, setError] = useState("");

  if (status !== "in_review" && status !== "closed") return null;

  async function finalize() {
    if (!window.confirm("Finalize this project? All assessment data will become read-only.")) return;
    setState("saving");
    setError("");
    try {
      await onFinalize();
      setState("idle");
    } catch (cause) {
      setState("error");
      setError(cause instanceof Error ? cause.message : "Could not finalize project");
    }
  }

  if (status === "closed") {
    return (
      <section className="project-finalization-panel panel finalized-panel" aria-label="Project finalization">
        <div><span className="eyebrow">Final state</span><h2>Project is finalized</h2><p>Assessment data is read-only. Use the reports for the audit handoff.</p></div>
        <div className="finalization-actions"><button className="secondary" type="button" onClick={onOpenReport}>Open Final Report</button><button className="primary" type="button" onClick={onOpenAudit}>Open Audit Package</button></div>
      </section>
    );
  }

  const ready = includedCount > 0 && approvedCount === includedCount && remaining.length === 0;
  const outstanding = includedCount - approvedCount;
  return (
    <section className="project-finalization-panel panel" aria-label="Project finalization">
      <div className="finalization-copy"><span className="eyebrow">Final gate</span><h2>Finalize project</h2><p>Reviewer approval must be complete before the assessment can be locked and handed to Audit.</p></div>
      <div className="finalization-readiness"><strong>{approvedCount} / {includedCount}</strong><span>included outcomes approved</span></div>
      {!ready && <div className="finalization-blocker" role="status"><strong>{outstanding === 1 ? "1 outcome remains" : `${Math.max(outstanding, remaining.length)} outcomes remain`}</strong><ul>{remaining.slice(0, 5).map((outcome) => <li key={outcome.code}><span>{outcome.code}</span> — {outcome.reason}</li>)}</ul></div>}
      {error && <p className="error" role="alert">{error}</p>}
      <button className="primary finalization-button" type="button" disabled={!ready || state === "saving"} onClick={() => void finalize()}>{state === "saving" ? "Finalizing…" : "Finalize Project"}</button>
    </section>
  );
}
