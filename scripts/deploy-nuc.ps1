# Deploy GK Circle to NUC stack from Windows (Docker -> NUC Postgres/Redis)
# Usage: .\scripts\deploy-nuc.ps1
# Requires: Docker Desktop, .env at repo root

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$envFile = Join-Path $Root ".env"
if (-not (Test-Path $envFile)) {
    Write-Error ".env not found. Copy .env.example to .env and configure for NUC."
}
if ((Get-Item $envFile).Length -eq 0) {
    Write-Error ".env is empty. Fill in values before deploying."
}

Get-Content $envFile | ForEach-Object {
    if ($_ -match '^\s*([^#=]+)=(.*)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim().Trim('"')
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

Write-Host "[deploy-nuc] Running Prisma migrate deploy against NUC database..."
Set-Location (Join-Path $Root "backend")
npx prisma migrate deploy
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Set-Location $Root
Write-Host "[deploy-nuc] Building and starting docker-compose.nuc.yml..."
docker compose --env-file .env -f docker-compose.nuc.yml up -d --build
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[deploy-nuc] Waiting for backend health..."
Start-Sleep -Seconds 15

Write-Host "[deploy-nuc] Health check (local nginx):"
try {
    $health = Invoke-RestMethod -Uri "http://127.0.0.1:8080/health" -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  API health: $($health | ConvertTo-Json -Compress)"
} catch {
    Write-Host "  Health check failed - see container logs:"
    docker compose --env-file .env -f docker-compose.nuc.yml ps
    docker compose --env-file .env -f docker-compose.nuc.yml logs backend --tail 30
}

Write-Host "[deploy-nuc] Done. Manual test URLs:"
Write-Host "  Local:  http://127.0.0.1:8080"
Write-Host "  Public: https://gkcircle.com (requires Cloudflare tunnel on NUC host)"
