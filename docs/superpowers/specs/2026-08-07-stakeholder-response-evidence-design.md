# Stakeholder Response and Evidence Design

## Goal

Separate client-supplied information from Counselor assessment. Stakeholders answer each existing NIST CSF Subcategory and attach supporting documents; Reviewers check those submissions; Counselors alone decide scope, Priority, Coverage, and target assessment.

## Scope

Version 1 uses each existing NIST CSF Subcategory as the question. It does not add a custom question bank, evidence requests, assignments, deadlines, notifications, document versioning, antivirus infrastructure, or cloud object storage.

## Roles and permissions

### Counselor Admin and Counselor

- Access every client organization and project.
- Read Stakeholder responses, review comments, and documents.
- Edit the existing Counselor assessment fields: included scope, rationale, current and target Priority, current and target Coverage, current activities, policies, target approach, notes, and considerations.
- Do not submit or review a response on behalf of a Stakeholder in Version 1.

### Organization Admin and Assessor

- Access only their own organization.
- Create and edit the textual response for a NIST Subcategory.
- Upload and delete documents belonging to that response.
- Save a response as draft and submit it for review.
- A submitted response becomes editable again only after a Reviewer marks it `needs_more_info`.

### Reviewer

- Access only their own organization.
- Read responses and download documents.
- Mark a submitted response as `reviewed` or `needs_more_info` and add a review comment.
- Cannot edit Stakeholder response text, documents, Priority, or Coverage.

### Viewer

- Access only their own organization.
- Read responses, review state, Counselor assessment, and documents.
- Cannot mutate any assessment or response data.

The backend enforces these permissions independently of the frontend. Existing profile update authorization changes so Stakeholder roles, including `assessor`, cannot update Counselor assessment fields.

## Workflow

1. Counselor creates a Project under a client Organization.
2. Organization Admin or Assessor opens a NIST outcome, writes the client response, and optionally uploads documents.
3. The Stakeholder saves drafts until ready, then submits the response.
4. Reviewer checks the response and documents.
5. Reviewer chooses `Reviewed` or `Need more info` and writes an optional comment.
6. `Need more info` reopens the response for Organization Admin and Assessor editing; `Reviewed` keeps it locked.
7. Counselor reads the client material and independently sets scope, Priority, Coverage, and target assessment.

## Data model

### `stakeholder_responses`

- `id uuid primary key`
- `project_id uuid not null references projects(id) on delete cascade`
- `subcategory_id uuid not null references subcategories(id)`
- `response_text text not null default ''`
- `status text not null default 'draft'` constrained to `draft`, `submitted`, `reviewed`, `needs_more_info`
- `responded_by uuid references users(id)`
- `submitted_at timestamptz`
- `review_comment text not null default ''`
- `reviewed_by uuid references users(id)`
- `reviewed_at timestamptz`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`
- Unique `(project_id, subcategory_id)`

### `response_documents`

- `id uuid primary key`
- `response_id uuid not null references stakeholder_responses(id) on delete cascade`
- `original_name text not null`
- `storage_key text not null unique`
- `mime_type text not null`
- `size_bytes bigint not null`
- `uploaded_by uuid not null references users(id)`
- `created_at timestamptz not null default now()`

Document binaries live below an application-owned directory mounted as a Docker named volume. The database stores metadata only.

Project and Organization deletion first enumerate their response document storage keys, remove the database records transactionally, and then remove the binaries. Missing binaries are ignored. A failed binary cleanup is logged with the storage key so it can be retried operationally; it does not restore already deleted business data.

## API

All project routes first verify the current user can access the Project's Organization.

- `GET /api/projects/{projectID}/responses` returns all response summaries for the Project.
- `PUT /api/projects/{projectID}/responses/{subcategoryID}` saves `responseText` as draft. Allowed for Organization Admin and Assessor only while status is `draft` or `needs_more_info`.
- `POST /api/projects/{projectID}/responses/{subcategoryID}/submit` changes draft or needs-more-info to submitted.
- `POST /api/projects/{projectID}/responses/{subcategoryID}/review` accepts `status: reviewed | needs_more_info` and `comment`. Allowed for Reviewer only while submitted.
- `POST /api/projects/{projectID}/responses/{subcategoryID}/documents` uploads one multipart file.
- `GET /api/projects/{projectID}/responses/{subcategoryID}/documents/{documentID}` downloads an authorized document.
- `DELETE /api/projects/{projectID}/responses/{subcategoryID}/documents/{documentID}` deletes metadata and binary. Allowed for Organization Admin or Assessor while response editing is open.

Invalid lifecycle transitions return `409 Conflict`. Unauthorized roles return `403 Forbidden`. Missing projects, outcomes, responses, or documents return `404 Not Found`.

## File handling

- Maximum file size: 20 MB.
- Allowed formats: PDF, DOCX, XLSX, PNG, and JPEG.
- Validate extension and detected MIME type; never trust the client-provided filename or MIME type alone.
- Generate a random opaque storage key and never use the original filename as a path.
- Normalize the original name for display and reject empty or control-character names.
- Write uploads to a temporary file in the evidence volume, then rename atomically after validation and metadata creation.
- If metadata creation fails, remove the temporary file.
- Download responses use `Content-Disposition: attachment` with a safely encoded original filename and `X-Content-Type-Options: nosniff`.
- Deletion removes the database row and binary; a missing binary is treated as already removed so metadata can still be cleaned up.

This local-volume design is intended for local and single-server deployment. Moving to S3-compatible storage remains a future adapter boundary.

## Frontend

Each NIST outcome card contains three clearly separated regions:

1. **Client response and documents** — response text, draft/submit actions, document list, upload, and delete controls for Organization Admin or Assessor; read-only for other roles.
2. **Review** — current status and comment; review actions only for Reviewer when submitted.
3. **Counselor assessment** — the existing scope, Current Profile, and Target Profile form; editable only for Counselor Admin and Counselor.

The page loads profile rows and response summaries in parallel. Upload, response, review, and assessment errors are scoped to the affected outcome so one failure does not discard text in another form.

## Docker and configuration

- Add a named volume mounted into the API container at `/data/evidence`.
- Configure `EVIDENCE_DIR=/data/evidence` with the same path as the development default.
- The API creates the directory on startup when it does not exist and fails startup with a clear error when it cannot write there.
- `docker compose down` preserves evidence; `docker compose down -v` removes PostgreSQL and evidence volumes.

## Testing

- Domain tests cover valid and invalid response lifecycle transitions.
- HTTP authorization tests cover every role for response save, submit, review, upload, download, and delete.
- Store integration tests cover response persistence and organization/project isolation.
- File service tests use a temporary directory and cover size limit, allowed formats, random storage keys, cleanup after failure, download, and deletion.
- Frontend component tests cover role-specific controls, draft/submit behavior, review actions, and error preservation.
- End-to-end browser verification covers Stakeholder response submission, Reviewer review, and Counselor-only Priority/Coverage editing.

## Migration and compatibility

Add an idempotent `004_stakeholder_responses.sql` migration. Existing Projects and assessment profiles remain valid; responses are created lazily when a Stakeholder first saves or uploads a document. Existing Counselor assessment data is unchanged.
