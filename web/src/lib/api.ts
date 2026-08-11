import type { FunctionNode, Invitation, Organization, ProfilePatch, ProfileRow, Project, ProjectCreateInput, ResponseDocument, Role, StakeholderResponse, Summary, User } from "./types";
const base = process.env.NEXT_PUBLIC_API_BASE_URL || "";
export class APIError extends Error { constructor(message:string,public status:number){super(message)} }
async function request<T>(path: string, init?: RequestInit): Promise<T> { const isFormData = typeof FormData !== "undefined" && init?.body instanceof FormData; const headers = isFormData ? { ...(init?.headers || {}) } : { "Content-Type": "application/json", ...(init?.headers || {}) }; const response = await fetch(`${base}${path}`, { ...init, headers }); if (!response.ok) { const body = await response.json().catch(() => null); throw new APIError(body?.error?.message || `Request failed (${response.status})`,response.status); } if(response.status===204)return undefined as T; return response.json() as Promise<T>; }
async function download(path: string): Promise<Blob> { const response = await fetch(`${base}${path}`); if (!response.ok) { const body = await response.json().catch(() => null); throw new APIError(body?.error?.message || `Request failed (${response.status})`,response.status); } return response.blob(); }
export const api = {
  login: (input:{email:string;password:string}) => request<User>("/api/auth/login",{method:"POST",body:JSON.stringify(input)}),
  logout: () => request<void>("/api/auth/logout",{method:"POST"}),
  me: () => request<User>("/api/auth/me"),
  getOrganizations: () => request<Organization[]>("/api/organizations"),
  getOrganizationBySlug: (slug: string) => request<Organization>(`/api/organizations/by-slug/${encodeURIComponent(slug)}`),
  createOrganization: (input:{name:string}) => request<Organization>("/api/organizations",{method:"POST",body:JSON.stringify(input)}),
  deleteOrganization: (id:string) => request<void>(`/api/organizations/${id}`,{method:"DELETE"}),
  getOrganizationProjects: (id:string) => request<Project[]>(`/api/organizations/${id}/projects`),
  getOrganizationProjectBySlug: (organizationID: string, slug: string) => request<Project>(`/api/organizations/${organizationID}/projects/by-slug/${encodeURIComponent(slug)}`),
  createOrganizationProject: (id:string,input:ProjectCreateInput) => request<Project>(`/api/organizations/${id}/projects`,{method:"POST",body:JSON.stringify(input)}),
  getOrganizationUsers: (id:string) => request<User[]>(`/api/organizations/${id}/users`),
  updateOrganizationUser: (organizationID:string,userID:string,input:{role:Role;status:"active"|"disabled"}) => request<User>(`/api/organizations/${organizationID}/users/${userID}`,{method:"PATCH",body:JSON.stringify(input)}),
  createInvitation: (organizationID:string,input:{email:string;role:Role}) => request<Invitation>(`/api/organizations/${organizationID}/invitations`,{method:"POST",body:JSON.stringify(input)}),
  getCounselors: () => request<User[]>("/api/counselors"),
  updateCounselor: (userID:string,input:{role:Role;status:"active"|"disabled"}) => request<User>(`/api/counselors/${userID}`,{method:"PATCH",body:JSON.stringify(input)}),
  createCounselorInvitation: (input:{email:string;role:Role}) => request<Invitation>("/api/counselor-invitations",{method:"POST",body:JSON.stringify(input)}),
  acceptInvitation: (token:string,input:{name:string;password:string}) => request<User>(`/api/invitations/${token}/accept`,{method:"POST",body:JSON.stringify(input)}),
  getFunctions: () => request<FunctionNode[]>("/api/functions"),
  getProjects: () => request<Project[]>("/api/projects"),
  createProject: (input: { name: string; organizationName: string }) => request<Project>("/api/projects", { method: "POST", body: JSON.stringify(input) }),
  deleteProject: (id: string) => request<void>(`/api/projects/${id}`, { method: "DELETE" }),
  getProject: (id: string) => request<Project>(`/api/projects/${id}`),
  getProfile: (id: string) => request<ProfileRow[]>(`/api/projects/${id}/profile`),
  updateProfile: (projectID: string, subcategoryID: string, patch: ProfilePatch) => request<ProfileRow>(`/api/projects/${projectID}/profile/${subcategoryID}`, { method: "PUT", body: JSON.stringify(patch) }),
  getResponses: (id: string) => request<StakeholderResponse[]>(`/api/projects/${id}/responses`),
  saveResponse: (projectID: string, subcategoryID: string, responseText: string) => request<StakeholderResponse>(`/api/projects/${projectID}/responses/${subcategoryID}`, { method: "PUT", body: JSON.stringify({ responseText }) }),
  submitResponse: (projectID: string, subcategoryID: string) => request<StakeholderResponse>(`/api/projects/${projectID}/responses/${subcategoryID}/submit`, { method: "POST", body: JSON.stringify({}) }),
  reviewResponse: (projectID: string, subcategoryID: string, input: { status: "reviewed" | "needs_more_info"; comment: string }) => request<StakeholderResponse>(`/api/projects/${projectID}/responses/${subcategoryID}/review`, { method: "POST", body: JSON.stringify(input) }),
  uploadResponseDocument: (projectID: string, subcategoryID: string, file: File) => { const form = new FormData(); form.append("file", file); return request<ResponseDocument>(`/api/projects/${projectID}/responses/${subcategoryID}/documents`, { method: "POST", body: form }); },
  downloadResponseDocument: (projectID: string, subcategoryID: string, documentID: string) => download(`/api/projects/${projectID}/responses/${subcategoryID}/documents/${documentID}`),
  deleteResponseDocument: (projectID: string, subcategoryID: string, documentID: string) => request<void>(`/api/projects/${projectID}/responses/${subcategoryID}/documents/${documentID}`, { method: "DELETE" }),
  getSummary: (id: string) => request<Summary>(`/api/projects/${id}/summary`),
};
