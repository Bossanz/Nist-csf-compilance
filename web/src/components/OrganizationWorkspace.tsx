"use client";

import { useState } from "react";
import type { Organization, Project, ProjectCreateInput, Role, User } from "../lib/types";

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
  onCreateProject: (input: ProjectCreateInput) => void | Promise<void>;
  onDeleteProject: (project: Project) => Promise<void>;
  onInvite: (input: { email: string; role: Role }) => void | Promise<void>;
  onUpdateUser?: (userID: string, input: { role: Role; status: "active" | "disabled" }) => void | Promise<void>;
};

type StakeholderRole = "org_admin" | "assessor" | "reviewer" | "viewer";
const stakeholderRoles: StakeholderRole[] = ["org_admin", "assessor", "reviewer", "viewer"];
const stakeholderRoleLabels: Record<StakeholderRole, string> = {
  org_admin: "Organization admin: manage access and responses",
  assessor: "Assessor: complete assigned outcomes and evidence",
  reviewer: "Reviewer: review submitted responses",
  viewer: "Viewer: read only",
};
type StakeholderAccess = { role: Role; status: "active" | "disabled" };

function roleName(role: Role) {
  return role === "org_admin" ? "Organization admin" : role.replaceAll("_", " ");
}

export function OrganizationWorkspace({ user, organization, projects, users, loading, error, invitationURL, onBack, onOpen, onCreateProject, onDeleteProject, onInvite, onUpdateUser }: Props) {
  const [projectName, setProjectName] = useState("");
  const [projectObjective, setProjectObjective] = useState("");
  const [assessmentPeriod, setAssessmentPeriod] = useState("");
  const [targetCompletionDate, setTargetCompletionDate] = useState("");
  const [scopeBoundary, setScopeBoundary] = useState("");
  const [complianceDriver, setComplianceDriver] = useState("");
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<Role>("viewer");
  const [projectSaving, setProjectSaving] = useState(false);
  const [inviteSaving, setInviteSaving] = useState(false);
  const [inviteError, setInviteError] = useState("");
  const [pendingDelete, setPendingDelete] = useState<Project | null>(null);
  const [confirmation, setConfirmation] = useState("");
  const [deleteError, setDeleteError] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [memberAccess, setMemberAccess] = useState<Record<string, StakeholderAccess>>({});
  const [memberSaving, setMemberSaving] = useState("");
  const [memberError, setMemberError] = useState("");
  const counselor = user.role === "counselor_admin" || user.role === "counselor";
  const canInvite = counselor || user.role === "org_admin";
  const canManageUsers = counselor || user.role === "org_admin";

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

  function getMemberAccess(member: User): StakeholderAccess {
    const draft = memberAccess[member.id];
    return draft || { role: member.role, status: member.status === "disabled" ? "disabled" : "active" };
  }

  function setMemberField(member: User, field: keyof StakeholderAccess, value: string) {
    const current = getMemberAccess(member);
    setMemberAccess((rows) => ({ ...rows, [member.id]: { ...current, [field]: value as Role | "active" | "disabled" } as StakeholderAccess }));
  }

  async function saveMember(member: User) {
    if (!onUpdateUser) return;
    const access = getMemberAccess(member);
    setMemberSaving(member.id);
    setMemberError("");
    try {
      await onUpdateUser(member.id, access);
    } catch (cause) {
      setMemberError(cause instanceof Error ? cause.message : "Could not update stakeholder access");
    } finally {
      setMemberSaving("");
    }
  }

  async function createInvitation() {
    setInviteSaving(true);
    setInviteError("");
    try {
      await onInvite({ email: email.trim().toLowerCase(), role });
    } catch (cause) {
      setInviteError(cause instanceof Error ? cause.message : "Could not create invitation");
    } finally {
      setInviteSaving(false);
    }
  }

  async function createProject() {
    setProjectSaving(true);
    try {
      await onCreateProject({
        name: projectName.trim(),
        objective: projectObjective.trim(),
        assessmentPeriod: assessmentPeriod.trim(),
        targetCompletionDate,
        scopeBoundary: scopeBoundary.trim(),
        complianceDriver: complianceDriver.trim(),
      });
      setProjectName("");
      setProjectObjective("");
      setAssessmentPeriod("");
      setTargetCompletionDate("");
      setScopeBoundary("");
      setComplianceDriver("");
    } catch {
      // The page-level handler renders the API error; keep the form values for retry.
    } finally {
      setProjectSaving(false);
    }
  }

  return (
    <main className="main dashboard">
      <button className="text-button back-button" type="button" onClick={onBack}>Back to organizations</button>

      <header className="workspace-hero">
        <div>
          <div className="context-line"><span>Organization</span><span aria-hidden="true">/</span><span>{organization.type}</span></div>
          <h1>{organization.name}</h1>
          <p className="muted">Create projects and manage stakeholder access for this organization.</p>
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
        {loading ? <div className="panel muted" role="status" aria-live="polite" aria-busy="true">Loading workspace…</div> : projects.length === 0 ? (
          <div className="empty-state">No projects yet. Create one below to start an assessment.</div>
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
          <form className="panel inline-creator project-create-form" onSubmit={(event) => { event.preventDefault(); void createProject(); }}>
            <div className="project-create-heading"><h3>Create a project</h3><p className="muted">Define the assessment context before stakeholders add responses.</p></div>
            <label className="field"><span>Project name</span><input required value={projectName} onChange={(event) => setProjectName(event.target.value)} /></label>
            <label className="field"><span>Assessment objective</span><textarea rows={3} value={projectObjective} onChange={(event) => setProjectObjective(event.target.value)} /></label>
            <label className="field"><span>Assessment period</span><input placeholder="e.g. 2026" value={assessmentPeriod} onChange={(event) => setAssessmentPeriod(event.target.value)} /></label>
            <label className="field"><span>Target completion date</span><input type="date" value={targetCompletionDate} onChange={(event) => setTargetCompletionDate(event.target.value)} /></label>
            <label className="field"><span>Scope boundary</span><small className="field-help">Systems, teams, or locations included in this assessment.</small><textarea rows={3} value={scopeBoundary} onChange={(event) => setScopeBoundary(event.target.value)} /></label>
            <label className="field"><span>Compliance driver</span><small className="field-help">The requirement or goal that prompted this assessment.</small><textarea rows={3} value={complianceDriver} onChange={(event) => setComplianceDriver(event.target.value)} /></label>
            <button className="primary" type="submit" disabled={loading || projectSaving}>{projectSaving ? "Creating project…" : "Create project"}</button>
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
          {users.length === 0 ? <div className="empty-state">No active stakeholders yet. Invite someone below to begin.</div> : users.map((member) => {
            const access = getMemberAccess(member);
            return (
              <div className="person-row account-row" key={member.id}>
                <div><strong>{member.name}</strong><span>{member.email}</span></div>
                {canManageUsers && onUpdateUser ? (
                  <div className="account-controls">
                    <label className="field account-control"><span>Role for {member.email}</span><select aria-label={`Role for ${member.email}`} value={access.role} onChange={(event) => setMemberField(member, "role", event.target.value)}>{stakeholderRoles.map((item) => <option value={item} key={item}>{stakeholderRoleLabels[item]}</option>)}</select></label>
                    <label className="field account-control"><span>Status for {member.email}</span><select aria-label={`Status for ${member.email}`} value={access.status} onChange={(event) => setMemberField(member, "status", event.target.value)}><option value="active">active</option><option value="disabled" disabled={member.id === user.id}>disabled</option></select></label>
                    <button className="secondary account-save" type="button" aria-label={`Save access for ${member.email}`} disabled={memberSaving === member.id} onClick={() => void saveMember(member)}>{memberSaving === member.id ? "Saving..." : "Save stakeholder access"}</button>
                  </div>
                ) : (
                  <span className="role-chip">{roleName(member.role)}</span>
                )}
              </div>
            );
          })}
        </div>
        {memberError && <div className="error account-error" role="alert">{memberError}</div>}
        {inviteError && <div className="error account-error" role="alert">{inviteError}</div>}
        {canInvite && (
          <>
            <div className="section-heading invite-heading">
              <h3 id="invite-stakeholder-heading">Invite a stakeholder</h3>
              <p className="muted">They will receive a one-time link to set a password and join this organization.</p>
            </div>
            <form className="panel invite-form" aria-labelledby="invite-stakeholder-heading" onSubmit={(event) => { event.preventDefault(); void createInvitation(); }}>
            <label className="field"><span>Email</span><input type="email" required value={email} onChange={(event) => setEmail(event.target.value)} /></label>
            <label className="field"><span>Access role</span><select value={role} onChange={(event) => setRole(event.target.value as Role)}>{stakeholderRoles.map((item) => <option value={item} key={item}>{stakeholderRoleLabels[item]}</option>)}</select></label>
            <button className="primary" type="submit" disabled={inviteSaving}>{inviteSaving ? "Creating invitation…" : "Create invitation"}</button>
            {invitationURL && <label className="field invitation-result"><span>One-time invite link (share with stakeholder)</span><input readOnly value={invitationURL} /></label>}
          </form>
          </>
        )}
      </section>
    </main>
  );
}
