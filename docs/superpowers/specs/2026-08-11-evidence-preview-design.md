# Evidence Preview Design

**Status:** Approved for implementation

## Goal

Let authorized users read common evidence files in the assessment page without first downloading them to disk.

## Scope

- PDF files are previewed in an inline viewer.
- PNG, JPG, and JPEG files are previewed inline.
- DOCX and XLSX remain download-only in v1 because reliable browser rendering would add a conversion service or a heavy client dependency.
- The existing authenticated document endpoint and role checks remain the source of truth.
- Download remains available for every file.

## User experience

Each evidence row keeps the filename, size, and delete action where permitted. Supported files show a `Preview` action. Clicking it loads the file through the existing authenticated API and opens a preview panel under the evidence list. The panel includes a loading state, an error state, and a close action. Unsupported files continue to use the existing filename download action.

## Technical approach

The page fetches the document as a `Blob` using the existing API client, creates a temporary object URL, and passes the preview URL and MIME type to the response panel. The panel renders PDFs with an `iframe` and images with an `img`. The page revokes the object URL when the preview changes or closes, preventing stale browser memory. No database or storage schema change is required.

## Access and failure behavior

Preview requests use the same project, subcategory, document identifiers and session cookie as downloads, so Counselor, assigned Stakeholder, Reviewer, and Viewer permissions are unchanged. A failed request displays an inline error and does not remove the evidence row. Download remains the fallback for unsupported MIME types.

## Verification

- Component tests cover preview action visibility, loading the supported document, rendering the inline preview, closing it, and retaining download behavior for unsupported files.
- Existing document handler and full web test suites must remain green.
- Production build and Docker health checks must pass.
