# Back up the local GK Circle PostgreSQL database.

$ErrorActionPreference = "Stop"
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$backupDir = Join-Path $PSScriptRoot "backups"
$backupFile = Join-Path $backupDir "gk_circle_db_$timestamp.sql"
New-Item -ItemType Directory -Force -Path $backupDir | Out-Null

docker compose exec -T db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner --no-acl' |
    Out-File -FilePath $backupFile -Encoding utf8

if ($LASTEXITCODE -ne 0) { throw "Database backup failed." }
Write-Host "Backup saved to $backupFile"
