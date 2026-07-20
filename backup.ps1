# jovVix — Database Backup Script
# Usage: .\backup.ps1

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$backupDir = "$PSScriptRoot\backups"
$backupFile = "$backupDir\jovvix_db_$timestamp.sql"

New-Item -ItemType Directory -Force -Path $backupDir | Out-Null

Write-Host "Backing up jovVix PostgreSQL database..." -ForegroundColor Cyan

docker compose exec -T db pg_dump `
  -U jovvix `
  -d jovvix `
  --clean `
  --if-exists `
  --no-owner `
  --no-acl `
  | Out-File -FilePath $backupFile -Encoding UTF8

if ($LASTEXITCODE -eq 0) {
    Write-Host "[SUCCESS] Backup saved to: $backupFile" -ForegroundColor Green
} else {
    Write-Host "[FAILED] Backup failed. Is the stack running?" -ForegroundColor Red
}
