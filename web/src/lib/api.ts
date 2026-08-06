import type { FunctionNode, ProfilePatch, ProfileRow, Project, Summary } from "./types";
const base = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
async function request<T>(path: string, init?: RequestInit): Promise<T> { const response = await fetch(`${base}${path}`, { headers: { "Content-Type": "application/json", ...(init?.headers || {}) }, ...init }); if (!response.ok) { const body = await response.json().catch(() => null); throw new Error(body?.error?.message || `Request failed (${response.status})`); } return response.json() as Promise<T>; }
export const api = {
  getFunctions: () => request<FunctionNode[]>("/api/functions"),
  getProjects: () => request<Project[]>("/api/projects"),
  createProject: (input: { name: string; organizationName: string }) => request<Project>("/api/projects", { method: "POST", body: JSON.stringify(input) }),
  getProject: (id: string) => request<Project>(`/api/projects/${id}`),
  getProfile: (id: string) => request<ProfileRow[]>(`/api/projects/${id}/profile`),
  updateProfile: (projectID: string, subcategoryID: string, patch: ProfilePatch) => request<ProfileRow>(`/api/projects/${projectID}/profile/${subcategoryID}`, { method: "PUT", body: JSON.stringify(patch) }),
  getSummary: (id: string) => request<Summary>(`/api/projects/${id}/summary`),
};
