# provision-qa-accounts.ps1
# Provisions production QA accounts on the NUC for deployment verification.
#
# Usage:
#   .\scripts\provision-qa-accounts.ps1
#
# Requirements:
#   - Passwordless SSH to NUC ('ssh nuc' must work)
#   - Backend container must be running (deploy first)
#   - NUC project directory: ~/apps/gkcircle

$ErrorActionPreference = "Stop"
$NucProject = "/home/randhir/apps/gkcircle"

# ─── Crypto-random password generator ────────────────────────────────────────

function New-CryptoPassword {
    param([int]$Length = 24)

    # Simple ASCII characters to avoid any parsing/encoding issues
    $chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz0123456789".ToCharArray()
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $bytes = New-Object byte[] $Length
    $rng.GetBytes($bytes)

    $password = -join ($bytes | ForEach-Object { $chars[$_ % $chars.Length] })
    $rng.Dispose()
    return $password
}

# ─── Preflight check ─────────────────────────────────────────────────────────

Write-Host ""
Write-Host "************************************************************"
Write-Host " GK Circle - QA Account Provisioner"
Write-Host "************************************************************"
Write-Host ""
Write-Host "[provision] Checking NUC connectivity..."

# Redirect PowerShell stderr stream (2>$null) to prevent docker/ssh stderr warnings from triggering $ErrorActionPreference = "Stop"
$preflightResult = ssh nuc "if [ -d $NucProject ]; then echo 'OK'; else echo 'MISSING'; fi" 2>$null
if ($LASTEXITCODE -ne 0 -or $preflightResult -notmatch "OK") {
    Write-Error "NUC preflight failed. Ensure 'ssh nuc' is configured and $NucProject exists."
    exit 1
}
Write-Host "[provision] NUC reachable. Project directory confirmed."

# ─── Check backend container is running ──────────────────────────────────────

Write-Host "[provision] Checking backend container status..."
$containerCheck = ssh nuc "cd $NucProject; docker compose ps 2>/dev/null" 2>$null | Out-String
if ($containerCheck -notmatch "backend.*Up") {
    Write-Error "Backend container is not running on the NUC. Deploy first, then run provisioning."
    Write-Host "  Run: .\scripts\deploy-nuc-remote.ps1"
    exit 1
}
Write-Host "[provision] Backend container is running. Proceeding."

# ─── Generate passwords ───────────────────────────────────────────────────────

Write-Host "[provision] Generating QA account passwords..."

$studentPw = New-CryptoPassword
$creatorPw = New-CryptoPassword
$adminPw   = New-CryptoPassword
$superAdminPw = New-CryptoPassword

Write-Host "[provision] Passwords generated (not stored anywhere)."

# ─── Run seed script on NUC ───────────────────────────────────────────────────

Write-Host "[provision] Running seed-qa-users.ts inside backend container..."
Write-Host ""

# Construct a single-line command to avoid heredoc parsing/piping issues in PowerShell
$seedCmd = "cd $NucProject; docker compose exec -T -e QA_STUDENT_PASSWORD='$studentPw' -e QA_CREATOR_PASSWORD='$creatorPw' -e QA_ADMIN_PASSWORD='$adminPw' -e QA_SUPERADMIN_PASSWORD='$superAdminPw' backend npx tsx scripts/qa/seed-qa-users.ts"

$prevEap = $ErrorActionPreference
$ErrorActionPreference = "Continue"
$seedOutput = ssh nuc $seedCmd 2>$null | Out-String
$seedExit = $LASTEXITCODE
$ErrorActionPreference = $prevEap

Write-Host $seedOutput

if ($seedExit -ne 0) {
    $studentPw = $null; $creatorPw = $null; $adminPw = $null; $superAdminPw = $null
    [System.GC]::Collect()
    Write-Error "QA seed failed (exit $seedExit). See output above."
    exit 1
}

# ─── Print credentials ONCE ──────────────────────────────────────────────────

Write-Host ""
Write-Host "************************************************************"
Write-Host "         GK Circle - QA Account Credentials"
Write-Host "                                                            "
Write-Host "   WARNING: COPY THESE NOW - THEY WILL NOT BE SHOWN AGAIN"
Write-Host "************************************************************"
Write-Host "                                                            "
Write-Host "  qa.student@gkcircle.com"
Write-Host "  Password: $studentPw"
Write-Host "                                                            "
Write-Host "  qa.creator@gkcircle.com"
Write-Host "  Password: $creatorPw"
Write-Host "                                                            "
Write-Host "  qa.admin@gkcircle.com"
Write-Host "  Password: $adminPw"
Write-Host "                                                            "
Write-Host "  qa.superadmin@gkcircle.com"
Write-Host "  Password: $superAdminPw"
Write-Host "                                                            "
Write-Host "************************************************************"
Write-Host ""
Write-Host "  Store in: 1Password / KeePass / secure notes"
Write-Host "  These accounts are used for production deployment verification."
Write-Host "  Do NOT share with real users."
Write-Host ""

# ─── Clear passwords from memory ─────────────────────────────────────────────

$studentPw = $null; $creatorPw = $null; $adminPw = $null; $superAdminPw = $null
[System.GC]::Collect()

Write-Host "[provision] Credentials cleared from memory."
Write-Host "[provision] QA provisioning complete."
Write-Host ""
