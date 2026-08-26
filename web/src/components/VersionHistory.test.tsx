import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import type { Project } from "../lib/types";
import { VersionHistory } from "./VersionHistory";

const project: Project = {
  id: "project-2",
  organizationID: "org-1",
  organizationName: "Acme",
  name: "Readiness",
  slug: "readiness-v2",
  status: "setup",
  createdAt: "2026-08-20T00:00:00Z",
  versionNumber: 2,
  isLatest: true,
};

const previousProject: Project = {
  ...project,
  id: "project-1",
  slug: "readiness",
  status: "closed",
  versionNumber: 1,
  isLatest: false,
};

afterEach(() => vi.restoreAllMocks());

test("shows version history and opens a selected version", () => {
  const onOpenVersion = vi.fn();
  render(<VersionHistory currentProject={project} versions={[project, previousProject]} canCreate={false} onCreateVersion={vi.fn()} onOpenVersion={onOpenVersion} />);

  expect(screen.getByRole("heading", { name: /version history/i })).toBeTruthy();
  expect(screen.getByText(/assessment v2/i)).toBeTruthy();
  expect(screen.getByText(/assessment v1/i)).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /open assessment v1/i }));
  expect(onOpenVersion).toHaveBeenCalledWith(previousProject);
});

test("asks for confirmation before creating the next version", async () => {
  const onCreateVersion = vi.fn().mockResolvedValue(undefined);
  render(<VersionHistory currentProject={previousProject} versions={[previousProject]} canCreate onCreateVersion={onCreateVersion} onOpenVersion={vi.fn()} />);

  fireEvent.click(screen.getByRole("button", { name: /start new assessment/i }));
  expect(screen.getByRole("dialog", { name: /start a new assessment/i })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: /confirm start/i }));

  await waitFor(() => expect(onCreateVersion).toHaveBeenCalledOnce());
});

test("does not offer a new version action when versioning is unavailable", () => {
  render(<VersionHistory currentProject={project} versions={[project]} canCreate={false} onCreateVersion={vi.fn()} onOpenVersion={vi.fn()} />);

  expect(screen.queryByRole("button", { name: /start new assessment/i })).toBeNull();
});
