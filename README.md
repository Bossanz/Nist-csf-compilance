# NIST CSF Compliance Web App

Lean vertical slice using Next.js, Go, and PostgreSQL.

## Run

Create a local `.env` from `.env.example` and replace the bootstrap password with a password of at least 12 characters. The bootstrap account is the first `counselor_admin` account.

```bash
cp .env.example .env
docker compose up --build
```

- Web: http://localhost:3000
- API: http://localhost:8080/healthz
- PostgreSQL: localhost:5432

The database initializes the NIST CSF 2.0 catalog from the supplied workbook with 6 Functions, 22 Categories, and 106 Subcategories.

If the PostgreSQL volume existed before authentication was added, apply the new migration once:

```bash
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/003_auth_rbac.sql
docker compose restart api
```

If the PostgreSQL volume existed before stakeholder responses were added, apply this migration once as well:

```bash
docker compose exec -T postgres psql -U compliance -d compliance -f /docker-entrypoint-initdb.d/004_stakeholder_responses.sql
docker compose restart api
```

Then sign in at http://localhost:3000 using `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` from `.env`.

## Version 1 workflow

1. A Counselor Admin signs in and creates a client organization.
2. A Counselor opens that organization and creates assessment projects.
3. A Counselor or Organization Admin creates stakeholder invitation links.
4. Stakeholders activate their accounts from `/invite/{token}`.
5. Organization Admins and Assessors answer NIST outcomes and attach evidence.
6. Reviewers mark submitted responses as Reviewed or Needs more information; Viewers have read-only access.
7. Counselors can view every client project and are the only stakeholder-side users who edit Priority, Coverage, and assessment fields.

Roles: `counselor_admin`, `counselor`, `org_admin`, `assessor`, `reviewer`, and `viewer`.

Permanent organization deletion is restricted to Counselor Admin and removes all organization-owned projects, assessments, stakeholders, invitations, and response metadata after exact-name confirmation. Evidence files are stored in the Docker named volume `evidence_data` and cleaned up when their project or organization is deleted.

## Verify

```bash
docker compose config
cd api && go test ./...
cd web && npm run typecheck && npm run build
```

Version 1 includes email/password authentication, server-side sessions, organization-scoped projects, invitation-based account creation, role access control, counselor profile editing, stakeholder responses, local evidence storage, review status, and coverage summary. Password reset, notifications, reports, and full action planning are intentionally deferred.
