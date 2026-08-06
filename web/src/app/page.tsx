"use client";
import { useEffect, useMemo, useState } from "react";
import { api } from "../lib/api";
import type { FunctionNode, ProfileRow, Project, Summary } from "../lib/types";
import { FunctionSidebar } from "../components/FunctionSidebar";
import { SummaryCards } from "../components/SummaryCards";
import { ProfileEditor } from "../components/ProfileEditor";
import { ProjectDashboard } from "../components/ProjectDashboard";

const emptySummary: Summary = {coveragePct:0,includedCount:0,pendingCount:0,rejectedCount:0,functions:[]};

export default function Home() {
  const [functions,setFunctions]=useState<FunctionNode[]>([]); const [projects,setProjects]=useState<Project[]>([]); const [project,setProject]=useState<Project|null>(null); const [profile,setProfile]=useState<ProfileRow[]>([]); const [summary,setSummary]=useState<Summary>(emptySummary); const [selected,setSelected]=useState(""); const [loadingProjects,setLoadingProjects]=useState(true); const [openingID,setOpeningID]=useState(""); const [error,setError]=useState("");
  useEffect(()=>{Promise.all([api.getFunctions(),api.getProjects()]).then(([catalog,items])=>{setFunctions(catalog);setSelected(catalog[0]?.code||"");setProjects(items)}).catch(e=>setError(e instanceof Error?e.message:"Could not load projects")).finally(()=>setLoadingProjects(false))},[]);

  async function openProject(item:Project){setOpeningID(item.id);setError("");try{const [rows,nextSummary]=await Promise.all([api.getProfile(item.id),api.getSummary(item.id)]);setProfile(rows);setSummary(nextSummary);setProject(item)}catch(e){setError(e instanceof Error?e.message:"Could not open project")}finally{setOpeningID("")}}
  async function createProject(input:{name:string;organizationName:string}){setOpeningID("new");setError("");try{const item=await api.createProject(input);setProjects(items=>[item,...items]);const [rows,nextSummary]=await Promise.all([api.getProfile(item.id),api.getSummary(item.id)]);setProfile(rows);setSummary(nextSummary);setProject(item)}catch(e){setError(e instanceof Error?e.message:"Could not create project")}finally{setOpeningID("")}}
  const visible=useMemo(()=>profile.filter(row=>row.functionCode===selected),[profile,selected]);
  async function save(id:string,patch:Parameters<typeof api.updateProfile>[2]){if(!project)return;const updated=await api.updateProfile(project.id,id,patch);setProfile(rows=>rows.map(row=>row.subcategoryID===id?updated:row));setSummary(await api.getSummary(project.id));}

  if(!project)return <ProjectDashboard projects={projects} loading={loadingProjects} openingID={openingID} error={error} onOpen={openProject} onCreate={createProject}/>;
  return <div className="shell"><FunctionSidebar functions={functions} selectedCode={selected} onSelect={setSelected}/><main className="main"><button className="secondary back-button" onClick={()=>{setProject(null);setError("")}}>← Back to projects</button><div className="eyebrow">Project / {project.status}</div><h1>{project.name}</h1><p className="muted">Current and Target Profile assessment</p><SummaryCards summary={summary}/><section className="workspace-heading"><div><span className="section-index">ASSESS / {selected}</span><h2>Outcome assessments</h2><p className="muted">Open an outcome to document scope, current practice, and the target state.</p></div><div className="outcome-count"><strong>{visible.length}</strong><span>outcomes</span></div></section><ProfileEditor rows={visible} onSave={save}/></main></div>;
}
