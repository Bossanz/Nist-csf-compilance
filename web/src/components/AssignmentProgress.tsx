type Props = {
  included: number;
  assigned: number;
  unassigned: number;
};

export function AssignmentProgress({ included, assigned, unassigned }: Props) {
  return (
    <section className="assignment-progress" aria-label="Assignment progress">
      <strong className="assignment-progress-heading">Assignment progress</strong>
      <div className="assignment-progress-items">
        <span><b>{included}</b> Included in scope</span>
        <span><b>{assigned}</b> Assigned to stakeholder</span>
        <span><b>{unassigned}</b> Waiting for stakeholder assignment</span>
      </div>
    </section>
  );
}
