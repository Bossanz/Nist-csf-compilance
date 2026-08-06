import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vitest";
import { ProjectDashboard } from "./ProjectDashboard";
import type { Project } from "../lib/types";

const project: Project = {
  id: "project-1",
  organizationID: "org-1",
  organizationName: "Acme",
  name: "Readiness Review",
  status: "setup",
  createdAt: "2026-08-06T03:00:00Z",
};

describe("ProjectDashboard", () => {
  test("renders persisted projects and opens the selected project", () => {
    const onOpen = vi.fn();
    render(<ProjectDashboard projects={[project]} loading={false} openingID="" error="" onOpen={onOpen} onCreate={vi.fn()} />);

    expect(screen.getByText("Readiness Review")).toBeTruthy();
    expect(screen.getByText("Acme")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: /open readiness review/i }));

    expect(onOpen).toHaveBeenCalledWith(project);
  });

  test("renders a readable empty state", () => {
    render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={vi.fn()} />);

    expect(screen.getByText(/no projects yet/i)).toBeTruthy();
  });

  test("submits trimmed project details", () => {
    const onCreate = vi.fn();
    render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={onCreate} />);

    fireEvent.change(screen.getByLabelText(/project name/i), { target: { value: "  Readiness Review  " } });
    fireEvent.change(screen.getByLabelText(/organization name/i), { target: { value: "  Acme  " } });
    fireEvent.click(screen.getByRole("button", { name: /create project/i }));

    expect(onCreate).toHaveBeenCalledWith({ name: "Readiness Review", organizationName: "Acme" });
  });
});
