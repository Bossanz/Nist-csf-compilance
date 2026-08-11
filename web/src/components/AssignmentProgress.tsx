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
        <span><b>{included}</b> Included</span>
        <span><b>{assigned}</b> Assigned</span>
        <span><b>{unassigned}</b> Waiting for assignment</span>
      </div>
    </section>
  );
}
