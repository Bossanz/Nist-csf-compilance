import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { OrganizationWorkspace } from "./OrganizationWorkspace";
import type { Organization, Project, User } from "../lib/types";

const counselor:User={id:"user-1",organizationID:null,name:"Consultant",email:"c@example.com",userType:"counselor",role:"counselor",status:"active"};
const orgAdmin:User={id:"user-2",organizationID:"org-1",name:"Customer Admin",email:"a@acme.test",userType:"stakeholder",role:"org_admin",status:"active"};
const organization:Organization={id:"org-1",name:"Acme",slug:"acme",type:"client"};
const project:Project={id:"project-1",organizationID:"org-1",organizationName:"Acme",name:"Readiness",slug:"readiness",status:"setup",createdAt:"2026-08-06T03:00:00Z"};

test("counselor creates a project with assessment context",()=>{const onCreate=vi.fn();render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={onCreate} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()}/>);fireEvent.change(screen.getByLabelText(/project name/i),{target:{value:"  Gap Review  "}});fireEvent.change(screen.getByLabelText(/objective/i),{target:{value:"Prepare for registration"}});fireEvent.change(screen.getByLabelText(/assessment period/i),{target:{value:"2026"}});fireEvent.change(screen.getByLabelText(/target completion date/i),{target:{value:"2026-09-30"}});fireEvent.change(screen.getByLabelText(/scope boundary/i),{target:{value:"Registration systems"}});fireEvent.change(screen.getByLabelText(/compliance driver/i),{target:{value:"Regulatory requirement"}});fireEvent.click(screen.getByRole("button",{name:/create project/i}));expect(onCreate).toHaveBeenCalledWith({name:"Gap Review",objective:"Prepare for registration",assessmentPeriod:"2026",targetCompletionDate:"2026-09-30",scopeBoundary:"Registration systems",complianceDriver:"Regulatory requirement"})});
test("organization admin can invite but cannot create projects",()=>{render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()}/>);expect(screen.queryByRole("button",{name:/create project/i})).toBeNull();expect(screen.getByRole("button",{name:/create invitation/i})).toBeTruthy()});
test("requires an Auditor project assignment before creating invitation", async()=>{const onInvite=vi.fn();render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={onInvite}/>);fireEvent.change(screen.getByLabelText(/access role/i),{target:{value:"auditor"}});fireEvent.change(screen.getByLabelText(/email/i),{target:{value:"auditor@example.com"}});fireEvent.click(screen.getByRole("button",{name:/create invitation/i}));expect((await screen.findByRole("alert")).textContent).toMatch(/select at least one project/i);fireEvent.click(screen.getByRole("checkbox",{name:/readiness/i}));fireEvent.click(screen.getByRole("button",{name:/create invitation/i}));await waitFor(()=>expect(onInvite).toHaveBeenCalledWith({email:"auditor@example.com",role:"auditor",projectIDs:[project.id]}))});
test("explains stakeholder invitation roles",()=>{render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()}/>);expect(screen.getByRole("heading",{name:/invite a stakeholder/i})).toBeTruthy();expect(screen.getByRole("option",{name:/assessor.*complete assigned outcomes/i})).toBeTruthy()});
test("keeps Auditor invitations with the organization admin",()=>{render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()}/>);expect(screen.queryByRole("option",{name:/auditor.*read-only review/i})).toBeNull()});

test("labels the projects and stakeholders sections",()=>{render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()}/>);expect(screen.getByRole("region",{name:/projects/i})).toBeTruthy();expect(screen.getByRole("region",{name:/stakeholders/i})).toBeTruthy()});

test("requires confirmation before deleting a project",async()=>{const onDelete=vi.fn().mockResolvedValue(undefined);render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={onDelete} onInvite={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:"Delete Readiness"}));expect(screen.getByRole("dialog",{name:/delete readiness/i})).toBeTruthy();const confirm=screen.getByRole("button",{name:/delete project/i}) as HTMLButtonElement;expect(confirm.disabled).toBe(true);fireEvent.change(screen.getByLabelText("Type Readiness to confirm"),{target:{value:"Readiness"}});fireEvent.click(confirm);await waitFor(()=>expect(onDelete).toHaveBeenCalledWith(project))});

test("shows a project deletion error and keeps the confirmation open",async()=>{const onDelete=vi.fn().mockRejectedValue(new Error("Could not delete project"));render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={onDelete} onInvite={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:"Delete Readiness"}));fireEvent.change(screen.getByLabelText("Type Readiness to confirm"),{target:{value:"Readiness"}});fireEvent.click(screen.getByRole("button",{name:/delete project/i}));expect((await screen.findByRole("alert")).textContent).toContain("Could not delete project");expect(screen.getByRole("dialog",{name:/delete readiness/i})).toBeTruthy()});

test("counselor updates a stakeholder role and status", async () => {
  const onUpdateUser = vi.fn().mockResolvedValue(undefined);
  render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[{ ...orgAdmin, role: "assessor" }]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()} onUpdateUser={onUpdateUser} />);
  fireEvent.change(screen.getByLabelText(/role for a@acme.test/i), { target: { value: "reviewer" } });
  fireEvent.change(screen.getByLabelText(/status for a@acme.test/i), { target: { value: "disabled" } });
  fireEvent.click(screen.getByRole("button", { name: /save access for a@acme.test/i }));
  await waitFor(() => expect(onUpdateUser).toHaveBeenCalledWith("user-2", { role: "reviewer", status: "disabled" }));
});

test("current user cannot select disabled status", () => {
  render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()} onUpdateUser={vi.fn()} />);
  expect(screen.getByLabelText(/status for a@acme.test/i).querySelector('option[value="disabled"]')?.hasAttribute("disabled")).toBe(true);
});

test("shows an invitation error instead of leaving a rejected request unhandled", async () => {
  const onInvite = vi.fn().mockRejectedValue(new Error("An active user or invitation already exists"));
  render(<OrganizationWorkspace user={orgAdmin} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={vi.fn()} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={onInvite} />);

  fireEvent.change(screen.getByLabelText(/^email$/i), { target: { value: "member@acme.test" } });
  fireEvent.click(screen.getByRole("button", { name: /create invitation/i }));

  expect((await screen.findByRole("alert")).textContent).toContain("An active user or invitation already exists");
});

test("preserves project context when creation fails", async () => {
  const onCreateProject = vi.fn().mockRejectedValue(new Error("Could not create project"));
  render(<OrganizationWorkspace user={counselor} organization={organization} projects={[project]} users={[orgAdmin]} loading={false} error="" onBack={vi.fn()} onOpen={vi.fn()} onCreateProject={onCreateProject} onDeleteProject={vi.fn().mockResolvedValue(undefined)} onInvite={vi.fn()} />);

  const input = screen.getByLabelText(/project name/i) as HTMLInputElement;
  fireEvent.change(input, { target: { value: "Readiness review" } });
  fireEvent.click(screen.getByRole("button", { name: /create project/i }));

  await waitFor(() => expect(onCreateProject).toHaveBeenCalled());
  expect(input.value).toBe("Readiness review");
});
