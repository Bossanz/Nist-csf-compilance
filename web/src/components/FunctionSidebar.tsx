import type { FunctionNode } from "../lib/types";

export type FunctionProgress = {
  value: number;
  label: string;
  attention?: number;
  attentionLabel?: string;
};

type Props = {
  functions: FunctionNode[];
  selectedCode: string;
  onSelect: (code: string) => void;
  progressByFunction?: Record<string, FunctionProgress>;
};

export function FunctionSidebar({ functions, selectedCode, onSelect, progressByFunction }: Props) {
  return (
    <nav className="sidebar" aria-label="CSF functions">
      <div className="brand">CSF / Workspace</div>
      <div>
        {functions.map((fn) => (
          <button className={"nav-item " + (selectedCode === fn.code ? "active" : "")} type="button" aria-current={selectedCode === fn.code ? "page" : undefined} key={fn.code} onClick={() => onSelect(fn.code)}>
            <span className="nav-label">{fn.code} <span className="muted">{fn.name}</span></span>
            {progressByFunction?.[fn.code] && (
              <span
                className="nav-meta"
              >
                <strong>{progressByFunction[fn.code]!.value}</strong>
                <small>{progressByFunction[fn.code]!.label}</small>
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
