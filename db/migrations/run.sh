#!/bin/sh
set -eu

psql_base="psql -v ON_ERROR_STOP=1 -h postgres -U ${POSTGRES_USER} -d ${POSTGRES_DB}"

$psql_base -c "CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"

for migration in /migrations/*.sql; do
  [ -f "$migration" ] || continue
  version="$(basename "$migration" .sql)"
  applied="$($psql_base -tAc "SELECT 1 FROM schema_migrations WHERE version='$version'")"
  if [ "$(printf '%s' "$applied" | tr -d '[:space:]')" = "1" ]; then
    echo "Migration $version already applied"
    continue
  fi
  echo "Applying migration $version"
  $psql_base -f "$migration"
  $psql_base -c "INSERT INTO schema_migrations(version) VALUES ('$version')"
done
