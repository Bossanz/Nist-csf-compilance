# Password Reset and Email Notifications

## Context

The application has email/password login, invitation-based account activation, role-based access, and the complete assessment workflow. Accounts can currently be invited and activated, but an active user cannot recover or change a password. Invitations and assessment events are visible in the UI and audit log, but there is no email delivery path for users who are not watching the workspace.

This feature adds the smallest useful v1 identity recovery and email notification layer without introducing production deployment work, a message queue, or an in-app notification center.

## Goals

1. Let an active user request a password reset without revealing whether an email exists.
2. Let a signed-in user change their password after providing the current password.
3. Send invitation and important assessment workflow emails through one replaceable email-sender interface.
4. Keep local development usable without SMTP by logging the generated links and messages.
5. Revoke existing sessions after a successful password reset or password change.
6. Keep current invitation and assessment mutations successful when a notification delivery attempt fails; delivery failure must be visible in server logs and must not lose the assessment data.

## Non-goals

- Production deployment, HTTPS, secret management, or database backup.
- MFA, SSO, email verification, password history, or account self-service deletion.
- A notification inbox, read/unread state, queue, retry worker, or template management system.
- HTML email design; v1 sends readable plain-text messages.
- Changing the existing invitation expiry or assessment workflow semantics.

## Workflow

```text
Forgot password
  -> user enters email
  -> server always returns the same acknowledgement
  -> active account receives/logs a one-time reset link
  -> user sets a new password
  -> all existing sessions are revoked
  -> user signs in again

Change password
  -> authenticated user enters current + new password
  -> server verifies the current password
  -> password hash is replaced
  -> all sessions are revoked
  -> user signs in again

Assessment events
  -> domain mutation succeeds
  -> notification recipient list is resolved
  -> EmailSender sends or logs a plain-text message
  -> delivery failures are logged without rolling back the domain mutation
```

## Password recovery design

### Data model

Add `db/init/009_password_reset_tokens.sql`:

```sql
CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user
  ON password_reset_tokens(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expiry
  ON password_reset_tokens(expires_at)
  WHERE used_at IS NULL;
```

Only the SHA-256 token hash is stored. Raw tokens exist only in the reset URL and the local log email. A new request invalidates previous unused tokens for that user. Tokens expire after 30 minutes and can be consumed once.

### API

`POST /api/auth/password-reset/request`

Request:

```json
{ "email": "person@example.com" }
```

Response is always `202` with the same message for an existing, unknown, invited, or disabled email:

```json
{ "message": "If an active account exists, a password reset link has been sent." }
```

This prevents account enumeration. Only active users receive a reset message.

`POST /api/auth/password-reset/confirm`

Request:

```json
{ "token": "raw-token", "password": "NewPassword!2026" }
```

The endpoint returns `204` on success, `400 invalid_reset_token` for an expired/used/unknown token, and the existing password validation error for a weak password. The token is consumed and every session for the user is deleted in the same transaction as the password update.

`PUT /api/auth/password`

Authenticated request:

```json
{ "currentPassword": "OldPassword!2026", "newPassword": "NewPassword!2026" }
```

The endpoint verifies the current password, replaces the hash, deletes all sessions for the user, and returns `204`. The browser sends the user back to `/login` because the current session is intentionally revoked.

### Password rules

Reuse the existing password policy and error (`ErrWeakPassword`) for invitation activation, reset, and change. Do not introduce a second password policy in this feature.

## Email delivery design

### Sender boundary

Add a small `EmailSender` interface in the HTTP/API layer:

```go
type EmailMessage struct {
    To      string
    Subject string
    Text    string
}

type EmailSender interface {
    Send(context.Context, EmailMessage) error
}
```

Implement two senders:

- `logEmailSender`: default for local development; writes recipient, subject, and plain-text body to API logs. Reset URLs are therefore testable without exposing tokens in a normal API response.
- `smtpEmailSender`: enabled by configuration and uses the standard library SMTP client. It sends the same plain-text message through the configured SMTP server.

Configuration:

```dotenv
EMAIL_MODE=log
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=no-reply@example.com
```

`APP_ORIGIN` remains the source for invitation and reset links. Missing SMTP configuration never prevents the API from starting when `EMAIL_MODE=log` is selected.

### Message types and recipients

Messages use stable subjects and plain-text content with the project or outcome code where relevant:

| Event | Recipient | Subject intent |
| --- | --- | --- |
| Invitation created | Invited email | Join the NIST CSF workspace |
| Password reset requested | Active account email | Reset your password |
| Response submitted | Active `reviewer` users in the project organization | Outcome ready for review |
| Response approved | Assigned assessor email | Outcome approved |
| Needs more information | Assigned assessor email | Outcome needs more information |
| Project finalized | Active `org_admin` and `reviewer` users in the organization | Project finalized |

The current counselor is not stored as a project member, so v1 does not guess a counselor recipient. The counselor already receives the final state in the workspace and Final Report.

### Failure behavior

Email delivery is best-effort after the domain write. An invitation, response submission/review, or finalization is not rolled back because SMTP is unavailable. The API logs the failure with event type and recipient count, and the existing invitation response continues to expose the one-time invitation URL as the local/manual fallback.

The email sender is injectable in tests. No raw password, password hash, reset token, or SMTP credential is written to audit metadata.

## Frontend behavior

Add these routes:

```text
/forgot-password
/reset-password?token={token}
/account/password
```

- Login adds a `Forgot password?` link.
- Forgot-password form shows one generic success message after submit, regardless of email existence.
- Reset-password form reads the token from the query string, validates the new password, and returns to Login after success.
- Account password form requires the current and new password and returns to Login after success.
- The Organizations identity block exposes `Change password` for authenticated users.
- Invitation creation keeps the current one-time URL display for local testing; SMTP mode also delivers it by email.
- Email delivery errors are shown as a non-blocking warning where the current mutation already succeeded.

## Authorization and privacy

- Password reset request is unauthenticated but deliberately non-enumerating.
- Password reset confirmation is authorized only by a valid one-time token.
- Password change requires the current authenticated session.
- Reset and change operations revoke sessions; invitation acceptance behavior stays unchanged.
- Notification recipients are resolved from active users in the same organization/project visibility boundary.
- Email bodies contain links and workflow context only; they do not contain response text, evidence content, passwords, or private storage keys.

## Testing strategy

### Store/auth tests

- Reset token is stored hashed, expires after 30 minutes, is invalidated by a newer request, and cannot be reused.
- Confirming a valid token updates the password and deletes all sessions atomically.
- Changing a password rejects the wrong current password and deletes all sessions on success.

### HTTP tests

- Reset request returns the same `202` response for known and unknown emails.
- Confirm maps invalid/expired/used tokens to `400 invalid_reset_token`.
- Password change requires authentication and maps the current-password failure clearly.
- Invitation, response submission/review, and finalization call the injected sender with the intended recipient and subject.
- Sender failure does not turn a successful domain mutation into a `500` response.

### Frontend tests

- Login renders the reset link.
- Forgot-password renders the generic success state.
- Reset-password submits the token and redirects to login.
- Change-password submits current/new credentials and redirects to login.
- Existing invitation, assessment, and finalization tests remain green.

## Rollout

1. Add the password reset migration and auth/store services.
2. Add the injectable email sender and local log mode.
3. Add password reset/change endpoints and UI routes.
4. Connect invitation and assessment events to notifications.
5. Update README with local log instructions and SMTP configuration.
6. Run Go tests, frontend tests, TypeScript, production build, and Docker health checks.
