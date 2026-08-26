$ErrorActionPreference = 'Stop'

$markers = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM schema_migrations WHERE version IN ('010_remediation_actions','011_auditor_invitation_audit')" 2>$null
if ($LASTEXITCODE -ne 0 -or $markers.Trim() -ne '2') { throw 'remediation and auditor migrations are not recorded' }

$versionMarker = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM schema_migrations WHERE version='012_project_versions'" 2>$null
if ($LASTEXITCODE -ne 0 -or $versionMarker.Trim() -ne '1') { throw 'project version migration is not recorded' }

$versionColumns = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM information_schema.columns WHERE table_name='projects' AND column_name IN ('version_group_id','version_number','previous_version_id')" 2>$null
if ($LASTEXITCODE -ne 0 -or [int]$versionColumns.Trim() -ne 3) { throw 'project version columns are missing' }

$tables = docker compose exec -T postgres psql -U compliance -d compliance -tAc "SELECT count(*) FROM pg_class WHERE relname IN ('remediation_actions','remediation_evidence','invitation_project_access','project_auditor_access')"
if ([int]$tables.Trim() -ne 4) { throw 'remediation or auditor access tables are missing' }

Write-Output 'migration smoke test passed'
