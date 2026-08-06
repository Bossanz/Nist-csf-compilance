"use client";

import { useEffect, useMemo, useState } from "react";
import { APIError, api } from "../lib/api";
import type { FunctionNode, Organization, ProfilePatch, ProfileRow, Project, Role, Summary, User } from "../lib/types";
import { LoginForm } from "../components/LoginForm";
import { OrganizationDashboard } from "../components/OrganizationDashboard";
import { OrganizationWorkspace } from "../components/OrganizationWorkspace";
import { FunctionSidebar } from "../components/FunctionSidebar";
import { SummaryCards } from "../components/SummaryCards";
import { ProfileEditor } from "../components/ProfileEditor";

const emptySummary: Summary = { coveragePct: 0, includedCount: 0, pendingCount: 0, rejectedCount: 0, functions: [] };

export default function Home() {
  const [user, setUser] = useState<User | null>(null);
  const [authChecked, setAuthChecked] = useState(false);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [members, setMembers] = useState<User[]>([]);
  const [project, setProject] = useState<Project | null>(null);
  const [functions, setFunctions] = useState<FunctionNode[]>([]);
  const [profile, setProfile] = useState<ProfileRow[]>([]);
  const [summary, setSummary] = useState<Summary>(emptySummary);
  const [selected, setSelected] = useState("");
  const [invitationURL, setInvitationURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    void initialize();
  }, []);

  async function initialize() {
    try {
      const currentUser = await api.me();
      const [organizationRows, functionRows] = await Promise.all([
        api.getOrganizations(),
        api.getFunctions(),
      ]);
      setUser(currentUser);
      setOrganizations(organizationRows);
      setFunctions(functionRows);
      setSelected(functionRows[0]?.code ?? "");
    } catch (cause) {
      if (!(cause instanceof APIError) || cause.status !== 401) {
        setError(messageOf(cause));
      }
    } finally {
      setAuthChecked(true);
    }
  }

  async function login(input: { email: string; password: string }) {
    setLoading(true);
    setError("");
    try {
      const currentUser = await api.login(input);
      const [organizationRows, functionRows] = await Promise.all([
        api.getOrganizations(),
        api.getFunctions(),
      ]);
      setUser(currentUser);
      setOrganizations(organizationRows);
      setFunctions(functionRows);
      setSelected(functionRows[0]?.code ?? "");
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  async function logout() {
    await api.logout();
    setUser(null);
    setOrganization(null);
    setProject(null);
    setOrganizations([]);
    setError("");
  }

  async function createOrganization(input: { name: string }) {
    setLoading(true);
    setError("");
    try {
      const created = await api.createOrganization(input);
      setOrganizations((rows) => [...rows, created]);
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  async function openOrganization(nextOrganization: Organization) {
    setLoading(true);
    setError("");
    try {
      const [projectRows, memberRows] = await Promise.all([
        api.getOrganizationProjects(nextOrganization.id),
        api.getOrganizationUsers(nextOrganization.id),
      ]);
      setOrganization(nextOrganization);
      setProjects(projectRows);
      setMembers(memberRows);
      setInvitationURL("");
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  async function createProject(input: { name: string }) {
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

  async function openProject(nextProject: Project) {
    setLoading(true);
    setError("");
    try {
      const [nextProfile, nextSummary] = await Promise.all([
        api.getProfile(nextProject.id),
        api.getSummary(nextProject.id),
      ]);
      setProfile(nextProfile);
      setSummary(nextSummary);
      setProject(nextProject);
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  async function deleteProject(item: Project) {
    await api.deleteProject(item.id);
    setProjects((rows) => rows.filter((row) => row.id !== item.id));
  }

  async function invite(input: { email: string; role: Role }) {
    if (!organization) return;
    const invitation = await api.createInvitation(organization.id, input);
    setInvitationURL(invitation.invitationURL);
  }

  async function saveProfile(subcategoryID: string, patch: ProfilePatch) {
    if (!project) return;
    const updated = await api.updateProfile(project.id, subcategoryID, patch);
    setProfile((rows) => rows.map((row) => row.subcategoryID === subcategoryID ? updated : row));
    setSummary(await api.getSummary(project.id));
  }

  const visibleProfile = useMemo(
    () => profile.filter((row) => row.functionCode === selected),
    [profile, selected],
  );

  if (!authChecked) return <main className="screen-center">Loading…</main>;

  if (!user) {
    return <LoginForm loading={loading} error={error} onSubmit={login} />;
  }

  if (project && profile) {
    return (
      <div className="shell">
        <FunctionSidebar functions={functions} selectedCode={selected} onSelect={setSelected} />
        <main className="main">
          <button className="secondary back-button" onClick={() => setProject(null)}>← Back to organization</button>
          <div className="eyebrow">{organization?.name} / {project.status}</div>
          <h1>{project.name}</h1>
          <p className="muted">Current and Target Profile assessment</p>
          {error && <div className="error" role="alert">{error}</div>}
          <SummaryCards summary={summary} />
          <section className="workspace-heading">
            <div><span className="section-index">ASSESS / {selected}</span><h2>Outcome assessments</h2></div>
            <div className="outcome-count"><strong>{visibleProfile.length}</strong><span>outcomes</span></div>
          </section>
          <ProfileEditor
            rows={visibleProfile}
            onSave={saveProfile}
            readOnly={!['counselor_admin', 'counselor', 'assessor'].includes(user.role)}
          />
        </main>
      </div>
    );
  }

  if (organization) {
    return (
      <OrganizationWorkspace
        user={user}
        organization={organization}
        projects={projects}
        users={members}
        loading={loading}
        error={error}
        invitationURL={invitationURL}
        onBack={() => setOrganization(null)}
        onOpen={openProject}
        onCreateProject={createProject}
        onDeleteProject={deleteProject}
        onInvite={invite}
      />
    );
  }

  return (
    <OrganizationDashboard
      user={user}
      organizations={organizations}
      loading={loading}
      error={error}
      onSelect={openOrganization}
      onCreate={createOrganization}
      onLogout={logout}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
