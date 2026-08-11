import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { StakeholderResponsePanel } from "./StakeholderResponsePanel";
import type { ResponseDocument, StakeholderResponse } from "../lib/types";

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
const pdfDocument: ResponseDocument = { id: "doc-pdf", responseID: "response-1", originalName: "evidence.pdf", mimeType: "application/pdf", sizeBytes: 1024, uploadedBy: "assessor-1", createdAt: "2026-08-07T00:00:00Z" };
const imageDocument: ResponseDocument = { id: "doc-image", responseID: "response-1", originalName: "diagram.png", mimeType: "image/png", sizeBytes: 2048, uploadedBy: "assessor-1", createdAt: "2026-08-07T00:00:00Z" };

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

test("Reviewer final decision can complete a submitted outcome", async () => {
  const onReview = vi.fn().mockResolvedValue(undefined);
  render(<StakeholderResponsePanel role="reviewer" response={{ ...response, status: "submitted" }} onSave={noop} onSubmit={noop} onReview={onReview} onUpload={noop} onDelete={noop} onDownload={noop} />);

  expect(screen.getByRole("heading", { name: /reviewer final decision/i })).toBeTruthy();
  fireEvent.change(screen.getByLabelText(/review status/i), { target: { value: "reviewed" } });
  fireEvent.change(screen.getByLabelText(/review comment/i), { target: { value: "Evidence is sufficient." } });
  fireEvent.click(screen.getByRole("button", { name: /save review/i }));

  await waitFor(() => expect(onReview).toHaveBeenCalledWith("reviewed", "Evidence is sufficient."));
});

test("viewer sees the response but no mutation controls", () => {
  render(<StakeholderResponsePanel role="viewer" response={response} onSave={noop} onSubmit={noop} onReview={noop} onUpload={noop} onDelete={noop} onDownload={noop} />);

  expect((screen.getByDisplayValue(response.responseText) as HTMLTextAreaElement).disabled).toBe(true);
  expect(screen.queryByRole("button", { name: /save response/i })).toBeNull();
  expect(screen.queryByRole("button", { name: /submit response/i })).toBeNull();
  expect(screen.queryByLabelText(/upload evidence/i)).toBeNull();
});

test("previews supported evidence without using the download action", async () => {
  const onPreview = vi.fn().mockResolvedValue(undefined);
  const onDownload = vi.fn().mockResolvedValue(undefined);
  render(<StakeholderResponsePanel role="viewer" response={{ ...response, documents: [pdfDocument] }} onSave={noop} onSubmit={noop} onReview={noop} onUpload={noop} onDelete={noop} onDownload={onDownload} onPreview={onPreview} onClosePreview={noop} />);

  fireEvent.click(screen.getByRole("button", { name: /preview evidence\.pdf/i }));
  await waitFor(() => expect(onPreview).toHaveBeenCalledWith(pdfDocument));
  expect(onDownload).not.toHaveBeenCalled();
});

test("renders and closes an inline image preview", () => {
  const onClosePreview = vi.fn();
  render(<StakeholderResponsePanel role="viewer" response={{ ...response, documents: [imageDocument] }} onSave={noop} onSubmit={noop} onReview={noop} onUpload={noop} onDelete={noop} onDownload={noop} onPreview={noop} preview={{ subcategoryID: response.subcategoryID, documentID: imageDocument.id, url: "blob:image-preview", mimeType: imageDocument.mimeType }} onClosePreview={onClosePreview} />);

  expect(screen.getByRole("img", { name: /diagram\.png preview/i }).getAttribute("src")).toBe("blob:image-preview");
  fireEvent.click(screen.getByRole("button", { name: /close preview/i }));
  expect(onClosePreview).toHaveBeenCalled();
});

test("keeps unsupported evidence as download-only", () => {
  const document: ResponseDocument = { ...pdfDocument, id: "doc-word", originalName: "policy.docx", mimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document" };
  const onDownload = vi.fn().mockResolvedValue(undefined);
  render(<StakeholderResponsePanel role="viewer" response={{ ...response, documents: [document] }} onSave={noop} onSubmit={noop} onReview={noop} onUpload={noop} onDelete={noop} onDownload={onDownload} onPreview={noop} onClosePreview={noop} />);

  expect(screen.queryByRole("button", { name: /preview policy\.docx/i })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: /policy\.docx/i }));
  expect(onDownload).toHaveBeenCalledWith(document);
});
