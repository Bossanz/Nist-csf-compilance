"use client";

import { useState } from "react";
import type { Project } from "../lib/types";

type NewProject = { name: string; organizationName: string };
type Props = {
  projects: Project[];
  loading: boolean;
  openingID: string;
  error: string;
  onOpen: (project: Project) => void;
  onCreate: (input: NewProject) => void;
  onDelete: (project: Project) => void;
};

function projectVersionLabel(project: Project) {
  return `Assessment v${project.versionNumber ?? 1}`;
}

export function ProjectDashboard({ projects, loading, openingID, error, onOpen, onCreate, onDelete }: Props) {
  const [name, setName] = useState("");
  const [organizationName, setOrganizationName] = useState("");

  function submit(event: React.FormEvent) {
    event.preventDefault();
    onCreate({ name: name.trim(), organizationName: organizationName.trim() });
  }

  return <main className="main dashboard">
    <header className="dashboard-header">
      <div><div className="context-line"><span>NIST CSF 2.0</span><span aria-hidden="true">/</span><span>Workspace</span></div><h1>Compliance projects</h1><p className="muted">Continue an assessment or start a new organizational profile.</p></div>
      <div className="project-total"><strong>{projects.length}</strong><span>projects</span></div>
    </header>

    {error && <div className="error" role="alert">{error}</div>}
    <section aria-labelledby="existing-projects">
      <div className="section-heading"><h2 id="existing-projects">Existing projects</h2><p className="muted">Your active assessment workspaces.</p></div>
      {loading ? <div className="panel muted" role="status" aria-live="polite" aria-busy="true">Loading projects…</div> : projects.length === 0 ?
        <div className="empty-state" role="status">No projects yet. Create one to begin an assessment.</div> :
        <div className="project-grid">{projects.map(project => <article className="project-card" key={project.id}>
          <div className="project-card-top"><div className="project-card-labels"><span className="status-chip">{project.status.replaceAll("_", " ")}</span><span className="project-version-label">{projectVersionLabel(project)}</span></div><time dateTime={project.createdAt}>{new Intl.DateTimeFormat("en-GB", { dateStyle: "medium" }).format(new Date(project.createdAt))}</time></div>
          <div><h3>{project.name}</h3><p className="muted">{project.organizationName}</p></div>
          <div className="project-actions"><button className="secondary" disabled={Boolean(openingID)} onClick={() => onOpen(project)} aria-label={`Open ${project.name}`}>{openingID === project.id ? "Working…" : "Open project"}</button><button className="danger" disabled={Boolean(openingID)} onClick={() => {if(window.confirm(`Delete ${project.name}? Its assessment data will be permanently deleted.`))onDelete(project)}} aria-label={`Delete ${project.name}`}>Delete</button></div>
        </article>)}</div>}
    </section>

    <section className="panel create-project" aria-labelledby="create-project">
      <div><h2 id="create-project">Create project</h2><p className="muted">A complete 106-outcome assessment workspace is prepared automatically.</p></div>
      <form className="form" onSubmit={submit}>
        <label className="field"><span>Project name</span><input required value={name} onChange={event => setName(event.target.value)} /></label>
        <label className="field"><span>Organization name</span><input value={organizationName} onChange={event => setOrganizationName(event.target.value)} /></label>
        <button className="primary" disabled={Boolean(openingID)}>Create project</button>
      </form>
    </section>
  </main>;
}
