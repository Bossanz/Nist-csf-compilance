import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { OrganizationDashboard } from "./OrganizationDashboard";
import type { Organization, User } from "../lib/types";

const admin:User={id:"user-1",organizationID:null,name:"Admin",email:"admin@example.com",userType:"counselor",role:"counselor_admin",status:"active"};
const organization:Organization={id:"org-1",name:"Acme",type:"client"};

test("selects an existing organization",()=>{const onSelect=vi.fn();render(<OrganizationDashboard user={admin} organizations={[organization]} loading={false} error="" onSelect={onSelect} onCreate={vi.fn()} onLogout={vi.fn()}/>);fireEvent.click(screen.getByRole("button",{name:/open acme/i}));expect(onSelect).toHaveBeenCalledWith(organization)});
test("counselor admin creates a trimmed organization",()=>{const onCreate=vi.fn();render(<OrganizationDashboard user={admin} organizations={[]} loading={false} error="" onSelect={vi.fn()} onCreate={onCreate} onLogout={vi.fn()}/>);fireEvent.change(screen.getByLabelText(/organization name/i),{target:{value:"  Acme  "}});fireEvent.click(screen.getByRole("button",{name:/create organization/i}));expect(onCreate).toHaveBeenCalledWith({name:"Acme"})});
