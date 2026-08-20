import type { AuditTrailEntry } from "../lib/types";

type Props = { events: AuditTrailEntry[] };

const actionLabels: Record<string, string> = {
  "audit_logs.viewed": "Audit activity viewed",
  "audit_package.downloaded": "Audit package downloaded",
  "audit_package.viewed": "Audit package viewed",
  "evidence.deleted": "Evidence deleted",
  "evidence.downloaded": "Evidence downloaded",
  "evidence.uploaded": "Evidence uploaded",
  "organization.created": "Organization created",
  "profile.updated": "Assessment profile updated",
  "project.created": "Project created",
  "project.finalized": "Project finalized",
  "project.scope_submitted": "Project scope submitted",
  "report.viewed": "Final report viewed",
  "response.reviewed": "Response approved",
  "response.submitted": "Response sent for review",
  "response.updated": "Response updated",
  "user.invitation_cancelled": "Invitation cancelled",
  "user.invitation_resent": "Invitation resent",
  "user.invited": "Invitation created",
  "user.role_changed": "User access changed",
};

export function AuditTimeline({ events }: Props) {
  return (
    <section className="panel audit-activity" aria-labelledby="activity-trail-heading">
      <div className="section-heading">
        <div><span className="eyebrow">Transparency</span><h2 id="activity-trail-heading">Activity trail</h2></div>
        <span className="muted">{events.length} {events.length === 1 ? "event" : "events"}</span>
      </div>
      {events.length === 0 ? <p className="muted">No audit activity has been recorded yet.</p> : (
        <ol className="audit-timeline workspace-audit-timeline">
          {events.map((event) => (
            <li key={event.id}>
              <div>
                <strong>{actionLabels[event.action] ?? event.action}</strong>
                <span>{event.entityType}{event.entityID ? " · " + event.entityID : ""}</span>
                <span>{event.actorName || event.actorEmail || "System"}{event.actorRole ? " · " + event.actorRole.replaceAll("_", " ") : ""} · {event.result || "success"}</span>
                {event.requestID && <span>Request {event.requestID}</span>}
              </div>
              <div>
                <time dateTime={event.createdAt}>{formatDate(event.createdAt)}</time>
                {event.ipAddress && <span>{event.ipAddress}</span>}
              </div>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.valueOf()) ? value : new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short" }).format(date);
}
