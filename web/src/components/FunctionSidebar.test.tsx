import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
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

  expect(screen.getByRole("button", { name: /gv govern.*2 submitted/i })).toBeTruthy();
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
