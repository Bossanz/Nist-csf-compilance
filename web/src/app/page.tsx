"use client";
import { useEffect, useMemo, useState } from "react";
import { api } from "../lib/api";
import type { FunctionNode, ProfileRow, Project, Summary } from "../lib/types";
import { FunctionSidebar } from "../components/FunctionSidebar";
import { SummaryCards } from "../components/SummaryCards";
import { ProfileEditor } from "../components/ProfileEditor";

export default function Home() { const [functions,setFunctions]=useState<FunctionNode[]>([]); const [project,setProject]=useState<Project|null>(null); const [profile,setProfile]=useState<ProfileRow[]>([]); const [summary,setSummary]=useState<Summary>({coveragePct:0,includedCount:0,pendingCount:0,rejectedCount:0,functions:[]}); const [selected,setSelected]=useState(""); const [name,setName]=useState(""); const [organizationName,setOrganizationName]=useState(""); const [error,setError]=useState("");
  useEffect(()=>{api.getFunctions().then(items=>{setFunctions(items);setSelected(items[0]?.code||"")}).catch(e=>setError(e.message))},[]);
  async function createProject(e:React.FormEvent){e.preventDefault();setError("");try{const p=await api.createProject({name,organizationName});setProject(p);const [rows,s]=await Promise.all([api.getProfile(p.id),api.getSummary(p.id)]);setProfile(rows);setSummary(s)}catch(e){setError(e instanceof Error?e.message:"Could not create project")}}
  const visible=useMemo(()=>profile.filter(row=>row.functionCode===selected),[profile,selected]);
  async function save(id:string,patch:Parameters<typeof api.updateProfile>[2]){if(!project)return;const updated=await api.updateProfile(project.id,id,patch);setProfile(rows=>rows.map(row=>row.subcategoryID===id?updated:row));setSummary(await api.getSummary(project.id));}
  if(!project)return <main className="main"><div className="eyebrow">NIST CSF 2.0</div><h1>Start a compliance project</h1><p className="muted">Create a project to begin a Current / Target Profile assessment.</p><div className="panel" style={{maxWidth:560,marginTop:24}}><form className="form" onSubmit={createProject}><input required placeholder="Project name" value={name} onChange={e=>setName(e.target.value)}/><input placeholder="Organization name" value={organizationName} onChange={e=>setOrganizationName(e.target.value)}/><button className="primary">Create project</button></form>{error&&<div className="error">{error}</div>}</div></main>;
  return <div className="shell"><FunctionSidebar functions={functions} selectedCode={selected} onSelect={setSelected}/><main className="main"><div className="eyebrow">Project / {project.status}</div><h1>{project.name}</h1><p className="muted">Current and Target Profile assessment</p><SummaryCards summary={summary}/><section className="workspace-heading"><div><span className="section-index">ASSESS / {selected}</span><h2>Outcome assessments</h2><p className="muted">Open an outcome to document scope, current practice, and the target state.</p></div><div className="outcome-count"><strong>{visible.length}</strong><span>outcomes</span></div></section><ProfileEditor rows={visible} onSave={save}/></main></div>;
}
