import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { AssignmentProgress } from "./AssignmentProgress";

test("shows included, assigned, and waiting counts", () => {
  render(<AssignmentProgress included={12} assigned={5} unassigned={7} />);

  const progress = screen.getByRole("region", { name: /assignment progress/i });
  expect(progress.textContent).toContain("12");
  expect(progress.textContent).toContain("5");
  expect(progress.textContent).toContain("7");
  expect(progress.textContent).toContain("Waiting for assignment");
});
