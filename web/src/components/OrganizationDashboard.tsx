"use client";

import { useState } from "react";
import type { Organization, User } from "../lib/types";

type Props = {
  user: User;
  organizations: Organization[];
  loading: boolean;
  error: string;
  onSelect: (organization: Organization) => void;
  onCreate: (input: { name: string }) => void;
  onDelete: (organization: Organization) => void;
  onLogout: () => void;
};

export function OrganizationDashboard({ user, organizations, loading, error, onSelect, onCreate, onDelete, onLogout }: Props) {
  const [name, setName] = useState("");
  const [pendingDelete, setPendingDelete] = useState<Organization | null>(null);
  const [confirmation, setConfirmation] = useState("");
  const isAdmin = user.role === "counselor_admin";

  function closeDeleteConfirmation() {
    setPendingDelete(null);
    setConfirmation("");
  }

  return (
    <main className="main dashboard">
      <header className="product-header">
        <div>
          <span className="eyebrow">NIST CSF 2.0 / ORGANIZATIONS</span>
          <h1>Client organizations</h1>
          <p className="muted">Choose a client workspace before creating or continuing a project.</p>
        </div>
        <div className="identity-block">
          <span className="role-chip">{user.role.replaceAll("_", " ")}</span>
          <strong>{user.name}</strong>
          <button className="text-button" onClick={onLogout}>Sign out</button>
        </div>
      </header>

      {error && <div className="error" role="alert">{error}</div>}

      <section aria-labelledby="organizations-heading">
        <div className="section-heading">
          <span className="section-index">CLIENT INDEX</span>
          <h2 id="organizations-heading">Organizations</h2>
        </div>
        {loading ? <div className="panel muted">Loading organizations…</div> : organizations.length === 0 ? (
          <div className="empty-state">No organizations yet.</div>
        ) : (
          <div className="organization-grid">
            {organizations.map((organization, index) => (
              <article className="organization-card" key={organization.id}>
                <div className="organization-number">{String(index + 1).padStart(2, "0")}</div>
                <div><h3>{organization.name}</h3><p className="muted">Client organization</p></div>
                <div className="organization-actions">
                  <button className="secondary" aria-label={`Open ${organization.name}`} onClick={() => onSelect(organization)}>Open workspace</button>
                  {isAdmin && <button className="danger" aria-label={`Delete ${organization.name}`} onClick={() => { setPendingDelete(organization); setConfirmation(""); }}>Delete</button>}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      {pendingDelete && (
        <section className="delete-confirmation" role="dialog" aria-labelledby="delete-organization-title">
          <div>
            <span className="section-index">PERMANENT DELETION</span>
            <h2 id="delete-organization-title">Delete {pendingDelete.name}?</h2>
            <p>This removes every project, assessment, stakeholder, and invitation owned by this organization. It cannot be undone.</p>
          </div>
          <label className="field">
            <span>Type {pendingDelete.name} to confirm</span>
            <input autoFocus value={confirmation} onChange={(event) => setConfirmation(event.target.value)} />
          </label>
          <div className="organization-actions">
            <button className="secondary" onClick={closeDeleteConfirmation}>Cancel</button>
            <button className="danger" disabled={confirmation !== pendingDelete.name || loading} onClick={() => { onDelete(pendingDelete); closeDeleteConfirmation(); }}>Delete permanently</button>
          </div>
        </section>
      )}

      {isAdmin && (
        <section className="panel create-organization" aria-labelledby="create-organization">
          <div><span className="section-index">NEW CLIENT</span><h2 id="create-organization">Create organization</h2><p className="muted">Projects and stakeholders will live inside this workspace.</p></div>
          <form className="form" onSubmit={(event) => { event.preventDefault(); onCreate({ name: name.trim() }); setName(""); }}>
            <label className="field"><span>Organization name</span><input required value={name} onChange={(event) => setName(event.target.value)} /></label>
            <button className="primary">Create organization</button>
          </form>
        </section>
      )}
    </main>
  );
}
