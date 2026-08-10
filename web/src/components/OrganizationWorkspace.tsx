"use client";

import { useState } from "react";
import type { Organization, Project, Role, User } from "../lib/types";

type Props = {
  user: User;
  organization: Organization;
  projects: Project[];
  users: User[];
  loading: boolean;
  error: string;
  invitationURL?: string;
  onBack: () => void;
  onOpen: (project: Project) => void;
  onCreateProject: (input: { name: string }) => void;
  onDeleteProject: (project: Project) => Promise<void>;
  onInvite: (input: { email: string; role: Role }) => void;
};

const stakeholderRoles: Role[] = ["org_admin", "assessor", "reviewer", "viewer"];

export function OrganizationWorkspace({ user, organization, projects, users, loading, error, invitationURL, onBack, onOpen, onCreateProject, onDeleteProject, onInvite }: Props) {
  const [projectName, setProjectName] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<Role>("viewer");
  const [pendingDelete, setPendingDelete] = useState<Project | null>(null);
  const [confirmation, setConfirmation] = useState("");
  const [deleteError, setDeleteError] = useState("");
  const [deleting, setDeleting] = useState(false);
  const counselor = user.role === "counselor_admin" || user.role === "counselor";
  const canInvite = counselor || user.role === "org_admin";

  function closeDeleteConfirmation() {
    setPendingDelete(null);
    setConfirmation("");
    setDeleteError("");
  }

  async function confirmDelete() {
    if (!pendingDelete) return;
    setDeleting(true);
    setDeleteError("");
    try {
      await onDeleteProject(pendingDelete);
      closeDeleteConfirmation();
    } catch (cause) {
      setDeleteError(cause instanceof Error ? cause.message : "Could not delete project");
    } finally {
      setDeleting(false);
    }
  }

  return (
    <main className="main dashboard">
      <button className="text-button back-button" type="button" onClick={onBack}>Organizations</button>

      <header className="workspace-hero">
        <div>
          <div className="context-line"><span>Organization</span><span aria-hidden="true">/</span><span>{organization.type}</span></div>
          <h1>{organization.name}</h1>
          <p className="muted">Projects and customer access are managed in one place.</p>
        </div>
        <div className="workspace-stats" aria-label="Workspace totals">
          <div><strong>{projects.length}</strong><span>projects</span></div>
          <div><strong>{users.length}</strong><span>people</span></div>
        </div>
      </header>

      {error && <div className="error" role="alert">{error}</div>}

      <section aria-labelledby="projects-heading">
        <div className="section-heading">
          <h2 id="projects-heading">Projects</h2>
          <p className="muted">Assessment workspaces inside {organization.name}.</p>
        </div>
        {loading ? <div className="panel muted">Loading workspace…</div> : projects.length === 0 ? (
          <div className="empty-state">No projects in this organization yet.</div>
        ) : (
          <div className="project-grid">
            {projects.map((project) => (
              <article className="project-card" key={project.id}>
                <div className="project-card-top">
                  <span className="status-chip">{project.status.replaceAll("_", " ")}</span>
                  <time dateTime={project.createdAt}>{new Intl.DateTimeFormat("en-GB", { dateStyle: "medium" }).format(new Date(project.createdAt))}</time>
                </div>
                <div><h3>{project.name}</h3><p className="muted">{organization.name}</p></div>
                <div className="project-actions">
                  <button className="secondary" type="button" aria-label={`Open ${project.name}`} onClick={() => onOpen(project)}>Open project</button>
                  {counselor && <button className="danger" type="button" aria-label={`Delete ${project.name}`} onClick={() => { setPendingDelete(project); setConfirmation(""); setDeleteError(""); }}>Delete</button>}
                </div>
              </article>
            ))}
          </div>
        )}

        {counselor && (
          <form className="panel inline-creator" onSubmit={(event) => { event.preventDefault(); onCreateProject({ name: projectName.trim() }); setProjectName(""); }}>
            <div><h3>New project</h3><p className="muted">Create a fresh assessment inside this organization.</p></div>
            <label className="field"><span>Project name</span><input required value={projectName} onChange={(event) => setProjectName(event.target.value)} /></label>
            <button className="primary">Create project</button>
          </form>
        )}
      </section>

      {pendingDelete && (
        <section className="delete-confirmation" role="dialog" aria-labelledby="delete-project-title">
          <div>
            <h2 id="delete-project-title">Delete {pendingDelete.name}?</h2>
            <p>This removes the project and its assessment responses. It cannot be undone.</p>
          </div>
          <label className="field">
            <span>Type {pendingDelete.name} to confirm</span>
            <input autoFocus value={confirmation} onChange={(event) => setConfirmation(event.target.value)} />
          </label>
          {deleteError && <div className="error" role="alert">{deleteError}</div>}
          <div className="confirmation-actions">
            <button className="secondary" type="button" onClick={closeDeleteConfirmation}>Cancel</button>
            <button className="danger" type="button" disabled={confirmation !== pendingDelete.name || deleting} onClick={() => void confirmDelete()}>{deleting ? "Deleting…" : "Delete project"}</button>
          </div>
        </section>
      )}

      <section className="people-section" aria-labelledby="stakeholders-heading">
        <div className="section-heading">
          <h2 id="stakeholders-heading">Stakeholders</h2>
          <p className="muted">People who can access this organization and its projects.</p>
        </div>
        <div className="people-list">
          {users.length === 0 ? <div className="empty-state">No active stakeholders yet.</div> : users.map((member) => (
            <div className="person-row" key={member.id}>
              <div><strong>{member.name}</strong><span>{member.email}</span></div>
              <span className="role-chip">{member.role.replaceAll("_", " ")}</span>
            </div>
          ))}
        </div>
        {canInvite && (
          <form className="panel invite-form" onSubmit={(event) => { event.preventDefault(); onInvite({ email: email.trim().toLowerCase(), role }); }}>
            <label className="field"><span>Email</span><input type="email" required value={email} onChange={(event) => setEmail(event.target.value)} /></label>
            <label className="field"><span>Role</span><select value={role} onChange={(event) => setRole(event.target.value as Role)}>{stakeholderRoles.map((item) => <option value={item} key={item}>{item.replaceAll("_", " ")}</option>)}</select></label>
            <button className="primary">Create invitation</button>
            {invitationURL && <label className="field invitation-result"><span>One-time invitation link</span><input readOnly value={invitationURL} /></label>}
          </form>
        )}
      </section>
    </main>
  );
}
