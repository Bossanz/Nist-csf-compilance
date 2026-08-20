$ErrorActionPreference = 'Stop'

$marker = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT 1 FROM schema_migrations WHERE version='010_remediation_actions'" 2>$null
if ($LASTEXITCODE -ne 0 -or $marker.Trim() -ne '1') { throw 'migration 010 is not recorded' }

$tables = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM pg_class WHERE relname IN ('remediation_actions','remediation_evidence')"
if ([int]$tables.Trim() -ne 2) { throw 'remediation tables are missing' }

Write-Output 'migration smoke test passed'
