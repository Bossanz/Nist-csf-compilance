"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { APIError, api } from "../../../lib/api";
import type { Invitation, Organization, Project, ProjectCreateInput, Role, User } from "../../../lib/types";
import { OrganizationWorkspace } from "../../../components/OrganizationWorkspace";
import { projectPath } from "../../../lib/routes";

export default function OrganizationPage() {
  const router = useRouter();
  const params = useParams<{ organizationSlug: string }>();
  const organizationSlug = params.organizationSlug;
  const [user, setUser] = useState<User | null>(null);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [members, setMembers] = useState<User[]>([]);
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [invitationURL, setInvitationURL] = useState("");
  const [authChecked, setAuthChecked] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [organizationSlug, retryCount]);

  async function initialize(active: boolean) {
    setLoading(true);
    setError("");
    setNotFound(false);
    try {
      const [currentUser, nextOrganization] = await Promise.all([
        api.me(),
        api.getOrganizationBySlug(organizationSlug),
      ]);
      const canManageInvitations = currentUser.role === "counselor_admin"
        || currentUser.role === "counselor"
        || currentUser.role === "org_admin";
      const [projectRows, memberRows, invitationRows] = await Promise.all([
        api.getOrganizationProjects(nextOrganization.id),
        api.getOrganizationUsers(nextOrganization.id),
        canManageInvitations
          ? api.getOrganizationInvitations(nextOrganization.id)
          : Promise.resolve<Invitation[]>([]),
      ]);
      if (!active) return;
      setUser(currentUser);
      setOrganization(nextOrganization);
      setProjects(projectRows);
      setMembers(memberRows);
      setInvitations(invitationRows);
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 401) {
        router.replace("/login");
      } else if (active) {
        setNotFound(cause instanceof APIError && cause.status === 404);
        setError(messageOf(cause));
      }
    } finally {
      if (active) {
        setLoading(false);
        setAuthChecked(true);
      }
    }
  }

  async function createProject(input: ProjectCreateInput) {
    if (!organization) return;
    setLoading(true);
    setError("");
    try {
      const created = await api.createOrganizationProject(organization.id, input);
      setProjects((rows) => [...rows, created]);
    } catch (cause) {
      setError(messageOf(cause));
      throw cause;
    } finally {
      setLoading(false);
    }
  }

  async function deleteProject(project: Project) {
    await api.deleteProject(project.id);
    setProjects((rows) => rows.filter((row) => row.id !== project.id));
  }

  async function invite(input: { email: string; role: Role; projectIDs?: string[] }) {
    if (!organization) return;
    const invitation: Invitation = await api.createInvitation(organization.id, input);
    setInvitationURL(invitation.invitationURL || "");
    setInvitations(await api.getOrganizationInvitations(organization.id));
  }

  async function resendInvitation(invitation: Invitation) {
    if (!organization) return;
    const replacement = await api.resendInvitation(organization.id, invitation.id);
    setInvitationURL(replacement.invitationURL || "");
    setInvitations(await api.getOrganizationInvitations(organization.id));
  }

  async function cancelInvitation(invitation: Invitation) {
    if (!organization) return;
    const cancelled = await api.cancelInvitation(organization.id, invitation.id);
    setInvitations((rows) => rows.map((row) => row.id === cancelled.id ? cancelled : row));
  }

  async function updateUser(userID: string, input: { role: Role; status: "active" | "disabled" }) {
    if (!organization) return;
    const updated = await api.updateOrganizationUser(organization.id, userID, input);
    setMembers((rows) => rows.map((row) => row.id === updated.id ? updated : row));
  }

  function retryLoad() {
    setUser(null);
    setOrganization(null);
    setProjects([]);
    setMembers([]);
    setInvitations([]);
    setError("");
    setNotFound(false);
    setAuthChecked(false);
    setRetryCount((value) => value + 1);
  }

  if (!authChecked || loading && !organization) return <main className="screen-center" role="status" aria-live="polite" aria-busy="true">Loading workspace…</main>;
  if (!user && error && !notFound) {
    return (
      <main className="screen-center" aria-labelledby="organization-load-error">
        <section className="empty-state" role="alert">
          <h1 id="organization-load-error">Could not load organization</h1>
          <p>{error}</p>
          <button className="primary" type="button" onClick={retryLoad}>Try again</button>
        </section>
      </main>
    );
  }
  if (notFound || !organization) {
    return (
      <main className="main dashboard">
        <button className="text-button back-button" type="button" onClick={() => router.push("/organizations")}>Organizations</button>
        <section className="empty-state" aria-labelledby="organization-not-found">
          <h1 id="organization-not-found">Organization not found</h1>
          <p>{error || "This workspace is unavailable."}</p>
        </section>
      </main>
    );
  }
  if (!user) return <main className="screen-center" role="status" aria-live="polite">Redirecting to sign in…</main>;

  return (
    <OrganizationWorkspace
      user={user}
      organization={organization}
      projects={projects}
      users={members}
      invitations={invitations}
      loading={loading}
      error={error}
      invitationURL={invitationURL}
      onBack={() => router.push("/organizations")}
      onOpen={(project) => router.push(projectPath(organization, project))}
      onCreateProject={createProject}
      onDeleteProject={deleteProject}
      onInvite={invite}
      onResendInvitation={resendInvitation}
      onCancelInvitation={cancelInvitation}
      onUpdateUser={updateUser}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
