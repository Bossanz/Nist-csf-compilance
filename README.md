# NIST CSF Compliance Web App

Lean vertical slice using Next.js, Go, and PostgreSQL.

## Run

Copy `.env.example` to `.env` if local overrides are needed, then run:

```bash
docker compose up --build
```

- Web: http://localhost:3000
- API: http://localhost:8080/healthz
- PostgreSQL: localhost:5432

The database initializes the NIST CSF 2.0 catalog from the supplied workbook with 6 Functions, 22 Categories, and 106 Subcategories.

## Verify

```bash
docker compose config
cd api && go test ./...
cd web && npm run typecheck && npm run build
```

Version 1 includes project creation, profile editing, and coverage summary. Authentication, notifications, evidence storage, reports, and full action planning are intentionally deferred.
