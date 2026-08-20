$ErrorActionPreference = 'Stop'
$config = docker compose config --format json | ConvertFrom-Json
foreach ($name in @('web','api','postgres','migrate')) {
  if (-not $config.services.$name) { throw "Missing service: $name" }
}
Write-Output 'compose services verified'
