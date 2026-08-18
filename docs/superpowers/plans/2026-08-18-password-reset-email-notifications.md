# Password Reset and Email Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Add v1 password recovery/change and plain-text invitation/workflow email notifications while keeping local Docker development usable without SMTP.

**Architecture:** Add a hashed, single-use password-reset token table and a focused auth service for request/confirm/change operations. Add an injectable email sender with log and SMTP implementations; domain mutations remain successful when delivery fails. Reuse the existing handlers, audit events, session table, invitation flow, and assessment event points instead of adding a queue or notification center.

**Tech Stack:** Go 1.24, net/http, PostgreSQL 16, pgx, existing bcrypt/token helpers, Go standard-library SMTP, Next.js 16 App Router, React 19, TypeScript, Vitest, Docker Compose.

**Spec:** docs/superpowers/specs/2026-08-18-password-reset-email-notifications-design.md

## Global Constraints

- Password reset tokens are stored only as SHA-256 hashes, expire after 30 minutes, and are single-use.
- Password-reset request always returns the same 202 response for known and unknown emails.
- Password reset and password change revoke every session for the user.
- EMAIL_MODE=log is the default local mode; SMTP is optional and configured through environment variables.
- Invitation, response, review, and finalization mutations are not rolled back when email delivery fails.
- Email bodies never contain passwords, response text, evidence content, private storage keys, or SMTP credentials.
- Do not add a notification table, queue, retry worker, notification inbox, HTML template engine, MFA, SSO, or production deployment work.
- Do not stage output/ or tmp/ artifacts.

---

### Task 1: Add password-reset token persistence and session revocation

**Files:**
- Create: db/init/009_password_reset_tokens.sql
- Modify: api/internal/store/models.go
- Create: api/internal/store/password_resets.go
- Create: api/internal/store/password_resets_test.go

**Interfaces:**
- Produces store.ErrInvalidPasswordResetToken and store.ErrInvalidCurrentPassword.
- Produces Store.FindActiveUserByEmail(ctx, email) (User, error).
- Produces Store.CreatePasswordResetToken(ctx, userID, tokenHash string, expiresAt time.Time) error.
- Produces Store.ConsumePasswordResetToken(ctx, tokenHash, passwordHash string, now time.Time) (User, error).
- Produces Store.ChangePassword(ctx, userID, passwordHash string) error.
- Produces Store.RevokeUserSessions(ctx, userID string) error.

- [ ] **Step 1: Write the failing store integration tests**

Add tests guarded by TEST_DATABASE_URL for these behaviors:

~~~
func TestPasswordResetTokenIsSingleUseAndRevokesSessions(t *testing.T) {
    // Create an active user, two sessions, and one reset token.
    // Consume the token with a new bcrypt hash.
    // Assert the user password changed, both sessions were deleted, and a second consume returns ErrInvalidPasswordResetToken.
}

func TestPasswordResetTokenExpires(t *testing.T) {
    // Create an expired token and assert ConsumePasswordResetToken returns ErrInvalidPasswordResetToken.
}

func TestChangePasswordRevokesAllSessions(t *testing.T) {
    // Create an active user with the old hash and two sessions.
    // ChangePassword must update the hash and leave zero sessions.
}
~~~

- [ ] **Step 2: Run the focused tests and confirm the expected red state**

Run from C:/Acuitmesh/NIST-CSF-Compliance/api:

~~~
go test ./internal/store -run PasswordReset -count=1
~~~

Expected result: the new store methods/types are missing or the integration tests skip when TEST_DATABASE_URL is not set. The test must compile after the test helpers are complete; a skipped database test is acceptable in the local environment.

- [ ] **Step 3: Add the migration**

Create db/init/009_password_reset_tokens.sql with:

~~~
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
~~~

- [ ] **Step 4: Implement token creation and atomic consumption**

Implement CreatePasswordResetToken by deleting existing unused tokens for the user and inserting the new hash. Implement ConsumePasswordResetToken in one transaction:

1. SELECT the token row with FOR UPDATE.
2. Reject missing, used, or expired rows with ErrInvalidPasswordResetToken.
3. Update the user password hash and mark used_at.
4. Delete all sessions for that user.
5. Commit and return the updated user without exposing PasswordHash in JSON.

Implement ChangePassword as one transaction that updates the hash and deletes all sessions for the user. Return pgx.ErrNoRows when the user does not exist.

- [ ] **Step 5: Run store tests and format the Go files**

~~~
gofmt -w internal/store/models.go internal/store/password_resets.go internal/store/password_resets_test.go
go test ./internal/store -run PasswordReset -count=1
go test ./internal/store -count=1
~~~

Expected result: focused tests pass when a test database is configured; the full store package remains green.

- [ ] **Step 6: Commit the persistence slice**

~~~
git add db/init/009_password_reset_tokens.sql api/internal/store/models.go api/internal/store/password_resets.go api/internal/store/password_resets_test.go
git commit -m "feat: add password reset token storage"
~~~

---

### Task 2: Add password-reset and password-change auth services

**Files:**
- Create: api/internal/auth/password_reset.go
- Create: api/internal/auth/password_reset_test.go

**Interfaces:**
- Consumes the store methods from Task 1.
- Produces PasswordService.Request(ctx, email) (user store.User, rawToken string, found bool, err error).
- Produces PasswordService.Confirm(ctx, rawToken, newPassword string) error.
- Produces PasswordService.Change(ctx, userID, currentPassword, newPassword string) error.

- [ ] **Step 1: Write failing service tests**

Use an in-memory fake repository that records token hashes, password updates, and session revocations. Add one behavior per test:

~~~
func TestRequestPasswordResetReturnsNoUserForUnknownEmail(t *testing.T) {}
func TestRequestPasswordResetCreatesHashedTokenForActiveUser(t *testing.T) {}
func TestConfirmPasswordResetRejectsWeakPassword(t *testing.T) {}
func TestConfirmPasswordResetPassesOnlyTokenHashToRepository(t *testing.T) {}
func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {}
func TestChangePasswordUpdatesHashForCorrectCurrentPassword(t *testing.T) {}
~~~

- [ ] **Step 2: Run the service tests and confirm red**

~~~
go test ./internal/auth -run "PasswordReset|ChangePassword" -count=1
~~~

Expected result: the service and methods do not exist yet.

- [ ] **Step 3: Implement the service minimally**

Use the existing NewToken, HashToken, HashPassword, CheckPassword, and ErrWeakPassword helpers. Request must normalize the email, return found=false for missing/disabled users without an error, and return a raw token only for an active user. Confirm validates the password before touching the database and passes only HashToken(rawToken) to the repository. Change validates the current hash and new password before calling the repository.

- [ ] **Step 4: Run service and auth tests**

~~~
gofmt -w internal/auth/password_reset.go internal/auth/password_reset_test.go
go test ./internal/auth -count=1
~~~

- [ ] **Step 5: Commit the auth service slice**

~~~
git add api/internal/auth/password_reset.go api/internal/auth/password_reset_test.go
git commit -m "feat: add password recovery service"
~~~

---

### Task 3: Add injectable local-log and SMTP email senders

**Files:**
- Create: api/internal/notifications/email.go
- Create: api/internal/notifications/email_test.go
- Modify: api/cmd/server/main.go
- Modify: api/internal/httpapi/handler.go
- Modify: .env.example
- Modify: docker-compose.yml

**Interfaces:**
- Produces notifications.EmailMessage with To, Subject, and Text.
- Produces notifications.EmailSender with Send(context.Context, EmailMessage) error.
- Produces notifications.NewFromEnv(logger *log.Logger, getenv func(string) string) EmailSender.

- [ ] **Step 1: Write failing sender tests**

Add tests for:

~~~
func TestLogEmailSenderWritesRecipientSubjectAndBody(t *testing.T) {}
func TestSMTPEmailSenderBuildsPlainTextMessage(t *testing.T) {}
func TestLogModeIsSelectedByDefault(t *testing.T) {}
~~~

The SMTP test must inject a fake SMTP transport or a local test listener; it must not contact an external mail server.

- [ ] **Step 2: Run sender tests and confirm red**

~~~
go test ./internal/notifications -count=1
~~~

- [ ] **Step 3: Implement the senders and configuration**

Implement logEmailSender using the standard logger. Implement SMTP with net/smtp, SMTP_HOST, SMTP_PORT (default 587), SMTP_USERNAME, SMTP_PASSWORD, and SMTP_FROM. If EMAIL_MODE is empty or log, choose log mode. If EMAIL_MODE=smtp and required settings are missing, return a clear startup error rather than silently pretending to send.

Add these variables to .env.example and pass them through the API service in docker-compose.yml:

~~~
EMAIL_MODE=log
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=no-reply@example.com
~~~

Instantiate the sender in main.go and pass it into httpapi.New. Keep EMAIL_MODE=log as the Docker default.

- [ ] **Step 4: Run sender tests and compile the API**

~~~
gofmt -w internal/notifications/email.go internal/notifications/email_test.go cmd/server/main.go
go test ./internal/notifications ./cmd/server
~~~

- [ ] **Step 5: Commit the email boundary**

~~~
git add api/internal/notifications api/cmd/server/main.go api/internal/httpapi/handler.go .env.example docker-compose.yml
git commit -m "feat: add local and smtp email senders"
~~~

---

### Task 4: Expose password recovery/change API endpoints

**Files:**
- Modify: api/internal/httpapi/handler.go
- Create: api/internal/httpapi/password_handler.go
- Create: api/internal/httpapi/password_handler_test.go

**Interfaces:**
- POST /api/auth/password-reset/request
- POST /api/auth/password-reset/confirm
- PUT /api/auth/password
- Existing Handler receives the password service and email sender from httpapi.New.

- [ ] **Step 1: Write failing HTTP tests**

Add tests for:

~~~
func TestPasswordResetRequestDoesNotRevealUnknownEmail(t *testing.T) {}
func TestPasswordResetConfirmReturnsInvalidToken(t *testing.T) {}
func TestPasswordChangeRequiresAuthentication(t *testing.T) {}
func TestPasswordChangeRevokesTheCurrentSession(t *testing.T) {}
~~~

Assert response status and stable error codes:

- Request: 202 and identical JSON for known/unknown email.
- Confirm: 400 with invalid_reset_token for bad token.
- Change: 401 without a session; 204 with valid current password.

- [ ] **Step 2: Run focused HTTP tests and confirm red**

~~~
go test ./internal/httpapi -run Password -count=1
~~~

- [ ] **Step 3: Add public and authenticated routing**

Dispatch the two reset routes before session authentication. Dispatch PUT /api/auth/password after authenticate. Keep the existing Origin validation and JSON decoding rules. The reset-request handler always writes the generic 202 response.

- [ ] **Step 4: Send reset email only when an active user exists**

Build the reset link as \${AppOrigin}/reset-password?token=\${rawToken}. Use subject Reset your CSF Compliance password. Pass delivery through the injected sender and log a delivery failure without changing the 202 response.

- [ ] **Step 5: Map auth errors**

Map invalid reset tokens to 400 invalid_reset_token, weak passwords to 400 validation_error, wrong current password to 400 invalid_current_password, and unknown internal failures to 500 internal_error. On successful password change/reset, do not set a new session cookie; the frontend redirects to login.

- [ ] **Step 6: Run all API tests**

~~~
gofmt -w internal/httpapi/password_handler.go internal/httpapi/password_handler_test.go internal/httpapi/handler.go
go test ./internal/httpapi ./internal/auth ./internal/store
~~~

- [ ] **Step 7: Commit the password API**

~~~
git add api/internal/httpapi/handler.go api/internal/httpapi/password_handler.go api/internal/httpapi/password_handler_test.go
git commit -m "feat: expose password recovery endpoints"
~~~

---

### Task 5: Connect invitation and assessment events to email notifications

**Files:**
- Create: api/internal/httpapi/notifications.go
- Create: api/internal/httpapi/notifications_test.go
- Modify: api/internal/httpapi/invitations_handler.go
- Modify: api/internal/httpapi/responses_handler.go
- Modify: api/internal/httpapi/handler.go
- Create: api/internal/store/notification_recipients.go
- Create: api/internal/store/notification_recipients_test.go

**Interfaces:**
- Produces Store.ListProjectReviewerEmails(ctx, projectID) ([]string, error).
- Produces Store.GetAssignedAssessorEmail(ctx, projectID, subcategoryID) (string, error).
- Produces Store.ListOrganizationEmailsByRoles(ctx, organizationID string, roles []string) ([]string, error).
- Produces a handler helper that sends a message to each recipient and logs failures without returning a domain mutation error.

- [ ] **Step 1: Write failing recipient and notification tests**

Test these concrete behaviors:

~~~
func TestInvitationSendsJoinEmail(t *testing.T) {}
func TestSubmittedResponseNotifiesReviewers(t *testing.T) {}
func TestApprovedResponseNotifiesAssignedAssessor(t *testing.T) {}
func TestReturnedResponseNotifiesAssignedAssessor(t *testing.T) {}
func TestFinalizedProjectNotifiesOrgAdminsAndReviewers(t *testing.T) {}
func TestEmailFailureDoesNotFailTheDomainMutation(t *testing.T) {}
~~~

Use a fake sender that records messages and can return an injected error. Use the existing fake stores and authenticated handler helpers; do not send real mail in tests.

- [ ] **Step 2: Run notification tests and confirm red**

~~~
go test ./internal/httpapi ./internal/store -run "Notification|Invitation|Response|Finalize" -count=1
~~~

- [ ] **Step 3: Implement recipient queries**

Resolve only active users. Reviewers are limited to the project organization. Assigned assessor lookup uses the included profile assignment and returns an empty recipient when no valid active assessor exists. Finalization recipients are active org_admin and reviewer users in the organization, deduplicated by normalized email.

- [ ] **Step 4: Implement message builders and best-effort delivery**

Use stable subjects:

~~~
Join the NIST CSF Compliance workspace
Outcome ready for review: {subcategoryCode}
Outcome approved: {subcategoryCode}
Outcome needs more information: {subcategoryCode}
Project finalized: {projectName}
~~~

Include only the invitation/reset/workspace link, project name, outcome code, and next action. Never include response text or evidence data. notify must log a failure with event type and count and then return nil to its caller.

- [ ] **Step 5: Connect all event points**

Call the invitation sender after CreateInvitation. Call reviewer notification after a response is submitted. Call assigned-assessor notification after reviewed or needs_more_info. Call organization notification after successful finalization. Keep existing audit events and API response bodies unchanged except for non-blocking logs.

- [ ] **Step 6: Run API tests**

~~~
gofmt -w internal/httpapi/notifications.go internal/httpapi/notifications_test.go internal/httpapi/invitations_handler.go internal/httpapi/responses_handler.go internal/store/notification_recipients.go internal/store/notification_recipients_test.go
go test ./internal/httpapi ./internal/store
~~~

- [ ] **Step 7: Commit notification wiring**

~~~
git add api/internal/httpapi/notifications.go api/internal/httpapi/notifications_test.go api/internal/httpapi/invitations_handler.go api/internal/httpapi/responses_handler.go api/internal/httpapi/handler.go api/internal/store/notification_recipients.go api/internal/store/notification_recipients_test.go
git commit -m "feat: notify users about assessment workflow events"
~~~

---

### Task 6: Add frontend password recovery and account password change

**Files:**
- Modify: web/src/lib/api.ts
- Modify: web/src/lib/types.ts only if a response type is needed
- Create: web/src/components/ForgotPasswordForm.tsx
- Create: web/src/components/ForgotPasswordForm.test.tsx
- Create: web/src/components/ResetPasswordForm.tsx
- Create: web/src/components/ResetPasswordForm.test.tsx
- Create: web/src/components/ChangePasswordForm.tsx
- Create: web/src/components/ChangePasswordForm.test.tsx
- Create: web/src/app/forgot-password/page.tsx
- Create: web/src/app/reset-password/page.tsx
- Create: web/src/app/account/password/page.tsx
- Modify: web/src/components/LoginForm.tsx
- Modify: web/src/components/LoginForm.test.tsx
- Modify: web/src/components/OrganizationDashboard.tsx
- Modify: web/src/components/OrganizationDashboard.test.tsx
- Modify: web/src/app/globals.css

**Interfaces:**
- api.requestPasswordReset(email: string).
- api.confirmPasswordReset(token: string, password: string).
- api.changePassword(currentPassword: string, newPassword: string).
- The three forms expose controlled submit callbacks and readable loading/error/success states.

- [ ] **Step 1: Write failing API/client and component tests**

Add fetch assertions for:

~~~
POST /api/auth/password-reset/request
POST /api/auth/password-reset/confirm
PUT  /api/auth/password
~~~

Add UI assertions that Login shows Forgot password?, Forgot Password shows the generic success message, Reset Password reads the token, Change Password submits current/new values, and the Organizations identity block links to /account/password.

- [ ] **Step 2: Run focused frontend tests and confirm red**

From C:/Acuitmesh/NIST-CSF-Compliance/web:

~~~
npm.cmd test -- --run src/components/ForgotPasswordForm.test.tsx src/components/ResetPasswordForm.test.tsx src/components/ChangePasswordForm.test.tsx src/lib/api.test.ts src/components/LoginForm.test.tsx
~~~

- [ ] **Step 3: Implement typed API methods and form components**

Keep the success copy generic for reset requests. Use router.replace("/login") after a successful reset or password change. Preserve API error messages for invalid token, wrong current password, and weak password. Do not put reset tokens in client logs.

- [ ] **Step 4: Add the three App Router pages**

forgot-password/page.tsx is public. reset-password/page.tsx is public and reads searchParams.get("token"). account/password/page.tsx calls api.me() and redirects to /login on 401; it renders the authenticated change form and returns to /login after success.

- [ ] **Step 5: Add navigation and focused styles**

Add the reset link to LoginForm. Add a Change password link beside the existing identity/sign-out actions in OrganizationDashboard. Reuse the existing form, panel, error, and primary button classes, and add page layout rules for the three new auth routes.

- [ ] **Step 6: Run frontend tests and TypeScript**

~~~
npm.cmd test -- --run src/components/ForgotPasswordForm.test.tsx src/components/ResetPasswordForm.test.tsx src/components/ChangePasswordForm.test.tsx src/app/login/page.test.tsx src/app/organizations/page.test.tsx
npx.cmd tsc --noEmit --incremental false
~~~

- [ ] **Step 7: Commit the frontend auth flow**

~~~
git add web/src/lib/api.ts web/src/lib/types.ts web/src/components/ForgotPasswordForm.tsx web/src/components/ForgotPasswordForm.test.tsx web/src/components/ResetPasswordForm.tsx web/src/components/ResetPasswordForm.test.tsx web/src/components/ChangePasswordForm.tsx web/src/components/ChangePasswordForm.test.tsx web/src/app/forgot-password/page.tsx web/src/app/reset-password/page.tsx web/src/app/account/password/page.tsx web/src/components/LoginForm.tsx web/src/components/LoginForm.test.tsx web/src/components/OrganizationDashboard.tsx web/src/components/OrganizationDashboard.test.tsx web/src/app/globals.css
git commit -m "feat: add password recovery screens"
~~~

---

### Task 7: Document local email testing and verify the complete feature

**Files:**
- Modify: README.md
- Modify: DESIGN.md only if the UI map needs the new routes
- Modify: .env.example if Task 3 did not already update it

**Interfaces:**
- README documents local EMAIL_MODE=log, where to find reset/invitation links, SMTP variables, new auth routes, and the non-enumerating reset behavior.
- README removes the now-inaccurate Password reset, Real email notifications, and Report export entries from the not-included list; server-side PDF remains explicitly out of scope.

- [ ] **Step 1: Write documentation assertions/checks**

Search the documentation after edits and verify it contains the exact route/config terms:

~~~
rg -n "password-reset|account/password|EMAIL_MODE|SMTP_HOST|local log|server-side PDF" README.md DESIGN.md .env.example
~~~

- [ ] **Step 2: Update README and DESIGN**

Document the user flow, local reset-link testing through API logs, invitation manual-link fallback, notification recipients, and the fact that assessment mutations do not roll back when email delivery fails.

- [ ] **Step 3: Apply the new migration to a running local volume**

Do not delete volumes. With Docker Compose running, apply the idempotent migration and restart the API:

~~~
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/009_password_reset_tokens.sql
docker compose restart api
~~~

- [ ] **Step 4: Run complete verification**

From api:

~~~
go test ./...
~~~

From web:

~~~
npm.cmd test -- --run
npx.cmd tsc --noEmit --incremental false
npm.cmd run build
~~~

From the repository root:

~~~
git diff --check
docker compose up --build -d
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:3000
~~~

Expected: all tests/builds pass; both health checks return 200; reset requests do not reveal account existence; EMAIL_MODE=log prints links without sending external mail; no output/tmp files are staged.

- [ ] **Step 5: Commit documentation and final verification**

~~~
git add README.md DESIGN.md .env.example
git commit -m "docs: document password recovery and email notifications"
~~~
