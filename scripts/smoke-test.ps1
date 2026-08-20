param(
  [Parameter(Mandatory = $true)] [string] $CounselorAdminEmail,
  [Parameter(Mandatory = $true)] [string] $CounselorAdminPassword,
  [string] $ApiBaseUrl = 'http://localhost:8080',
  [string] $AppOrigin = 'http://localhost:3000',
  [switch] $KeepData
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Invoke-Api {
  param(
    [Parameter(Mandatory = $true)] [Microsoft.PowerShell.Commands.WebRequestSession] $Session,
    [Parameter(Mandatory = $true)] [string] $Method,
    [Parameter(Mandatory = $true)] [string] $Path,
    [object] $Body
  )

  $request = @{
    Uri = "$ApiBaseUrl$Path"
    Method = $Method
    WebSession = $Session
    UseBasicParsing = $true
    Headers = @{ Origin = $AppOrigin }
  }
  if ($PSBoundParameters.ContainsKey('Body')) {
    $request.ContentType = 'application/json'
    $request.Body = ($Body | ConvertTo-Json -Depth 10 -Compress)
  }
  try {
    return Invoke-RestMethod @request
  } catch {
    $detail = $_.ErrorDetails.Message
    if ([string]::IsNullOrWhiteSpace($detail)) { $detail = $_.Exception.Message }
    throw "API $Method $Path failed: $detail"
  }
}

function Login {
  param([string] $Email, [string] $Password)
  $session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  try {
    Invoke-Api -Session $session -Method Post -Path '/api/auth/login' -Body @{ email = $Email; password = $Password } | Out-Null
  } catch {
    throw "Login failed for $Email. Reset it with: go run ./cmd/local-password --database-url <url> --email $Email --password <new-password>. $($_.Exception.Message)"
  }
  return $session
}

function Assert-Equal {
  param([object] $Actual, [object] $Expected, [string] $Message)
  if ("$Actual" -ne "$Expected") { throw "$Message. Expected '$Expected', got '$Actual'." }
}

$adminSession = $null
$organizationID = $null
$projectID = $null
$suffix = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
$organizationAdminEmail = "p1-org-admin-$suffix@example.com"
$assessorEmail = "p1-assessor-$suffix@example.com"
$reviewerEmail = "p1-reviewer-$suffix@example.com"
$auditorEmail = "p1-auditor-$suffix@example.com"
$cancelledInvitationEmail = "p1-cancelled-$suffix@example.com"
$organizationAdminPassword = 'OrgAdmin!2026'
$assessorPassword = 'Assessor!2026'
$reviewerPassword = 'Reviewer!2026'
$auditorPassword = 'Auditor!2026'

try {
  $health = Invoke-RestMethod "$ApiBaseUrl/healthz"
  Assert-Equal $health.status 'ok' 'API health check failed'

  $adminSession = Login $CounselorAdminEmail $CounselorAdminPassword
  $organization = Invoke-Api -Session $adminSession -Method Post -Path '/api/organizations' -Body @{ name = "P1 Smoke Organization $suffix" }
  $organizationID = $organization.id

  $organizationAdminInvitation = Invoke-Api -Session $adminSession -Method Post -Path "/api/organizations/$organizationID/invitations" -Body @{ email = $organizationAdminEmail; role = 'org_admin' }
  $assessorInvitation = Invoke-Api -Session $adminSession -Method Post -Path "/api/organizations/$organizationID/invitations" -Body @{ email = $assessorEmail; role = 'assessor' }
  $reviewerInvitation = Invoke-Api -Session $adminSession -Method Post -Path "/api/organizations/$organizationID/invitations" -Body @{ email = $reviewerEmail; role = 'reviewer' }
  $organizationAdminToken = ([System.Uri]$organizationAdminInvitation.invitationURL).Segments[-1]
  $assessorToken = ([System.Uri]$assessorInvitation.invitationURL).Segments[-1]
  $reviewerToken = ([System.Uri]$reviewerInvitation.invitationURL).Segments[-1]

  $anonymousSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  $organizationAdminUser = Invoke-Api -Session $anonymousSession -Method Post -Path "/api/invitations/$organizationAdminToken/accept" -Body @{ name = 'P1 Smoke Organization Admin'; password = $organizationAdminPassword }
  $assessorUser = Invoke-Api -Session $anonymousSession -Method Post -Path "/api/invitations/$assessorToken/accept" -Body @{ name = 'P1 Smoke Assessor'; password = $assessorPassword }
  $reviewerUser = Invoke-Api -Session $anonymousSession -Method Post -Path "/api/invitations/$reviewerToken/accept" -Body @{ name = 'P1 Smoke Reviewer'; password = $reviewerPassword }
  $organizationAdminSession = Login $organizationAdminEmail $organizationAdminPassword

  $project = Invoke-Api -Session $adminSession -Method Post -Path "/api/organizations/$organizationID/projects" -Body @{
    name = "P1 Workflow Smoke $suffix"
    objective = 'Verify the complete NIST CSF assessment and remediation workflow.'
    assessmentPeriod = 'Q3 2026'
    targetCompletionDate = '2026-09-30'
    scopeBoundary = 'Production application and supporting operations.'
    complianceDriver = 'Internal audit readiness.'
  }
  $projectID = $project.id
  $auditorInvitation = Invoke-Api -Session $organizationAdminSession -Method Post -Path "/api/organizations/$organizationID/invitations" -Body @{ email = $auditorEmail; role = 'auditor'; projectIDs = @($projectID) }
  $cancelledInvitation = Invoke-Api -Session $organizationAdminSession -Method Post -Path "/api/organizations/$organizationID/invitations" -Body @{ email = $cancelledInvitationEmail; role = 'viewer' }
  $cancelled = Invoke-Api -Session $organizationAdminSession -Method Post -Path "/api/organizations/$organizationID/invitations/$($cancelledInvitation.id)/cancel" -Body @{}
  Assert-Equal $cancelled.status 'cancelled' 'Invitation cancellation did not update the lifecycle status'
  $resentAuditorInvitation = Invoke-Api -Session $organizationAdminSession -Method Post -Path "/api/organizations/$organizationID/invitations/$($auditorInvitation.id)/resend" -Body @{}
  Assert-Equal $resentAuditorInvitation.status 'pending' 'Invitation resend did not create a pending replacement'
  $auditorToken = ([System.Uri]$resentAuditorInvitation.invitationURL).Segments[-1]
  $auditorUser = Invoke-Api -Session $anonymousSession -Method Post -Path "/api/invitations/$auditorToken/accept" -Body @{ name = 'P1 Smoke Auditor'; password = $auditorPassword }

  $invitationList = Invoke-Api -Session $organizationAdminSession -Method Get -Path "/api/organizations/$organizationID/invitations"
  if (-not (@($invitationList | Where-Object { $_.status -eq 'cancelled' }).Count -ge 1)) { throw 'Cancelled invitation was not visible in the lifecycle list' }
  if (-not (@($invitationList | Where-Object { $_.status -eq 'superseded' }).Count -ge 1)) { throw 'Superseded invitation was not visible in the lifecycle list' }
  $profilePayload = Invoke-Api -Session $adminSession -Method Get -Path "/api/projects/$projectID/profile"
  $profile = $profilePayload
  if ($profile -is [System.Array]) { $profile = $profile[0] }
  if ($profile -is [System.Array]) { $profile = $profile[0] }
  if ([string]::IsNullOrWhiteSpace($profile.subcategoryID)) { throw 'Project has no usable profile rows' }

  $scope = Invoke-Api -Session $adminSession -Method Put -Path "/api/projects/$projectID/profile/$($profile.subcategoryID)" -Body @{
    included = $true
    rationale = 'Selected to verify governance and response readiness.'
    assignedUserID = $assessorUser.id
  }
  if (-not $scope.included) { throw 'Scope outcome was not included' }
  $submittedScope = Invoke-Api -Session $adminSession -Method Post -Path "/api/projects/$projectID/scope/submit"
  Assert-Equal $submittedScope.status 'in_review' 'Scope submission did not move the project to review'

  $auditorSession = Login $auditorEmail $auditorPassword
  $auditorProject = Invoke-Api -Session $auditorSession -Method Get -Path "/api/projects/$projectID"
  Assert-Equal $auditorProject.id $projectID 'Auditor could not read the assigned Project'
  $auditorAuditLogs = Invoke-Api -Session $auditorSession -Method Get -Path "/api/projects/$projectID/audit-logs"
  if (@($auditorAuditLogs).Count -lt 1) { throw 'Auditor could not read the assigned Project audit trail' }

  $assessorSession = Login $assessorEmail $assessorPassword
  Invoke-Api -Session $assessorSession -Method Put -Path "/api/projects/$projectID/profile/$($profile.subcategoryID)" -Body @{
    currentPriority = 'high'
    currentCoverageLevel = 'partial'
    currentStatusText = 'Security events are reviewed quarterly.'
    currentPoliciesText = 'Security monitoring policy.'
    targetPriority = 'high'
    targetCoverageLevel = 'full'
    targetApproachText = 'Centralize monitoring and document the operating procedure.'
    notes = 'Smoke workflow response.'
    considerations = 'Confirm production ownership.'
  } | Out-Null
  Invoke-Api -Session $assessorSession -Method Put -Path "/api/projects/$projectID/responses/$($profile.subcategoryID)" -Body @{ responseText = 'The organization reviews security events quarterly and is formalizing centralized monitoring.' } | Out-Null
  $submittedResponse = Invoke-Api -Session $assessorSession -Method Post -Path "/api/projects/$projectID/responses/$($profile.subcategoryID)/submit" -Body @{}
  Assert-Equal $submittedResponse.status 'submitted' 'Assessor response was not submitted'

  $reviewerSession = Login $reviewerEmail $reviewerPassword
  $approvedResponse = Invoke-Api -Session $reviewerSession -Method Post -Path "/api/projects/$projectID/responses/$($profile.subcategoryID)/review" -Body @{ status = 'reviewed'; comment = 'Smoke response accepted.' }
  Assert-Equal $approvedResponse.status 'reviewed' 'Reviewer did not approve the response'

  $finalized = Invoke-Api -Session $adminSession -Method Post -Path "/api/projects/$projectID/finalize" -Body @{}
  Assert-Equal $finalized.status 'closed' 'Project did not finalize'

  $action = Invoke-Api -Session $adminSession -Method Post -Path "/api/projects/$projectID/remediation-actions" -Body @{
    subcategoryID = $profile.subcategoryID
    title = 'Centralize security monitoring'
    description = 'Forward application and API security events to a searchable log store.'
    desiredResult = 'Security events are searchable and retained.'
    priority = 'high'
    ownerUserID = $assessorUser.id
    dueDate = '2026-09-30'
  }
  Assert-Equal $action.status 'open' 'Remediation Action was not created'

  Invoke-Api -Session $assessorSession -Method Patch -Path "/api/projects/$projectID/remediation-actions/$($action.id)/progress" -Body @{ progressNote = 'Centralized monitoring configuration is ready for review.' } | Out-Null
  $awaitingReview = Invoke-Api -Session $assessorSession -Method Post -Path "/api/projects/$projectID/remediation-actions/$($action.id)/submit" -Body @{}
  Assert-Equal $awaitingReview.status 'awaiting_review' 'Remediation Action was not submitted'
  $closedAction = Invoke-Api -Session $adminSession -Method Post -Path "/api/projects/$projectID/remediation-actions/$($action.id)/review" -Body @{ decision = 'close'; comment = 'Smoke remediation verified.' }
  Assert-Equal $closedAction.status 'closed' 'Remediation Action was not closed'

  $report = Invoke-Api -Session $adminSession -Method Get -Path "/api/projects/$projectID/final-report"
  $audit = Invoke-Api -Session $adminSession -Method Get -Path "/api/projects/$projectID/audit-package"
  if (@($report.remediationActions).Count -ne 1 -or @($audit.remediationActions).Count -ne 1) { throw 'Reports did not include the remediation Action' }
  Assert-Equal $report.remediationActions[0].status 'closed' 'Final Report has the wrong remediation status'
  Assert-Equal $audit.remediationActions[0].status 'closed' 'Audit Package has the wrong remediation status'

  Write-Output "authenticated smoke test passed: organization=$organizationID project=$projectID"
  if ($KeepData) {
    Write-Output 'sample data kept for manual review'
    Write-Output "project URL: http://localhost:3000/organizations/$($organization.slug)/projects/$($project.slug)"
  }
} finally {
  if (-not $KeepData -and $organizationID -and $adminSession) {
    try {
      Invoke-Api -Session $adminSession -Method Delete -Path "/api/organizations/$organizationID" | Out-Null
      Write-Output 'temporary smoke organization removed'
    } catch {
      Write-Warning "Could not remove temporary smoke organization ${organizationID}: $($_.Exception.Message)"
    }
  }
}
