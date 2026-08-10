import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vitest";
import { ProjectDashboard } from "./ProjectDashboard";
import type { Project } from "../lib/types";

const project: Project = {
  id: "project-1",
  organizationID: "org-1",
  organizationName: "Acme",
  name: "Readiness Review",
  slug: "readiness-review",
  status: "setup",
  createdAt: "2026-08-06T03:00:00Z",
};

describe("ProjectDashboard", () => {
  afterEach(() => vi.restoreAllMocks());

  test("renders persisted projects and opens the selected project", () => {
    const onOpen = vi.fn();
    render(<ProjectDashboard projects={[project]} loading={false} openingID="" error="" onOpen={onOpen} onCreate={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByRole("heading", { name: /existing projects/i })).toBeTruthy();
    expect(screen.getByText("Readiness Review")).toBeTruthy();
    expect(screen.getByText("Acme")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: /open readiness review/i }));

    expect(onOpen).toHaveBeenCalledWith(project);
  });

  test("renders a readable empty state", () => {
    render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByRole("heading", { name: /create project/i })).toBeTruthy();
    expect(screen.getByText(/no projects yet/i)).toBeTruthy();
  });

  test("submits trimmed project details", () => {
    const onCreate = vi.fn();
    render(<ProjectDashboard projects={[]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={onCreate} onDelete={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/project name/i), { target: { value: "  Readiness Review  " } });
    fireEvent.change(screen.getByLabelText(/organization name/i), { target: { value: "  Acme  " } });
    fireEvent.click(screen.getByRole("button", { name: /create project/i }));

    expect(onCreate).toHaveBeenCalledWith({ name: "Readiness Review", organizationName: "Acme" });
  });

  test("confirms before deleting a project", () => {
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);
    const onDelete = vi.fn();
    render(<ProjectDashboard projects={[project]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={vi.fn()} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole("button", { name: /delete readiness review/i }));

    expect(confirm).toHaveBeenCalledWith(expect.stringContaining("Readiness Review"));
    expect(onDelete).toHaveBeenCalledWith(project);
  });

  test("keeps the project when deletion is cancelled", () => {
    vi.spyOn(window, "confirm").mockReturnValue(false);
    const onDelete = vi.fn();
    render(<ProjectDashboard projects={[project]} loading={false} openingID="" error="" onOpen={vi.fn()} onCreate={vi.fn()} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole("button", { name: /delete readiness review/i }));

    expect(onDelete).not.toHaveBeenCalled();
  });

  test("disables project actions while an operation is active", () => {
    render(<ProjectDashboard projects={[project]} loading={false} openingID="project-1" error="" onOpen={vi.fn()} onCreate={vi.fn()} onDelete={vi.fn()} />);

    expect((screen.getByRole("button", { name: /open readiness review/i }) as HTMLButtonElement).disabled).toBe(true);
    expect((screen.getByRole("button", { name: /delete readiness review/i }) as HTMLButtonElement).disabled).toBe(true);
  });
});
