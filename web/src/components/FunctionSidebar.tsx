import type { FunctionNode } from "../lib/types";

export function FunctionSidebar({ functions, selectedCode, onSelect }: { functions: FunctionNode[]; selectedCode: string; onSelect: (code: string) => void }) {
  return (
    <nav className="sidebar" aria-label="CSF functions">
      <div className="brand">CSF / Workspace</div>
      <div>{functions.map((fn) => <button className={`nav-item ${selectedCode === fn.code ? "active" : ""}`} key={fn.code} onClick={() => onSelect(fn.code)}>{fn.code} <span className="muted">{fn.name}</span></button>)}</div>
    </nav>
  );
}
