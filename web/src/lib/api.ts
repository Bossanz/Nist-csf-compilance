import type { FunctionNode, Invitation, Organization, ProfilePatch, ProfileRow, Project, Role, Summary, User } from "./types";
const base = process.env.NEXT_PUBLIC_API_BASE_URL || "";
export class APIError extends Error { constructor(message:string,public status:number){super(message)} }
async function request<T>(path: string, init?: RequestInit): Promise<T> { const response = await fetch(`${base}${path}`, { headers: { "Content-Type": "application/json", ...(init?.headers || {}) }, ...init }); if (!response.ok) { const body = await response.json().catch(() => null); throw new APIError(body?.error?.message || `Request failed (${response.status})`,response.status); } if(response.status===204)return undefined as T; return response.json() as Promise<T>; }
export const api = {
  login: (input:{email:string;password:string}) => request<User>("/api/auth/login",{method:"POST",body:JSON.stringify(input)}),
  logout: () => request<void>("/api/auth/logout",{method:"POST"}),
  me: () => request<User>("/api/auth/me"),
  getOrganizations: () => request<Organization[]>("/api/organizations"),
  createOrganization: (input:{name:string}) => request<Organization>("/api/organizations",{method:"POST",body:JSON.stringify(input)}),
  deleteOrganization: (id:string) => request<void>(`/api/organizations/${id}`,{method:"DELETE"}),
  getOrganizationProjects: (id:string) => request<Project[]>(`/api/organizations/${id}/projects`),
  createOrganizationProject: (id:string,input:{name:string}) => request<Project>(`/api/organizations/${id}/projects`,{method:"POST",body:JSON.stringify(input)}),
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
  getSummary: (id: string) => request<Summary>(`/api/projects/${id}/summary`),
};
