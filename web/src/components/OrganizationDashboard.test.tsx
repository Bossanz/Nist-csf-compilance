import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { OrganizationDashboard } from "./OrganizationDashboard";
import type { Organization, User } from "../lib/types";

const admin:User={id:"user-1",organizationID:null,name:"Admin",email:"admin@example.com",userType:"counselor",role:"counselor_admin",status:"active"};
const counselor:User={...admin,id:"user-2",role:"counselor"};
const organization:Organization={id:"org-1",name:"Acme",type:"client"};

test("selects an existing organization",()=>{const onSelect=vi.fn();render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={onSelect} onCreate={vi.fn()} onDelete={vi.fn()} onLogout={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:/open acme/i}));expect(onSelect).toHaveBeenCalledWith(organization)});
test("counselor admin creates a trimmed organization",()=>{const onCreate=vi.fn();render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={onCreate} onDelete={vi.fn()} onLogout={vi.fn()}/>);fireEvent.change(screen.getByLabelText(/organization name/i),{target:{value:"  Acme  "}});fireEvent.click(screen.getByRole("button",{name:/create organization/i}));expect(onCreate).toHaveBeenCalledWith({name:"Acme"})});

test("counselor admin confirms permanent deletion with the exact organization name",()=>{const onDelete=vi.fn();render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={onDelete} onLogout={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:"Delete Acme"}));const finalButton=screen.getByRole("button",{name:/delete permanently/i}) as HTMLButtonElement;expect(finalButton.disabled).toBe(true);fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"acme"}});expect(finalButton.disabled).toBe(true);fireEvent.change(screen.getByLabelText(/type acme to confirm/i),{target:{value:"Acme"}});fireEvent.click(finalButton);expect(onDelete).toHaveBeenCalledWith(organization)});

test("regular counselor cannot see organization deletion",()=>{render(<OrganizationDashboard user={counselor} organizations={[organization]} loading={false} error="" onSelect={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn()} onLogout={vi.fn()}/>);expect(screen.queryByRole("button",{name:"Delete Acme"})).toBeNull()});
