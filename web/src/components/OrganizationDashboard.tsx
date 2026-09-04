"use client";

import { useState } from "react";
import type { Organization, Role, User } from "../lib/types";
import { useDialogFocus } from "../lib/useDialogFocus";
import { getOrganizationPortfolioMetrics } from "../lib/workspaceMetrics";

type Props = {
  user: User;
  organizations: Organization[];
  loading: boolean;
  error: string;
  onSelect: (organization: Organization) => void;
  onCreate: (input: { name: string }) => void | Promise<void>;
  onDelete: (organization: Organization) => Promise<void>;
  onLogout: () => void;
  counselors?: User[];
  counselorInvitationURL?: string;
  onInviteCounselor?: (input: { email: string; role: "counselor" | "counselor_admin" }) => void | Promise<void>;
  onUpdateCounselor?: (userID: string, input: { role: "counselor" | "counselor_admin"; status: "active" | "disabled" }) => void | Promise<void>;
};

type CounselorRole = Extract<Role, "counselor" | "counselor_admin">;
type CounselorAccess = { role: CounselorRole; status: "active" | "disabled" };

export function OrganizationDashboard({ user, organizations, loading, error, onSelect, onCreate, onDelete, onLogout, counselors = [], counselorInvitationURL, onInviteCounselor, onUpdateCounselor }: Props) {
  const [name, setName] = useState("");
  const [creating, setCreating] = useState(false);
  const [counselorEmail, setCounselorEmail] = useState("");
  const [counselorRole, setCounselorRole] = useState<CounselorRole>("counselor");
  const [counselorAccess, setCounselorAccess] = useState<Record<string, CounselorAccess>>({});
  const [counselorSaving, setCounselorSaving] = useState("");
  const [counselorError, setCounselorError] = useState("");
  const [pendingDelete, setPendingDelete] = useState<Organization | null>(null);
  const [confirmation, setConfirmation] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const isAdmin = user.role === "counselor_admin";

  function closeDeleteConfirmation() {
    setPendingDelete(null);
    setConfirmation("");
    setDeleteError("");
  }

  const deleteDialogFocus = useDialogFocus(Boolean(pendingDelete), closeDeleteConfirmation);
  const portfolioMetrics = getOrganizationPortfolioMetrics(organizations);

  async function confirmDelete() {
    if (!pendingDelete) return;
    setDeleting(true);
    setDeleteError("");
    try {
      await onDelete(pendingDelete);
      closeDeleteConfirmation();
    } catch (cause) {
      setDeleteError(cause instanceof Error ? cause.message : "Could not delete organization");
    } finally {
      setDeleting(false);
    }
  }

  function getCounselorAccess(item: User): CounselorAccess {
    const draft = counselorAccess[item.id];
    return draft || { role: item.role as CounselorRole, status: item.status === "disabled" ? "disabled" : "active" };
  }

  function setCounselorField(item: User, field: keyof CounselorAccess, value: string) {
    const current = getCounselorAccess(item);
    setCounselorAccess((rows) => ({ ...rows, [item.id]: { ...current, [field]: value } as CounselorAccess }));
  }

  async function saveCounselor(item: User) {
    if (!onUpdateCounselor) return;
    const access = getCounselorAccess(item);
    setCounselorSaving(item.id);
    setCounselorError("");
    try {
      await onUpdateCounselor(item.id, access);
    } catch (cause) {
      setCounselorError(cause instanceof Error ? cause.message : "Could not update counselor access");
    } finally {
      setCounselorSaving("");
    }
  }

  async function createOrganization() {
    setCreating(true);
    try {
      await onCreate({ name: name.trim() });
      setName("");
    } catch {
      // The page-level handler renders the API error; keep the form values for retry.
    } finally {
      setCreating(false);
    }
  }

  return (
    <main className="main dashboard">
      <header className="product-header">
        <div>
          <div className="context-line"><span>NIST CSF 2.0</span><span aria-hidden="true">/</span><span>Organizations</span></div>
          <h1>Client organizations</h1>
          <p className="muted">Choose a client workspace before creating or continuing a project.</p>
        </div>
        <div className="identity-block">
          <span className="role-chip">{user.role.replaceAll("_", " ")}</span>
          <strong>{user.name}</strong>
          <a className="auth-link" href="/account/password">Change password</a>
          <button className="text-button" onClick={onLogout}>Sign out</button>
        </div>
      </header>

      {error && <div className="error" role="alert">{error}</div>}

      <section className="portfolio-summary" aria-label="Portfolio posture">
        <div className="portfolio-summary-heading">
          <div>
            <p className="section-context">Portfolio posture</p>
            <h2 id="portfolio-posture-heading">Client workspaces</h2>
          </div>
          <p className="portfolio-summary-hint">A focused starting point for every client assessment. Open a workspace to see its projects, owners, evidence, and review state.</p>
        </div>
        <dl className="portfolio-metrics">
          <div>
            <dt>Client workspaces</dt>
            <dd>{portfolioMetrics.total}</dd>
            <small>Available in this portfolio</small>
          </div>
          <div>
            <dt>Ready to open</dt>
            <dd>{portfolioMetrics.active}</dd>
            <small>Open a workspace to continue</small>
          </div>
          <div>
            <dt>Project status</dt>
            <dd className="portfolio-metric-text">Inside workspace</dd>
            <small>Review status lives with the client</small>
          </div>
          <div>
            <dt>Next action</dt>
            <dd className="portfolio-metric-text">Open a workspace</dd>
            <small>Start with the client register below</small>
          </div>
        </dl>
      </section>

      <section aria-labelledby="organizations-heading">
        <div className="section-heading">
          <h2 id="organizations-heading">Organizations</h2>
          <p className="muted">Client workspaces available to your account.</p>
        </div>
        <section className="organization-register" aria-labelledby="client-workspace-register-heading">
          <div className="register-heading">
            <div>
              <h3 id="client-workspace-register-heading">Client workspace register</h3>
              <p className="muted">Choose a workspace to continue the assessment.</p>
            </div>
            <span className="register-count">{organizations.length} workspace{organizations.length === 1 ? "" : "s"}</span>
          </div>
          {loading ? <div className="panel muted" role="status" aria-live="polite" aria-busy="true">Loading organizations…</div> : organizations.length === 0 ? (
            <div className="empty-state">No client organizations yet. Create one below to begin.</div>
          ) : (
            <div className="organization-register-list">
              {organizations.map((organization, index) => (
                <article className="organization-register-row" key={organization.id}>
                  <div className="register-index" aria-hidden="true">{String(index + 1).padStart(2, "0")}</div>
                  <div className="register-primary">
                    <div className="register-title-line"><h4>{organization.name}</h4><span className="status-chip">Available</span></div>
                    <p>{organization.type === "client" ? "Client workspace" : "Counselor workspace"}</p>
                    <small>Open to review projects and workspace activity</small>
                  </div>
                  <div className="register-meta"><span>Access</span><strong>Available to you</strong></div>
                  <div className="organization-actions register-actions">
                    <button className="secondary" aria-label={`Open ${organization.name}`} onClick={() => onSelect(organization)}>Open client workspace</button>
                    {isAdmin && <button className="danger" aria-label={`Delete ${organization.name}`} onClick={(event) => { deleteDialogFocus.triggerRef.current = event.currentTarget; setPendingDelete(organization); setConfirmation(""); setDeleteError(""); }}>Delete</button>}
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </section>

      {pendingDelete && (
        <section ref={deleteDialogFocus.dialogRef} className="delete-confirmation" role="dialog" aria-modal="true" aria-labelledby="delete-organization-title">
          <div>
            <h2 id="delete-organization-title">Delete {pendingDelete.name}?</h2>
            <p>The organization, its projects, assessment data, stakeholder access, and invitations will be permanently deleted.</p>
          </div>
          <label className="field">
            <span>Type {pendingDelete.name} to confirm</span>
            <input autoFocus value={confirmation} onChange={(event) => setConfirmation(event.target.value)} />
          </label>
          {deleteError && <div className="error" role="alert">{deleteError}</div>}
          <div className="confirmation-actions">
            <button className="secondary" onClick={closeDeleteConfirmation}>Cancel</button>
            <button className="danger" disabled={confirmation !== pendingDelete.name || loading || deleting} onClick={() => void confirmDelete()}>{deleting ? "Deleting…" : "Delete permanently"}</button>
          </div>
        </section>
      )}

      {isAdmin && (
        <>
          <section className="panel create-organization" aria-labelledby="create-organization">
            <div><h2 id="create-organization">Create organization</h2><p className="muted">Projects and stakeholders will live inside this workspace.</p></div>
            <form className="form" onSubmit={(event) => { event.preventDefault(); void createOrganization(); }}>
              <label className="field"><span>Organization name</span><input required value={name} onChange={(event) => setName(event.target.value)} /></label>
              <button className="primary" type="submit" disabled={loading || creating}>{creating ? "Creating…" : "Create organization"}</button>
            </form>
          </section>

          <section className="people-section" aria-labelledby="counselors-heading">
            <div className="section-heading">
              <h2 id="counselors-heading">Counselors</h2>
              <p className="muted">Manage the consulting accounts that can access every client organization.</p>
            </div>
            <div className="people-list">
              {counselors.length === 0 ? <div className="empty-state">No counselor accounts yet. Create an invitation below to add one.</div> : counselors.map((item) => {
                const access = getCounselorAccess(item);
                return (
                  <div className="person-row account-row" key={item.id}>
                    <div><strong>{item.name}</strong><span>{item.email}</span></div>
                    <div className="account-controls">
                      <label className="field account-control"><span>Role for {item.email}</span><select aria-label={`Role for ${item.email}`} value={access.role} onChange={(event) => setCounselorField(item, "role", event.target.value)}><option value="counselor">counselor</option><option value="counselor_admin">counselor admin</option></select></label>
                      <label className="field account-control"><span>Status for {item.email}</span><select aria-label={`Status for ${item.email}`} value={access.status} onChange={(event) => setCounselorField(item, "status", event.target.value)}><option value="active">active</option><option value="disabled" disabled={item.id === user.id}>disabled</option></select></label>
                      <button className="secondary account-save" type="button" aria-label={`Save access for ${item.email}`} disabled={counselorSaving === item.id} onClick={() => void saveCounselor(item)}>{counselorSaving === item.id ? "Saving…" : "Save counselor access"}</button>
                    </div>
                  </div>
                );
              })}
            </div>
            {counselorError && <div className="error account-error" role="alert">{counselorError}</div>}
            {onInviteCounselor && (
              <form className="panel invite-form counselor-invite-form" onSubmit={(event) => { event.preventDefault(); void onInviteCounselor({ email: counselorEmail.trim().toLowerCase(), role: counselorRole }); }}>
                <label className="field"><span>Counselor email</span><input type="email" required value={counselorEmail} onChange={(event) => setCounselorEmail(event.target.value)} /></label>
                <label className="field"><span>Counselor role</span><select value={counselorRole} onChange={(event) => setCounselorRole(event.target.value as CounselorRole)}><option value="counselor">counselor</option><option value="counselor_admin">counselor admin</option></select></label>
                <button className="primary" type="submit">Create counselor invitation</button>
                {counselorInvitationURL && <label className="field invitation-result"><span>One-time counselor invitation link</span><input readOnly value={counselorInvitationURL} /></label>}
              </form>
            )}
          </section>
        </>
      )}
    </main>
  );
}
