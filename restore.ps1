# Restore a GK Circle SQL backup into the local stack.

param([Parameter(Mandatory = $true)][string]$BackupFile)

$ErrorActionPreference = "Stop"
$resolved = Resolve-Path -LiteralPath $BackupFile -ErrorAction Stop
Write-Host "Restore target: local GK Circle database"
Write-Host "Backup: $resolved"
if ((Read-Host "Type RESTORE to continue") -cne "RESTORE") {
    Write-Host "Restore cancelled."
    exit 0
}

Get-Content -LiteralPath $resolved | docker compose exec -T db sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
if ($LASTEXITCODE -ne 0) { throw "Database restore failed." }
Write-Host "Restore completed."
