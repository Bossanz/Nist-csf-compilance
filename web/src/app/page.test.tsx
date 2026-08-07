import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import Home from "./page";
import { APIError, api } from "../lib/api";
import type { FunctionNode, Organization, ProfileRow, Project, Summary, User } from "../lib/types";

vi.mock("../lib/api",()=>({APIError:class APIError extends Error{constructor(message:string,public status:number){super(message)}},api:{login:vi.fn(),logout:vi.fn(),me:vi.fn(),getOrganizations:vi.fn(),createOrganization:vi.fn(),deleteOrganization:vi.fn(),getOrganizationProjects:vi.fn(),createOrganizationProject:vi.fn(),getOrganizationUsers:vi.fn(),createInvitation:vi.fn(),getFunctions:vi.fn(),getProfile:vi.fn(),getSummary:vi.fn(),deleteProject:vi.fn(),updateProfile:vi.fn()}}));

const user:User={id:"user-1",organizationID:null,name:"Consultant",email:"c@example.com",userType:"counselor",role:"counselor",status:"active"};
const admin:User={...user,name:"Admin",email:"admin@example.com",role:"counselor_admin"};
const organization:Organization={id:"org-1",name:"Acme",type:"client"};
const project:Project={id:"project-1",organizationID:"org-1",organizationName:"Acme",name:"Readiness",status:"setup",createdAt:"2026-08-06T03:00:00Z"};
const functions:FunctionNode[]=[{id:"function-1",code:"GV",name:"Govern",description:"Governance",categories:[]}];
const profile:ProfileRow[]=[];
const summary:Summary={coveragePct:0,includedCount:0,pendingCount:0,rejectedCount:0,functions:[]};

beforeEach(()=>{vi.clearAllMocks();vi.mocked(api.me).mockResolvedValue(user);vi.mocked(api.getOrganizations).mockResolvedValue([organization]);vi.mocked(api.getOrganizationProjects).mockResolvedValue([project]);vi.mocked(api.getOrganizationUsers).mockResolvedValue([]);vi.mocked(api.getFunctions).mockResolvedValue(functions);vi.mocked(api.getProfile).mockResolvedValue(profile);vi.mocked(api.getSummary).mockResolvedValue(summary)});

test("shows login when session restoration returns 401",async()=>{vi.mocked(api.me).mockRejectedValue(new APIError("Authentication required",401));render(<Home/>);expect(await screen.findByRole("heading",{name:/sign in/i})).toBeTruthy()});

test("navigates organizations then projects",async()=>{render(<Home/>);fireEvent.click(await screen.findByRole("button",{name:/open acme/i}));expect(await screen.findByRole("heading",{name:"Acme"})).toBeTruthy();fireEvent.click(screen.getByRole("button",{name:/open readiness/i}));expect(await screen.findByRole("heading",{name:"Readiness"})).toBeTruthy();fireEvent.click(screen.getByRole("button",{name:/back to organization/i}));expect(screen.getByRole("heading",{name:"Acme"})).toBeTruthy()});

test("creates a project inside the selected organization",async()=>{vi.mocked(api.getOrganizationProjects).mockResolvedValue([]);vi.mocked(api.createOrganizationProject).mockResolvedValue(project);render(<Home/>);fireEvent.click(await screen.findByRole("button",{name:/open acme/i}));fireEvent.change(await screen.findByLabelText(/project name/i),{target:{value:"Gap Review"}});fireEvent.click(screen.getByRole("button",{name:/create project/i}));await waitFor(()=>expect(api.createOrganizationProject).toHaveBeenCalledWith("org-1",{name:"Gap Review"}))});

test("counselor admin permanently deletes an organization",async()=>{vi.mocked(api.me).mockResolvedValue(admin);vi.mocked(api.deleteOrganization).mockResolvedValue(undefined);render(<Home/>);fireEvent.click(await screen.findByRole("button",{name:"Delete Acme"}));fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"Acme"}});fireEvent.click(screen.getByRole("button",{name:/delete permanently/i}));await waitFor(()=>expect(screen.queryByRole("heading",{name:"Acme"})).toBeNull())});

test("keeps the organization and shows an API deletion error",async()=>{vi.mocked(api.me).mockResolvedValue(admin);vi.mocked(api.deleteOrganization).mockRejectedValue(new APIError("Could not delete organization",500));render(<Home/>);fireEvent.click(await screen.findByRole("button",{name:"Delete Acme"}));fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"Acme"}});fireEvent.click(screen.getByRole("button",{name:/delete permanently/i}));expect((await screen.findByRole("alert")).textContent).toContain("Could not delete organization");expect(screen.getByRole("heading",{name:"Acme"})).toBeTruthy()});
