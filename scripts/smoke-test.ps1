$ErrorActionPreference = 'Stop'
$health = Invoke-RestMethod http://localhost:8080/healthz
if ($health.status -ne 'ok') { throw 'health check failed' }
$catalog = Invoke-RestMethod http://localhost:8080/api/functions
if ($catalog.Count -ne 6) { throw "expected 6 functions, got $($catalog.Count)" }
$project = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/projects -ContentType 'application/json' -Body (@{name='Smoke Project'; organizationName='Smoke Org'} | ConvertTo-Json)
$projects = Invoke-RestMethod http://localhost:8080/api/projects
$listed = $projects | Where-Object { $_.id -eq $project.id }
if (-not $listed) { throw 'created project was not returned by project list' }
if ($listed.organizationName -ne 'Smoke Org') { throw 'project list omitted organization name' }
$profile = Invoke-RestMethod http://localhost:8080/api/projects/$($project.id)/profile
if ($profile.Count -ne 106) { throw "expected 106 profile rows, got $($profile.Count)" }
$first = $profile[0]
$updated = Invoke-RestMethod -Method Put -Uri http://localhost:8080/api/projects/$($project.id)/profile/$($first.subcategoryID) -ContentType 'application/json' -Body (@{
  included=$true
  rationale='Critical business outcome'
  currentPriority='High'
  currentCoverageLevel='partial'
  currentStatusText='Quarterly review'
  currentPoliciesText='Risk policy'
  targetPriority='High'
  targetCoverageLevel='full'
  targetApproachText='Formalize governance review'
  notes='Owner to confirm'
  considerations='Align with ERM'
} | ConvertTo-Json)
if (-not $updated.included -or $updated.currentCoverageLevel -ne 'partial' -or $updated.rationale -ne 'Critical business outcome' -or $updated.targetApproachText -ne 'Formalize governance review') { throw 'complete assessment update failed' }
$summary = Invoke-RestMethod http://localhost:8080/api/projects/$($project.id)/summary
if ($summary.includedCount -ne 1 -or [math]::Abs($summary.coveragePct - 33.3333333333333) -gt 0.001) { throw "summary calculation failed: $($summary | ConvertTo-Json -Compress)" }
Write-Output 'smoke test passed'
