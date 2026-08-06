# Authentication and RBAC V1 Design

## Goal

Add email/password login, separate organizations from projects, and enforce a small role model for counselors and customer stakeholders without adding SSO, email delivery, MFA, or counselor assignment.

## User model

`user_type` describes which side a user belongs to:

- `counselor`: the consulting company operating the software.
- `stakeholder`: a customer user belonging to one organization.

`role` describes permissions:

- `counselor_admin`: access all organizations and projects; create and delete organizations; manage counselor accounts.
- `counselor`: access all organizations and projects; create projects and edit assessments; cannot manage system-level accounts or delete organizations.
- `org_admin`: access every project in their organization; invite stakeholders and assign stakeholder roles.
- `assessor`: view projects in their organization and edit assessments.
- `reviewer`: view projects and review results. Approval workflow is deferred.
- `viewer`: read-only access to projects and assessments.

Counselor users have no `organization_id`. Stakeholders must have exactly one `organization_id`. An organization admin cannot create, modify, or grant counselor accounts or counselor roles.

## Authentication architecture

The Go API owns authentication, sessions, and authorization. Next.js renders the login and invitation pages and calls same-origin `/api/auth/...` endpoints through the existing proxy.

Passwords are stored only as bcrypt hashes using `golang.org/x/crypto/bcrypt`. Successful login creates an opaque, cryptographically random session token. PostgreSQL stores only the token hash. The raw token is returned in a cookie with `HttpOnly`, `SameSite=Lax`, `Path=/`, and `Secure` outside local development. Session expiry is enforced by the API.

State-changing requests require a same-origin `Origin` header and JSON content type. Login attempts use a small in-memory throttle keyed by normalized email and client address. This keeps V1 independent of Redis while limiting basic credential guessing; a distributed rate limiter is deferred until the API runs multiple replicas.

The first counselor administrator is bootstrapped from `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD`. Startup creates this user only when no `counselor_admin` exists. Existing administrators are never overwritten by later environment changes. Production deployment must supply non-default values through secret configuration; credentials are not committed to source control.

## Data model

Extend `users` with:

- `email` unique and case-insensitive.
- `password_hash` nullable while an invitation is pending.
- `role` constrained to the six V1 roles.
- `status` constrained to `invited`, `active`, or `disabled`.
- `created_at` and `updated_at`.

Keep the existing `organization_id` and `user_type`. Add database constraints that require stakeholder users to have an organization and counselor users to have none.

Add `sessions`:

- `id`, `user_id`, `token_hash`, `expires_at`, `last_seen_at`, `created_at`.

Add `invitations`:

- `id`, nullable `organization_id`, `email`, `user_type`, `role`, `token_hash`, `invited_by`, `expires_at`, `accepted_at`, `created_at`.

Add `audit_logs`:

- `id`, `actor_user_id`, `organization_id`, `project_id`, `action`, `entity_type`, `entity_id`, `metadata`, `created_at`.

V1 keeps one organization per stakeholder and does not add an organization-membership join table. Multi-organization stakeholders and counselor assignment are deferred.

## Application flow

### Login

1. User submits email and password to `POST /api/auth/login`.
2. The API returns the same generic error for an unknown email, wrong password, disabled account, or inactive invited account.
3. On success the API creates a session, sets the cookie, and returns the current user.
4. `GET /api/auth/me` restores the session after page refresh.
5. `POST /api/auth/logout` revokes the session and clears the cookie.

### Organization and project navigation

1. After login, the home page lists organizations visible to the current user.
2. Counselor roles see all organizations. Stakeholders see only their organization.
3. Selecting an organization opens an organization workspace containing organization details, projects, and stakeholders.
4. Counselor roles create projects inside the selected organization.
5. Project creation never creates an organization implicitly.

### Invitation

1. A counselor creates an organization and its first `org_admin` invitation.
2. An `org_admin` can invite additional stakeholders with `org_admin`, `assessor`, `reviewer`, or `viewer` roles.
3. A `counselor_admin` can invite counselor users with `counselor_admin` or `counselor` roles; these invitations have no organization.
4. The API returns a one-time invitation URL. V1 provides a Copy button; it does not send email.
5. The recipient opens the URL, sets a password, and activates the account.
6. Invitation tokens are single-use, expire server-side, and are stored only as hashes.

## API authorization

All protected API handlers obtain the current user from authentication middleware. Handlers derive organization scope from the authenticated user and database records; they do not trust a browser-supplied `organization_id` by itself.

V1 endpoints include:

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`
- `GET /api/organizations`
- `POST /api/organizations`
- `GET /api/organizations/:id`
- `GET /api/organizations/:id/projects`
- `POST /api/organizations/:id/projects`
- `GET /api/organizations/:id/users`
- `PATCH /api/organizations/:id/users/:userID`
- `POST /api/organizations/:id/invitations`
- `GET /api/counselors`
- `PATCH /api/counselors/:userID`
- `POST /api/counselor-invitations`
- `POST /api/invitations/:token/accept`

Existing project, profile, summary, and deletion endpoints become protected and enforce organization and role checks. UI visibility mirrors permissions for usability, but the Go API remains the security boundary.

## Error handling

- Login failures use one generic message to avoid account enumeration.
- Repeated login failures return `429` after the configured V1 threshold without revealing whether an account exists.
- Duplicate organization names produce a validation error but names are not treated as a global identity key.
- Duplicate active or pending stakeholder emails within an organization return a conflict.
- Expired, accepted, or invalid invitation tokens return the same invalid-invitation response.
- Unauthorized access returns `401`; authenticated users without permission receive `403`; inaccessible tenant resources return `404` where revealing their existence would leak customer data.
- Disabled users and revoked or expired sessions cannot access protected endpoints.

## Testing

- Database tests cover constraints, bootstrap idempotency, session expiry, invitation single-use behavior, and organization scoping.
- API tests cover login/logout/me, generic login failure, each role boundary, cross-organization denial, and protected existing endpoints.
- Frontend tests cover login, session restoration, organization-first navigation, role-based controls, invitation-link copying, and unauthorized responses.
- Docker verification covers bootstrap login, organization creation, first org-admin invitation, stakeholder activation, project creation under an existing organization, and access denial across roles.

## V1 exclusions

- Email delivery
- Password reset
- Microsoft or Google login
- MFA
- Multi-organization stakeholders
- Counselor-to-customer assignment
- Reviewer approval workflow
- Fine-grained per-project stakeholder permissions
- Redis or a separate session service

## Security references

- OWASP Session Management Cheat Sheet: server-side session state, random opaque identifiers, cookie attributes, and server-enforced expiry.
- Go `golang.org/x/crypto/bcrypt`: password hashing and comparison.
