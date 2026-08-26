import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { FunctionSidebar } from "./FunctionSidebar";

test("labels the function navigation landmark", () => {
  render(
    <FunctionSidebar
      functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]}
      selectedCode="GV"
      onSelect={() => undefined}
    />,
  );

  expect(screen.getByRole("navigation", { name: /csf functions/i })).toBeTruthy();
  expect(screen.getByRole("button", { name: /gv govern/i }).getAttribute("aria-current")).toBe("page");
});

test("shows role-relevant progress beside each Function", () => {
  render(
    <FunctionSidebar
      functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]}
      selectedCode="GV"
      onSelect={() => undefined}
      progressByFunction={{ GV: { value: 2, label: "submitted" } }}
    />,
  );

  expect(screen.getByRole("button", { name: /gv govern.*2 reviewing/i })).toBeTruthy();
});

test("shows Function coverage as a percentage when supplied", () => {
  render(
    <FunctionSidebar
      functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]}
      selectedCode="GV"
      onSelect={() => undefined}
      progressByFunction={{ GV: { value: 2, label: "included", coveragePct: 67, includedCount: 2 } }}
    />,
  );

  expect(screen.getByRole("button", { name: /gv govern.*67% 2 included/i })).toBeTruthy();
  expect(screen.getByText("67%")).toBeTruthy();
  expect(screen.getByText("2 included")).toBeTruthy();
});

test("exposes the active role mode alongside the Function index", () => {
  render(
    <FunctionSidebar
      functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]}
      selectedCode="GV"
      onSelect={() => undefined}
      mode="Scope & Assignment"
    />,
  );

  expect(screen.getByText("Scope & Assignment")).toBeTruthy();
});

test("organizes the workspace into overview, assignment, log, and action plan navigation", () => {
  const onSelectSurface = vi.fn();
  const onSelect = vi.fn();
  render(
    <FunctionSidebar
      functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]}
      selectedCode="GV"
      onSelect={onSelect}
      onSelectSurface={onSelectSurface}
      onBack={vi.fn()}
    />,
  );

  expect(screen.getByRole("button", { name: "Overview" })).toBeTruthy();
  expect(screen.getByRole("button", { name: "Assignment" })).toBeTruthy();
  expect(screen.getByRole("button", { name: "Log" })).toBeTruthy();
  expect(screen.getByRole("button", { name: "Action Plan" })).toBeTruthy();
  expect(screen.getByRole("button", { name: /gv govern/i })).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: "Assignment" }));
  expect(screen.queryByRole("button", { name: /gv govern/i })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Assignment" }));
  fireEvent.click(screen.getByRole("button", { name: /gv govern/i }));
  fireEvent.click(screen.getByRole("button", { name: "Overview" }));
  fireEvent.click(screen.getByRole("button", { name: "Log" }));
  fireEvent.click(screen.getByRole("button", { name: "Action Plan" }));

  expect(onSelect).toHaveBeenCalledWith("GV");
  expect(onSelectSurface).toHaveBeenCalledWith("assignment");
  expect(onSelectSurface).toHaveBeenCalledWith("overview");
  expect(onSelectSurface).toHaveBeenCalledWith("log");
  expect(onSelectSurface).toHaveBeenCalledWith("actions");
  expect(screen.getByRole("button", { name: /back to organization/i })).toBeTruthy();
});
