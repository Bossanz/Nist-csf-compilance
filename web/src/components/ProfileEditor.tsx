import type { ProfilePatch, ProfileRow } from "../lib/types";
import { AssessmentCard } from "./AssessmentCard";

export function ProfileEditor({
  rows,
  onSave,
}: {
  rows: ProfileRow[];
  onSave: (id: string, patch: ProfilePatch) => Promise<void>;
}) {
  if (rows.length === 0) {
    return <div className="empty-state">No outcomes found for this Function.</div>;
  }

  return (
    <div className="assessment-list">
      {rows.map((row) => <AssessmentCard key={row.id} row={row} onSave={onSave} />)}
    </div>
  );
}
