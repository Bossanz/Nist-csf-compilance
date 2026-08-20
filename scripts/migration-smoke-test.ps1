$ErrorActionPreference = 'Stop'

$markers = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM schema_migrations WHERE version IN ('010_remediation_actions','011_auditor_invitation_audit')" 2>$null
if ($LASTEXITCODE -ne 0 -or $markers.Trim() -ne '2') { throw 'remediation and auditor migrations are not recorded' }

$tables = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM pg_class WHERE relname IN ('remediation_actions','remediation_evidence','invitation_project_access','project_auditor_access')"
if ([int]$tables.Trim() -ne 4) { throw 'remediation or auditor access tables are missing' }

Write-Output 'migration smoke test passed'
