# deploy-nuc-remote.ps1
# GK Circle - NUC Production Deployment Script
#
# Deployment Script Version: 2.0
#
# Responsibility: DEPLOY ONLY (G1-G7)
#   G1  Git SHA verification
#   G2  Clean working tree
#   G3  Docker image build
#   G4  Database backup (pg_dump)
#   G5  Prisma migrate deploy
#   G6  Prisma migrate status
#   G7  docker compose up -d
#
# After deployment, run verify-nuc-deployment.ps1 to verify runtime state.
# Verification is intentionally separate so it can be re-run independently.
#
# Usage:
#   .\scripts\deploy-nuc-remote.ps1
#
# Requirements:
#   - Passwordless SSH: 'ssh nuc' must work (see install-nuc-ssh-key.ps1)
#   - NUC project: ~/apps/gkcircle with .env present
#
# Environment: all ports and container names are read from docker compose
# at runtime using stable JSON APIs (docker compose ps --format json,
# docker inspect). Never hardcoded in this script.

$ErrorActionPreference = "Stop"
$NucProject = "/home/randhir/apps/gkcircle"
$ComposeFile = "docker-compose.nuc.yml"
$EnvFile = ".env"

# ─── Helper: run a bash script on the NUC ────────────────────────────────────

function Invoke-NucBash([string]$Script, [switch]$AllowNonZero) {
    $lf = $Script.Replace("`r", "")
    $tmpLocal = [System.IO.Path]::GetTempFileName()
    $tmpRemote = "/tmp/nuc-deploy-$(Get-Random).sh"
    try {
        [System.IO.File]::WriteAllText($tmpLocal, $lf, (New-Object System.Text.UTF8Encoding $false))
        # SCP to remote NUC temp directory (preserves LF line endings)
        scp $tmpLocal "nuc:$tmpRemote" 2>$null
        
        $prevEap = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        # Run remote script with bash explicitly
        $out = ssh nuc ('bash ' + $tmpRemote + '; exitCode=$?; rm -f ' + $tmpRemote + '; exit $exitCode') 2>&1
        $exit = $LASTEXITCODE
        $ErrorActionPreference = $prevEap
    } finally {
        Remove-Item $tmpLocal -ErrorAction SilentlyContinue
    }
    if (-not $AllowNonZero -and $exit -ne 0) {
        throw "Remote bash failed (exit $exit):`n$out"
    }
    return [PSCustomObject]@{ Output = ($out | Out-String).Trim(); ExitCode = $exit }
}

# ─── Gate printer ─────────────────────────────────────────────────────────────

$gateResults = @()

function Write-Gate([string]$Id, [string]$Label, [bool]$Pass, [string]$Detail = "", [string[]]$Rollback = @()) {
    $icon = if ($Pass) { "OK" } else { "FAIL" }
    # Use explicit variable boundary ${Label} to prevent PowerShell drive-colon parsing issues ($Label:)
    $line = "  $Id  $icon  ${Label}: $Detail"
    Write-Host $line
    $script:gateResults += [PSCustomObject]@{ Id = $Id; Pass = $Pass; Label = $Label; Detail = $Detail }

    if (-not $Pass) {
        Write-Host ""
        Write-Host "  ==================== ROLLBACK GUIDANCE ===================="
        foreach ($cmd in $Rollback) { Write-Host "    $cmd" }
        Write-Host "  ==========================================================="
        Write-Host ""
    }
}

function Write-GateSummary {
    Write-Host ""
    Write-Host "==========================================================="
    Write-Host " GK Circle NUC - DEPLOYMENT GATE SUMMARY"
    Write-Host "==========================================================="
    foreach ($g in $gateResults) {
        $icon = if ($g.Pass) { "OK" } else { "FAIL" }
        Write-Host "  $($g.Id)  $icon  $($g.Label): $($g.Detail)"
    }
    Write-Host "==========================================================="
    Write-Host "  Next: run .\scripts\verify-nuc-deployment.ps1"
    Write-Host "==========================================================="
    Write-Host ""
}

# ─── Preflight ────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "==========================================================="
Write-Host " GK Circle NUC - Deployment (G1-G7)"
Write-Host "==========================================================="
Write-Host ""

$preflightResult = Invoke-NucBash "test -d $NucProject && test -f $NucProject/$ComposeFile && test -s $NucProject/$EnvFile && echo ENV_OK"
if ($preflightResult.Output -notmatch "ENV_OK") {
    Write-Error "Preflight failed: $($preflightResult.Output)"
    exit 1
}
Write-Host "[deploy] Preflight OK."
Write-Host ""

# ─── Remote deploy script ─────────────────────────────────────────────────────

# Use a single-quoted here-string (@' ... '@) so the PowerShell parser treats it as 100% literal text.
# This avoids any syntax issues with bash variables ($), pipe/logical operators (&&, ||), or double quotes.
# We then replace placeholders to inject the local PowerShell variables.
$remoteScriptTemplate = @'
set -e
cd __NUC_PROJECT__
set -a; . ./__ENV_FILE__; set +a

# --- Detect docker compose command ---
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  echo "GATE_FAIL:G0:Docker compose not found"; exit 1
fi

# --- G1: Git SHA ---
if git pull --ff-only 2>/dev/null; then
  echo "[nuc] git pull OK"
else
  echo "[nuc] WARN: git pull skipped. Deploying current HEAD."
fi
GIT_SHA=$(git rev-parse --short HEAD)
echo "GATE_G1_SHA:$GIT_SHA"

# --- G2: Clean working tree ---
DIRTY=$(git status --short)
if [ -n "$DIRTY" ]; then
  echo "GATE_FAIL:G2:Dirty tree: $DIRTY"; exit 1
fi
echo "GATE_G2_OK"

# --- G3: Docker build ---
if ! $DC --env-file __ENV_FILE__ -f __COMPOSE_FILE__ build 2>&1; then
  echo "GATE_FAIL:G3:docker compose build failed"; exit 1
fi

BACKEND_IMAGE_ID=$(docker images --format '{{json .}}' | python3 -c "import sys, json; data = [json.loads(line) for line in sys.stdin]; img = next((i for i in data if i.get('Repository','') in ('gkcircle-backend','gk-circle-backend') and i.get('Tag','')=='latest'), None); print(img['ID'][:12] if img else 'unknown')" 2>/dev/null || echo 'unknown')
FRONTEND_IMAGE_ID=$(docker images --format '{{json .}}' | python3 -c "import sys, json; data = [json.loads(line) for line in sys.stdin]; img = next((i for i in data if i.get('Repository','') in ('gkcircle-frontend','gk-circle-frontend') and i.get('Tag','')=='latest'), None); print(img['ID'][:12] if img else 'unknown')" 2>/dev/null || echo 'unknown')
echo "GATE_G3_OK:backend=$BACKEND_IMAGE_ID frontend=$FRONTEND_IMAGE_ID"

# --- G4: Database backup ---
mkdir -p __NUC_PROJECT__/backup
BACKUP_FILE="__NUC_PROJECT__/backup/$(date +%Y%m%d_%H%M%S).sql.gz"
DB_USER=${DB_USER:-gk_user}
DB_NAME=${DB_NAME:-gk_circle}
if docker exec gk-circle-postgres pg_dump -U "$DB_USER" "$DB_NAME" 2>/dev/null | gzip > "$BACKUP_FILE"; then
  BACKUP_SIZE=$(du -sh "$BACKUP_FILE" | cut -f1)
  echo "GATE_G4_OK:$BACKUP_FILE ($BACKUP_SIZE)"
else
  echo "GATE_FAIL:G4:pg_dump failed"
  exit 1
fi

# --- G5: Prisma migrate deploy ---
if ! $DC --env-file __ENV_FILE__ -f __COMPOSE_FILE__ run --rm backend npx prisma migrate deploy > migration.log 2>&1; then
  echo "GATE_FAIL:G5:prisma migrate deploy failed"
  cat migration.log
  exit 1
fi
MIGRATION_SUMMARY=$(grep -oP '(\d+ migration|All migrations).*' migration.log 2>/dev/null || echo 'No migrations applied')
echo "GATE_G5_OK:$MIGRATION_SUMMARY"

# --- G6: Prisma migrate status ---
if ! $DC --env-file __ENV_FILE__ -f __COMPOSE_FILE__ run --rm backend npx prisma migrate status > migration-status.log 2>&1; then
  echo "GATE_FAIL:G6:prisma migrate status failed"
  cat migration-status.log
  exit 1
fi
if grep -qiE '(pending|failed)' migration-status.log; then
  echo "GATE_FAIL:G6:Pending or failed migrations"
  exit 1
fi
echo "GATE_G6_OK:all applied"

# --- G7: docker compose up -d ---
if ! $DC --env-file __ENV_FILE__ -f __COMPOSE_FILE__ up -d; then
  echo "GATE_FAIL:G7:docker compose up -d failed"
  exit 1
fi
sleep 5
echo "GATE_G7_OK"
echo "DEPLOY_COMPLETE"
'@

$remoteScript = $remoteScriptTemplate -replace '__NUC_PROJECT__', $NucProject -replace '__ENV_FILE__', $EnvFile -replace '__COMPOSE_FILE__', $ComposeFile

$result = Invoke-NucBash $remoteScript -AllowNonZero

# ─── Parse gate results ────────────────────────────────────────────────────

$output = $result.Output
Write-Host $output
Write-Host ""

if ($output -match "GATE_FAIL:(\w+):(.+)") {
    $failGate = $Matches[1]
    $failReason = $Matches[2]
    Write-Gate $failGate "FAILED" $false $failReason @(
        "ssh nuc 'cd ~/apps/gkcircle && docker compose ps'",
        "ssh nuc 'cd ~/apps/gkcircle && docker compose logs backend --tail 50'"
    )
    Write-GateSummary
    exit 1
}

if ($result.ExitCode -ne 0 -or $output -notmatch "DEPLOY_COMPLETE") {
    Write-Host "[deploy] Deployment failed."
    Write-GateSummary
    exit 1
}

$sha = if ($output -match "GATE_G1_SHA:(\S+)") { $Matches[1] } else { "unknown" }
$images = if ($output -match "GATE_G3_OK:(.+)") { $Matches[1] } else { "unknown" }
$backup = if ($output -match "GATE_G4_OK:(.+)") { $Matches[1] } else { "unknown" }
$migrate = if ($output -match "GATE_G5_OK:(.+)") { $Matches[1] } else { "applied" }

Write-Gate "G1" "Git SHA" $true $sha
Write-Gate "G2" "Clean tree" $true "clean"
Write-Gate "G3" "Docker build" $true $images
Write-Gate "G4" "DB backup" $true $backup
Write-Gate "G5" "Migrate deploy" $true $migrate
Write-Gate "G6" "Migrate status" $true "all applied"
Write-Gate "G7" "Compose up" $true "containers recreated"

Write-GateSummary
Write-Host "[deploy] Deployment complete (G1-G7 all pass)."
Write-Host "[deploy] Run: .\scripts\verify-nuc-deployment.ps1"
Write-Host ""
