import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { ProjectFinalizationPanel } from "./ProjectFinalizationPanel";

afterEach(() => vi.restoreAllMocks());

test("blocks finalization until every included outcome is approved", () => {
  render(<ProjectFinalizationPanel status="in_review" includedCount={2} approvedCount={1} remaining={[{ code: "GV.OC-02", reason: "Reviewing" }]} onFinalize={vi.fn()} onOpenReport={vi.fn()} onOpenAudit={vi.fn()} />);

  expect(screen.getByText(/1 outcome remains/i)).toBeTruthy();
  expect((screen.getByRole("button", { name: /finalize project/i }) as HTMLButtonElement).disabled).toBe(true);
  expect(screen.getByText("GV.OC-02")).toBeTruthy();
});

test("asks for confirmation and finalizes when all outcomes are approved", async () => {
  vi.spyOn(window, "confirm").mockReturnValue(true);
  const onFinalize = vi.fn().mockResolvedValue(undefined);
  render(<ProjectFinalizationPanel status="in_review" includedCount={2} approvedCount={2} remaining={[]} onFinalize={onFinalize} onOpenReport={vi.fn()} onOpenAudit={vi.fn()} />);

  fireEvent.click(screen.getByRole("button", { name: /finalize project/i }));

  await waitFor(() => expect(onFinalize).toHaveBeenCalledOnce());
});

test("explains that remediation does not block assessment finalization", () => {
  render(<ProjectFinalizationPanel status="in_review" includedCount={1} approvedCount={1} remaining={[]} onFinalize={vi.fn()} onOpenReport={vi.fn()} onOpenAudit={vi.fn()} />);

  expect(screen.getByText(/Action Plan work is separate and can continue after finalization/i)).toBeTruthy();
});

test("shows read-only links after the project is finalized", () => {
  const onOpenReport = vi.fn();
  const onOpenAudit = vi.fn();
  render(<ProjectFinalizationPanel status="closed" includedCount={2} approvedCount={2} remaining={[]} onFinalize={vi.fn()} onOpenReport={onOpenReport} onOpenAudit={onOpenAudit} />);

  expect(screen.getByText(/project is finalized/i)).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: /open final report/i }));
  fireEvent.click(screen.getByRole("button", { name: /open audit package/i }));
  expect(onOpenReport).toHaveBeenCalled();
  expect(onOpenAudit).toHaveBeenCalled();
});
