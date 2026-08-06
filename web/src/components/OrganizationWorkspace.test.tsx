import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { OrganizationWorkspace } from "./OrganizationWorkspace";
import type { Organization, Project, User } from "../lib/types";

const counselor:User={id:"user-1",organizationID:null,name:"Consultant",email:"c@example.com",userType:"counselor",role:"counselor",status:"active"};
const orgAdmin:User={id:"user-2",organizationID:"org-1",name:"Customer Admin",email:"a@acme.test",userType:"stakeholder",role:"org_admin",status:"active"};
const organization:Organization={id:"org-1",name:"Acme",type:"client"};
const project:Project={id:"project-1",organizationID:"org-1",organizationName:"Acme",name:"Readiness",status:"setup",createdAt:"2026-08-06T03:00:00Z"};

test("counselor creates a project in the selected organization",()=>{const onCreate=vi.fn();render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={onCreate} onDeleteProject={vi.fn()} onInvite={vi.fn()}/>);fireEvent.change(screen.getByLabelText(/project name/i),{target:{value:"  Gap Review  "}});fireEvent.click(screen.getByRole("button",{name:/create project/i}));expect(onCreate).toHaveBeenCalledWith({name:"Gap Review"})});
test("organization admin can invite but cannot create projects",()=>{render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn()} onInvite={vi.fn()}/>);expect(screen.queryByRole("button",{name:/create project/i})).toBeNull();expect(screen.getByRole("button",{name:/create invitation/i})).toBeTruthy()});
