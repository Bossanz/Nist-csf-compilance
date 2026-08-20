"use client";

import type { Invitation, Project } from "../lib/types";

type Props = {
  invitations: Invitation[];
  projects: Project[];
  busyInvitationID: string;
  canManageLifecycle: boolean;
  onResend: (invitation: Invitation) => void | Promise<void>;
  onCancel: (invitation: Invitation) => void | Promise<void>;
};

export function InvitationList({ invitations, projects, busyInvitationID, canManageLifecycle, onResend, onCancel }: Props) {
  const projectNames = new Map(projects.map((project) => [project.id, project.name]));

  return (
    <section className="panel invitation-list" aria-labelledby="invitations-heading">
      <div className="section-heading">
        <div><span className="eyebrow">Access control</span><h3 id="invitations-heading">Invitations</h3></div>
        <span className="muted">{invitations.length} total</span>
      </div>
      {invitations.length === 0 ? <p className="muted">No invitations have been created.</p> : (
        <div className="invitation-rows">
          {invitations.map((invitation) => {
            const busy = busyInvitationID === invitation.id;
            const scopedProjects = (invitation.projectIDs ?? []).map((projectID) => projectNames.get(projectID) || projectID);
            const canResend = canManageLifecycle && (invitation.status === "pending" || invitation.status === "expired");
            const canCancel = canManageLifecycle && invitation.status === "pending";
            return (
              <article className="invitation-row" key={invitation.id}>
                <div className="invitation-main">
                  <strong>{invitation.email}</strong>
                  <span className="muted">{invitation.role.replaceAll("_", " ")}</span>
                  {invitation.role === "auditor" && <span className="muted">Projects: {scopedProjects.length ? scopedProjects.join(", ") : "not assigned"}</span>}
                </div>
                <div className="invitation-meta">
                  <span className={`status-chip invitation-status invitation-status-${invitation.status}`}>{invitation.status}</span>
                  {invitation.status === "pending" && <time dateTime={invitation.expiresAt}>Expires {formatDate(invitation.expiresAt)}</time>}
                </div>
                <div className="invitation-actions">
                  {canResend && <button className="secondary" type="button" aria-label={`Resend invitation for ${invitation.email}`} disabled={busy} onClick={() => void onResend(invitation)}>{busy ? "Working…" : "Resend"}</button>}
                  {canCancel && <button className="danger" type="button" aria-label={`Cancel invitation for ${invitation.email}`} disabled={busy} onClick={() => void onCancel(invitation)}>Cancel</button>}
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) return "—";
  return new Intl.DateTimeFormat("en-GB", { dateStyle: "medium" }).format(date);
}
