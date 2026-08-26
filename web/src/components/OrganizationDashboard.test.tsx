import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { OrganizationDashboard } from "./OrganizationDashboard";
import type { Organization, User } from "../lib/types";

const admin:User={id:"user-1",organizationID:null,name:"Admin",email:"admin@example.com",userType:"counselor",role:"counselor_admin",status:"active"};
const counselor:User={...admin,id:"user-2",role:"counselor"};
const organization:Organization={id:"org-1",name:"Acme",slug:"acme",type:"client"};
const managedCounselor:User={...admin,id:"user-2",name:"Consultant",email:"counselor@example.com",role:"counselor",status:"active"};

test("selects an existing organization",()=>{const onSelect=vi.fn();render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={onSelect} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()}/>);expect(screen.getByRole("heading",{name:"Organizations"})).toBeTruthy();fireEvent.click(screen.getByRole("button",{name:/open acme/i}));expect(onSelect).toHaveBeenCalledWith(organization)});
test("links the signed-in user to password settings",()=>{render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()}/>);expect(screen.getByRole("link",{name:/change password/i}).getAttribute("href")).toBe("/account/password")});
test("explains how to start when no client organizations exist",()=>{render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()}/>);expect(screen.getByText("No client organizations yet. Create one below to begin.")).toBeTruthy()});
test("counselor admin creates a trimmed organization",()=>{const onCreate=vi.fn();render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={onCreate} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()}/>);fireEvent.change(screen.getByLabelText(/organization name/i),{target:{value:"  Acme  "}});fireEvent.click(screen.getByRole("button",{name:/create organization/i}));expect(onCreate).toHaveBeenCalledWith({name:"Acme"})});

test("counselor admin confirms permanent deletion with the exact organization name",()=>{const onDelete=vi.fn().mockResolvedValue(undefined);render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={onDelete} onLogout={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:"Delete Acme"}));expect(screen.getByRole("dialog",{name:/delete acme/i}).getAttribute("aria-modal")).toBe("true");const finalButton=screen.getByRole("button",{name:/delete permanently/i}) as HTMLButtonElement;expect(finalButton.disabled).toBe(true);fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"acme"}});expect(finalButton.disabled).toBe(true);fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"Acme"}});fireEvent.click(finalButton);expect(onDelete).toHaveBeenCalledWith(organization)});

test("keeps keyboard focus inside the organization deletion dialog", () => {
  render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Delete Acme" }));

  const dialog = screen.getByRole("dialog", { name: /delete acme/i });
  const confirmation = screen.getByLabelText(/type acme to confirm/i);
  const cancel = screen.getByRole("button", { name: "Cancel" });
  expect(document.activeElement).toBe(confirmation);

  (cancel as HTMLButtonElement).focus();
  fireEvent.keyDown(dialog, { key: "Tab" });
  expect(document.activeElement).toBe(confirmation);

  (confirmation as HTMLInputElement).focus();
  fireEvent.keyDown(dialog, { key: "Tab", shiftKey: true });
  expect(document.activeElement).toBe(cancel);
});

test("regular counselor cannot see organization deletion",()=>{render(<OrganizationDashboard user={counselor} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()}/>);expect(screen.queryByRole("button",{name:"Delete Acme"})).toBeNull()});

test("keeps the organization confirmation open when deletion fails",async()=>{const onDelete=vi.fn().mockRejectedValue(new Error("Could not delete organization"));render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={onDelete} onLogout={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:"Delete Acme"}));fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"Acme"}});fireEvent.click(screen.getByRole("button",{name:/delete permanently/i}));expect((await screen.findByRole("alert")).textContent).toContain("Could not delete organization");expect(screen.getByRole("dialog",{name:/delete acme/i})).toBeTruthy()});

test("counselor admin updates a counselor role and status", async () => {
  const onUpdateCounselor = vi.fn().mockResolvedValue(undefined);
  render(<OrganizationDashboard user={admin} organizations={[]} counselors={[managedCounselor]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} onUpdateCounselor={onUpdateCounselor} />);
  fireEvent.change(screen.getByLabelText(/role for counselor@example.com/i), { target: { value: "counselor_admin" } });
  fireEvent.change(screen.getByLabelText(/status for counselor@example.com/i), { target: { value: "disabled" } });
  fireEvent.click(screen.getByRole("button", { name: /save access for counselor@example.com/i }));
  await waitFor(() => expect(onUpdateCounselor).toHaveBeenCalledWith("user-2", { role: "counselor_admin", status: "disabled" }));
});

test("uses the shared ellipsis while counselor access saves", () => {
  let release!: () => void;
  const onUpdateCounselor = vi.fn(() => new Promise<void>((resolve) => { release = resolve; }));
  render(<OrganizationDashboard user={admin} organizations={[]} counselors={[managedCounselor]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} onUpdateCounselor={onUpdateCounselor} />);

  fireEvent.click(screen.getByRole("button", { name: /save access for counselor@example.com/i }));

  expect(screen.getByRole("button", { name: /save access for counselor@example.com/i }).textContent).toContain("Saving…");
  release();
});

test("counselor admin creates a counselor invitation", () => {
  const onInviteCounselor = vi.fn();
  render(<OrganizationDashboard user={admin} organizations={[]} counselors={[]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} onInviteCounselor={onInviteCounselor} />);
  fireEvent.change(screen.getByLabelText(/counselor email/i), { target: { value: "  new@example.com " } });
  fireEvent.change(screen.getByLabelText(/counselor role/i), { target: { value: "counselor_admin" } });
  fireEvent.click(screen.getByRole("button", { name: /create counselor invitation/i }));
  expect(onInviteCounselor).toHaveBeenCalledWith({ email: "new@example.com", role: "counselor_admin" });
});

test("preserves the organization name when creation fails", async () => {
  const onCreate = vi.fn().mockRejectedValue(new Error("Could not create organization"));
  render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={onCreate} onDelete={vi.fn().mockResolvedValue(undefined)} onLogout={vi.fn()} />);

  const input = screen.getByLabelText(/organization name/i) as HTMLInputElement;
  fireEvent.change(input, { target: { value: "Acme" } });
  fireEvent.click(screen.getByRole("button", { name: /create organization/i }));

  await waitFor(() => expect(onCreate).toHaveBeenCalledWith({ name: "Acme" }));
  expect(input.value).toBe("Acme");
});
