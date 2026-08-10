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
