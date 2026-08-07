import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { StakeholderResponsePanel } from "./StakeholderResponsePanel";
import type { StakeholderResponse } from "../lib/types";

const response: StakeholderResponse = {
  id: "response-1",
  projectID: "project-1",
  subcategoryID: "subcategory-1",
  responseText: "We review access quarterly.",
  status: "draft",
  respondedBy: "assessor-1",
  submittedAt: null,
  reviewComment: "",
  reviewedBy: null,
  reviewedAt: null,
  createdAt: "2026-08-07T00:00:00Z",
  updatedAt: "2026-08-07T00:00:00Z",
  documents: [],
};

const noop = vi.fn().mockResolvedValue(undefined);

test("assessor can save, submit, and upload evidence", async () => {
  const onSave = vi.fn().mockResolvedValue(undefined);
  const onSubmit = vi.fn().mockResolvedValue(undefined);
  const onUpload = vi.fn().mockResolvedValue(undefined);
  render(<StakeholderResponsePanel role="assessor" response={response} onSave={onSave} onSubmit={onSubmit} onReview={noop} onUpload={onUpload} onDelete={noop} onDownload={noop} />);

  fireEvent.change(screen.getByLabelText(/client response/i), { target: { value: "Updated answer" } });
  fireEvent.click(screen.getByRole("button", { name: /save response/i }));
  await waitFor(() => expect(onSave).toHaveBeenCalledWith("Updated answer"));
  fireEvent.click(screen.getByRole("button", { name: /submit response/i }));
  await waitFor(() => expect(onSubmit).toHaveBeenCalled());

  const file = new File(["%PDF-1.7"], "evidence.pdf", { type: "application/pdf" });
  fireEvent.change(screen.getByLabelText(/upload evidence/i), { target: { files: [file] } });
  await waitFor(() => expect(onUpload).toHaveBeenCalledWith(file));
});

test("reviewer can mark a submitted response as needing more information", async () => {
  const onReview = vi.fn().mockResolvedValue(undefined);
  render(<StakeholderResponsePanel role="reviewer" response={{ ...response, status: "submitted" }} onSave={noop} onSubmit={noop} onReview={onReview} onUpload={noop} onDelete={noop} onDownload={noop} />);

  fireEvent.change(screen.getByLabelText(/review status/i), { target: { value: "needs_more_info" } });
  fireEvent.change(screen.getByLabelText(/review comment/i), { target: { value: "Please add the access review record." } });
  fireEvent.click(screen.getByRole("button", { name: /save review/i }));

  await waitFor(() => expect(onReview).toHaveBeenCalledWith("needs_more_info", "Please add the access review record."));
});

test("viewer sees the response but no mutation controls", () => {
  render(<StakeholderResponsePanel role="viewer" response={response} onSave={noop} onSubmit={noop} onReview={noop} onUpload={noop} onDelete={noop} onDownload={noop} />);

  expect((screen.getByDisplayValue(response.responseText) as HTMLTextAreaElement).disabled).toBe(true);
  expect(screen.queryByRole("button", { name: /save response/i })).toBeNull();
  expect(screen.queryByRole("button", { name: /submit response/i })).toBeNull();
  expect(screen.queryByLabelText(/upload evidence/i)).toBeNull();
});
