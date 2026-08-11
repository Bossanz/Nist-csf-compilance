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
  const [invitationURL, setInvitationURL] = useState("");
  const [authChecked, setAuthChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [organizationSlug]);

  async function initialize(active: boolean) {
    setLoading(true);
    setError("");
    setNotFound(false);
    try {
      const currentUser = await api.me();
      const nextOrganization = await api.getOrganizationBySlug(organizationSlug);
      const [projectRows, memberRows] = await Promise.all([
        api.getOrganizationProjects(nextOrganization.id),
        api.getOrganizationUsers(nextOrganization.id),
      ]);
      if (!active) return;
      setUser(currentUser);
      setOrganization(nextOrganization);
      setProjects(projectRows);
      setMembers(memberRows);
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
    } finally {
      setLoading(false);
    }
  }

  async function deleteProject(project: Project) {
    await api.deleteProject(project.id);
    setProjects((rows) => rows.filter((row) => row.id !== project.id));
  }

  async function invite(input: { email: string; role: Role }) {
    if (!organization) return;
    const invitation: Invitation = await api.createInvitation(organization.id, input);
    setInvitationURL(invitation.invitationURL);
  }

  if (!authChecked || !user || loading && !organization) return <main className="screen-center">Loading...</main>;
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

  return (
    <OrganizationWorkspace
      user={user}
      organization={organization}
      projects={projects}
      users={members}
      loading={loading}
      error={error}
      invitationURL={invitationURL}
      onBack={() => router.push("/organizations")}
      onOpen={(project) => router.push(projectPath(organization, project))}
      onCreateProject={createProject}
      onDeleteProject={deleteProject}
      onInvite={invite}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
